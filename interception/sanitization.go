package interception

import (
	"context"
	"net/url"
	"regexp"

	"github.com/bearer/go-agent/events"
)

// Filtered is a well-known string replacing filtered-out content.
const Filtered = `[FILTERED]`

// DefaultSensitiveKeys is the expression used for sensitive keys if no other value is set.
var DefaultSensitiveKeys = regexp.MustCompile(`(?i)^authorization$|^password$|^secret$|^passwd$|^api.?key$|^access.?token$|^auth.?token$|^credentials$|^mysql_pwd$|^stripetoken$|^card.?number.?$|^secret$|^client.?id$|,^client.?secret$`)

// DefaultSensitiveData is the expression used for sensitive data if no other value is set.
var DefaultSensitiveData = regexp.MustCompile("(?i)[a-z0-9]{1}[a-z0-9.!#$%&’*+=?^_\"{|}~-]+@[a-z0-9-]+(?:\\.[a-z0-9-]+)*|(?:\\d[ -]*?){13,16}")

// SanitizationProvider is an events.Listener provider returning listeners based
// on the sensitive keys and regexps.
type SanitizationProvider struct {
	SensitiveKeys    []*regexp.Regexp
	SensitiveRegexps []*regexp.Regexp
}

// Listeners implements the events.ListenerProvider interface.
func (p SanitizationProvider) Listeners(e events.Event) []events.Listener {
	if e.Topic() != TopicReport {
		return nil
	}

	return []events.Listener{
		p.SanitizeQueryAndPaths,
		p.SanitizeRequestHeaders,
		p.SanitizeResponseHeaders,
	}
}

// To avoid overwriting original values, sanitizeURL returns a new URL.
func (p SanitizationProvider) sanitizeURL(u *url.URL) (*url.URL, error) {
	sanU, err := url.ParseRequestURI(u.String())
	if err != nil {
		return nil, err
	}
	q := u.Query()
	sanQ := url.Values{}

Query:
	for k, vs := range q {
		// Filter on keys, erasing all values.
		for _, sk := range p.SensitiveKeys {
			if sk.MatchString(k) {
				sanQ[k] = []string{Filtered}
				continue Query
			}
		}

		// If the key didn't match replace the matching values.
		// TODO this only replaces the first matching regexp. Agent spec
		//      doesn't clarify what to do if multiple regexps would match.
	Values:
		for _, v := range vs {
			for _, sr := range p.SensitiveRegexps {
				if sr.MatchString(v) {
					sanQ.Add(k, sr.ReplaceAllLiteralString(v, Filtered))
					continue Values
				}
				sanQ.Add(k, v)
			}
		}
	}
	sanU.RawQuery = sanQ.Encode()

	// TODO this only replaces the first matching regexp. Agent spec doesn't
	//      clarify what to do if multiple regexps would match.
	for _, r := range p.SensitiveRegexps {
		if r.MatchString(sanU.Path) {
			sanU.Path = r.ReplaceAllLiteralString(sanU.Path, Filtered)
			break
		}
	}
	return sanU, nil
}

// SanitizeQueryAndPaths sanitizes the URL query parameters and paths in both the
// original request and the request present in the response, which may or may
// not be the same.
func (p SanitizationProvider) SanitizeQueryAndPaths(_ context.Context, e events.Event) error {
	request := e.Request()
	// To avoid overwriting original values, sanitizeRequestURL returns a new request.
	req := request.Clone(request.Context())
	u, err := p.sanitizeURL(req.URL)
	if err != nil {
		return err
	}
	// Not valid in a client request, so just clean it.
	req.RequestURI = ``
	req.URL = u
	e.SetRequest(req)

	response := e.Response()
	// e.Response values contain a copy of the Request, which needs to be
	// sanitized too. We just did it if the Request object was reused.
	if response.Request == request {
		response.Request = req
		return nil
	}
	req = response.Request.Clone(response.Request.Context())
	u, err = p.sanitizeURL(req.URL)
	if err != nil {
		return err
	}
	req.RequestURI = ``
	req.URL = u
	// The request is cloned, so it need to be set back on the Response.
	response.Request = req
	e.SetResponse(response)

	return nil
}

// SanitizeRequestHeaders sanitizes Request headers and trailers.
func (p SanitizationProvider) SanitizeRequestHeaders(_ context.Context, e events.Event) error {
	return nil
}

// SanitizeResponseHeaders sanitizes Response headers and trailers.
func (p SanitizationProvider) SanitizeResponseHeaders(_ context.Context, e events.Event) error {
	return nil
}
