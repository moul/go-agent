package filters

import (
	"fmt"
	"reflect"
)

// Matcher is the interface shared by all matchers.
type Matcher interface {
	Matches(x interface{}) bool
}

// Like reflect.Value.IsNil, but return false instead of panicking on non-nil-able
// values.
func isNilValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return value.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

// Allows checking a value for nil-ness regardless of its dynamic type.
//
// Only use for interface types with unknown implementation: for other cases,
// use a type assertion and concrete type nil check.
func isNilInterface(x interface{}) bool {
	return isNilValue(reflect.ValueOf(x))
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
