package filters

import (
	"fmt"

	"github.com/bearer/go-agent/events"
)

// YesFilter provides a filter accepting any input, even nil.
type YesFilter struct{}

// Type is part of the Filter interface.
func (*YesFilter) Type() FilterType {
	return yesInternalFilter
}

// MatchesCall is part of the Filter interface.
func (*YesFilter) MatchesCall(events.Event) bool {
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

// AddChildren is part of the FilterSet interface.
func (f *YesFilter) AddChildren(...Filter) FilterSet { return f }

// Children is part of the FilterSet interface.
func (*YesFilter) Children() []Filter { return nil }
