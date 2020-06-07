package interception

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/filters"
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
		portMap := map[string]string{
			"http":   "80",
			"https":  "443",
			"socks5": "1080",
		}
		sPort, ok = portMap[ce.Scheme]
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
	e := &ResponseEvent{
		error: err,
	}
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

// RoundTrip implements the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var logLevel LogLevel
	var err error
	ctx := request.Context()

	if logLevel, err = rt.stageConnect(ctx, request.URL); err != nil {
		e := &ReportEvent{
			apiEvent: apiEvent{
				EventBase: *(&events.EventBase{}).SetRequest(request),
				logLevel:  logLevel,
			},
			Stage:    filters.StageConnect,
			Error:    err,
		}
		_, _ = rt.Dispatch(ctx, e)
		return nil, err
	}

	if logLevel, err = rt.stageRequest(logLevel, request); err != nil {
		e := &ReportEvent{
			apiEvent: apiEvent{
				EventBase: *(&events.EventBase{}).SetRequest(request),
				logLevel:  logLevel,
			},
			Stage:    filters.StageRequest,
			Error:    err,
		}
		_, _ = rt.Dispatch(ctx, e)
		return nil, err
	}

	response, err := rt.Underlying.RoundTrip(request)
	_, err = rt.stageResponse(ctx, logLevel, request, response, err)
	e := &ReportEvent{
		apiEvent: apiEvent{
			EventBase: *(&events.EventBase{}).SetRequest(request).SetResponse(response),
			logLevel:  logLevel,
		},
		Stage:    filters.StageResponse,
		Error:    err,
	}
	_, _ = rt.Dispatch(ctx, e)
	return response, err
}
