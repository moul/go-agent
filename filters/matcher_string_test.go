package filters

import (
	"errors"
	"testing"
)

func Test_stringMatcher_IgnoresCase(t *testing.T) {
	tests := []struct {
		name       string
		useDefault bool
		ignoreCase bool
		want       bool
	}{
		{"happy true", false, true, true},
		{"happy false", false, false, false},
		{"happy default", true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m *stringMatcher
			if tt.name == "happy default" {
				m = &stringMatcher{}
			} else {
				m = &stringMatcher{ignoreCase: tt.ignoreCase}
			}

			if got := m.IgnoresCase(); got != tt.want {
				t.Errorf("IgnoresCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringMatcher_Matches(t *testing.T) {
	tests := []struct {
		name string
		s    string
		x    interface{}
		want bool
	}{
		{"happy string", foo, foo, true},
		{"happy error", foo, errors.New(foo), true},
		{"happy stringer", foo, kvStringer(foo), true},
		{"sad string", foo, bar, false},
		{"sad non stringable", foo, 42, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &stringMatcher{s: tt.s}
			if got := m.Matches(tt.x); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
