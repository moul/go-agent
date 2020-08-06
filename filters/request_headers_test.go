package filters

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/events"
)

func TestRequestHeadersFilter_MatchesCall(t *testing.T) {
	noMatcher := regexp.MustCompile(`no matcher`)
	tests := []struct {
		name                   string
		keyRegexp, valueRegexp *regexp.Regexp
		header                 http.Header
		want                   bool
	}{
		{"happy single", reFoo, reBar, http.Header{foo: []string{bar}}, true},
		{"happy multi", reFoo, reBar, http.Header{foo: []string{foo, bar}}, true},
		{"happy no key", nil, reBar, http.Header{foo: []string{bar}}, true},
		{"happy no filter", nil, nil, make(http.Header), true},
		{"happy no matcher", nil, nil, make(http.Header), true},
		{"sad no matcher but nil", nil, nil, nil, false},
		{"sad no matching value", reFoo, reBar, http.Header{foo: []string{foo}}, false},
		{"sad no matching key", reFoo, reBar, http.Header{bar: []string{bar}}, false},
		{"sad no params", reFoo, reBar, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &RequestHeadersFilter{}
			if !noMatcher.MatchString(tt.name) {
				// This is not a test for SetMatcher.
				_ = f.SetMatcher(NewKeyValueMatcher(tt.keyRegexp, tt.valueRegexp))

			}
			e := (&events.EventBase{}).SetRequest(&http.Request{Header: tt.header})
			if got := f.MatchesCall(e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestHeadersFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name string
		KeyValueMatcher
		wantErr bool
	}{
		{"happy", NewKeyValueMatcher(nil, nil), false},
		{"sad nil", (*keyValueMatcher)(nil), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &RequestHeadersFilter{}
			if err := f.SetMatcher(tt.KeyValueMatcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestHeadersFilter_Type(t *testing.T) {
	expected := RequestHeadersFilterType.String()
	var f RequestHeadersFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}
