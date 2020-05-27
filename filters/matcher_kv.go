package filters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unsafe"
)

// KeyValueMatcher is a matcher combining key and value matching for nested data types.
//
// Because it is reflection-based and can crawl complex structures, it may be
// very CPU-expensive.
type KeyValueMatcher interface {
	Matcher
	KeyRegexp() *regexp.Regexp
	ValueRegexp() *regexp.Regexp
}

type pMap map[unsafe.Pointer]byte

func (pm pMap) String() string {
	b := strings.Builder{}
	for k, v := range pm {
		b.WriteString(fmt.Sprintf("%p: %d\n", k, v))
	}
	return b.String()
}

// Track a value, returning true if it was already tracked.
func (pm pMap) track(x interface{}) bool {
	v := reflect.ValueOf(x)
	var usp unsafe.Pointer

	switch v.Kind() {
	// Within a given program, multiple different strings share a common Data
	// pointer, but each has a different StringHeader pointer, so we need to
	// track the Data pointers, not the string pointers themselves.
	case reflect.String:
		shp := (*reflect.StringHeader)(unsafe.Pointer(&x))
		usp = unsafe.Pointer(shp.Data)
	case reflect.Slice:
		// value.Pointer() on a slice returns a pointer to the first element, which
		// allows us to deduplicate on different slices sharing the same underlying
		// content.
		usp = unsafe.Pointer(v.Pointer())
	case reflect.Map:
		usp = unsafe.Pointer(v.Pointer())
	case reflect.Array:
		usp = unsafe.Pointer(&x)
	default:
		// Other data types are not expected to be tracked, because they cannot
		// match anyway.
		usp = unsafe.Pointer(v.Pointer())
	}

	return pm.trackRawPointer(usp)
}
func (pm pMap) trackRawPointer(usp unsafe.Pointer) bool {
	pm[usp]++
	return pm[usp] > 1
}

type keyValueMatcher struct {
	// mutex protects the seen map.
	mutex                  sync.RWMutex
	seen                   pMap
	keyRegexp, valueRegexp *regexp.Regexp
}

func (m *keyValueMatcher) KeyRegexp() *regexp.Regexp {
	return m.keyRegexp
}

func (m *keyValueMatcher) ValueRegexp() *regexp.Regexp {
	return m.valueRegexp
}

func (m *keyValueMatcher) Matches(x interface{}) (found bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.seen = pMap{}

	return m.doMatch(x, false)
}

func (m *keyValueMatcher) doMatch(x interface{}, ignoreKeyRegexp bool) bool {
	// Obtain a reflect.Value if we don't already have one.
	v, isValue := x.(reflect.Value)
	if !isValue {
		x = stringify(x)
		v = reflect.ValueOf(x)
	}
	if isNil(v) {
		return false
	}

	// Apply kind-specific matching for types supporting matching.
	switch kind := v.Kind(); kind {
	case reflect.String:
		return m.matchesString(x.(string), ignoreKeyRegexp)
	case reflect.Map:
		return m.matchesMap(x)
	case reflect.Slice:
		return m.matchesSlice(x, false)
	case reflect.Array:
		return m.matchesSlice(x, false)
	}

	// Other types cannot match.
	return false
}

func (m *keyValueMatcher) matchesString(s string, ignoreKeyRegexp bool) bool {
	if m.seen.track(s) {
		return false
	}

	if m.keyRegexp != nil && !ignoreKeyRegexp {
		return false
	}
	return m.valueRegexp == nil || m.valueRegexp.MatchString(s)
}

// matchesSlice matches against each element in a slice. The parameter must be
// a non-nil slice (nil-ness checked in doMatch).
func (m *keyValueMatcher) matchesSlice(x interface{}, ignoreKeyRegexp bool) bool {
	value := reflect.ValueOf(x)
	if !isElementMatchableKind(value) {
		return false
	}
	if m.seen.track(x) {
		return false
	}
	for i := 0; i < value.Len(); i++ {
		v := value.Index(i).Interface()
		if m.doMatch(v, ignoreKeyRegexp) {
			return true
		}
	}
	return false
}

// matchesArray matches against each element in an array. The parameter must be
// an array.
func (m *keyValueMatcher) matchesArray(x interface{}) bool {
	value := reflect.ValueOf(x)
	if value.Len() == 0 || !isElementMatchableKind(value) {
		return false
	}

	/* Note: arrays are not tracked as they are normally copied, not referenced. */

	for i := 0; i < value.Len(); i++ {
		if m.doMatch(value.Index(i).Interface(), false) {
			return true
		}
	}
	return false
}

// matchesMap matches against each key and value in a map. The parameter must be
// some kind of map.
func (m *keyValueMatcher) matchesMap(x interface{}) bool {
	value := reflect.ValueOf(x)

	if isNil(value) || value.Len() == 0 || !isElementMatchableKind(value) {
		return false
	}

	// x.Pointer() on a map returns the map address, meaning we may still
	// perform duplicate checks on children, which will then be tracked on their
	// own. It still protects against reused map references.
	if m.seen.track(x) {
		return false
	}

	mapIter := value.MapRange()
	for mapIter.Next() {
		if m.keyRegexp != nil {
			// If key doesn't match, no need to check x.
			key := mapIter.Key().Interface()
			// For stringable keys, use a plain regexp match: cycle detection does
			// not apply.
			switch key.(type) {
			case string, fmt.Stringer, error:
				// For these three types, s will always be a string.
				s := stringify(key)
				if !m.keyRegexp.MatchString(s.(string)) {
					continue
				}
			default:
				if !m.doMatch(key, false) {
					continue
				}
			}
		}

		// XXX Should it reject nil items instead, although there is no regex ?
		if m.valueRegexp == nil {
			return true
		}

		i := mapIter.Value().Interface()
		// If the value is a stringable, match it, else match the value normally.
		switch i.(type) {
		case string, fmt.Stringer, error:
			s := stringify(i)
			// For these three types, s will always be a string.
			if m.matchesString(s.(string), true) {
				return true
			}
		default:
			if m.doMatch(i, false) {
				return true
			}
		}
	}
	return false
}

// NewKeyValueMatcher creates a valid default KeyValueMatcher accepting any value.
func NewKeyValueMatcher() KeyValueMatcher {
	return &keyValueMatcher{
		seen: pMap{},
	}
}
