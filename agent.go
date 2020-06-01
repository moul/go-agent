package agent

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// SecretKeyName is the environment variable used to hold the Bearer secret key,
// specific to each client.
const SecretKeyName = `BEARER_SECRETKEY`

// SecretKeyPattern is the text for the regular expression validating the shape
// of submitted secret keys, before they are sent over to Bearer for value validation.
const SecretKeyPattern = `app_[0-9]{50}`

// ExampleWellFormedInvalidKey is a well-formed key known to be invalid. It may
// be used for integration test scenarios.
const ExampleWellFormedInvalidKey = `app_12345678901234567890123456789012345678901234567890`

// Agent is the type of the Bearer entry point for your programs.
type Agent struct {
	logger *zerolog.Logger
	config *Config
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application.
func NewAgent(secretKey string, opts ...Option) (*Agent, error) {
	c, err := NewConfig(append(opts, WithSecretKey(secretKey))...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}
	a := Agent{config: c}
	return &a, nil
}

func close() error {
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
	for _, client := range append(clients, http.DefaultClient) {
		client.Transport = Decorate(client.Transport)
	}
	return close
}
