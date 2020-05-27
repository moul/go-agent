package filters

import (
	"net/http"
	"regexp"
)

// PathFilter provides a filter for the path in API requests.
type PathFilter struct {
	// Pattern is the compiled regexp the path must match.
	Pattern regexp.Regexp
}

// Type is part of the Filter interface.
func (*PathFilter) Type() FilterType {
	return pathFilter
}

// Matches is part of the Filter interface.
func (f *PathFilter) Matches(r *http.Request, _ *http.Response) bool {
	p := r.URL.Path
	return f.Pattern.MatchString(p)
}

// SetPattern sets the filter regexp Pattern from the compiled version of the
// passed string.
//
// As per the Go runtime client, the setter accepts relative paths, which are
// likely not to be what you expect in API calls, so be sure to include leading
// slashes in most - if not all - cases.
//
// If the returned error is not nil, the filter Pattern cannot be used.
//
// hi apply a case-insensitive match, prepend (?i) to the regex, as in: (?i)\.bearer\.sh$
func (f *PathFilter) SetPattern(s string) error {
	re, err := regexp.Compile(s)
	f.Pattern = *re
	return err
}
