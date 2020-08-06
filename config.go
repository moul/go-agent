package agent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/filters"
	"github.com/bearer/go-agent/interception"
)

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
	fetchEndpoint     string
	fetchInterval     time.Duration
	ReportEndpoint    string
	ReportOutstanding uint

	// Internal runtime properties.
	fetcher *config.Fetcher
	*zerolog.Logger
	sync.Mutex
}

// optionDefaults is an always-on Option loading built-in values.
var optionDefaults Option = func(c *Config) error {
	c.Logger = config.DefaultLogger()
	c.fetchEndpoint = config.DefaultConfigEndpoint
	c.ReportEndpoint = config.DefaultReportEndpoint
	c.ReportOutstanding = config.DefaultReportOutstanding
	c.fetchInterval = config.DefaultFetchInterval
	c.sensitiveKeys = []*regexp.Regexp{interception.DefaultSensitiveKeys}
	c.sensitiveRegexes = []*regexp.Regexp{interception.DefaultSensitiveData}
	return nil
}

// optionEnvironment is an always-on Option loading values from the environment.
// In this version, it overrides the secret key passed manually if it is not
// well-formed, as a fallback security.
var optionEnvironment Option = func(c *Config) error {
	if config.IsSecretKeyWellFormed(c.secretKey) {
		return nil
	}
	if secretKey, ok := os.LookupEnv(SecretKeyName); ok {
		if config.IsSecretKeyWellFormed(secretKey) {
			c.secretKey = secretKey
		}
	}
	return nil
}

// WithDisabled is a functional Option to disable the agent
func WithDisabled(value bool) Option {
	return func(c *Config) error {
		c.isDisabled = value
		return nil
	}
}

// withError is a functional Option for errors.
func withError(err error) Option {
	return func(*Config) error {
		return err
	}
}

// WithLogger is a functional Option for the logger.
func WithLogger(logger io.Writer) Option {
	return func(c *Config) error {
		zl, ok := logger.(*zerolog.Logger)
		if !ok {
			l := zerolog.New(logger)
			zl = &l
		}
		if c != nil {
			c.Logger = zl
			return nil
		}
		return errors.New(`cannot set logger on nil Config`)
	}
}

// withRemote is an always-on functional Option loading values from Bearer platform configuration.
func withRemote(transport http.RoundTripper, version string) Option {
	return func(c *Config) error {
		c.fetcher = config.NewFetcher(transport, c.Logger, version, c.fetchEndpoint, c.fetchInterval, c.runtimeEnvironmentType, c.secretKey)
		d, err := c.fetcher.Fetch()
		if err != nil {
			c.isDisabled = true
			return nil
		}
		c.UpdateFromDescription(d)
		return nil
	}
}

// WithEnvironment is a functional Option configuring the runtime environment type.
//
// The environment type is a free-form tag for clients, allowing them to report
// which type of environment they are running in, like "development", "staging"
// or "production", for reporting in the Bearer UI.
//
// It allows clients to avoid the issues associated with having development and
// production metrics grouped together although they have different use profiles.
func WithEnvironment(rtet string) Option {
	return func(c *Config) error {
		c.runtimeEnvironmentType = rtet
		return nil
	}
}

// withSecretKey is a functional Option setting the secret key if it is well-formed.
func withSecretKey(secretKey string) Option {
	return func(c *Config) error {
		if c.secretKey == `` {
			if !config.IsSecretKeyWellFormed(secretKey) {
				c.isDisabled = true
				return errors.New(`ill-formed secret key`)
			}
			c.secretKey = secretKey
		}
		return nil
	}
}

