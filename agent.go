package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/rs/zerolog"
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
	m             sync.Mutex
	SecretKey     string
	logger        *zerolog.Logger
	config        *Config
	baseTransport http.RoundTripper
	transports    transportMap
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application.
func NewAgent(secretKey string, opts ...Option) (*Agent, error) {
	c, err := NewConfig(append(opts, WithSecretKey(secretKey))...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}
	a := Agent{
		SecretKey: secretKey,
		config:    c,
	}
	return &a, nil
}

// DefaultTransport returns the original implementation of the http.DefaultTransport,
// even if it was overridden by the Agent in the meantime.
func (a *Agent) DefaultTransport() http.RoundTripper {
	return a.baseTransport
}

// Decorate wraps a http.RoundTripper with Bearer instrumentation.
func (a *Agent) Decorate(rt http.RoundTripper) http.RoundTripper {
	return nil
}

// DecorateAll wraps the http.RoundTripper transports in all passed clients, as
// well as the runtime library DefaultClient, with Bearer instrumentation.
func (a *Agent) DecorateAll(clients []*http.Client) {
	a.baseTransport = http.DefaultClient.Transport
	if a.transports == nil {
		a.transports = make(transportMap)
	}

	allClients := append(clients, http.DefaultClient)

	a.m.Lock()
	defer a.m.Unlock()
	// Deduplicate transports to avoid multilayer decoration.
	for _, c := range allClients {
		rt, ok := a.transports[c.Transport]
		if ok {
			continue
		}
		a.transports[rt] = rt
	}

	// Decorate the deduplicated transports.
	for base, deco := range a.transports {
		a.transports[base] = a.Decorate(deco)
	}

	// Since we just built this map in a mutex lock consistency is guaranteed.
	for _, client := range clients {
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
	return a
}

// DefaultAgent is a preconfigured agent logging to os.Stderr.
//
// To ensure complete compatibility with the runtime "log" package, be sure to
// log.SetOutput(DefaultAgent.Logger()).
var DefaultAgent = (&Agent{
	transports: make(transportMap),
}).SetLogger(os.Stderr)

// TODO This is just a placeholder for future logic.
func close() error {
	log.Fatal()
	return nil
}

// Decorate wraps an existing transport (RoundTripper) with Bearer instrumentation.
func Decorate(rt http.RoundTripper) http.RoundTripper {
	return rt
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
