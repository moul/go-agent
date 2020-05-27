package filters

import (
	"regexp"
	"testing"
)

func Test_mapHasMatchingKey(t *testing.T) {
	type args struct {
		x  interface{}
		re *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"happy", args{map[string]func(){ foo: func(){}}, reFoo}, true},
		{"nil map", args{nil, reBar}, false},
		{"nil map and key", args{nil, nil}, false},
		{"nil key regex", args{map[string]int{foo: 42}, nil}, true},
		{"empty map", args{map[string](chan int){}, reBar}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapHasMatchingStringKey(tt.args.x, tt.args.re); got != tt.want {
				t.Errorf("mapHasMatchingStringKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
