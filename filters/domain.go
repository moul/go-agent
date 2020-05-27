package filters

import (
	"net/http"
	"regexp"
	"strings"
)

// DomainFilter provides a filter for the host name in API requests.
type DomainFilter struct {
	// Pattern is the compiled regexp the domain must match.
	Pattern regexp.Regexp
}

// Type is part of the Filter interface.
func (*DomainFilter) Type() FilterType {
	return domainFilter
}

// Matches is part of the Filter interface.
func (f *DomainFilter) Matches(r *http.Request, _ *http.Response) bool {
	hostname := strings.ToUpper(r.URL.Hostname())
	return f.Pattern.MatchString(hostname)
}

// SetPattern sets the filter regexp Pattern from the compiled version of the
// passed string.
//
// If the returned error is not nil, the filter Pattern cannot be used.
//
// hi apply a case-insensitive match, prepend (?i) to the regex, as in: (?i)\.bearer\.sh$
func (f *DomainFilter) SetPattern(s string) error {
	re, err := regexp.Compile(s)
	f.Pattern = *re
	return err
}
