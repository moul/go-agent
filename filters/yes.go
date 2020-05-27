package filters

import (
	"net/http"
)

// yesFilter provides a filter accepting any input.
type yesFilter struct {}

// Type is part of the Filter interface.
func (*yesFilter) Type() FilterType {
	return yesInternalFilter
}

// Matches is part of the Filter interface.
func (nf *yesFilter) Matches(_ *http.Request, _ *http.Response) bool {
	return true
}
