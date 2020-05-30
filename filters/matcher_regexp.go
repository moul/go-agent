package filters

import (
	"fmt"
	"regexp"
)

// EmptyRegexp is the compiled regexp for the empty string.
var EmptyRegexp = regexp.MustCompile(``)

// NewEmptyRegexpMatcher provides a default all-accepting matcher.
var NewEmptyRegexpMatcher = func() RegexpMatcher {
	return &regexMatcher{EmptyRegexp}
}

// RegexpMatcher provides the ability to match agains a Go regular expression.
//
// By default, it matches anything.
type RegexpMatcher interface {
	Matcher
	Regexp() *regexp.Regexp
}

type regexMatcher struct {
	Pattern *regexp.Regexp
}

func (m *regexMatcher) String() string {
	if m.Pattern == nil {
		return ``
	}
	return m.Pattern.String()
}

func (m *regexMatcher) Matches(x interface{}) bool {
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

func (m *regexMatcher) Regexp() *regexp.Regexp {
	return m.Pattern
}

// NewRegexpMatcher creates a RangeMatcher.
//   - If the regex is invalid, the matcher will be nil.
//   - Otherwise it will be a usable matcher.
func NewRegexpMatcher(s string) RegexpMatcher {
	re, err := regexp.Compile(s)
	if err != nil {
		return nil
	}
	rm := regexMatcher{
		Pattern: re,
	}
	return &rm
}
