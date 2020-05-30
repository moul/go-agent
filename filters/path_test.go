package filters

import (
	"net/http"
	"net/url"
	"testing"
)

const (
	path = `/foo`
	pathRE = `^/foo$`
)

func TestPathFilter_MatchesCall(t *testing.T) {
	const badRe = `[`
	tests := []struct {
		name    string
		matcher RegexpMatcher
		path    string
		want    bool
	}{
		{"empty", NewEmptyRegexMatcher(), ``, true},
		{"empty vs non-empty", NewEmptyRegexMatcher(), path, true},
		{"non-empty vs empty", NewRegexpMatcher(pathRE), ``, false},
		{"happy", NewRegexpMatcher(pathRE), path, true},
		{"sad good regexp", NewRegexpMatcher(`^/bar$`), path, false},
		// Bad regexps are replaced by a pass-all empty regexp.
		{"sad bad regexp", NewRegexpMatcher(badRe), path, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PathFilter{
				RegexpMatcher: tt.matcher,
			}
			url, _ := url.Parse(tt.path)
			if got := f.MatchesCall(&http.Request{URL: url}, nil); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathFilter_SetMatcher(t *testing.T) {
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
			f := &PathFilter{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathFilter_ensureMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher RegexpMatcher
	}{
		{"not set", nil},
		{"set", NewEmptyRegexMatcher()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PathFilter{RegexpMatcher: tt.matcher}
			f.ensureMatcher()
			if f.RegexpMatcher == nil {
				t.Fatal("ensureMatcher did not set a non-nil matcher")
			}
		})
	}
}

func TestPathFilter_Type(t *testing.T) {
	var f PathFilter
	actual := f.Type()
	if actual != pathFilter {
		t.Errorf("Type() = %v, want %v", actual, pathFilter)
	}
}
