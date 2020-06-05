package filters

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
)

type testString string

func (ts testString) String() string {
	return string(ts)
}

func TestRegexMatcher_Matches(t *testing.T) {
	var (
		BearerMatcher = regexMatcher{Pattern: regexp.MustCompile(BearerRE)}
	)
	tests := []struct {
		r RegexpMatcher
		// check needs to be a string, fmt.Stringer, or error to succeed.
		check    interface{}
		expected bool
	}{
		{NewEmptyRegexpMatcher(), ``, true},
		{NewEmptyRegexpMatcher(), BearerDomain, true},
		{&BearerMatcher, ``, false},
		{&BearerMatcher, BearerDomain, true},
		{&BearerMatcher, errors.New(BearerDomain), true},
		{&BearerMatcher, testString(BearerDomain), true},
		{&BearerMatcher, 42, false},
		{&regexMatcher{}, ``, true},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s.Contains(%d)", tt.r, tt.check)
		t.Run(name, func(t *testing.T) {
			actual := tt.r.Matches(tt.check)
			if actual != tt.expected {
				t.Errorf("%s.Contains(%d): expected %t, got %t", tt.r, tt.check, tt.expected, actual)
			}
		})
	}
}

func TestRegexMatcher_Regexp(t *testing.T) {
	matcher := NewRegexpMatcher(BearerRE)
	rm, ok := matcher.(*regexMatcher)
	if !ok {
		t.Fatalf("expected %T matcher, got %T", NewEmptyRegexpMatcher(), matcher)
	}
	actual := rm.Regexp().String()
	expected := regexp.MustCompile(BearerRE).String()
	if actual != expected {
		t.Fatalf("incorrect regexp:\n  wanted %s\n  got %s", expected, actual)
	}
}
