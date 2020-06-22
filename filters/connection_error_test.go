package filters

import (
	"io"
	"reflect"
	"testing"

	"github.com/bearer/go-agent/events"
)

func TestConnectionErrorFilter_Type(t *testing.T) {
	expected := ConnectionErrorFilterType.String()
	var f ConnectionErrorFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}

func Test_connectionErrorFilterFromDescription(t *testing.T) {
	// This type does not actually depend on the filter map and description.
	actual := connectionErrorFilterFromDescription(nil, nil)
	expected := &ConnectionErrorFilter{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("connectionErrorFilterFromDescription() = %v, want %v", actual, expected)
	}
}

func TestConnectionErrorFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{`happy`, nil, false},
		{`sad`, &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := &ConnectionErrorFilter{}
			if err := co.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionErrorFilter_MatchesCall(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{ `happy`, io.EOF, true},
		{ `sad`, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ConnectionErrorFilter{}
			e := events.EventBase{Error: tt.err}
			if got := f.MatchesCall(&e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}
