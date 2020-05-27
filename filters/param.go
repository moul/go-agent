package filters

import (
	"net/http"
	"regexp"
)

// ParamFilter provides a key-value filter for API request parameters.
type ParamFilter struct {
	KeyPattern   *regexp.Regexp
	ValuePattern *regexp.Regexp
}

// Type is part of the Filter interface.
func (*ParamFilter) Type() FilterType {
	return paramFilter
}

// MatchesCall is part of the Filter interface.
func (f *ParamFilter) MatchesCall(r *http.Request, _ *http.Response) bool {
	var keyMatch, valueMatch bool
	if f.KeyPattern == nil {
		keyMatch = true
	}
	if f.ValuePattern == nil {
		valueMatch = true
	}
	if keyMatch && valueMatch {
		return true
	}
	if mapHasMatchingStringKey(r.URL.Query(), f.KeyPattern) {
		keyMatch = true
	}
	return keyMatch && valueMatch
}

// SetKeyPattern sets the filter regexp Pattern from the compiled version of the
// passed string.
//
// If the returned error is not nil, the filter Pattern cannot be used.
//
// To apply a case-insensitive match, prepend (?i) to the regex, as in: (?i)\.bearer\.sh$
func (f *ParamFilter) SetKeyPattern(s string) error {
	re, err := regexp.Compile(s)
	f.KeyPattern = re
	return err
}
