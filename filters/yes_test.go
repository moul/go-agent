package filters

import (
	"net/http"
	"testing"
)

func TestYesFilter_MatchesCall(t *testing.T) {
	type args struct {
		in0 *http.Request
		in1 *http.Response
	}
	tests := []struct {
		name string
		args args
	}{
		{"both nil", args{nil, nil}},
		{"only request", args{&http.Request{}, nil}},
		{"only response", args{nil, &http.Response{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nf := &YesFilter{}
			if gotTrue := nf.MatchesCall(tt.args.in0, tt.args.in1); !gotTrue {
				t.Errorf("MatchesCall() = %v, want true", gotTrue)
			}
		})
	}
}

type yesMatcher struct{}

func (m *yesMatcher) Matches(x interface{}) bool {
	return true
}

func TestYesFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{ "nil", nil, false },
		{ "non nil", &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ye := &YesFilter{}
			if err := ye.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYesFilter_Type(t *testing.T) {
	actual := (&YesFilter{}).Type()
	if actual != yesInternalFilter {
		t.Errorf("Type() = %v, want %v", actual, yesInternalFilter)
	}
}
