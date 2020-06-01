package agent

import (
	"io"
	"net/http"
	"regexp"

	"github.com/bearer/go-agent/events"
)

const (
	// Connect is the earliest event triggered in an intercepted API.
	// It is used to validate the endpoint URL, regardless of the Request which
	// will be sent to it.
	Connect events.Topic = "connect"

	// Request is the second event triggered in an intercepted API.
	// It is used to validate the Request itself, as well as its context.
	Request events.Topic = "request"

	// Response is the third event triggered in an intercepted API.
	// It is used to react to the response headers and possibly start of body
	// being received. Note that at this point, there is no guarantee that either
	// the Request or Response bodies are actually entirely available, due to
	// HTTP advanced features like request/response interleaving. It is not
	// triggered if the round-trip returns an error, as the associated response
	// is not guaranteed to be well-formed.
	Response events.Topic = "response"

	// Bodies is the fourth and last event triggered in an intercepted API.
	// It is used once the bodies on both Request and Response have been closed
	// by the API client. It does NOT mean that these bodies are necessarily
	// complete, as a client may have closed a request early.
	Bodies events.Topic = "bodies"
)

// SchemeRegexp is the regular expression matching the RFC3986 grammar
// production for URL schemes.
var SchemeRegexp = regexp.MustCompile(`[\w][-+\.\w]+`)

// Filters provides the filter Listeners for the Connect and Request stages.
func Filters(e events.Event) []events.Listener {
	return nil
}

// DataCollectionRules provides the data collection rules Listeners for the
// Response and Bodies stages.
func DataCollectionRules(e events.Event) []events.Listener {
	return nil
}

// ConnectEvent is the type of events dispatched at the Connect stage.
type ConnectEvent struct {
	events.EventBase

	// Host is the host to which the request is sent. It may be an IPv6 braced address.
	Host string

	// Port is the TCP port number, in the uint16 range by RFC793.
	Port uint16

	// Scheme, also known as "protocol", is the first part of RFC3986 URL syntax.
	Scheme string
}

// Topic is part of the Event interface.
func (re ConnectEvent) Topic() events.Topic {
	return Connect
}

// RequestEvent is the type of events dispatched at the Request stages.
type RequestEvent struct {
	events.EventBase

	// Request is the Request being transferred to the API endpoint.
	Request *http.Request
}

// Topic is part of the Event interface.
func (re RequestEvent) Topic() events.Topic {
	return Request
}

// ResponseEvent is the type of events dispatched at the Response stage.
type ResponseEvent struct {
	events.EventBase
	Response *http.Response
}

// Topic is part of the Event interface.
func (ResponseEvent) Topic() events.Topic {
	return Response

}

// BodiesEvent is the type of events dispatched at the Bodies stage.
type BodiesEvent struct {
	events.EventBase

	RequestBody, ResponseBody io.ReadCloser
}

// Topic is part of the Event interface.
func (BodiesEvent) Topic() events.Topic {
	return Bodies
}
