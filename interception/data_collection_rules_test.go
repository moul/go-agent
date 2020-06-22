package interception

import (
	"fmt"
	"testing"
)

func TestDataCollectionRuleDescription_String(t *testing.T) {
	type fields struct {
		FilterHash string
		Params     struct {
			AggregationFilterHash string
			Buid                  string
			IsErrorTriggerfilter  bool
			TypeName              string
		}
		Config    DynamicConfigDescription
		Signature string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{`happy`, fields{
			FilterHash: `foo`,
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{``, ``, false, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   true,
			},
			Signature: "",
		}, "foo                         : ApiRule                - Active/All\n"},
		{`sad no filter`, fields{
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{``, ``, false, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   true,
			},
			Signature: "",
		}, "(unset)                     : ApiRule                - Active/All\n"},
		{`trigger filter`, fields{
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{``, ``, true, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   true,
			},
			Signature: "",
		}, "(unset)                     : ApiRule                - Active/All, IETF\n"},
		{`with BUID`, fields{
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{``, `bar`, false, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   true,
			},
			Signature: "",
		}, "(unset)                     : ApiRule                - Active/All, BUID: bar\n"},
		{`with aggregation hash`, fields{
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{`baz`, ``, false, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   true,
			},
			Signature: "",
		}, "(unset)                     : ApiRule                - Active/All, AH: baz\n"},
		{`sad inactive`, fields{
			FilterHash: `foo`,
			Params: struct {
				AggregationFilterHash string
				Buid                  string
				IsErrorTriggerfilter  bool
				TypeName              string
			}{``, ``, false, `ApiRule`},
			Config: DynamicConfigDescription{
				LogLevel: All.String(),
				Active:   false,
			},
			Signature: "",
		}, "foo                         : ApiRule                - Inactive\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DataCollectionRuleDescription{
				FilterHash: tt.fields.FilterHash,
				Params:     tt.fields.Params,
				Config:     tt.fields.Config,
				Signature:  tt.fields.Signature,
			}
			if testing.Verbose() {
				t.Log(fmt.Sprintf("%s%s%v\n%v", d.String(), tt.want, []byte(d.String()), []byte(tt.want)))
			}
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataCollectionRuleDescription_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		active interface{}
		want   bool
	}{
		{`active`, true, true},
		{`inactive`, false, false},
		{`nil`, nil, true},
		{`empty string`, ``, true},
		{`string false`, `false`, false},
		{`string true`, `True`, true},
		{`number`, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DataCollectionRuleDescription{
				Config: DynamicConfigDescription{Active: tt.active},
			}
			if got := d.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDynamicConfigDescription_String(t *testing.T) {
	type fields struct {
		LogLevel string
		Active   interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{`active/all`, fields{All.String(), true}, `Active/All`},
		{`inactive/all`, fields{All.String(), false}, `Inactive`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dcd := DynamicConfigDescription{
				LogLevel: tt.fields.LogLevel,
				Active:   tt.fields.Active,
			}
			if got := dcd.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
