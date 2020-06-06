package config

import (
	"fmt"
	"strings"
)

// DataCollectionRule represents a data collection rule.
// @FIXME Define actual type instead of placeholder.
type DataCollectionRule interface{}

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
	LogLevel string // ALL, RESTRICTED, or DETECTED.
	Active   bool
}

// String implements fmt.Stringer.
func (d DynamicConfigDescription) String() string {
	if d.Active {
		return fmt.Sprintf("Active/%s", d.LogLevel)
	}
	return "Inactive"
}

