package filters

import (
	"fmt"
	"testing"
)

func TestIntRange_Contains(t *testing.T) {
	var (
		Closed1_10     = NewRange().From(1).To(10)
		LeftOpen0_10   = NewRange().From(0).To(10).ExcludeFrom()
		GreaterEqual20 = NewRange().From(20)
		LessThan400    = NewRange().To(400).ExcludeTo()
		Open300_400    = NewRange().ExcludeFrom().From(300).ExcludeTo().To(400)
		Invalid        = NewRange().From(1).To(0)
	)
	tests := []struct {
		r        Range
		check    int
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
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s.Contains(%d)", tt.r, tt.check)
		t.Run(name, func(t *testing.T) {
			actual := tt.r.Contains(tt.check)
			if actual != tt.expected {
				t.Errorf("%s.Contains(%d): expected %t, got %t", tt.r, tt.check, tt.expected, actual)
			}
		})
	}
}