// WithSensitiveKeys is a functional Option configuring the sensitive regexps.
//
// It will return an error if any key is empty. Duplicate regexps will be reduced
// to unique values to limit filtering costs.
func WithSensitiveKeys(keys []string) Option {
	dups := make(map[string]int, len(keys))
	var reduced []*regexp.Regexp
	for _, key := range keys {
		if key == "" {
			return withError(errors.New("empty string may not be used as a sensitive key"))
		}
		dups[key]++
		if dups[key] > 1 {
			continue
		}
		reKey, err := regexp.Compile(key)
		if err != nil {
			return withError(fmt.Errorf("invalid sensitive key regexp: %s", key))
		}
		reduced = append(reduced, reKey)
	}
	// For non-nil empty slice, return a non-nil empty slice too.
	if keys != nil && reduced == nil {
		reduced = make([]*regexp.Regexp, 0)
	}
	return func(c *Config) error {
		c.sensitiveKeys = reduced
		return nil
	}
}

// WithSensitiveRegexps is a functional Option configuring the sensitive regular expressions.
//
// It will cause an error if any of the regular expressions is invalid.
func WithSensitiveRegexps(res []string) Option {
	dups := make(map[string]int, len(res))
	var reduced []*regexp.Regexp
	for _, re := range res {
		if re == "" {
			return withError(errors.New("empty string may not be used as a sensitive regex"))
		}
		dups[re]++
		if dups[re] > 1 {
			continue
		}
		rer, err := regexp.Compile(re)
		if err != nil {
			return withError(err)
		}
		reduced = append(reduced, rer)
	}

	// For non-nil empty slice, return a non-nil empty slice too.
	if res != nil && reduced == nil {
		reduced = []*regexp.Regexp{}
	}
	return func(c *Config) error {
		c.sensitiveRegexes = reduced
		return nil
	}
}

// WithEndpoints is an undocumented functional Option used for development
// purposes.
func WithEndpoints(fetchEndpoint string, reportEndpoint string) Option {
	return func(c *Config) error {
		c.fetchEndpoint = fetchEndpoint
		c.ReportEndpoint = reportEndpoint
		return nil
	}
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

// IsDisabled is a getter for isDisabled, also checking whether the key is plausible.
func (c *Config) IsDisabled() bool {
	return c == nil || c.isDisabled || !config.IsSecretKeyWellFormed(c.secretKey)
}

// Environment is a getter for runtimeEnvironmentType.
func (c *Config) Environment() string {
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

// DataCollectionRules returns the active DataCollectionRule instances.
func (c *Config) DataCollectionRules() []*interception.DataCollectionRule {
	return c.dataCollectionRules
}

// Option is the type use by functional options for configuration.
type Option func(*Config) error

// NewConfig is the default Config constructor: it builds a configuration from
// the builtin agent defaults, the environment, the Bearer platform configuration
// and any optional Option values passed by the caller.
func NewConfig(secretKey string, transport http.RoundTripper, version string, opts ...Option) (*Config, error) {
	alwaysOnBefore := []Option{
		optionDefaults,
		optionEnvironment,
		withSecretKey(secretKey),
	}

	alwaysOnAfter := []Option{
		withRemote(transport, version), // Sets Fetcher.
	}

	options := append(append(alwaysOnBefore, opts...), alwaysOnAfter...)
	c := &Config{}
	for _, withOption := range options {
		err := withOption(c)
		if err != nil {
			return nil, err
		}
	}
	if !c.IsDisabled() {
		c.fetcher.Start(func(description *config.Description) {
			c.UpdateFromDescription(description)
		})
	}
	if c.Logger == nil {
		_ = WithLogger(os.Stderr)(c)
	}
	return c, nil
}

// UpdateFromDescription overrides the Config with configuration generated from
// a configuration Description.
func (c *Config) UpdateFromDescription(description *config.Description) {
	c.Lock()
	defer c.Unlock()
	filterDescriptions, err := description.FilterDescriptions()
	if err != nil {
		c.Warn().Msgf(`invalid configuration received from config server: %v`, err)
		return
	}
	resolved, err := description.ResolveHashes(filterDescriptions)
	if err != nil {
		c.Warn().Msgf(`incorrect filter resolution in configuration received from config server: %v`, err)
		return
	}
	c.filters = resolved

	dcrs, err := description.ResolveDCRs(resolved)
	if err != nil {
		c.Warn().Err(err).Msg(`resolving data collection rules`)
		return
	}
	c.dataCollectionRules = dcrs
}
