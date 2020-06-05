package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/filters"
)

// OptionDefaults is an always-on Option loading built-in values.
var OptionDefaults Option = func(c *Config) error {
	c.runtimeEnvironmentType = DefaultRuntimeEnvironmentType
	c.configEndpoint = DefaultConfigEndpoint
	c.reportHost = DefaultReportHost
	c.reportOutstanding = DefaultReportOutstanding
	c.fetchInterval = DefaultConfigFetchInterval
	return nil
}

// OptionDisabled is a special non-configurable Option disabling the agent.
var OptionDisabled Option = func(c *Config) error {
	c.isDisabled = true
	return nil
}

// OptionEnvironment is an always-on Option loading values from the environment.
var OptionEnvironment Option = func(c *Config) error {
	if key, ok := os.LookupEnv(SecretKeyName); ok {
		if isSecretKeyWellFormed(key) {
			c.secretKey = key
		}
	}
	return nil
}

// WithDataCollectionRules is an Option configuring the data collection rules.
func WithDataCollectionRules(dcrs []DataCollectionRule) Option {
	return func(c *Config) error {
		c.dataCollectionRules = dcrs
		return nil
	}
}

// WithError is a functional Option for errors.
func WithError(err error) Option {
	return func(*Config) error {
		return err
	}
}

// WithFilters is a functional Option for filters.
func WithFilters(fs []filters.Filter) Option {
	return func(c *Config) error {
		c.filters = fs
		return nil
	}
}

// WithRemote is an always-on functional Option loading values from Bearer platform configuration.
func WithRemote(transport http.RoundTripper, logger *zerolog.Logger, version string) Option {
	return func(c *Config) error {
		fetcher := NewFetcher(transport, logger, version, c)
		fetcher.Fetch()
		return nil
	}
}

// WithRuntimeEnvironmentType is a functional Option configuring the runtime environment type.
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

// WithSecretKey is a functional Option setting the secret key if it is well-formed.
func WithSecretKey(secretKey string) Option {
	if !isSecretKeyWellFormed(secretKey) {
		return WithError(errors.New("secret key is not well-formed"))
	}
	return func(c *Config) error {
		c.secretKey = secretKey
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
			return WithError(errors.New("empty string may not be used as a sensitive key"))
		}
		dups[key]++
		if dups[key] > 1 {
			continue
		}
		reKey, err := regexp.Compile(key)
		if err != nil {
			return WithError(fmt.Errorf("invalid sensitive key regexp: %s", key))
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
			return WithError(errors.New("empty string may not be used as a sensitive regex"))
		}
		dups[re]++
		if dups[re] > 1 {
			continue
		}
		rer, err := regexp.Compile(re)
		if err != nil {
			return WithError(err)
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
