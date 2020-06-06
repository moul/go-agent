package filters

import (
	"net/http"
	"regexp"
	"testing"
)

func TestResponseHeadersFilter_MatchesCall(t *testing.T) {
	noMatcher := regexp.MustCompile(`no matcher`)
	tests := []struct {
		name       string
		key, value string
		header     http.Header
		want       bool
	}{
		{"happy single", foo, bar, http.Header{foo: []string{bar}}, true},
		{"happy multi", foo, bar, http.Header{foo: []string{foo, bar}}, true},
		{"happy no key", ``, bar, http.Header{foo: []string{bar}}, true},
		{"happy no filter", ``, ``, make(http.Header), true},
		{"happy no matcher", ``, ``, make(http.Header), true},
		{"sad no matcher but nil", ``, ``, nil, false},
		{"sad no matching value", foo, bar, http.Header{foo: []string{foo}}, false},
		{"sad no matching key", foo, bar, http.Header{bar: []string{bar}}, false},
		{"sad no params", foo, bar, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ResponseHeadersFilter{}
			if !noMatcher.MatchString(tt.name) {
				// This is not a test for SetMatcher.
				_ = f.SetMatcher(NewKeyValueMatcher(tt.key, tt.value))

			}
			response := &http.Response{Header: tt.header}
			if got := f.MatchesCall(nil, response); got != tt.want {
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
		{"happy", NewKeyValueMatcher(``, ``), false},
		{"sad", NewKeyValueMatcher(badRe, ``), true},
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
