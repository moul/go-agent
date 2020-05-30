package filters

import (
	"errors"
	"net/http"
)

// ParamFilter provides a key-value filter for API request parameters.
type ParamFilter struct {
	KeyValueMatcher
}

// Type is part of the Filter interface.
func (*ParamFilter) Type() FilterType {
	return paramFilter
}

// MatchesCall is part of the Filter interface.
func (f *ParamFilter) MatchesCall(r *http.Request, _ *http.Response) bool {
	m := NewKeyValueMatcher(f.KeyRegexp().String(), f.ValueRegexp().String())
	if r.URL == nil {
		return false
	}
	return m.Matches(r.URL.Query())
}

// SetMatcher sets the filter KeyValueMatcher.
//
// If the returned error is not nil, the filter cannot be used.
//
// To apply a case-insensitive match, prepend (?i) to the matcher regexps,
// as in: (?i)\.bearer\.sh$
func (f *ParamFilter) SetMatcher(m KeyValueMatcher) error {
	f.KeyValueMatcher = m
	if m == nil {
		return errors.New("set nil Key-Value matcher on Param filter")
	}
	return nil
}
