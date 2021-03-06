package filters

import (
	"net/http"
	"strings"
	"testing"

	"github.com/bearer/go-agent/events"
)

func TestHTTPMethodFilter_MatchesCall(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		ignoreCase bool
		req        *http.Request
		want       bool
	}{
		{"happy", http.MethodTrace, false, &http.Request{Method: http.MethodTrace}, true},
		{"happy no case", http.MethodPut, true, &http.Request{Method: strings.ToLower(http.MethodPut)}, true},
		{"sad for name", http.MethodOptions, false, &http.Request{Method: http.MethodTrace}, false},
		{"sad for case", http.MethodHead, false, &http.Request{Method: strings.ToLower(http.MethodHead)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &HTTPMethodFilter{NewStringMatcher(tt.method, tt.ignoreCase)}
			e := (&events.EventBase{}).SetRequest(tt.req)
			if got := f.MatchesCall(e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMethodFilter_SetMatcher(t *testing.T) {
	var invalidSlice = []byte{'P', 'O', 0xC2, 'S', 'T'}
	badString := string(invalidSlice)
	invalidSlice[3] = ' ' // The invalid UTF-8 will be fixed, but leave the space.
	badMethod := string(invalidSlice)

	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{"happy", http.MethodDelete, false},
		{"happy empty method", ``, false},
		{"happy ill-formed UTF-8", badString, false}, // Gets fixed.
		{"sad method", "PO ST", true},
		{"sad ill-formed UTF-8", badMethod, true}, // Gets fixed, but still sad
		{"bad matcher", http.MethodGet, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &HTTPMethodFilter{}
			var matcher Matcher
			if tt.name == "bad matcher" {
				matcher = NewEmptyRegexpMatcher()
			} else {
				matcher = NewStringMatcher(tt.method, true)
			}
			if err := f.SetMatcher(matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPMethodFilter_Type(t *testing.T) {
	expected := HTTPMethodFilterType.String()
	var f HTTPMethodFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}

func TestHTTPMethodFilter_SetMatcher1(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{`happy`, &stringMatcher{}, false},
		{`happy nil`, nil, false},
		{`sad bad matcher`, &regexpMatcher{}, true},
		{`sad bad method`, &stringMatcher{s: "po,st"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &HTTPMethodFilter{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_methodFilterFromDescription(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantNil bool
	}{
		{`happy`, http.MethodGet, false},
		{`sad bad method`, `po,st`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd := FilterDescription{Value: tt.value}
			if got := methodFilterFromDescription(nil, &fd); got == nil != tt.wantNil {
				t.Errorf("methodFilterFromDescription() = %v, want %t", got, tt.wantNil)
			}
		})
	}
}
