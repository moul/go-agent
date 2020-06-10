package config_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/config"
)

// This key is well-formed but invalid.
const wellFormedKey = "app_01234567890123456789012345678901234567890123456789"

func TestConfig_Default(t *testing.T) {
	actual, err := config.NewConfig(config.WithSecretKey(wellFormedKey))
	if err != nil {
		t.Errorf("failed building default config")
	}
	if actual.IsDisabled() {
		t.Errorf("incorrect default for isDisabled")
	}
	if actual.RuntimeEnvironmentType() != config.DefaultRuntimeEnvironmentType {
		t.Errorf("incorrect default for runtimeEnvironmentType")
	}
	if actual.SecretKey() != wellFormedKey {
		t.Errorf("incorrect default for secretKey")
	}
}

func TestConfigInvalidSecretKey(t *testing.T) {
	const key = "invalid key"
	actual, err := config.NewConfig(config.WithSecretKey(key))
	if err == nil {
		t.Errorf("failed building default config")
	}
	if actual != nil {
		t.Errorf("built config from invalid key")
	}
}

func TestConfig_WithoutKey(t *testing.T) {
	actual, err := config.NewConfig()
	if err != nil {
		t.Errorf("failed building config without a secret key")
	}
	if !actual.IsDisabled() {
		t.Errorf("incorrect default for isDisabled without a secret key")
	}
	if actual.SecretKey() != "" {
		t.Errorf("incorrect default for missing secretKey")
	}
}

func TestConfig_Disabled(t *testing.T) {
	actual, err := config.NewConfig(config.WithSecretKey(wellFormedKey), config.Disabled)
	if err != nil {
		t.Errorf("failed building disabled config")
	}
	if !actual.IsDisabled() {
		t.Errorf("incorrect isDisabled for disabled config")
	}
}

func TestConfig_WithRuntimeEnvironmentType(t *testing.T) {
	const expected = "production"
	c, err := config.NewConfig(
		config.WithSecretKey(wellFormedKey),
		config.WithRuntimeEnvironmentType(expected),
	)
	if err != nil {
		t.Errorf("failed building config with environment type")
	}
	actual := c.RuntimeEnvironmentType()
	if actual != expected {
		t.Errorf("incorrect environment type: expected %s, got %s", expected, actual)
	}
}

func TestConfig_WithSensitiveKeys(t *testing.T) {
	type testType struct {
		name     string
		keys     []string
		wantFail bool
		expected []string
	}
	tests := []testType{
		{"nil", nil, false, nil},
		{"empty", []string{}, false, []string{}},
		{"normal", []string{"one", "two"}, false, []string{"one", "two"}},
		{"duplicated", []string{"one", "two", "one"}, false, []string{"one", "two"}},
		{"contains empty", []string{"one", "", "two"}, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := config.NewConfig(
				config.WithSecretKey(wellFormedKey),
				config.WithSensitiveKeys(tt.keys),
			)
			if err != nil && !tt.wantFail {
				t.Fatal("failed building disabled config")
			} else if err == nil && tt.wantFail {
				t.Fatal("built config in spite of invalid sensitive key")
			}
			if tt.wantFail {
				return
			}
			actualKeys := c.SensitiveKeys()
			var actualStrings []string
			if actualKeys != nil {
				actualStrings = make([]string, len(actualKeys))
				for i := 0; i < len(actualKeys); i++ {
					actualStrings[i] = actualKeys[i].String()
				}
			}
			if !reflect.DeepEqual(actualStrings, tt.expected) {
				t.Errorf("for %s case, expected %v, but got %v", tt.name, tt.expected, actualKeys)
			}
		})
	}
}

func TestConfig_WithSensitiveRegexes(t *testing.T) {
	type testType struct {
		name     string
		regexps  []string
		wantFail bool
		expected []*regexp.Regexp
	}
	reOne := regexp.MustCompile("one")
	reTwo := regexp.MustCompile("two")
	tests := []testType{
		{"nil", nil, false, nil},
		{"empty", []string{}, false, []*regexp.Regexp{}},
		{"normal", []string{"one", "two"}, false, []*regexp.Regexp{reOne, reTwo}},
		{"duplicated", []string{"one", "two", "one"}, false, []*regexp.Regexp{reOne, reTwo}},
		{"contains empty", []string{"one", "", "two"}, true, nil},
		{"contains invalid", []string{"one", "t[wo"}, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := config.NewConfig(
				config.WithSecretKey(wellFormedKey),
				config.WithSensitiveRegexps(tt.regexps),
			)
			if err != nil && !tt.wantFail {
				t.Fatal("failed building disabled config")
			} else if err == nil && tt.wantFail {
				t.Fatal("built config in spite of invalid sensitive regex")
			}
			if tt.wantFail {
				return
			}
			actual := agent.SensitiveRegexps()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("for %s case, expected %v, but got %v", tt.name, tt.expected, actual)
			}
		})
	}
}