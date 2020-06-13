package interception

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"

	"github.com/bearer/go-agent/events"
)

// Filtered is a well-known string replacing filtered-out content.
const Filtered = `[FILTERED]`

// DefaultSensitiveKeys is the expression used for sensitive keys if no other value is set.
var DefaultSensitiveKeys = regexp.MustCompile(`(?i)^(authorization|password|secret|passwd|api.?key|access.?token|auth.?token|credentials|mysql_pwd|stripetoken|card.?number.?|secret|client.?id|client.?secret)$`)

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
		p.SanitizeRequestBody,
		p.SanitizeResponseBody,
	}
}

// sanitizeURL and sanitizeHeaders apply the same logical loop, but the methods
// invoked have differing implementations.
// To avoid overwriting original values, sanitizeURL returns a new URL.
func (p SanitizationProvider) sanitizeURL(u *url.URL) (*url.URL, error) {
	sanU, err := url.ParseRequestURI(u.String())
	if err != nil {
		return nil, err
	}
	in := u.Query()
	out := make(url.Values, len(in))

Name:
	for name, values := range in {
		// Filter on keys, erasing all values.
		for _, sk := range p.SensitiveKeys {
			if sk.MatchString(name) {
				out.Set(name, Filtered)
				continue Name
			}
		}

		// If the key didn't match replace the matching values.
		for _, value := range values {
			for _, sr := range p.SensitiveRegexps {
				if sr.MatchString(value) {
					value = sr.ReplaceAllLiteralString(value, Filtered)
				}
			}
			out.Add(name, value)
		}
	}
	sanU.RawQuery = out.Encode()

	for _, r := range p.SensitiveRegexps {
		if r.MatchString(sanU.Path) {
			sanU.Path = r.ReplaceAllLiteralString(sanU.Path, Filtered)
		}
	}
	return sanU, nil
}

// sanitizeHeaders and sanitizeURL apply the same logical loop, but the methods
// invoked have differing implementations.
// To avoid overwriting original values, sanitizeHeaders returns a new URL.
func (p SanitizationProvider) sanitizeHeaders(in http.Header) http.Header {
	out := make(http.Header, len(in))

Name:
	for name, values := range in {
		// Filter on keys, erasing all values.
		for _, sk := range p.SensitiveKeys {
			if sk.MatchString(name) {
				out.Set(name, Filtered)
				continue Name
			}
		}

		// If the key didn't match replace the matching values.
		for _, value := range values {
			for _, sr := range p.SensitiveRegexps {
				if sr.MatchString(value) {
					value = sr.ReplaceAllLiteralString(value, Filtered)
				}
			}
			out.Add(name, value)
		}
	}

	return out
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
	req := e.Request()
	req.Header = p.sanitizeHeaders(req.Header)
	e.SetRequest(req)

	res := e.Response()
	resReq := res.Request
	if resReq == req {
		return nil
	}
	resReq.Header = p.sanitizeHeaders(resReq.Header)
	res.Request = resReq
	e.SetResponse(res)
	return nil
}

// SanitizeResponseHeaders sanitizes Response headers and trailers.
func (p SanitizationProvider) SanitizeResponseHeaders(_ context.Context, e events.Event) error {
	res := e.Response()
	res.Header = p.sanitizeHeaders(res.Header)
	e.SetResponse(res)
	return nil
}

// SanitizeRequestBody sanitized the Request body in a ReportEvent.
func (p SanitizationProvider) SanitizeRequestBody(_ context.Context, e events.Event) error {
	re, ok := e.(*ReportEvent)
	if !ok {
		return fmt.Errorf(`expected ReportEvent, got %T`, e)
	}
	w := NewWalker(re.RequestBody)
	var accu interface{}
	err := w.Walk(&accu, p.BodySanitizer)
	if err != nil {
		return err
	}
	re.RequestBody = w.Value()
	return nil
}

// SanitizeResponseBody sanitizes the Response body in a ReportEvent.
func (p SanitizationProvider) SanitizeResponseBody(_ context.Context, e events.Event) error {
	re, ok := e.(*ReportEvent)
	if !ok {
		return fmt.Errorf(`expected ReportEvent, got %T`, e)
	}
	w := NewWalker(re.ResponseBody)
	var accu interface{}
	err := w.Walk(&accu, p.BodySanitizer)
	if err != nil {
		return err
	}
	re.ResponseBody = w.Value()
	return nil
}

// BodySanitizer applies sanitization rules to data.
func (p SanitizationProvider) BodySanitizer(k interface{}, v *interface{}, accu *interface{}) error {
	if k == nil {
		return nil
	}
	if sk, ok := k.(string); ok {
		for _, re := range p.SensitiveKeys {
			if re.MatchString(sk) {
				*v = Filtered
				return nil
			}
		}
	}

	if reflect.ValueOf(*v).Kind() == reflect.String {
		sv, _ := (*v).(string) // Cannot fail because of previous line.
		for _, re := range p.SensitiveRegexps {
			if re.MatchString(sv) {
				sv = re.ReplaceAllLiteralString(sv, Filtered)
			}
		}
		*v = sv
	}
	return nil
}
