package filters

import (
	"net/http"
)

// StatusValid is a range covering the valid HTTP status code range.
var StatusValid = NewRange().ExcludeTo().From(100).To(600)

// StatusCodeFilter provides a filter for the response status code in API requests.
type StatusCodeFilter struct {
	Range
}

// Type is part of the Filter interface.
func (*StatusCodeFilter) Type() FilterType {
	return statusCodeFilter
}

// Matches is part of the Filter interface.
func (f *StatusCodeFilter) Matches(_ *http.Request, s *http.Response) bool {
	return f.Range.Contains(s.StatusCode)
}

// SetRange sets the filter Range. A nil Range mean any valid StatusCode.
//
// If the returned error is not nil, the Range cannot be used. There is currently
// no condition causing an error.
func (f *StatusCodeFilter) SetRange(r Range) error {
	if r == nil {
		r = StatusValid
	}
	f.Range = r
	return nil
}
