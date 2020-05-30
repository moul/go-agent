package filters

import (
	"fmt"
	"strings"
)

// StringMatcher provides the ability to match against a string, possibly ignoring case.
//
// By default, it matches the empty string.
type StringMatcher interface {
	Matches(x interface{}) bool
	fmt.Stringer
	IgnoresCase() bool
}

type stringMatcher struct {
	s          string
	ignoreCase bool
}

func (m *stringMatcher) Matches(x interface{}) bool {
	var s string
	switch x.(type) {
	case string, fmt.Stringer, error:
		s = stringify(x).(string)
		if m.ignoreCase {
			return strings.EqualFold(s, m.s)
		}
		return s == m.s

	default:
		return false
	}
}

func (m *stringMatcher) IgnoresCase() bool {
	return m.ignoreCase
}

func (m *stringMatcher) String() string {
	return m.s
}

// NewStringMatcher creates a StringMatcher.
//
// It will never be invalid, its default value will match the empty string.
func NewStringMatcher(s string, ignoreCase bool) StringMatcher {
	return &stringMatcher{
		s:          strings.ToValidUTF8(s, ``),
		ignoreCase: ignoreCase,
	}
}
