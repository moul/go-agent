package interception

import (
	"io"
	"net/http"
	"net/url"

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

// ConnectEvent is the type of events dispatched at the Connect stage.
//
// Its Data() is a URL. Recommended use is to set the URL
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

// NewConnectEvent builds a ConnectEvent for a url.URL.
func NewConnectEvent(url *url.URL) *ConnectEvent {
	e := &ConnectEvent{}
	e.WithTopic(string(Connect))
	e.SetData(url)
	return e
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

// NewRequestEvent builds a RequestEvent for a *http.Request.
func NewRequestEvent(r *http.Request) *RequestEvent {
	e := &RequestEvent{}
	e.WithTopic(string(Request))
	e.SetData(r)
	return e
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
