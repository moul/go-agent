package filters

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParamFilter_MatchesCall(t *testing.T) {
	fooBar, _ := url.Parse("http://host.tld?foo=bar")
	fooBaz, _ := url.Parse("http://host.tld?foo=baz")
	quxBar, _ := url.Parse("http://host.tld?qux=bar")

	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{"happy", &http.Request{URL: fooBar}, true},
		{"sad bad key", &http.Request{URL: quxBar}, false},
		{"sad bad value", &http.Request{URL: fooBaz}, false},
		{"sad no query", &http.Request{URL: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ParamFilter{}
			// This is not a test for SetMatcher.
			_ = f.SetMatcher(NewKeyValueMatcher(foo, bar))

			if got := f.MatchesCall(tt.req, nil); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParamFilter_SetMatcher(t *testing.T) {
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
			f := &ParamFilter{}
			if err := f.SetMatcher(tt.KeyValueMatcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParamFilter_Type(t *testing.T) {
	expected := paramFilter
	var f ParamFilter
	actual := f.Type()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}
