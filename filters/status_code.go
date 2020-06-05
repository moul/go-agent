package filters

import (
	"fmt"
	"net/http"
)

// StatusCodeFilter provides a filter for the response status code in API requests.
type StatusCodeFilter struct {
	RangeMatcher
}

// Type is part of the Filter interface.
func (*StatusCodeFilter) Type() FilterType {
	return statusCodeFilter
}

func (f *StatusCodeFilter) ensureMatcher() {
	if f.RangeMatcher != nil {
		return
	}
	_ = f.SetMatcher(NewHTTPStatusMatcher())
}

// MatchesCall is part of the Filter interface.
func (f *StatusCodeFilter) MatchesCall(_ *http.Request, s *http.Response) bool {
	f.ensureMatcher()
	return f.Matches(s.StatusCode)
}

// SetMatcher sets the filter RangeMatcher. A nil RangeMatcher mean any valid StatusCode.
//
// If the returned error is not nil, the RangeMatcher is rejected.
func (f *StatusCodeFilter) SetMatcher(matcher Matcher) error {
	if matcher == nil {
		matcher = NewHTTPStatusMatcher()
	}
	rm, ok := matcher.(RangeMatcher)
	if !ok {
		f.ensureMatcher()
		return fmt.Errorf("the StatusCodeFilter only accepts RangeMatchers: got %T", matcher)
	}
	f.RangeMatcher = rm
	return nil
}
