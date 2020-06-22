package interception

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

const (
	// TopicConnect is the earliest event triggered in an intercepted API.
	// It is used to validate the endpoint URL, regardless of the Request which
	// will be sent to it.
	TopicConnect events.Topic = "connect"

	// TopicRequest is the second event triggered in an intercepted API.
	// It is used to validate the Request itself, as well as its context.
	TopicRequest events.Topic = "request"

	// TopicResponse is the third event triggered in an intercepted API.
	// It is used to react to the response headers and possibly start of body
	// being received. Note that at this point, there is no guarantee that either
	// the Request or Response bodies are actually entirely available, due to
	// HTTP advanced features like request/response interleaving. It is not
	// triggered if the round-trip returns an error, as the associated response
	// is not guaranteed to be well-formed.
	TopicResponse events.Topic = "response"

	// TopicBodies is the fourth and last event triggered in an intercepted API.
	// It is used once the bodies on both Request and Response have been closed
	// by the API client. It does NOT mean that these bodies are necessarily
	// complete, as a client may have closed a request early.
	TopicBodies events.Topic = "bodies"

	// TopicReport is the event used to request transmission of a ReportLog to
	// the logs platform. Unlike its four siblings, it can be triggered at any
	// stage of the API call lifecycle.
	TopicReport events.Topic = "report_log"
)

// APIEvent is the type common to all API call lifecycle events.
type APIEvent interface {
	events.Event
	LogLevel() LogLevel
	SetLogLevel(l LogLevel) APIEvent
}
type apiEvent struct {
	events.EventBase
	logLevel LogLevel
}

func (ae *apiEvent) LogLevel() LogLevel {
	return ae.logLevel
}

func (ae *apiEvent) SetLogLevel(l LogLevel) APIEvent {
	// Enforce value validation.
	ae.logLevel = LogLevelFromInt(int(l))
	return ae
}

// ConnectEvent is the type of events dispatched at the TopicConnect stage.
//
// Its Data() is a URL. Recommended use is to set the URL
type ConnectEvent struct {
	apiEvent

	// Host is the host to which the request is sent. It may be an IPv6 braced address.
	Host string

	// Port is the TCP port number, in the uint16 range by RFC793.
	Port uint16

	// Scheme, also known as "protocol", is the first part of RFC3986 URL syntax.
	Scheme string
}

// Request overrides the events.EventBase.Request method, building an on-the-fly
// request from the event fields.
func (re ConnectEvent) Request() *http.Request {
	req, _ := http.NewRequest(``, (&url.URL{
		Scheme: re.Scheme,
		Host:   fmt.Sprintf(`%s:%d`, re.Host, re.Port),
	}).String(), nil)
	return req
}

// Topic is part of the Event interface.
func (re ConnectEvent) Topic() events.Topic {
	return TopicConnect
}

// NewConnectEvent builds a ConnectEvent for a url.URL.
func NewConnectEvent(url *url.URL) *ConnectEvent {
	e := &ConnectEvent{}
	e.SetData(url)
	return e
}

// NewReportEvent builds a ReportEvent, empty but for logLevel, stage, and error.
func NewReportEvent(logLevel LogLevel, stage proxy.Stage, err error) *ReportEvent {
	be := &BodiesEvent{
		apiEvent: apiEvent{
			EventBase: events.EventBase{Error: err},
			logLevel:  logLevel,
		},
	}
	be.SetTopic(string(TopicRequest))

	return &ReportEvent{
		BodiesEvent: be,
		Stage: stage,
	}
}

// RequestEvent is the type of events dispatched at the TopicRequest stages.
type RequestEvent struct {
	apiEvent
}

// Topic is part of the Event interface.
func (re RequestEvent) Topic() events.Topic {
	return TopicRequest
}

// ResponseEvent is the type of events dispatched at the TopicResponse stage.
type ResponseEvent struct {
	apiEvent
}

// Topic is part of the Event interface.
func (ResponseEvent) Topic() events.Topic {
	return TopicResponse

}

// BodiesEvent is the type of events dispatched at the TopicBodies stage.
type BodiesEvent struct {
	apiEvent
	readTimestamp             time.Time
	RequestBody, ResponseBody interface{}
	RequestSha, ResponseSha   string
}

// ReportEvent is emitted to publish a call proxy.ReportLog.
type ReportEvent struct {
	*BodiesEvent
	proxy.Stage
	T0, T1 time.Time
}

// Topic is part of the Event interface.
func (ReportEvent) Topic() events.Topic {
	return TopicReport
}

// DCRProvider is an events.Listener provider returning listeners based on the
// active data collection rules.
type DCRProvider struct {
	DCRs []*DataCollectionRule
}

// Listeners implements the events.ListenerProvider interface.
func (p DCRProvider) Listeners(e events.Event) []events.Listener {
	switch e.Topic() {
	case TopicConnect, TopicRequest, TopicResponse, TopicBodies:
		return []events.Listener{func(ctx context.Context, e events.Event) error {
			var logLevel LogLevel
			ae, ok := e.(APIEvent)
			if !ok {
				return fmt.Errorf("topic %s used with non-APIEvent type %T", e.Topic(), e)
			}
			logLevel = ae.LogLevel()
			original := logLevel
			for _, dcr := range p.DCRs {
				// Maximum LogLevel reached: no need to check more rules.
				if logLevel == All {
					break
				}
				// Rule won't allow more logging, just skip it.
				if dcr.LogLevel <= logLevel {
					continue
				}
				// No filter: just apply the rule logLevel without running it
				if dcr.Filter == nil {
					logLevel = dcr.LogLevel
					continue
				}
				// Rule may increase logLevel if it matches: run it.
				if dcr.MatchesCall(e) {
					logLevel = dcr.LogLevel
				}
			}
			if logLevel != original {
				ae.SetLogLevel(logLevel)
			}
			return nil
		}}
	default:
		return nil
	}
}

// ProxyProvider is an events.ListenerProvider returning a proxy listener.
type ProxyProvider struct {
	*proxy.Sender
}

// Listeners implements the events.ListenerProvider interface.
func (p ProxyProvider) Listeners(e events.Event) []events.Listener {
	if e.Topic() != TopicReport {
		return nil
	}
	listener := func(ctx context.Context, e events.Event) error {
		re, ok := e.(*ReportEvent)
		if !ok {
			return fmt.Errorf("topic %s used with event type %T", e.Topic(), e)
		}
		ll := re.LogLevel()
		rl := ll.Prepare(re)
		p.Send(rl)
		return nil
	}

	return []events.Listener{listener}
}
