package filters

import (
	"fmt"
	"testing"
)

func TestIntRange_Contains(t *testing.T) {
	var (
		Closed1_10     = NewRangeMatcher().From(1).To(10)
		LeftOpen0_10   = NewRangeMatcher().From(0).To(10).ExcludeFrom()
		GreaterEqual20 = NewRangeMatcher().From(20)
		LessThan400    = NewRangeMatcher().To(400).ExcludeTo()
		Open300_400    = NewRangeMatcher().ExcludeFrom().From(300).ExcludeTo().To(400)
		Invalid        = NewRangeMatcher().From(1).To(0)
	)
	tests := []struct {
		r RangeMatcher
		// check needs to be an int to succeed.
		check    interface{}
		expected bool
	}{
		// { from: 1, to: 10 } ⇒ 1 <= VALUE <= 10
		{Closed1_10, 0, false},
		{Closed1_10, 1, true},
		{Closed1_10, 10, true},
		{Closed1_10, 11, false},

		// { from: 0, fromExclusive: true, to: 10 } ⇒ 0 < VALUE <= 10
		{LeftOpen0_10, -1, false},
		{LeftOpen0_10, 0, false},
		{LeftOpen0_10, 10, true},
		{LeftOpen0_10, 11, false},

		// { from: 20 } ⇒ VALUE >= 20
		{GreaterEqual20, 19, false},
		{GreaterEqual20, 20, true},
		{GreaterEqual20, 21, true},

		// { to: 400, toExclusive: true } ⇒ VALUE < 400
		{LessThan400, 399, true},
		{LessThan400, 400, false},
		{LessThan400, 401, false},

		// { from: 300, fromExclusive: true, to: 400, toExclusive: true } ⇒ 300 < VALUE < 400
		{Open300_400, 299, false},
		{Open300_400, 300, false},
		{Open300_400, 301, true},
		{Open300_400, 399, true},
		{Open300_400, 400, false},
		{Open300_400, 401, false},

		// Not in Agent Spec.
		{Invalid, 1, false},
		{NewHTTPStatusMatcher(), "not an int", false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s.Contains(%d)", tt.r, tt.check)
		t.Run(name, func(t *testing.T) {
			actual := tt.r.Matches(tt.check)
			if actual != tt.expected {
				t.Errorf("%s.Contains(%d): expected %t, got %t", tt.r, tt.check, tt.expected, actual)
			}
		})
	}
}

func TestRangeMatcherDescription_ToInt(t *testing.T) {
	tests := []struct {
		name  string
		mixed interface{}
		want  int
	}{
		{`happy int`, 42, 42},
		{`happy float`, 42.0, 42},
		{`happy string`, `42`, 42},
		{`sad`, 2i, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ra RangeMatcherDescription
			if got := ra.ToInt(tt.mixed); got != tt.want {
				t.Errorf("ToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeMatcherDescription_String(t *testing.T) {
	tests := []struct {
		name        string
		ExcludeFrom bool
		ExcludeTo   bool
		want        string
	}{
		{`closed`, false, false, `[1:2]`},
		{`half-open right`, false, true, `[1:2[`},
		{`half-open left`, true, false, `]1:2]`},
		{`open`, true, true, `]1:2[`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := RangeMatcherDescription{
				From:        1,
				To:          2,
				ExcludeFrom: tt.ExcludeFrom,
				ExcludeTo:   tt.ExcludeTo,
			}
			if got := d.String(); got != `Range: `+tt.want+"\n" {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
