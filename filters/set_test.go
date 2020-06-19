package filters

import (
	"net/http"
	"testing"

	"github.com/bearer/go-agent/events"
)

// Beware: this test can only run when the set_names.go file has been
// generated, so be sure to have run "go generate filters" first.
func TestFilterSetOperator_String(t *testing.T) {
	const badOp = 200
	tests := []struct {
		name string
		i    FilterSetOperator
		want string
	}{
		{ "happy", Any, `Any`},
		{ "sad", FilterSetOperator(badOp), `FilterSetOperator(200)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterSet_MatchesCall(t *testing.T) {
	noFilter := (&NotFilter{}).AddChildren(&YesFilter{})
	yesFilter := &YesFilter{}
	type fields struct {
		operator FilterSetOperator
		children []Filter
	}
	type args struct {
		request  *http.Request
		response *http.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{ "happy any yes", fields{Any, []Filter{&YesFilter{}}}, args{}, true},
		{ "happy all yes", fields{All, []Filter{&YesFilter{}}}, args{}, true},
		// NotFirst covered by NotFilter tests.
		{ "sad bad op", fields{FilterSetOperator(200), nil}, args{}, false},
		{ "sad any all false", fields{Any, []Filter{noFilter, noFilter}}, args{}, false},
		{ "sad all one false", fields{All, []Filter{yesFilter, noFilter}}, args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &filterSet{
				operator: tt.fields.operator,
				children: tt.fields.children,
			}
			e := (&events.EventBase{}).SetRequest(tt.args.request).SetResponse(tt.args.response)
			if got := f.MatchesCall(e); got != tt.want {
				t.Errorf("MatchesCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterSet_SetMatcher(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		wantErr bool
	}{
		{"nil", nil, false},
		{"non nil", &yesMatcher{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &filterSet{}
			if err := f.SetMatcher(tt.matcher); (err != nil) != tt.wantErr {
				t.Errorf("SetMatcher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetFilter_Type(t *testing.T) {
	expected := FilterSetFilterType.String()
	var f *filterSet
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}
