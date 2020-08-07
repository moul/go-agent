package filters

import (
	"fmt"

	"github.com/bearer/go-agent/events"
)

// StatusCodeFilter provides a filter for the response status code in API requests.
type StatusCodeFilter struct {
	RangeMatcher
}

// Type is part of the Filter interface.
func (*StatusCodeFilter) Type() FilterType {
	return StatusCodeFilterType
}

func (f *StatusCodeFilter) ensureMatcher() {
	if f.RangeMatcher != nil {
		return
	}
	_ = f.SetMatcher(NewHTTPStatusMatcher())
}

// MatchesCall is part of the Filter interface.
func (f *StatusCodeFilter) MatchesCall(e events.Event) bool {
	if e.Response() == nil {
		return false
	}
	f.ensureMatcher()
	return f.Matches(e.Response().StatusCode)
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

func statusCodeFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	r := fd.Range
	m := NewRangeMatcher()
	if r.From != nil {
		m.From(r.ToInt(r.From))
	}
	if r.To != nil {
		m.To(r.ToInt(r.To))
	}
	if r.ExcludeFrom {
		m.ExcludeFrom()
	}
	if r.ExcludeTo {
		m.ExcludeTo()
	}
	f := &StatusCodeFilter{}
	err := f.SetMatcher(m)
	if err != nil {
		return nil
	}
	return f
}
