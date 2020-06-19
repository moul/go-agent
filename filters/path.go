package filters

import (
	"fmt"

	"github.com/bearer/go-agent/events"
)

// PathFilter provides a filter for the path in API requests.
type PathFilter struct {
	RegexpMatcher
}

// Type is part of the Filter interface.
func (*PathFilter) Type() FilterType {
	return PathFilterType
}

func (f *PathFilter) ensureMatcher() {
	if f.RegexpMatcher != nil {
		return
	}
	_ = f.SetMatcher(NewEmptyRegexpMatcher())
}

// MatchesCall is part of the Filter interface.
func (f *PathFilter) MatchesCall(e events.Event) bool {
	f.ensureMatcher()
	criterium := e.Request().URL.Path
	return f.RegexpMatcher.Matches(criterium)
}

// SetMatcher sets the filter RegexpMatcher.
//
// If the returned error is not nil, the filter Matcher cannot be used.
//
// As per the Go runtime client, the setter accepts relative paths, which are
// likely not to be what you expect in API calls, so be sure to include leading
// slashes in most - if not all - cases.
//
// If the returned error is not nil, the filter Regex will accept any value.
//
// To apply a case-insensitive match, prepend (?i) to the regex, as in: (?i)\.bearer\.sh$
func (f *PathFilter) SetMatcher(matcher Matcher) error {
	if matcher == nil {
		matcher = NewEmptyRegexpMatcher()
	}
	rm, ok := matcher.(RegexpMatcher)
	if !ok {
		f.ensureMatcher()
		return fmt.Errorf("the PathFilter only accepts RegexMatchers: got %T", matcher)
	}
	f.RegexpMatcher = rm
	return nil
}

func pathFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	// FIXME apply RegexpMatcherDescription.Flags
	m := NewRegexpMatcher(fd.Pattern.Value)
	f := &PathFilter{}
	err := f.SetMatcher(m)
	if err != nil {
		return nil
	}
	return f
}
