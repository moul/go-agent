package filters

import "strconv"

// Range provides the ability to check whether an int value is within an integer range.
//
// By default, its lo and hi limits are the maximum representable values, and
// they are included in comparisons.
//
// The interface allows a fluent initialization, like: NewRange().hi(200).ExcludeTo(),
// which will define the semi-open interval [minInt, 200[.
type Range interface {
	Contains(int) bool
	From(int) Range
	To(int) Range
	ExcludeFrom() Range
	ExcludeTo() Range
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

// NewRange creates a valid default Range covering all int values.
func NewRange() Range {
	return &intRange{
		lo: minInt,
		hi: maxInt,
	}
}

// Contains implements the Range interface.
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

func (r *intRange) From(n int) Range {
	r.lo = n
	return r
}

func (r *intRange) To(n int) Range {
	r.hi = n
	return r
}

func (r *intRange) ExcludeFrom() Range {
	r.FromExclusive = true
	return r
}

func (r *intRange) ExcludeTo() Range {
	r.ToExclusive = true
	return r
}
