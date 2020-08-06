package interception

import (
	"fmt"
	"strings"

	"github.com/bearer/go-agent/filters"
	"github.com/bearer/go-agent/proxy"
)

// DataCollectionRule represents a data collection rule.
//
// Inactive rules descriptions generate nil *DataCollectionRule values.
type DataCollectionRule struct {
	filters.Filter
	*LogLevel
	IsActive   *bool
	FilterHash string
	Params     map[string]interface{}
	Signature  string
}

// NewDCRFromDescription creates a DataCollectionRule from a DataCollectionRuleDescription
// and a valid filters.FilterMap.
func NewDCRFromDescription(filterMap filters.FilterMap, d DataCollectionRuleDescription) *DataCollectionRule {
	var logLevel *LogLevel
	if d.Config.LogLevel != nil {
		ll := LogLevelFromString(*d.Config.LogLevel)
		logLevel = &ll
	}
	dcr := &DataCollectionRule{
		FilterHash: d.FilterHash,
		LogLevel:   logLevel,
		IsActive:   d.Config.Active,
		Params:     d.Params,
		Signature:  d.Signature,
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
	Params     map[string]interface{}
	Config     DynamicConfigDescription
	Signature  string
}

func (d DataCollectionRuleDescription) String() string {
	b := strings.Builder{}
	hash := d.FilterHash
	if hash == `` {
		hash = `(unset)`
	}
	b.WriteString(fmt.Sprintf("%-28s: %-22s\n", hash, d.Params[`TypeName`]))
	return b.String()
}

// DynamicConfigDescription provides a serialization-friendy description of DynamicConfig.
type DynamicConfigDescription struct {
	LogLevel *string // ALL, RESTRICTED, or DETECTED.
	Active   *bool
}

// PrepareTriggeredRulesForReport translates DataCollectionRule objects
// representing triggered rules into the format used for reporting
func PrepareTriggeredRulesForReport(triggeredRules []*DataCollectionRule) []proxy.ReportDataCollectionRule {
	result := make([]proxy.ReportDataCollectionRule, len(triggeredRules))
	for i, rule := range triggeredRules {
		result[i] = proxy.ReportDataCollectionRule{
			FilterHash: rule.FilterHash,
			Params:     rule.Params,
			Signature:  rule.Signature,
		}
	}
	return result
}
