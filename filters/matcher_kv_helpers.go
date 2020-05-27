package filters

import (
	"fmt"
	"reflect"
	"regexp"
)

// kvStringer is a private type used to provide a fmt.Stringer implementation
// for stringify purposes.
type kvStringer string

func (s kvStringer) String() string {
	return string(s)
}

// stringerType is a reflect Interface type for fmt.Stringer.
var zeroKVStringer fmt.Stringer = kvStringer("")
var stringerType = reflect.TypeOf(&zeroKVStringer).Elem()

// errorType is a reflect Interface type for error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

func isElementMatchableKind(v reflect.Value) bool {
	return isMatchableKind(v.Type().Elem())
}

func isMatchableKind(typ reflect.Type) bool {
	kind := typ.Kind()
	if kind == reflect.Interface && (typ.Implements(errorType) || typ.Implements(stringerType)) {
		return true
	}
	// Non-matchable runtime comparable kinds like int can be matchable if they
	// belong to defined types implementing a matchable interface like error or
	// fmt.Stringer.
	if typ.Comparable() {
		return true
	}

	matchable := map[reflect.Kind]bool{
		reflect.String: true,
		reflect.Slice:  true,
		reflect.Array:  true,
		reflect.Map:    true,
	}
	if _, isMatchable := matchable[kind]; isMatchable {
		return true
	}
	return false
}

// Is the passed type any kind of map by string, regardless of the element type?
func isMapByString(typ reflect.Type) bool {
	// We do not match fmt.Stringer implementations to keep matching costs low,
	// so the only accepted types are map[string]something.
	if  typ.Kind() == reflect.Map && typ.Key().Kind() == reflect.String {
		return true
	}
	return false
}

// Like reflect.Value.IsNil, but return false instead of panicking on non-nil-able
// values.
func isNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return value.IsNil()
	}
	return false
}

// Does the passed map by string contain any key matching the regexp?
func mapHasMatchingStringKey(x interface{}, re *regexp.Regexp) bool {
	if x == nil || !isMapByString(reflect.TypeOf(x)) {
		return false
	}
	if re == nil {
		return true
	}

	// We now know that x is a map[string]something.
	value := reflect.ValueOf(x)
	found := false
	for _, k := range value.MapKeys() {
		sk := k.String()
		if re.MatchString(sk) {
			found = true
			break
		}
	}
	return found
}

// stringify converts the well-known interfaces error and fmt.Stringer to
// strings, and does not modify any other value.
func stringify(x interface{}) interface{} {
	if s, ok := x.(error); ok {
		return s.Error()
	}
	if s, ok := x.(fmt.Stringer); ok {
		return s.String()
	}
	return x
}
