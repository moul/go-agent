package interception

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bearer/go-agent/proxy"
)

func TestDataCollectionRuleDescription_String(t *testing.T) {
	allLogLevel := All.String()
	trueVal := true
	falseVal := false
	type fields struct {
		FilterHash string
		Params     map[string]interface{}
		Config     DynamicConfigDescription
		Signature  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{`happy`, fields{
			FilterHash: `foo`,
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &trueVal,
			},
			Signature: "",
		}, "foo                         : ApiRule               \n"},
		{`no filter`, fields{
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &trueVal,
			},
			Signature: "",
		}, "(unset)                     : ApiRule               \n"},
		{`trigger filter`, fields{
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &trueVal,
			},
			Signature: "",
		}, "(unset)                     : ApiRule               \n"},
		{`with BUID`, fields{
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &trueVal,
			},
			Signature: "",
		}, "(unset)                     : ApiRule               \n"},
		{`with aggregation hash`, fields{
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &trueVal,
			},
			Signature: "",
		}, "(unset)                     : ApiRule               \n"},
		{`sad inactive`, fields{
			FilterHash: `foo`,
			Params: map[string]interface{}{
				`TypeName`: `ApiRule`,
			},
			Config: DynamicConfigDescription{
				LogLevel: &allLogLevel,
				Active:   &falseVal,
			},
			Signature: "",
		}, "foo                         : ApiRule               \n"},
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

func TestPrepareTriggeredRulesForReport(t *testing.T) {
	rules := []*DataCollectionRule{
		&DataCollectionRule{
			FilterHash: `hash1`,
			Params:     map[string]interface{}{`TypeName`: `T1`},
			Signature:  `sig1`,
		},
		&DataCollectionRule{
			FilterHash: `hash2`,
			Params:     map[string]interface{}{`TypeName`: `T2`},
			Signature:  `sig2`,
		},
	}

	expected := []proxy.ReportDataCollectionRule{
		proxy.ReportDataCollectionRule{
			FilterHash: `hash1`,
			Params:     map[string]interface{}{`TypeName`: `T1`},
			Signature:  `sig1`,
		},
		proxy.ReportDataCollectionRule{
			FilterHash: `hash2`,
			Params:     map[string]interface{}{`TypeName`: `T2`},
			Signature:  `sig2`,
		},
	}

	reportRules := PrepareTriggeredRulesForReport(rules)
	if !reflect.DeepEqual(reportRules, expected) {
		t.Errorf("Expected:\n%#v\n\nActual:\n%#v\n", expected, reportRules)
	}
}
