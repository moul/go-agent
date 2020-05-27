package filters

import (
	"net/http"
	"testing"
)

func TestNotFilter_MatchesCall(t *testing.T) {
	type args struct {
		r *http.Request
		s *http.Response
	}
	tests := []struct {
		name   string
		filter Filter
		args   args
		want   bool
	}{
		{"inverted yes", &YesFilter{}, args{nil, nil}, false},
		{"inverted nil", nil, args{nil, nil}, false},
		{"inverted no", &NotFilter{&YesFilter{}}, args{nil, nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &NotFilter{
				Filter: tt.filter,
			}
			if got := f.MatchesCall(tt.args.r, tt.args.s); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotFilter_SetFilter(t *testing.T) {
	tests := []struct {
		name    string
		f1, f2  Filter
		wantErr bool
	}{
		{"both nil", nil, nil, false},
		{"nil and non-nil", nil, &YesFilter{}, false},
		{"non-nil and nil", &YesFilter{}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &NotFilter{
				Filter: tt.f1,
			}
			if err := f.SetFilter(tt.f2); (err != nil) != tt.wantErr {
				t.Errorf("SetFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotFilter_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{"nil", nil, false},
		{"non-nil", &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &NotFilter{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotFilter_ensureFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    Filter
	}{
		{"nil", nil},
		{"non-nil", &NotFilter{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &NotFilter{
				Filter: tt.filter,
			}
			f.ensureFilter();
			gotNil := f.Filter == nil
			if gotNil {
				t.Errorf("ensureMatcher() = %v, want nil", f.Filter)
			}
		})
	}
}

func TestNotFilter_Type(t *testing.T) {
	var f NotFilter
	actual := f.Type()
	if actual != notFilter {
		t.Errorf("Type() = %v, want %v", actual, notFilter)
	}
}
