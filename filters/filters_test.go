package filters

import (
	"testing"
)

func Test_filterType_Values(t *testing.T) {
	tests := []struct {
		name                        string
		typ                         FilterType
		wantName                    string
		wantsRequest, wantsResponse bool
	}{
		{"not", NotFilterType, "NotFilter", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.Name(); got != tt.wantName {
				t.Errorf("Name() = %v, want %v", got, tt.wantName)
			}
			if got := tt.typ.WantsRequest(); got != tt.wantsRequest {
				t.Errorf("WantsRequest() = %v, expected %v", got, tt.wantsRequest)
			}
			if got := tt.typ.WantsResponse(); got != tt.wantsResponse {
				t.Errorf("WantsRes@ponse() = %v, expected %v", got, tt.wantsResponse)
			}
		})
	}
}
