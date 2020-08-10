package interception

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// RoundTripper is the instrumented implementation of http.RoundTripper.
//
// It triggers events for the TopicConnect, TopicRequest, and TopicResponse stages.
type RoundTripper struct {
	events.Dispatcher
	Underlying http.RoundTripper
}

// schemeRegexp is the regular expression matching the RFC3986 grammar
// production for URL schemes.
var schemeRegexp = regexp.MustCompile(`^[\w][-+.\w]+$`)

// RFCListener validates the destination URL under RFC793, RFC1384, RFC1738
// and RFC3986 before entering the standard Bearer multistage API wrapping.
//
// It is hard-coded in the round-tripper to avoid its being disabled.
func RFCListener(_ context.Context, e events.Event) error {
	ce, ok := e.(*ConnectEvent)
	if !ok {
		return errors.New(`the RFCListener is only used with ConnectEvent`)
	}
	data := ce.Data()
	url, ok := data.(*url.URL)
	if !ok {
		return errors.New(`no URL found in ConnectEvent`)
	}

	ce.Host = url.Hostname()

	// XXX As some point, we might want to include a host validation, following RFC1738 Sec. 3.1
	ce.Scheme = url.Scheme

	// RFC3986.
	if !schemeRegexp.MatchString(ce.Scheme) {
		return fmt.Errorf("invalid scheme [%s]", ce.Scheme)
	}

	sPort := url.Port()
	if sPort == `` {
		// Cf. Go runtime: src/net/http/transport.go
		PortMap := map[string]string{
			"http":   "80",
			"https":  "443",
			"socks5": "1080",
		}
		sPort, ok = PortMap[ce.Scheme]
		if !ok {
			return fmt.Errorf("ill-formed port specification in Host [%s]", url.Host)
		}
	}

	intPort, err := strconv.Atoi(sPort)
	if err != nil {
		// This might be a case for a panic, since URL.Port() is expected to
		// return an empty string if the port is not numeric.
		return fmt.Errorf("ill-formed port [%s]", sPort)
	}

	// RFC793 sec 3.1 and RFC1340 p.7.
	if intPort <= 0 || intPort > 2<<15-1 {
		return fmt.Errorf("invalid port [%d]", intPort)
	}
	ce.Port = uint16(intPort)

	return nil
}

// stageConnect implements the Bearer TopicConnect stage.
func (rt *RoundTripper) stageConnect(ctx context.Context, url *url.URL) (APIEvent, error) {
	e := NewConnectEvent(url)
	_, err := rt.Dispatch(ctx, e)
	if err != nil {
		return e, err
	}
	if err = ctx.Err(); err != nil {
		return e, err
	}
	return e, nil
}

func (rt *RoundTripper) stageRequest(prevEvent APIEvent, request *http.Request) (APIEvent, error) {
	if prevEvent == nil || !prevEvent.Config().IsActive {
		return nil, nil
	}

	ctx := request.Context()
	be := &RequestEvent{}
	be.SetTopic(string(TopicRequest))
	be.SetConfig(prevEvent.Config())
	be.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
	be.SetRequest(request)
	_, err := rt.Dispatch(ctx, be)
	if err != nil {
		return be, err
	}
	if err = ctx.Err(); err != nil {
		return be, err
	}

	return be, nil
}

func (rt *RoundTripper) stageResponse(ctx context.Context, prevEvent APIEvent, request *http.Request, response *http.Response, err error) (APIEvent, error) {
	if prevEvent == nil || !prevEvent.Config().IsActive {
		return nil, nil
	}
	if err != nil {
		return prevEvent, err
	}
	e := &ResponseEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}
	e.SetConfig(prevEvent.Config())
	e.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
	e.SetRequest(request).SetResponse(response)
	_, err = rt.Dispatch(ctx, e)
	if err != nil {
		return e, err
	}
	if err = ctx.Err(); err != nil {
		return e, err
	}

	return e, nil
}

func (rt *RoundTripper) stageBodies(ctx context.Context, prevEvent APIEvent, request *http.Request, response *http.Response, err error) *ReportEvent {
	if prevEvent == nil || !prevEvent.Config().IsActive {
		return nil
	}
	rev := NewReportEvent(proxy.StageBodies, err)
	e := &BodiesEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}
	e.SetTopic(string(TopicBodies))
	rev.BodiesEvent = e
	rev.SetConfig(prevEvent.Config())
	rev.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
	rev.SetRequest(request).SetResponse(response)
	if err != nil {
		rev.Error = err
		return rev
	}
	_, rev.Error = rt.Dispatch(ctx, rev.BodiesEvent)
	if rev.Error != nil {
		return rev
	}
	if rev.Error = ctx.Err(); rev.Error != nil {
		return rev
	}
	return rev
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var prevEvent APIEvent
	var err error
	var rev *ReportEvent
	var (
		// Ensure valid timestamps even on early returns.
		t0 = time.Now()
		t1 = t0
	)

	ctx := request.Context()

	defer func() {
		if rev == nil || !rev.Config().IsActive {
			return
		}
		rev.T0 = t0
		// If the t1 reset was not reached, us the time spent in the agent.
		if t1 == t0 {
			t1 = time.Now()
		}
		rev.T1 = t1
		_, _ = rt.Dispatch(ctx, rev)
	}()

	if prevEvent, err = rt.stageConnect(ctx, request.URL); err != nil {
		rev = NewReportEvent(proxy.StageConnect, err)
		rev.SetRequest(request)
		rev.SetConfig(prevEvent.Config())
		rev.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
		return nil, err
	}

	if prevEvent, err = rt.stageRequest(prevEvent, request); err != nil {
		rev = NewReportEvent(proxy.StageRequest, err)
		rev.SetRequest(request)
		rev.SetConfig(prevEvent.Config())
		rev.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
		return nil, err
	}

	if request.Body != nil {
		request.Body = NewBodyReadCloser(request.Body, MaximumBodySize+1)
	}

	// Perform and time the underlying API call, without resBody capture.
	t0 = time.Now()
	response, rtErr := rt.Underlying.RoundTrip(request)
	t1 = time.Now()

	if response != nil && response.Body != nil {
		response.Body = NewBodyReadCloser(response.Body, MaximumBodySize+1)
	}

	if prevEvent, err = rt.stageResponse(ctx, prevEvent, request, response, rtErr); err != nil {
		stage := proxy.StageResponse
		if response == nil {
			stage = proxy.StageRequest
		}
		rev = NewReportEvent(stage, err)
		rev.SetRequest(request).SetResponse(response)
		rev.SetConfig(prevEvent.Config())
		rev.SetTriggeredDataCollectionRules(prevEvent.TriggeredDataCollectionRules())
		return rev.Response(), err
	}

	rev = rt.stageBodies(ctx, prevEvent, request, response, err)
	if rev == nil {
		return response, rtErr
	}
	return rev.Response(), rev.Err()
}
