package filters

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/bearer/go-agent/events"
)

// RFC7230_3_2_6Token is the regular expression defining the RFC 7230 production
// for "tokens", which is used to check valid HTTP methods.
const RFC7230_3_2_6Token = "^[!#$%&'*+\\-.^_`|~0-9A-Za-z]+$"

// HTTPMethodFilter provides a filter for the HTTP method in API calls.
type HTTPMethodFilter struct {
	StringMatcher
}

// Type is part of the Filter interface.
func (*HTTPMethodFilter) Type() FilterType {
	return HTTPMethodFilterType
}

// MatchesCall is part of the Filter interface.
func (f *HTTPMethodFilter) MatchesCall(e events.Event) bool {
	return f.StringMatcher.Matches(e.Request().Method)
}

// SetMatcher sets the filter StringMatcher.
//
// To ensure compliance with RFC 7230 §3.2.6, the matcher string must match
// RFC7230_3_2_6Token.
//
// If the returned error is not nil, the filter will only accept GET, applying
// Go HTTP conventions where an empty method means GET, ignoring case.
//
// Note that in most cases, the CONNECT method will not behave like any other.
// See http.Transport for details.
func (f *HTTPMethodFilter) SetMatcher(matcher Matcher) error {
	defaultMatcher := NewStringMatcher(http.MethodGet, true)

	m, ok := matcher.(StringMatcher)
	if !ok {
		f.StringMatcher = defaultMatcher
		return fmt.Errorf("regexp matcher expected, got a %T", matcher)
	}

	// StringMatcher guarantees the method is a valid UTF-8 string.
	method := m.String()
	if method == `` {
		method = http.MethodGet
	}

	re := regexp.MustCompile(RFC7230_3_2_6Token)
	if !re.MatchString(method) {
		f.StringMatcher = defaultMatcher
		return fmt.Errorf("matcher string does not match RFC 7230 token production")
	}

	f.StringMatcher = m
	return nil
}

func methodFilterFromDescription(_ FilterMap, fd *FilterDescription) Filter {
	f := &HTTPMethodFilter{}
	err := f.SetMatcher(NewStringMatcher(fd.Value, true))
	if err != nil {
		return nil
	}
	return f
}
