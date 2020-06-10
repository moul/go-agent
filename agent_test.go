package agent

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// Perform decoration tests without going to the network.
func TestNewAgent(t *testing.T) {
	const expected = `test handler`

	tests := []struct {
		name               string
		scheme, host, port string
		wantErr            bool
	}{
		{`happy`, `https`, `example.com`, `443`, false},
		{`bad scheme`, `bea@rer`, `example.com`, `1023`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAgent(ExampleWellFormedInvalidKey)

			// Set up test server.
			ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte(expected))
			}))
			defer ts.Close()

			c := ts.Client()
			c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			a.DecorateClientTransports(c)

			// Don't use NewRequest or url.Parse, as they validate URLs,
			// catching errors before we do, and we still want to intercept
			// users manufacturing invalid Requests on their own.
			var u *url.URL
			if regexp.MustCompile(`^happy`).MatchString(tt.name) {
				u, _ = url.Parse(ts.URL)
			} else {
				u = &url.URL{
					Scheme: tt.scheme,
					Host:   strings.Trim(strings.Join([]string{tt.host, tt.port}, `:`), `:`),
				}
			}
			r := &http.Request{URL: u}
			res, err := c.Do(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			defer res.Body.Close()
			if actual, err := ioutil.ReadAll(res.Body); err != nil || string(actual) != expected {
				t.Errorf("Got incorrect response: %s instead of %s", actual, expected)
			}
		})
	}
}
