package filters

import (
	"fmt"
	"net/http"
)

// NotFilter provides a filter inverting the underlying filter.
type NotFilter struct {
	Filter
}

// Type is part of the Filter interface.
func (*NotFilter) Type() FilterType {
	return notFilter
}

// ensureMatcher makes sure the instance filter is set, and returns true if it
// had to set a new one itself.
func (f *NotFilter) ensureFilter() {
	if f.Filter != nil {
		return
	}
	f.Filter = &YesFilter{}
}

// MatchesCall is part of the Filter interface.
func (f *NotFilter) MatchesCall(r *http.Request, s *http.Response) bool {
	f.ensureFilter()
	return !f.Filter.MatchesCall(r, s)
}

// SetFilter sets the filter to invert.
//
// Passing a nil filter treats it as a Yes filter, returning a No filter.
func (f *NotFilter) SetFilter(filter Filter) error {
	if filter != nil {
		filter = &YesFilter{}
	}
	f.Filter = filter
	return nil
}

// SetMatcher only accepts a nil Matcher, because it applies the matcher found
// in its underlying Filter.
func (f *NotFilter) SetMatcher(matcher Matcher) error {
	if matcher != nil {
		return fmt.Errorf("instances of NotFilter only accept a nil Matcher, got %T", matcher)
	}
	return nil
}
