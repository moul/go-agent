package filters

import (
	"reflect"
	"testing"

	"github.com/bearer/go-agent/proxy"
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

func TestNewFilterFromDescription(t *testing.T) {
	tests := []struct {
		name string
		typ  FilterType
		want Filter
	}{
		{`bad type`, nil, nil},
		{`not`, NotFilterType, nil},
		{`set`, FilterSetFilterType, &filterSet{}},
		{`domain`, DomainFilterType, &DomainFilter{NewEmptyRegexpMatcher()}},
		{`method`, HTTPMethodFilterType, &HTTPMethodFilter{NewStringMatcher(``, true)}},
		{`param`, ParamFilterType, &ParamFilter{NewKeyValueMatcher(``, ``)}},
		{`path`, PathFilterType, &PathFilter{NewEmptyRegexpMatcher()}},
		{`request headers`, RequestHeadersFilterType, &RequestHeadersFilter{NewKeyValueMatcher(``, ``)}},
		{`response headers`, ResponseHeadersFilterType, &ResponseHeadersFilter{NewKeyValueMatcher(``, ``)}},
		{`status`, StatusCodeFilterType, &StatusCodeFilter{NewRangeMatcher().From(0).To(0)}},
		{`error`, ConnectionErrorFilterType, &ConnectionErrorFilter{}},
		{`yes`, YesInternalFilter, &YesFilter{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typeName string
			if tt.name == `bad type` {
				typeName = tt.name
			} else {
				typeName = tt.typ.Name()
			}
			fd := FilterDescription{
				TypeName: typeName,
			}
			if got := NewFilterFromDescription(nil, &fd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFilterFromDescription() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestFilterDescription_String(t *testing.T) {
	type fields struct {
		ChildHash            string
		Value                string
		Pattern              RegexpMatcherDescription
		FilterSetDescription FilterSetDescription
		KeyValueDescription  KeyValueDescription
		Range                RangeMatcherDescription
		StageType            string
		TypeName             string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{`happy`, fields{
			ChildHash: `ch`,
			Value:     `val`,
			StageType: string(proxy.StageConnect),
			TypeName:  `foo`,
		}, "foo                    - ConnectStage         - H: ch\nValue: val\n"},
		{`tiny`, fields{
			StageType: string(proxy.StageConnect),
			TypeName:  `bar`,
		}, "bar                    - ConnectStage         - \n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := FilterDescription{
				ChildHash:            tt.fields.ChildHash,
				Value:                tt.fields.Value,
				Pattern:              tt.fields.Pattern,
				FilterSetDescription: tt.fields.FilterSetDescription,
				KeyValueDescription:  tt.fields.KeyValueDescription,
				Range:                tt.fields.Range,
				StageType:            tt.fields.StageType,
				TypeName:             tt.fields.TypeName,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
