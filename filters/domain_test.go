package filters

import (
	"net/http"
	"net/url"
	"testing"
)

const (
	BearerDomain = `bearer.sh`
	BearerRE = `^bearer\.sh$`
)

func TestDomainFilter_MatchesCall(t *testing.T) {
	tests := []struct {
		name    string
		matcher RegexpMatcher
		domain  string
		want    bool
	}{
		{"empty", NewEmptyRegexMatcher(), "", true},
		{"empty vs non-empty", NewEmptyRegexMatcher(), BearerDomain, true},
		{"non-empty vs empty", NewRegexMatcher(BearerRE), "", false},
		{"happy", NewRegexMatcher(BearerRE), BearerDomain, true},
		{"sad good regexp", NewRegexMatcher(`^bearer.com$`), BearerDomain, false},
		// Bad regexps are replaced by a pass-all empty regexp.
		{"sad bad regexp", NewRegexMatcher(`[`), BearerDomain, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DomainFilter{
				RegexpMatcher: tt.matcher,
			}
			url, _ := url.Parse(`https://` + tt.domain)
			if got := f.MatchesCall(&http.Request{URL: url}, nil); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomainFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{"happy", NewEmptyRegexMatcher(), false},
		{"nil", nil, false},
		{"sad matcher", &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DomainFilter{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDomainFilter_ensureMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher RegexpMatcher
	}{
		{"not set", nil},
		{"set", NewEmptyRegexMatcher()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DomainFilter{RegexpMatcher: tt.matcher}
			f.ensureMatcher()
			if f.RegexpMatcher == nil {
				t.Fatal("ensureMatcher did not set a non-nil matcher")
			}
		})
	}
}

func TestDomainFilter_Type(t *testing.T) {
	var f DomainFilter
	actual := f.Type()
	if actual != domainFilter {
		t.Errorf("Type() = %v, want %v", actual, domainFilter)
	}
}
