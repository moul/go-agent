package filters

//go:generate stringer -type=FilterSetOperator -output set_names.go

import (
	"fmt"
	"net/http"
	"strings"
)

// FilterSetOperator represents the operators available in Filter sets.
type FilterSetOperator byte

const (
	// Any matches if any of the child Filters matches.
	Any FilterSetOperator = iota

	// All matches if all of the child Filters match.
	All

	// NotFirst is special: it only matches if the first child does not match,
	// ignoring any other children.
	NotFirst
)

// FilterSet is the type of compound Filters made of other Filters, which can
// themselves be FilterSet values.
type FilterSet interface {
	Filter
	// AddChildren adds any number of children filters. Nil filters are stripped
	// from the list.
	AddChildren(...Filter) FilterSet

	// Children returns the list of children.
	Children() []Filter
}

type filterSet struct {
	operator FilterSetOperator
	children []Filter
}

func (f *filterSet) Type() FilterType {
	return FilterSetFilterType
}

func (f *filterSet) MatchesCall(request *http.Request, response *http.Response) bool {
	switch op := f.operator; op {
	case Any:
		return f.matchAny(request, response)

	case All:
		return f.matchAll(request, response)

	case NotFirst:
		return f.matchNotFirst(request, response)
	}

	return false
}

func (f *filterSet) matchNotFirst(request *http.Request, response *http.Response) bool {
	if len(f.children) == 0 {
		return false
	}
	return !f.children[0].MatchesCall(request, response)
}

func (f *filterSet) matchAny(request *http.Request, response *http.Response) bool {
	for _, f := range f.children {
		if f.MatchesCall(request, response) {
			return true
		}
	}
	return false
}

func (f *filterSet) matchAll(request *http.Request, response *http.Response) bool {
	for _, f := range f.children {
		if !f.MatchesCall(request, response) {
			return false
		}
	}
	return true
}

func (f *filterSet) SetMatcher(matcher Matcher) error {
	if !isNilInterface(matcher) {
		return fmt.Errorf("instances of NotFilter only accept a nil Matcher, got %T", matcher)
	}
	return nil
}

func (f *filterSet) AddChildren(filters ...Filter) FilterSet {
	for _, filter := range filters {
		if !isNilInterface(filter) {
			f.children = append(f.children, filter)
		}
	}
	return f
}

func (f *filterSet) Children() []Filter {
	return f.children
}

// NewFilterSet builds a FilterSet from an operator and children Filter instances.
func NewFilterSet(operator FilterSetOperator, children ...Filter) FilterSet {
	fs := (&filterSet{operator: operator}).AddChildren(children...)
	return fs
}

// FilterSetDescription provides a serialization-friendly description of a FilterSet.
type FilterSetDescription struct {
	// ChildHashes is set on filters.FilterSet filters
	ChildHashes []string

	// Operator is set on filters.FilterSet filters. It may only be `ANY` or `ALL`.
	Operator string
}

// String implements fmt.Stringer.
func (d FilterSetDescription) String() string {
	if len(d.ChildHashes) > 0 || d.Operator != `` {
		return fmt.Sprintf("%s(%s)\n", d.Operator, strings.Join(d.ChildHashes, `, `))
	}
	return ``
}

func setFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	var op FilterSetOperator
	switch {
	case strings.EqualFold(fd.Operator, Any.String()):
		op = Any
	case strings.EqualFold(fd.Operator, All.String()):
		op = All
	case strings.EqualFold(fd.Operator, NotFirst.String()):
		op = NotFirst
	default:
		op = Any
	}
	children := make([]Filter, 0, len(fd.ChildHashes))
	for _, h := range fd.ChildHashes {
		var f Filter
		f, ok := filterMap[h]
		if !ok {
			return nil
		}
		children = append(children, f)
	}

	f := NewFilterSet(op, children...)
	return f
}
