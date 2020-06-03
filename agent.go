package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/interception"
)

// SecretKeyName is the environment variable used to hold the Bearer secret key,
// specific to each client. Fetching the secret key from the environment is a
// best practice in 12-factor application development.
const SecretKeyName = `BEARER_SECRETKEY`

// SecretKeyPattern is the text for the regular expression validating the shape
// of submitted secret keys, before they are sent over to Bearer for value validation.
const SecretKeyPattern = `app_[0-9]{50}`

// ExampleWellFormedInvalidKey is a well-formed key known to be invalid. It may
// be used for integration test scenarios.
const ExampleWellFormedInvalidKey = `app_12345678901234567890123456789012345678901234567890`

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
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application.
func NewAgent(secretKey string, opts ...config.Option) (*Agent, error) {
	c, err := config.NewConfig(append(opts, config.WithSecretKey(secretKey))...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}
	a := Agent{
		Dispatcher: events.NewDispatcher(),
		SecretKey:  secretKey,
		config:     c,
	}

	a.Dispatcher.AddProviders(interception.Connect,
		events.ListenerProviderFunc(a.Provider),
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
		Config:     a.config,
		Underlying: rt,
	}
}

// DecorateClientTransports wraps the http.RoundTripper transports in all passed clients, as
// well as the runtime library DefaultClient, with Bearer instrumentation.
func (a *Agent) DecorateClientTransports(clients ...*http.Client) {
	a.baseTransport = http.DefaultClient.Transport
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

// Init initializes the agent with a Bearer secret key, and specifies which HTTP
// clients it needs to decorate in addition to the http.DefaultClient.
func (a *Agent) Init(secretKey string, clients ...*http.Client) *Agent {
	reKey := regexp.MustCompile(SecretKeyPattern)
	if !reKey.MatchString(secretKey) {
		a.logger.Error().Msgf("attempting Init with ill-formed secret key: [%s]", secretKey)
		return nil
	}
	a.DecorateClientTransports(clients...)
	return a
}

// DefaultAgent is a preconfigured agent logging to os.Stderr.
//
// To ensure complete compatibility with the runtime "log" package, be sure to
// log.SetOutput(DefaultAgent.Logger()).
var DefaultAgent = (&Agent{
	Dispatcher: events.NewDispatcher(),
	transports: make(transportMap),
}).SetLogger(os.Stderr)

// TODO This is just a placeholder for future logic.
func close() error {
	log.Fatal(`End of Bearer agent operation`)
	return nil
}

// Decorate wraps any HTTP transport in Bearer instrumentation, returning an
// equivalent instrumented transport.
func Decorate(rt http.RoundTripper) http.RoundTripper {
	return DefaultAgent.Decorate(rt)
}

// Init initializes the Bearer agent:
//   - it validates the user secret key
//   - it decorates the transport of the default client and of the clients it may receive.
//   - it returns a closing function which will ensure orderly termination of the
//     app, including flushing the list of records not yet transmitted to Bearer.
func Init(secretKey string, clients ...*http.Client) func() error {
	DefaultAgent.Init(secretKey, clients...)
	return close
}

// Provider provides the default agent listeners:
//   - Connect: RFCListener, validating URL under RFC grammars.
//   - Request, Response, Bodies: none yet.
func (a *Agent) Provider(e events.Event) []events.Listener {
	var l []events.Listener
	switch e.Topic() {
	case interception.Connect:
		l = []events.Listener{interception.RFCListener}
	default:
		// TODO define and implement other build-in listeners, e.g DataCollectionListener.
	}

	return l
}
