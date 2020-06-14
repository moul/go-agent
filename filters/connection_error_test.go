package filters

import (
	"testing"
)

func TestConnectionErrorFilter_Type(t *testing.T) {
	expected := ConnectionErrorFilterType.String()
	var f ConnectionErrorFilter
	actual := f.Type().String()
	if actual != expected {
		t.Errorf("Type() = %v, want %v", actual, expected)
	}
}

