package filters

import (
	"fmt"

	"github.com/bearer/go-agent/events"
)

// ConnectionErrorFilter matches if any non-nil error is present.
type ConnectionErrorFilter struct{}

// Type is part of the Filter interface.
func (*ConnectionErrorFilter) Type() FilterType {
	return ConnectionErrorFilterType
}

// MatchesCall is part of the Filter interface.
func (f *ConnectionErrorFilter) MatchesCall(e events.Event) bool {
	return e.Err() != nil
}

// SetMatcher is part of the Filter interface. In ConnectionErrorFilter, is only accepts
// a nil matcher, as no underlying matcher is actually used.
func (*ConnectionErrorFilter) SetMatcher(matcher Matcher) error {
	if matcher != nil {
		return fmt.Errorf("instances of ConnectionErrorFilter only accept a nil Matcher, got %T", matcher)
	}
	return nil
}

func connectionErrorFilterFromDescription(FilterMap, *FilterDescription) Filter {
	return &ConnectionErrorFilter{}
}

