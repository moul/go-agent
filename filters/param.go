package filters

import (
	"errors"
	"fmt"

	"github.com/bearer/go-agent/events"
)

// ParamFilter provides a key-value filter for API request parameters.
type ParamFilter struct {
	KeyValueMatcher
}

// Type is part of the Filter interface.
func (*ParamFilter) Type() FilterType {
	return ParamFilterType
}

// MatchesCall is part of the Filter interface.
func (f *ParamFilter) MatchesCall(e events.Event) bool {
	m := NewKeyValueMatcher(f.KeyRegexp().String(), f.ValueRegexp().String())
	u := e.Request().URL
	if u == nil {
		return false
	}
	return m.Matches(u.Query())
}

// SetMatcher sets the filter KeyValueMatcher.
//
// If the returned error is not nil, the filter Regex will accept any value.
//
// To apply a case-insensitive match, prepend (?i) to the matcher regexps,
// as in: (?i)\.bearer\.sh$
func (f *ParamFilter) SetMatcher(matcher Matcher) error {
	defaultMatcher := NewKeyValueMatcher(``, ``)

	m, ok := matcher.(KeyValueMatcher)
	if !ok {
		f.KeyValueMatcher = defaultMatcher
		return fmt.Errorf("key-value matcher expected, got a %T", matcher)
	}

	if isNilInterface(m) {
		f.KeyValueMatcher = defaultMatcher
		return errors.New("set nil Key-Value matcher on Param filter")
	}

	f.KeyValueMatcher = m
	return nil
}

func paramFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	// FIXME apply RegexpMatcherDescription.Flags.
	m := NewKeyValueMatcher(fd.KeyPattern.Value, fd.ValuePattern.Value)
	if m == nil {
		return nil
	}
	f := &ParamFilter{}
	err := f.SetMatcher(m)
	if err != nil {
		return nil
	}
	return f
}
