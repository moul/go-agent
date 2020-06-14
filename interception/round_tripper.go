package interception

import (
	"context"
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
		return nil
	}
	url, ok := ce.Data().(*url.URL)
	if !ok {
		return nil
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
func (rt *RoundTripper) stageConnect(ctx context.Context, url *url.URL) (LogLevel, error) {
	e := NewConnectEvent(url)
	_, err := rt.Dispatch(ctx, e)
	if err != nil {
		return e.LogLevel(), err
	}
	if err = ctx.Err(); err != nil {
		return e.LogLevel(), err
	}
	return e.LogLevel(), nil
}

func (rt *RoundTripper) stageRequest(logLevel LogLevel, request *http.Request) (LogLevel, error) {
	ctx := request.Context()
	e := &RequestEvent{}
	e.SetLogLevel(logLevel)
	e.SetRequest(request)
	_, err := rt.Dispatch(ctx, e)
	if err != nil {
		return e.LogLevel(), err
	}
	if err = ctx.Err(); err != nil {
		return e.LogLevel(), err
	}

	return e.LogLevel(), nil
}

func (rt *RoundTripper) stageResponse(ctx context.Context, logLevel LogLevel, request *http.Request, response *http.Response, err error) (LogLevel, error) {
	e := &ResponseEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}
	e.SetLogLevel(logLevel)
	e.SetRequest(request).SetResponse(response)
	_, err = rt.Dispatch(ctx, e)
	if err != nil {
		return e.LogLevel(), err
	}
	if err = ctx.Err(); err != nil {
		return e.LogLevel(), err
	}

	return e.LogLevel(), nil
}

func (rt *RoundTripper) stageBodies(ctx context.Context, logLevel LogLevel, request *http.Request, response *http.Response, err error) *ReportEvent {
	rev := NewReportEvent(logLevel, proxy.StageBodies, err)
	rev.BodiesEvent = BodiesEvent{apiEvent: apiEvent{EventBase: events.EventBase{Error: err}}}

	rev.SetLogLevel(logLevel)
	rev.SetRequest(request).SetResponse(response)
	_, rev.Error = rt.Dispatch(ctx, &rev.BodiesEvent)
	if rev.Error != nil {
		return rev
	}
	if rev.Error = ctx.Err(); rev.Error != nil {
		return rev
	}
	rev.T1 = rev.readTimestamp
	return rev
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	logLevel := Detected
	var err error
	var rev *ReportEvent
	var (
		// Ensure valid timestamps even on early returns.
		t0 = time.Now()
		t1 = t0
	)

	// XXX To add a platform-driven timeout, use context.WithTimeout here.
	ctx := context.WithValue(request.Context(), LogLevelKey, logLevel)

	defer func() {
		rev.T0 = t0
		// If the t1 reset was not reached, us the time spent in the agent.
		if t1 == t0 {
			t1 = time.Now()
		}
		rev.T1 = t1
		_, _ = rt.Dispatch(ctx, rev)
	}()

	if logLevel, err = rt.stageConnect(ctx, request.URL); err != nil {
		rev = NewReportEvent(logLevel, proxy.StageConnect, err)
		rev.SetRequest(request)
		return nil, err
	}

	if logLevel, err = rt.stageRequest(logLevel, request); err != nil {
		rev = NewReportEvent(logLevel, proxy.StageRequest, err)
		rev.SetRequest(request)
		return nil, err
	}

	// Perform and time the underlying API call, without body capture.
	t0 = time.Now()
	response, err := rt.Underlying.RoundTrip(request)
	t1 = time.Now()

	if logLevel, err = rt.stageResponse(ctx, logLevel, request, response, err); err != nil {
		rev = NewReportEvent(logLevel, proxy.StageResponse, err)
		rev.SetRequest(request).SetResponse(response)
		return rev.Response(), err
	}

	// No need to check logLevel here: if we reached that point, logLevel is All.
	rev = rt.stageBodies(ctx, logLevel, request, response, err)
	return rev.Response(), rev.Err()
}
