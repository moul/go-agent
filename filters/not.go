package filters

import (
	"errors"
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

// Matches is part of the Filter interface.
func (f *NotFilter) Matches(r *http.Request, s *http.Response) bool {
	if f.Filter == nil {
		f.Filter = &yesFilter{}
	}
	return !f.Filter.Matches(r, s)
}

// SetFilter sets the filter to invert.
//
// Passing a nil filter to invert is an error.
func (f *NotFilter) SetFilter(base Filter) error {
	var err error
	if base == nil {
		base = &yesFilter{}
		err = errors.New("refusing to apply Not to a nil filter")
	}
	f.Filter = base
	return err
}
