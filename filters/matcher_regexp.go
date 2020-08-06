package filters

import (
	"fmt"
	"regexp"
)

// EmptyRegexp is the compiled regexp for the empty string.
var EmptyRegexp = regexp.MustCompile(``)

// NewEmptyRegexpMatcher provides a default all-accepting matcher.
var NewEmptyRegexpMatcher = func() RegexpMatcher {
	return &regexpMatcher{EmptyRegexp}
}

// RegexpMatcher provides the ability to match agains a Go regular expression.
//
// By default, it matches anything.
type RegexpMatcher interface {
	Matcher
	Regexp() *regexp.Regexp
}

type regexpMatcher struct {
	Pattern *regexp.Regexp
}

func (m *regexpMatcher) String() string {
	if m.Pattern == nil {
		return ``
	}
	return m.Pattern.String()
}

func (m *regexpMatcher) Matches(x interface{}) bool {
	if m.Pattern == nil {
		return true
	}

	switch y := x.(type) {
	case string:
		return m.Pattern.MatchString(y)
	case fmt.Stringer:
		return m.Pattern.MatchString(y.String())
	case error:
		return m.Pattern.MatchString(y.Error())
	}

	return false
}

func (m *regexpMatcher) Regexp() *regexp.Regexp {
	return m.Pattern
}

// NewRegexpMatcher creates a RangeMatcher.
func NewRegexpMatcher(re *regexp.Regexp) RegexpMatcher {
	return &regexpMatcher{
		Pattern: re,
	}
}

// RegexpMatcherDescription is a serialization-friendly description of a RegexpMatcher.
type RegexpMatcherDescription struct {
	// Flags is a string of the regexp flags
	Flags string
	// Value is the string form of the regexp.
	Value string
}

// String implements fmt.Stringer.
func (d RegexpMatcherDescription) String() string {
	if d.Value == `` {
		return ``
	}
	return fmt.Sprintf("Regexp: /%s/%s\n", d.Value, d.Flags)
}
