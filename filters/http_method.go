package filters

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// HTTPMethodFilter provides a filter for the HTTP method in API calls.
type HTTPMethodFilter struct {
	// Method is the uppercase version of the method name accepted by the filter.
	// It can be
	// Be aware that the CONNECT method is not supported by Go HTTP client.
	Method string
}

// Type is part of the Filter interface.
func (*HTTPMethodFilter) Type() FilterType {
	return httpMethodFilter
}

// Matches is part of the Filter interface.
func (f *HTTPMethodFilter) Matches(r *http.Request, _ *http.Response) bool {
	// No need to sanitize method, as we want a byte for byte comparaison.
	return r.Method == f.Method
}

// SetMethod sets the filter Method from the uppercase version of the passed
// string, if it is valid. As per Go conventions, an empty string means GET.
//
// If the returned error is not nil, the filter Method is set to GET.
//
// Note that in most cases, the CONNECT method will not behave like any other.
// See http.Transport for details.
//
// As per RFC7231 Sec. 4, Method matches the RFC 7230 Sec. 3.2.6 "token" production.
func (f *HTTPMethodFilter) SetMethod(s string) error {
	const RFC7230_3_2_6Token = "[!#$%&'*+\\-.^_`|~0-9A-Za-z]+"
	// Shortcut for empty string, as applied by the Go HTTP Client.
	if s == "" {
		f.Method = "GET"
		return nil
	}

	// Sanitize passed method.
	s = strings.ToValidUTF8(s, "")
	re := regexp.MustCompile(RFC7230_3_2_6Token)
	if !re.MatchString(s) {
		f.Method = "GET"
		return fmt.Errorf("bad HTTP Method: %s", s)
	}

	f.Method = strings.ToUpper(s)
	return nil
}

