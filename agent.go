package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/interception"
	"github.com/bearer/go-agent/proxy"
)

// ExampleWellFormedInvalidKey is a well-formed key known to be invalid. It may
// be used for integration test scenarios.
const ExampleWellFormedInvalidKey = `app_12345678901234567890123456789012345678901234567890`

// Version is the semantic agent version.
const Version = `0.0.1`

type transportMap map[http.RoundTripper]http.RoundTripper

// Agent is the type of the Bearer entry point for your programs.
type Agent struct {
	m sync.Mutex
	events.Dispatcher
	SecretKey     string
	logger        *zerolog.Logger
	config        *config.Config
	baseTransport http.RoundTripper
	transports    transportMap
	*proxy.Sender
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application,
// and pass a zerolog.Logger instance for the logger.
func NewAgent(secretKey string, logger io.Writer, opts ...config.Option) (*Agent, error) {
	if !config.IsSecretKeyWellFormed(secretKey) {
		return nil, fmt.Errorf("secret key %s is not well-formed", secretKey)
	}
	a := Agent{
		baseTransport: unwrapTransport(http.DefaultClient.Transport),
		Dispatcher:    events.NewDispatcher(),
		SecretKey:     secretKey,
		transports:    make(transportMap),
	}
	a.SetLogger(logger)

	secretKeyOption := config.WithSecretKey(secretKey)
	c, err := config.NewConfig(
		unwrapTransport(http.DefaultClient.Transport),
		a.logger,
		Version,
		append([]config.Option{secretKeyOption}, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}
	a.config = c
	a.Sender = proxy.NewSender(c.ReportOutstanding, c.ReportEndpoint, Version,
		c.SecretKey(), c.RuntimeEnvironmentType(),
		a.DefaultTransport(), a.Logger())
	go a.Sender.Start()

	dcrp := interception.DCRProvider{DCRs: a.config.DataCollectionRules()}
	a.Dispatcher.AddProviders(interception.TopicConnect, events.ListenerProviderFunc(a.Provider), dcrp)
	a.Dispatcher.AddProviders(interception.TopicRequest, dcrp)
	a.Dispatcher.AddProviders(interception.TopicResponse, dcrp)
	a.Dispatcher.AddProviders(interception.TopicBodies, interception.BodyParsingProvider{},dcrp)
	a.Dispatcher.AddProviders(interception.TopicReport,
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
		Dispatcher: a.Dispatcher,
		Underlying: rt,
	}
}

// DecorateClientTransports wraps the http.RoundTripper transports in all passed
// clients, as well as the runtime library DefaultClient, with Bearer
// instrumentation.
func (a *Agent) DecorateClientTransports(clients ...*http.Client) {
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
func uninitializedClose(err error) func () error {
	return func () error {
		log.Println(err)
		return err
	}
}

// XXX A placeholder for future logic, as in #BG-14 prevent early termination.
func close(a *Agent) func () error {
	return func() error {
		log.Printf(`End of Bearer agent operation with %d API calls logged`, a.Sender.Counter)
		return nil
	}
}

// Init initializes the Bearer agent:
//   - it validates the user secret key
//   - it decorates the transport of the default client and of the clients it may receive.
//   - it returns a closing function which will ensure orderly termination of the
//     app, including flushing the list of records not yet transmitted to Bearer.
func Init(secretKey string, opts ...config.Option) func() error {
	var err error
	DefaultAgent, err = NewAgent(secretKey, os.Stderr, opts...)
	if err != nil {
		err = fmt.Errorf("could not initialize Bearer agent: %w", err)
		log.Println(err)
		return uninitializedClose(err)
	}
	DefaultAgent.DecorateClientTransports()
	return close(DefaultAgent)
}

// Provider provides the default agent listeners:
//   - TopicConnect: RFCListener, validating URL under RFC grammars.
//   - TopicRequest, TopicResponse, TopicBodies: no.
func (a *Agent) Provider(e events.Event) []events.Listener {
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
	return rt
}
