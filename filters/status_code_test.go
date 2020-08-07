package filters

import (
	"net/http"
	"testing"

	"github.com/bearer/go-agent/events"
)

func TestStatusCodeFilter_MatchesCall(t *testing.T) {
	tests := []struct {
		name         string
		matcher      RangeMatcher
		withResponse bool
		statusCode   int
		want         bool
	}{
		{"default high", NewRangeMatcher(), true, maxInt, true},
		{"default low", NewRangeMatcher(), true, minInt, true},
		{"hhtp ok", NewHTTPStatusMatcher(), true, http.StatusOK, true},
		{"http too high", NewHTTPStatusMatcher(), true, 600, false},
		{"http too low", NewHTTPStatusMatcher(), true, minInt, false},
		{"no response", NewHTTPStatusMatcher(), false, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &StatusCodeFilter{
				RangeMatcher: tt.matcher,
			}
			e := &events.EventBase{}
			if tt.withResponse {
				e.SetResponse(&http.Response{StatusCode: tt.statusCode})
			}
			if got := f.MatchesCall(e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCodeFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{"happy", NewHTTPStatusMatcher(), false},
		{"nil", nil, false},
		{"sad matcher", &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &StatusCodeFilter{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatusCodeFilter_ensureMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher RangeMatcher
	}{
		{"not set", nil},
		{"set", NewHTTPStatusMatcher()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &StatusCodeFilter{RangeMatcher: tt.matcher}
			f.ensureMatcher()
			if f.RangeMatcher == nil {
				t.Fatal("ensureMatcher did not set a non-nil matcher")
			}
		})
	}
}

func TestStatusCodeFilter_Type(t *testing.T) {
	expected := StatusCodeFilterType.String()
	var f StatusCodeFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}
