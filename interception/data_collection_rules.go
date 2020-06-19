package interception

import (
	"fmt"
	"strings"

	"github.com/bearer/go-agent/filters"
)

// DataCollectionRule represents a data collection rule.
//
// Inactive rules descriptions generate nil *DataCollectionRule values.
type DataCollectionRule struct {
	filters.Filter
	LogLevel
}

// NewDCRFromDescription creates a DataCollectionRule from a DataCollectionRuleDescription
// and a valid filters.FilterMap.
func NewDCRFromDescription(filterMap filters.FilterMap, d DataCollectionRuleDescription) *DataCollectionRule {
	if !d.IsActive() {
		return nil
	}

	dcr := &DataCollectionRule{
		LogLevel: LogLevelFromString(d.Config.LogLevel),
	}
	if d.FilterHash != `` {
		f, ok := filterMap[d.FilterHash]
		if ok {
			dcr.Filter = f
		}
	}
	return dcr
}

// DataCollectionRuleDescription is a serialization-friendly description for a
// data collection rule.
type DataCollectionRuleDescription struct {
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

// IsActive checks whether the DataCollectionRuleDescription is active.
//
// Its assume rules are active by default, which is not explicit from Agent Spec.
func (d DataCollectionRuleDescription) IsActive() bool {
	return d.Config.IsActive()
}

func (d DataCollectionRuleDescription) String() string {
	b := strings.Builder{}
	hash := d.FilterHash
	if hash == `` {
		hash = `(unset)`
	}
	b.WriteString(fmt.Sprintf("%-28s: %-22s - ", hash, d.Params.TypeName))
	params := []string{d.Config.String()}
	if d.Params.IsErrorTriggerfilter {
		params = append(params, `IETF`)
	}
	if d.Params.Buid != `` {
		params = append(params, `BUID: `+d.Params.Buid)
	}
	if d.Params.AggregationFilterHash != `` {
		params = append(params, `AH: `+d.Params.AggregationFilterHash)
	}
	b.WriteString(strings.Join(params, `, `) + "\n")
	return b.String()
}

// DynamicConfigDescription provides a serialization-friendy description of DynamicConfig.
type DynamicConfigDescription struct {
	LogLevel string      // ALL, RESTRICTED, or DETECTED.
	Active   interface{} // Accept booleans only, but default to true.
}

// IsActive checks whether the DynamicConfigurationDescription is active.
//
// Its assume rules are active by default, which is not explicit from Agent Spec.
func (dcd DynamicConfigDescription) IsActive() bool {
	switch x := dcd.Active.(type) {
	case nil:
		return true // Inverse of default Go value.
	case bool:
		return x
	case string:
		return x == `` || strings.EqualFold(x, `true`)
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (dcd DynamicConfigDescription) String() string {
	if dcd.IsActive() {
		return fmt.Sprintf("Active/%s", LogLevelFromString(dcd.LogLevel))
	}
	return "Inactive"
}
