package agent

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	// DefaultRuntimeEnvironmentType is the default environment type.
	DefaultRuntimeEnvironmentType = "default"

	// DefaultConfigHost is the default configuration server for Bearer.
	DefaultConfigHost = "config.bearer.sh"
	// DefaultReportHost is the default reporting host for Bearer.
	DefaultReportHost = "agent.bearer.sh"
	// SecretKeyRegex is the format of Bearer secret regexps.
	SecretKeyRegex = `^app_[[:xdigit:]]{50}$`
)

// DataCollectionRule represents a data collection rule.
// @FIXME Define actual type instead of placeholder.
type DataCollectionRule interface{}

// Filter implements a filtering rule.
// @FIXME Define actual type instead of placeholder.
type Filter interface{}

// Config represents the Agent configuration.
type Config struct {
	// Generation options
	isDisabled             bool
	runtimeEnvironmentType string
	secretKey              string

	// Sanitization options.
	sensitiveRegexes []*regexp.Regexp
	sensitiveKeys    []*regexp.Regexp

	// Rules.
	dataCollectionRules []DataCollectionRule
	filters             []Filter

	// Internal dev. options.
	configHost string
	reportHost string
}

// SecretKey is a getter for secretKey.
func (c *Config) SecretKey() string {
	return c.secretKey
}

func isSecretKeyWellFormed(k string) bool {
	re := regexp.MustCompile(SecretKeyRegex)
	return re.MatchString(k)
}

// IsDisabled is a getter for isDisabled, also checking whether the key is plausible.
func (c *Config) IsDisabled() bool {
	return c.isDisabled || !isSecretKeyWellFormed(c.secretKey)
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

// Option is the type use by functional options for configuration.
type Option func(*Config) error

// NewConfig is the Config constructor.
func NewConfig(opts ...Option) (*Config, error) {
	var err error
	c := Config{
		runtimeEnvironmentType: DefaultRuntimeEnvironmentType,
		configHost:             DefaultConfigHost,
		reportHost:             DefaultReportHost,
	}
	for _, opt := range opts {
		err = opt(&c)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Disabled is an Option disabling the agent.
func Disabled(c *Config) error {
	c.isDisabled = true
	return nil
}

// WithSecretKey is an Option setting the secret key if it is well-formed.
func WithSecretKey(secretKey string) Option {
	if !isSecretKeyWellFormed(secretKey) {
		return errorOption(errors.New("secret key is not well-formed"))
	}
	return func(c *Config) error {
		c.secretKey = secretKey
		return nil
	}
}

// WithRuntimeEnvironmentType is an Option configuring the runtime environment type.
//
// The environment type is a free-form tag for clients, allowing them to report
// which type of environment they are running in, like "development", "staging"
// or "production", for reporting in the Bearer UI.
//
// It allows clients to avoid the issues associated with having development and
// production metrics grouped together although they have different use profiles.
func WithRuntimeEnvironmentType(rtet string) Option {
	return func(c *Config) error {
		c.runtimeEnvironmentType = rtet
		return nil
	}
}

// WithDataCollectionRules is an Option configuring the data collection rules.
func WithDataCollectionRules(dcrs []DataCollectionRule) Option {
	return func(c *Config) error {
		c.dataCollectionRules = dcrs
		return nil
	}
}

// WithFilters is an Option configuring the filters.
func WithFilters(fs []Filter) Option {
	return func(c *Config) error {
		c.filters = fs
		return nil
	}
}

func errorOption(err error) Option {
	return func(*Config) error {
		return err
	}
}

// WithSensitiveKeys is an Option configuring the sensitive regexps.
//
// It will return an error if any key is empty. Duplicate regexps will be reduced
// to unique values to limit filtering costs.
func WithSensitiveKeys(keys []string) Option {
	dups := make(map[string]int, len(keys))
	var reduced []*regexp.Regexp
	for _, key := range keys {
		if key == "" {
			return errorOption(errors.New("empty string may not be used as a sensitive key"))
		}
		dups[key]++
		if dups[key] > 1 {
			continue
		}
		reKey, err := regexp.Compile(key)
		if err != nil {
			return errorOption(fmt.Errorf("invalid sensitive key regexp: %s", key))
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

// WithSensitiveRegexps is an Option configuring the sensitive regular expressions.
//
// It will cause an error if any of the regular expressions is invalid.
func WithSensitiveRegexps(res []string) Option {
	dups := make(map[string]int, len(res))
	var reduced []*regexp.Regexp
	for _, re := range res {
		if re == "" {
			return errorOption(errors.New("empty string may not be used as a sensitive regex"))
		}
		dups[re]++
		if dups[re] > 1 {
			continue
		}
		rer, err := regexp.Compile(re)
		if err != nil {
			return errorOption(err)
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
