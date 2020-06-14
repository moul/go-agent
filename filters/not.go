package filters

import (
	"fmt"

	"github.com/bearer/go-agent/events"
)

// NotFilter provides a filter inverting the underlying filter.
type NotFilter struct {
	*filterSet
}

// Type is part of the Filter interface.
func (*NotFilter) Type() FilterType {
	return NotFilterType
}

func (f *NotFilter) ensureFilter() {
	if isNilInterface(f.filterSet) {
		f.filterSet = &filterSet{}
	}
	f.filterSet.operator = NotFirst
}

// MatchesCall is part of the Filter interface.
func (f *NotFilter) MatchesCall(e events.Event) bool {
	f.ensureFilter()
	return f.filterSet.MatchesCall(e)
}

// SetFilter sets the filter to invert.
//
// Passing a nil filter treats it as a Yes filter, returning a No filter.
func (f *NotFilter) SetFilter(filter Filter) error {
	f.ensureFilter()
	if !isNilInterface(filter) {
		f.filterSet.AddChildren(filter)
	}
	return nil
}

// SetMatcher only accepts a nil Matcher, because it applies the matcher found
// in its underlying Filter.
func (f *NotFilter) SetMatcher(matcher Matcher) error {
	if !isNilInterface(matcher) {
		return fmt.Errorf("instances of NotFilter only accept a nil Matcher, got %T", matcher)
	}
	return nil
}

// AddChildren overrides the embedded filterSet method to have one child at most.
func (f *NotFilter) AddChildren(filters ...Filter) FilterSet {
	f.ensureFilter()
	// Don't add one if the Not filters already has one.
	if len(f.children) > 0 {
		return f
	}
	for _, filter := range filters {
		// Only insert the first non-nil filter.
		if !isNilInterface(filter) {
			f.children = append(f.children, filters[0])
			break
		}
	}
	return f
}

func notFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	child, ok := filterMap[fd.ChildHash]
	if !ok {
		return nil
	}
	f := filterSet{operator: NotFirst}
	f.AddChildren(child)
	return &f
}
