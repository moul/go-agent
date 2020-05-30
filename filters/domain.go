package filters

import (
	"fmt"
	"net/http"
)

// DomainFilter provides a filter for the host name in API requests.
type DomainFilter struct {
	RegexpMatcher
}

// Type is part of the Filter interface.
func (*DomainFilter) Type() FilterType {
	return domainFilter
}

func (f *DomainFilter) ensureMatcher() {
	if f.RegexpMatcher != nil {
		return
	}
	if err := f.SetMatcher(NewEmptyRegexpMatcher()); err != nil {
		// Should not happen, by code structure.
		panic(err)
	}
}

// MatchesCall is part of the Filter interface.
func (f *DomainFilter) MatchesCall(r *http.Request, _ *http.Response) bool {
	f.ensureMatcher()
	criterium := r.URL.Hostname()
	return f.RegexpMatcher.Matches(criterium)
}

// SetMatcher sets the filter RegexpMatcher.
//
// If the returned error is not nil, the filter Matcher cannot be used.
//
// If the returned error is not nil, the filter Regex will accept any value.
//
// To apply a case-insensitive match, prepend (?i) to the regex, as in: (?i)\.bearer\.sh$
// DomainFilter should always use a case-insensitive match.
func (f *DomainFilter) SetMatcher(matcher Matcher) error {
	if matcher == nil {
		matcher = NewEmptyRegexpMatcher()
	}
	rm, ok := matcher.(RegexpMatcher)
	if !ok {
		f.ensureMatcher()
		return fmt.Errorf("the DomainFilter only accepts RegexMatchers: got %T", matcher)
	}
	f.RegexpMatcher = rm
	return nil
}
