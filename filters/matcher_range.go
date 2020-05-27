package filters

import "strconv"

// NewHTTPStatusMatcher returns a range matcher for the valid HTTP status code range.
func NewHTTPStatusMatcher() RangeMatcher {
	return NewRangeMatcher().ExcludeTo().From(100).To(600)
}

// RangeMatcher provides the ability to check whether an int value is within an integer range.
//
// By default, its lo and hi limits are the maximum representable values, and
// they are included in comparisons.
//
// The interface allows a fluent initialization, like: NewRangeMatcher().hi(200).ExcludeTo(),
// which will define the semi-open interval [minInt, 200[.
type RangeMatcher interface {
	Matcher
	Contains(int) bool
	From(int) RangeMatcher
	To(int) RangeMatcher
	ExcludeFrom() RangeMatcher
	ExcludeTo() RangeMatcher
}

type intRange struct {
	lo, hi                     int
	FromExclusive, ToExclusive bool
}

// String implements fmt.Stringer.
func (r intRange) String() string {
	var (
		leftBrace  = "["
		rightBrace = "]"
	)

	if r.FromExclusive {
		leftBrace = "]"
	}
	if r.ToExclusive {
		rightBrace = "["
	}
	var sLo, sHi string
	if r.lo > minInt && r.lo < maxInt {
		sLo = strconv.Itoa(r.lo)
	}
	if r.hi > minInt && r.hi < maxInt {
		sHi = strconv.Itoa(r.hi)
	}
	return leftBrace + sLo + ":" + sHi + rightBrace
}

const (
	maxInt = int(^uint(0) >> 1)
	minInt = -maxInt - 1
)

// Contains implements the RangeMatcher interface.
func (r intRange) Contains(n int) bool {
	// Shortcut invalid ranges.
	if r.lo > r.hi {
		return false
	}

	low := r.lo
	if r.FromExclusive {
		low++
	}
	if n < low {
		return false
	}

	hi := r.hi
	if r.ToExclusive {
		hi--
	}
	if n > hi {
		return false
	}
	return true
}

func (r *intRange) From(n int) RangeMatcher {
	r.lo = n
	return r
}

func (r *intRange) To(n int) RangeMatcher {
	r.hi = n
	return r
}

func (r *intRange) ExcludeFrom() RangeMatcher {
	r.FromExclusive = true
	return r
}

func (r *intRange) ExcludeTo() RangeMatcher {
	r.ToExclusive = true
	return r
}

func (r *intRange) Matches(x interface{}) bool {
	n, ok := x.(int)
	if !ok {
		return false
	}
	return r.Contains(n)
}

// NewRangeMatcher creates a valid default RangeMatcher covering all int values.
func NewRangeMatcher() RangeMatcher {
	return &intRange{
		lo: minInt,
		hi: maxInt,
	}
}
