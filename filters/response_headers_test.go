package filters

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/events"
)

func TestResponseHeadersFilter_MatchesCall(t *testing.T) {
	noMatcher := regexp.MustCompile(`no matcher`)
	tests := []struct {
		name                   string
		keyRegexp, valueRegexp *regexp.Regexp
		header                 http.Header
		withResponse           bool
		want                   bool
	}{
		{"happy single", reFoo, reBar, http.Header{foo: []string{bar}}, true, true},
		{"happy multi", reFoo, reBar, http.Header{foo: []string{foo, bar}}, true, true},
		{"happy no key", nil, reBar, http.Header{foo: []string{bar}}, true, true},
		{"happy no filter", nil, nil, make(http.Header), true, true},
		{"happy no matcher", nil, nil, make(http.Header), true, true},
		{"sad no matcher but nil", nil, nil, nil, true, false},
		{"sad no matching value", reFoo, reBar, http.Header{foo: []string{foo}}, true, false},
		{"sad no matching key", reFoo, reBar, http.Header{bar: []string{bar}}, true, false},
		{"sad no params", reFoo, reBar, nil, true, false},
		{"sad no response", reFoo, reBar, nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ResponseHeadersFilter{}
			if !noMatcher.MatchString(tt.name) {
				// This is not a test for SetMatcher.
				_ = f.SetMatcher(NewKeyValueMatcher(tt.keyRegexp, tt.valueRegexp))

			}
			e := &events.EventBase{}
			if tt.withResponse {
				e.SetResponse(&http.Response{Header: tt.header})
			}
			if got := f.MatchesCall(e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseHeadersFilter_SetMatcher(t *testing.T) {
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
			f := &ResponseHeadersFilter{}
			if err := f.SetMatcher(tt.KeyValueMatcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseHeadersFilter_Type(t *testing.T) {
	expected := ResponseHeadersFilterType.String()
	var f ResponseHeadersFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}
