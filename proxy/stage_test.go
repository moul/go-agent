package proxy

import "testing"

func TestStage_Next(t *testing.T) {
	tests := []struct {
		name string
		s    Stage
		want Stage
	}{
		{"undefined", StageUndefined, StageInvalid},
		{"connect", StageConnect, StageRequest},
		{"request", StageRequest, StageResponse},
		{"response", StageResponse, StageBodies},
		{"bodies", StageBodies, StageInvalid},
		{"invalid", StageInvalid, StageInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Next(); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
