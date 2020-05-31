package filters

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"unsafe"
)

const AlreadyChecked = "already checked"

var foo = "foo"
var bar = "bar"
const badRe = `[`
var reFoo = regexp.MustCompile(foo)
var reBar = regexp.MustCompile(bar)
var noneSeen = func() pMap { return pMap{} }

func Test_keyValueMatcher_matchesString(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}

	tests := []struct {
		name   string
		fields fields
		s      string
		want   bool
	}{
		{"happy", fields{nil, reBar}, bar, true},
		{"key regex", fields{reFoo, nil}, bar, false},
		{"different string", fields{nil, reBar}, foo, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if got := m.matchesString(tt.s, false); got != tt.want {
				t.Errorf("matchesString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyValueMatcher_Matches(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}
	tests := []struct {
		name      string
		fields    *fields
		x         interface{}
		wantFound bool
	}{
		{"nil", &fields{nil, nil}, (*chan int)(nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if gotFound := m.Matches(tt.x); gotFound != tt.wantFound {
				t.Errorf("Matches() = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func Test_keyValueMatcher_ValueRegexp(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields *fields
		want   *regexp.Regexp
	}{
		{"nil", &fields{nil, nil}, nil},
		{"non-nil", &fields{nil, reBar}, reBar},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if got := m.ValueRegexp(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValueRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyValueMatcher_matchesMap(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}

	tests := []struct {
		name   string
		fields *fields
		value  interface{} // Must be some sort of map type.
		want   bool
	}{
		{"happy string key", &fields{reFoo, reBar}, map[string]string{foo: bar}, true},
		{"happy stringer key", &fields{reFoo, reBar}, map[fmt.Stringer]string{kvStringer(foo): bar}, true},
		{"happy error key", &fields{reFoo, reBar}, map[error]string{errors.New(foo): bar}, true},
		{"happy no key", &fields{nil, reBar}, map[int]fmt.Stringer{42: kvStringer(bar)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if got := m.Matches(tt.value); got != tt.want {
				t.Errorf("matchesMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyValueMatcher_matchesSlice(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields *fields
		value  interface{} // Needs to be some sort of slice.
		want   bool
	}{
		{"happy", &fields{nil, reBar}, []string{bar}, true},
		{"match prevented by key", &fields{reFoo, reBar}, []string{bar}, false},
		{"nil slice", &fields{nil, reBar}, []int(nil), false},
		{"empty slice", &fields{nil, reBar}, []string{}, false},
		{"other slice", &fields{nil, reBar}, []complex64{2i}, false},
		{AlreadyChecked, &fields{nil, reBar}, []string{bar}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			var got bool
			// Matches() initializes the seen pMap so to check a hit, we need to
			// to call the underlying implementation.
			if tt.name == AlreadyChecked {
				m.seen.track(tt.value)
				got = m.matchesSlice(tt.value, false)
			} else {
				got = m.Matches(tt.value)
			}

			if got != tt.want {
				t.Errorf("matchesSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyValueMatcher_matchesArray(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields *fields
		value  interface{} // Needs to be some sort of array.
		want   bool
	}{
		{"happy", &fields{nil, reBar}, [...]string{bar}, true},
		{"empty array", &fields{nil, reBar}, [...]float64{}, false},
		{"other array type", &fields{nil, reBar}, [...]complex64{2i}, false},
		{"other array values", &fields{nil, reBar}, [...]string{foo}, false},
		// No AlreadyChecked: arrays are normally copied, not referenced.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}

			if got := m.Matches(tt.value); got != tt.want {
				t.Errorf("Matches(array) = %v, want %v", got, tt.want)
			}

			m = &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if got := m.matchesArray(tt.value); got != tt.want {
				t.Errorf("Matches(array) = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_keyValueMatcher_track(t *testing.T) {
	p := pMap{}
	for i := 0; i < 3; i++ {
		want := i > 0
		if got := p.track(&i); got != want {
			t.Errorf("track() = %v, want %v", got, want)
		}
	}
}

func TestNewKeyValueMatcher(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantNil bool
	}{
		{"happy", foo, bar, false},
		{"sad key", badRe, bar, true},
		{"sad value", foo, badRe, true},
		{"sad both", badRe, badRe, true},
		{"empty key", ``, bar, false},
		{"empty value", foo, ``, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKeyValueMatcher(tt.key, tt.value); (got == nil) != tt.wantNil {
				t.Errorf("NewKeyValueMatcher() = %v, want %v", got, tt.wantNil)
			}
		})
	}
}

func Test_keyValueMatcher_KeyRegexp(t *testing.T) {
	type fields struct {
		keyRegexp   *regexp.Regexp
		valueRegexp *regexp.Regexp
	}
	tests := []struct {
		name   string
		fields *fields
		want   *regexp.Regexp
	}{
		{"nil", &fields{nil, nil}, nil},
		{"non-nil", &fields{reFoo, nil}, reFoo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &keyValueMatcher{
				seen:        noneSeen(),
				keyRegexp:   tt.fields.keyRegexp,
				valueRegexp: tt.fields.valueRegexp,
			}
			if got := m.KeyRegexp(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pMap_String(t *testing.T) {
	m := pMap{unsafe.Pointer(&bar): 42}
	tests := []struct {
		name string
		pm   pMap
		want string
	}{
		{"nil", nil, ``},
		{"empty", make(pMap), ``},
		{"single", m, fmt.Sprintf("%p: %d\n", &bar, 42)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
