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
func (rt *RoundTripper) stageConnect(ctx context.Context, url *url.URL) (*APIEventConfig, error) {
	e := NewConnectEvent(url)
	_, err := rt.Dispatch(ctx, e)
	if err != nil {
		return e.Config(), err
	}
	if err = ctx.Err(); err != nil {
		return e.Config(), err
	}
	return e.Config(), nil
}

func (rt *RoundTripper) stageRequest(apiEventConfig *APIEventConfig, request *http.Request) (*APIEventConfig, error) {
	if apiEventConfig == nil || !apiEventConfig.IsActive {
		return apiEventConfig, nil
	}

	ctx := request.Context()
	be := &BodiesEvent{}
	be.SetTopic(string(TopicRequest))
	be.SetConfig(apiEventConfig)
	be.SetRequest(request)
	_, err := rt.Dispatch(ctx, be)
	if err != nil {
		return be.Config(), err
	}
	if err = ctx.Err(); err != nil {
		return be.Config(), err
	}

	return be.Config(), nil
}

func (rt *RoundTripper) stageResponse(ctx context.Context, apiEventConfig *APIEventConfig, request *http.Request, response *http.Response, err error) (*APIEventConfig, error) {
	if apiEventConfig == nil || !apiEventConfig.IsActive {
		return apiEventConfig, nil
	}
	if err != nil {
		return apiEventConfig, err
	}
	e := &ResponseEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}
	e.SetConfig(apiEventConfig)
	e.SetRequest(request).SetResponse(response)
	_, err = rt.Dispatch(ctx, e)
	if err != nil {
		return e.Config(), err
	}
	if err = ctx.Err(); err != nil {
		return e.Config(), err
	}

	return e.Config(), nil
}

func (rt *RoundTripper) stageBodies(ctx context.Context, apiEventConfig *APIEventConfig, request *http.Request, response *http.Response, err error) (*APIEventConfig, *ReportEvent) {
	if apiEventConfig == nil || !apiEventConfig.IsActive {
		return apiEventConfig, nil
	}
	rev := NewReportEvent(apiEventConfig.LogLevel, proxy.StageBodies, err)
	e := &BodiesEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}
	e.SetTopic(string(TopicBodies))
	rev.BodiesEvent = e
	e.SetConfig(apiEventConfig)
	rev.SetConfig(apiEventConfig)
	rev.SetRequest(request).SetResponse(response)
	if err != nil {
		rev.Error = err
		return e.Config(), rev
	}
	_, rev.Error = rt.Dispatch(ctx, rev.BodiesEvent)
	if rev.Error != nil {
		return e.Config(), rev
	}
	if rev.Error = ctx.Err(); rev.Error != nil {
		return e.Config(), rev
	}
	rev.T1 = rev.readTimestamp
	return e.Config(), rev
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var apiEventConfig *APIEventConfig
	var err error
	var rev *ReportEvent
	var (
		// Ensure valid timestamps even on early returns.
		t0 = time.Now()
		t1 = t0
	)

	ctx := request.Context()

	defer func() {
		if rev == nil {
			return
		}
		rev.T0 = t0
		// If the t1 reset was not reached, us the time spent in the agent.
		if t1 == t0 {
			t1 = time.Now()
		}
		rev.T1 = t1
		if err == nil && apiEventConfig != nil && apiEventConfig.IsActive {
			_, _ = rt.Dispatch(ctx, rev)
		}
	}()

	if apiEventConfig, err = rt.stageConnect(ctx, request.URL); err != nil {
		rev = NewReportEvent(apiEventConfig.LogLevel, proxy.StageConnect, err)
		rev.SetRequest(request)
		return nil, err
	}

	if apiEventConfig, err = rt.stageRequest(apiEventConfig, request); err != nil {
		rev = NewReportEvent(apiEventConfig.LogLevel, proxy.StageRequest, err)
		rev.SetRequest(request)
		return nil, err
	}

	// Perform and time the underlying API call, without resBody capture.
	t0 = time.Now()
	response, rtErr := rt.Underlying.RoundTrip(request)
	t1 = time.Now()

	if apiEventConfig, err = rt.stageResponse(ctx, apiEventConfig, request, response, rtErr); err != nil {
		rev = NewReportEvent(apiEventConfig.LogLevel, proxy.StageResponse, err)
		rev.SetRequest(request).SetResponse(response)
		return rev.Response(), err
	}

	// No need to check logLevel here: if we reached that point, logLevel is All.
	apiEventConfig, rev = rt.stageBodies(ctx, apiEventConfig, request, response, err)
	if rev == nil {
		return response, rtErr
	}
	return rev.Response(), rev.Err()
}
