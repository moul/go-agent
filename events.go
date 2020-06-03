package agent

import (
	"github.com/bearer/go-agent/events"
)

// Filters provides the filter Listeners for the Connect and Request stages.
func Filters(e events.Event) []events.Listener {
	return nil
}

// DataCollectionRules provides the data collection rules Listeners for the
// Response and Bodies stages.
func DataCollectionRules(e events.Event) []events.Listener {
	return nil
}

