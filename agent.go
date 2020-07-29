package agent

//go:generate sh generate_sha.sh

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/interception"
	"github.com/bearer/go-agent/proxy"
)

const (
	// ExampleWellFormedInvalidKey is a well-formed key known to be invalid. It may
	// be used for integration test scenarios.
	ExampleWellFormedInvalidKey = `app_12345678901234567890123456789012345678901234567890`

	// SecretKeyName is the environment variable used to hold the Bearer secret key,
	// specific to each client. Fetching the secret key from the environment is a
	// best practice in 12-factor application development.
	SecretKeyName = `BEARER_SECRETKEY`

	// Version is the semantic agent version.
	Version = `0.0.1`
)

type transportMap map[http.RoundTripper]http.RoundTripper

// Agent is the type of the Bearer entry point for your programs.
type Agent struct {
	m             sync.Mutex
	dispatcher    events.Dispatcher
	SecretKey     string
	config        *Config
	baseTransport http.RoundTripper
	transports    transportMap
	*proxy.Sender
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application,
// and pass a config.WithLogger(some *zerolog.Logger) config.Option.
func NewAgent(secretKey string, opts ...Option) (*Agent, error) {
	if !config.IsSecretKeyWellFormed(secretKey) {
		return nil, fmt.Errorf("secret key %s is not well-formed", secretKey)
	}
	a := Agent{
		baseTransport: unwrapTransport(http.DefaultClient.Transport),
		dispatcher:    events.NewDispatcher(),
		SecretKey:     secretKey,
		transports:    make(transportMap),
	}

	c, err := NewConfig(secretKey, a.baseTransport, Version, opts...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}

	a.config = c
	if c.IsDisabled() {
		return &a, errors.New(`remote config unavailable`)
	}

	a.Sender = proxy.NewSender(c.ReportOutstanding, c.ReportEndpoint, Version,
		c.SecretKey(), c.Environment(),
		a.DefaultTransport(), a.Logger())
	go a.Sender.Start()

	dcrp := interception.DCRProvider{DCRs: a.config.DataCollectionRules()}
	a.dispatcher.AddProviders(interception.TopicConnect, events.ListenerProviderFunc(a.Provider), dcrp)
	a.dispatcher.AddProviders(interception.TopicRequest, dcrp)
	a.dispatcher.AddProviders(interception.TopicResponse, dcrp)
	a.dispatcher.AddProviders(interception.TopicBodies, interception.BodyParsingProvider{}, dcrp)
	a.dispatcher.AddProviders(interception.TopicReport,
		interception.SanitizationProvider{
			SensitiveKeys:    a.config.SensitiveKeys(),
			SensitiveRegexps: a.config.SensitiveRegexps(),
		},
		interception.ProxyProvider{Sender: a.Sender},
	)
	return &a, nil
}

// DefaultTransport returns the original implementation of the http.DefaultTransport,
// even if it was overridden by the Agent in the meantime.
func (a *Agent) DefaultTransport() http.RoundTripper {
	return a.baseTransport
}

// Decorate wraps a http.RoundTripper with Bearer instrumentation.
func (a *Agent) Decorate(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &interception.RoundTripper{
		Dispatcher: a.dispatcher,
		Underlying: rt,
	}
}

// DecorateClientTransports wraps the http.RoundTripper transports in all passed
// clients, as well as the runtime library DefaultClient, with Bearer
// instrumentation.
func (a *Agent) DecorateClientTransports(clients ...*http.Client) {
	if a.config.IsDisabled() {
		return
	}
	if a.transports == nil {
		a.transports = make(transportMap)
	}

	allClients := append(clients, http.DefaultClient)

	a.m.Lock()
	defer a.m.Unlock()
	// Deduplicate transports to avoid multilayer decoration.
	for _, c := range allClients {
		ct := c.Transport
		_, ok := a.transports[ct]
		if ok {
			continue
		}
		a.transports[ct] = ct
	}

	// Decorate the deduplicated transports.
	for base := range a.transports {
		a.transports[base] = a.Decorate(base)
	}

	// Since we just built this map in a mutex lock consistency is guaranteed.
	for _, client := range allClients {
		client.Transport = a.transports[client.Transport]
	}
}

// DefaultAgent is a preconfigured agent logging to os.Stderr.
//
// To ensure complete compatibility with the runtime "log" package, be sure to
// log.SetOutput(DefaultAgent.Logger()).
var DefaultAgent *Agent

// uninitializedClose provides a final warning that the Bearer agent was not
// initialized during the application execution.
func uninitializedClose(err error) func() error {
	return func() error {
		log.Println(err)
		return err
	}
}

// XXX A placeholder for future logic, as in #BG-14 prevent early termination.
func close(a *Agent) func() error {
	return func() error {
		a.Trace(fmt.Sprintf(`End of Bearer agent operation with %d API calls logged`, a.Sender.Counter), nil)
		return nil
	}
}

// Init initializes the Bearer agent:
//   - it validates the user secret key
//   - it decorates the transport of the default client and of the clients it may receive.
//   - it returns a closing function which will ensure orderly termination of the
//     app, including flushing the list of records not yet transmitted to Bearer.
func Init(secretKey string, opts ...Option) func() error {
	var err error
	DefaultAgent, err = NewAgent(secretKey, opts...)
	if err != nil || DefaultAgent.config == nil || DefaultAgent.config.IsDisabled() {
		err = fmt.Errorf("did not initialize Bearer agent: %w", err)
		log.Println(err)
		return uninitializedClose(err)
	}
	DefaultAgent.DecorateClientTransports()
	return close(DefaultAgent)
}

// Provider provides the default agent listeners:
//   - TopicConnect: RFCListener, validating URL under RFC grammars.
//   - TopicRequest, TopicResponse, TopicBodies: no.
func (*Agent) Provider(e events.Event) []events.Listener {
	var l []events.Listener
	switch topic := e.Topic(); topic {
	case interception.TopicConnect:
		l = []events.Listener{
			interception.RFCListener,
		}
	}

	return l
}

func unwrapTransport(rt http.RoundTripper) http.RoundTripper {
	for {
		if base, ok := rt.(*interception.RoundTripper); ok {
			rt = base.Underlying
			continue
		}
		break
	}
	// If the underlying transport, some other package may modify it, so we cannot
	// rely on it being correct afterwards, so provide a default Transport matching
	// the standard values of the http.DefaultTransport.
	if rt == nil {
		rt = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	return rt
}
