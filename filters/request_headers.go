package filters

import (
	"errors"
	"fmt"
	"net/http"
)

// RequestHeadersFilter provides a filter for API Request headers.
type RequestHeadersFilter struct {
	KeyValueMatcher
}

// Type is part of the Filter interface.
func (f *RequestHeadersFilter) Type() FilterType {
	return RequestHeadersFilterType
}

func (f *RequestHeadersFilter) ensureMatcher() {
	if !isNilInterface(f.KeyValueMatcher) {
		return
	}
	_ = f.SetMatcher(NewKeyValueMatcher(``, ``))
}

// MatchesCall is part of the Filter interface.
func (f *RequestHeadersFilter) MatchesCall(request *http.Request, _ *http.Response) bool {
	f.ensureMatcher()
	return f.KeyValueMatcher.Matches(request.Header)
}

// SetMatcher sets the filter KeyValueMatcher.
//
// If the returned error is not nil, the filter will accept any value except nil.
//
// To apply a case-insensitive match, prepend (?i) to the matcher regexps,
// as in: (?i)\.bearer\.sh$
func (f *RequestHeadersFilter) SetMatcher(matcher Matcher) error {
	defaultMatcher := NewKeyValueMatcher(``, ``)

	m, ok := matcher.(KeyValueMatcher)
	if !ok {
		f.KeyValueMatcher = defaultMatcher
		return fmt.Errorf("key-value matcher expected, got a %T", matcher)
	}

	if isNilInterface(m) {
		f.KeyValueMatcher = defaultMatcher
		return errors.New("set nil Key-Value matcher on RequestHeaders filter")
	}

	f.KeyValueMatcher = m
	return nil
}

func requestFilterHeadersFromDescription(filterMap FilterMap, d interface{}) Filter {
	kvd, ok := d.(KeyValueDescription)
	if !ok {
		return nil
	}
	// FIXME apply RegexpMatcherDescription.Flags
	m := NewKeyValueMatcher(kvd.KeyPattern.Value, kvd.ValuePattern.Value)
	if m == nil {
		return nil
	}
	f := &RequestHeadersFilter{}
	err := f.SetMatcher(m)
	if err != nil {
		return nil
	}
	return f
}
