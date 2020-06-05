package agent

import (
	"net/http"

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

// RequestEvent is the type of events dispatched at the Connect and Request stages.
type RequestEvent struct {
	events.EventBase
	Req http.Request
}

// Topic implements the Event interface.
func (re RequestEvent) Topic() events.Topic {
	return "request"
}

// ResponseEvent is the type of events dispatched at the Response and Bodies stages.
type ResponseEvent struct {
	events.EventBase
	Res http.Response
}

// Topic implements the Event interface.
func (ResponseEvent) Topic() events.Topic {
	return "response"
}

