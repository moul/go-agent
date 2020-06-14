package config

import (
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/filters"
	"github.com/bearer/go-agent/interception"
)

const (

	// SecretKeyName is the environment variable used to hold the Bearer secret key,
	// specific to each client. Fetching the secret key from the environment is a
	// best practice in 12-factor application development.
	SecretKeyName = `BEARER_SECRETKEY`

	// DefaultRuntimeEnvironmentType is the default environment type.
	DefaultRuntimeEnvironmentType = "development" // "default"

	// DefaultConfigEndpoint is the default configuration endpoint for Bearer.
	DefaultConfigEndpoint = "https://config.bearer.sh/config"

	// DefaultConfigFetchInterval is the default rate at which the Agent will
	// asynchronously fetch configuration refreshes from Bearer.
	DefaultConfigFetchInterval = 5 * time.Second

	// DefaultReportEndpoint is the default reporting endpoint for Bearer.
	DefaultReportEndpoint = "https://agent.bearer.sh/logs"

	// DefaultReportOutstanding it the default maximum number of pending data
	// collection writes in flight at any given time. When that limit is
	// exceeded, records are no longer sent to Bearer to avoid saturating the
	// client.
	DefaultReportOutstanding = 1000
)

// SecretKeyRegex is the format of Bearer secret keys.
// It is used to verify the shape of submitted secret keys, before they are
// sent over to Bearer for value validation.
var SecretKeyRegex = regexp.MustCompile(`^app_[[:xdigit:]]{50}$`)

// Config represents the Agent configuration.
type Config struct {
	// Generation options
	isDisabled             bool
	runtimeEnvironmentType string
	secretKey              string

	// Sanitization options.
	sensitiveRegexes []*regexp.Regexp // Named per Agent spec, although Go uses "regexp".
	sensitiveKeys    []*regexp.Regexp

	// Rules.
	dataCollectionRules []*interception.DataCollectionRule
	Rules               []interface{} // XXX Agent spec defines the field but no use for it.
	filters             filters.FilterMap

	// Internal dev. options.
	configEndpoint    string
	fetchInterval     time.Duration
	ReportEndpoint    string
	ReportOutstanding uint

	// Internal runtime properties.
	fetcher *Fetcher
	logger  *zerolog.Logger
	sync.Mutex
}

// DisableRemote stops the goroutine updating the Agent configuration periodically.
func (c *Config) DisableRemote() {
	if c.fetcher == nil {
		return
	}
	c.fetcher.Stop()
	c.fetcher = nil
}

// SecretKey is a getter for secretKey.
func (c *Config) SecretKey() string {
	return c.secretKey
}

// IsSecretKeyWellFormed verifies whether the secret key matches the expected format.
func IsSecretKeyWellFormed(secretKey string) bool {
	return SecretKeyRegex.MatchString(secretKey)
}

// IsDisabled is a getter for isDisabled, also checking whether the key is plausible.
func (c *Config) IsDisabled() bool {
	return c.isDisabled || !IsSecretKeyWellFormed(c.secretKey)
}

// RuntimeEnvironmentType is a getter for runtimeEnvironmentType.
func (c *Config) RuntimeEnvironmentType() string {
	return c.runtimeEnvironmentType
}

// SensitiveKeys is a getter for sensitiveKeys.
func (c *Config) SensitiveKeys() []*regexp.Regexp {
	return c.sensitiveKeys
}

// SensitiveRegexps is a getter for sensitiveRegexps.
func (c *Config) SensitiveRegexps() []*regexp.Regexp {
	return c.sensitiveRegexes
}

// UpdateFromDescription overrides the Config with configuration generated from
// a configuration Description.
func (c *Config) UpdateFromDescription(description Description) {
	c.Lock()
	defer c.Unlock()
	filterDescriptions, err := description.filterDescriptions()
	if err != nil {
		c.logger.Warn().Msgf(`invalid configuration received from config server: %v`, err)
		return
	}
	resolved, err := description.resolveHashes(filterDescriptions)
	if err != nil {
		c.logger.Warn().Msgf(`incorrect filter resolution in configuration received from config server: %v`, err)
		return
	}
	c.filters = resolved

	dcrs, err := description.resolveDCRs(resolved)
	if err != nil {
		c.logger.Warn().Err(err).Msg(`resolving data collection rules`)
		return
	}
	c.dataCollectionRules = dcrs
}

// DataCollectionRules returns the active DataCollectionRule instances.
func (c *Config) DataCollectionRules() []*interception.DataCollectionRule {
	return c.dataCollectionRules
}

// Option is the type use by functional options for configuration.
type Option func(*Config) error

// NewConfig is the default Config constructor: it builds a configuration from
// the builtin agent defaults, the environment, the Bearer platform configuration
// and any optional Option values passed by the caller.
func NewConfig(transport http.RoundTripper, logger *zerolog.Logger, version string, opts ...Option) (*Config, error) {
	alwaysOn := []Option{
		OptionDefaults,
		OptionEnvironment,
		WithRemote(transport, logger, version), // Sets Fetcher.
	}

	options := append(alwaysOn, opts...)
	c := &Config{}
	for _, withOption := range options {
		err := withOption(c)
		if err != nil {
			return nil, err
		}
	}
	c.fetcher.Start()
	return c, nil
}
