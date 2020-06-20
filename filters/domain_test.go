package filters

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/bearer/go-agent/events"
)

const (
	BearerDomain = `bearer.sh`
	BearerRE     = `^bearer\.sh$`
)

func TestDomainFilter_MatchesCall(t *testing.T) {
	tests := []struct {
		name    string
		matcher RegexpMatcher
		domain  string
		want    bool
	}{
		{"empty", NewEmptyRegexpMatcher(), ``, true},
		{"empty vs non-empty", NewEmptyRegexpMatcher(), BearerDomain, true},
		{"non-empty vs empty", NewRegexpMatcher(BearerRE), ``, false},
		{"happy", NewRegexpMatcher(BearerRE), BearerDomain, true},
		{"sad good regexp", NewRegexpMatcher(`^bearer.com$`), BearerDomain, false},
		// Bad regexps are replaced by a pass-all empty regexp.
		{"sad bad regexp", NewRegexpMatcher(badRe), BearerDomain, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DomainFilter{
				RegexpMatcher: tt.matcher,
			}
			url, _ := url.Parse(`https://` + tt.domain)

			e := &events.EventBase{}
			e.SetRequest(&http.Request{URL: url})
			if got := f.MatchesCall(e); got != tt.want {
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
		{"happy", NewEmptyRegexpMatcher(), false},
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
		{"set", NewEmptyRegexpMatcher()},
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
	expected := DomainFilterType.String()
	var f DomainFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}

func Test_domainFilterFromDescription(t *testing.T) {
	type args struct {
		in0 FilterMap
		fd  *FilterDescription
	}
	tests := []struct {
		name string
		args args
		want Filter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domainFilterFromDescription(tt.args.in0, tt.args.fd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("domainFilterFromDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}
