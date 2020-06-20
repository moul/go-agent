package interception

import (
	"fmt"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  string
	}{
		{`all`, All, `All`},
		{`sad`, All + 1, fmt.Sprintf(`LogLevel(%d)`, All+1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
