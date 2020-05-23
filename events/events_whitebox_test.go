package events

import (
	"testing"
)

func TestNewDispatcher(t *testing.T) {
	di := NewDispatcher()
	if di == nil {
		t.Fatalf("got nil dispatcher")
	}
	d, ok := di.(*dispatcher)
	if !ok {
		t.Fatalf("unexpected new dispatcher type %T", d)
	}
	if d.providers == nil {
		t.Fatal("new dispatcher contains no providers map")
	}
}

// Test_dispatcher_SetProvider tests implementation details, which the black box
// test can not verify.
func Test_dispatcher_SetProvider_wb(t *testing.T) {
	d := dispatcher{}
	d.SetProvider("topic", nil)
	if d.providers == nil || len(d.providers) == 0 {
		t.Fatal("empty providers list in spite of SetProvider")
	}
}
