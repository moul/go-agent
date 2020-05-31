package filters

import (
	"fmt"
	"reflect"
)

// kvStringer is a private type used to provide a fmt.Stringer implementation
// to use when checking whether the type of a reflect.Value implements fmt.Stringer.
type kvStringer string

func (s kvStringer) String() string {
	return string(s)
}

func isElementMatchableKind(v reflect.Value) bool {
	return isMatchableKind(v.Type().Elem())
}

func isMatchableKind(typ reflect.Type) bool {
	// stringerType is a reflect Interface type for fmt.Stringer.
	var zeroKVStringer fmt.Stringer = kvStringer(``)
	var stringerType = reflect.TypeOf(&zeroKVStringer).Elem()

	// errorType is a reflect Interface type for error.
	var errorType = reflect.TypeOf((*error)(nil)).Elem()

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
