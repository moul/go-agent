package filters

import (
	"fmt"
	"net/http"
)

// YesFilter provides a filter accepting any input, even nil.
type YesFilter struct {}

// Type is part of the Filter interface.
func (*YesFilter) Type() FilterType {
	return yesInternalFilter
}

// MatchesCall is part of the Filter interface.
func (nf *YesFilter) MatchesCall(_ *http.Request, _ *http.Response) bool {
	return true
}

// SetMatcher is part of the Filter interface. In YesFilter, is only accepts
// a nil matcher, as no underlying matcher is actually used.
func (*YesFilter) SetMatcher(matcher Matcher) error {
	if matcher != nil {
		return fmt.Errorf("instances of YesFilter only accept a nil Matcher, got %T", matcher)
	}
	return nil
}
