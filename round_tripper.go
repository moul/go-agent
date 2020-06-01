package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bearer/go-agent/events"
)

// RoundTripper is the instrumented implementation of http.RoundTripper.
//
// It triggers events for the Connect, Request, and Response stages.
type RoundTripper struct {
	events.Dispatcher
	Config
	underlying http.RoundTripper
}

// stageValidate validates the destination URL under RFC793, RFC1384, RFC1738
// and RFC3986 before entering the standard Bearer multistage API wrapping.
//
// It is hard-coded in the round-tripper to avoid its being disabled.
func (rt *RoundTripper) stageValidate(url *url.URL) (host string, port uint16, scheme string, err error) {
	// XXX As some point, we might want to include a host validation, following RFC1738 Sec. 3.1
	sPort, scheme := url.Port(), url.Scheme

	// RFC3986.
	if !SchemeRegexp.MatchString(scheme) {
		return ``, 0, ``, fmt.Errorf("invalid scheme [%s]", scheme)
	}

	if sPort == `` {
		return ``, 0, ``, fmt.Errorf("ill-formed port specification in Host [%s]", url.Host)
	}

	intPort, err := strconv.Atoi(sPort)
	if err != nil {
		// This might be a case for a panic, since URL.Port() is expected to
		// return an empty string if the port is not numeric.
		return ``, 0, ``, fmt.Errorf("ill-formed port [%s]", sPort)
	}

	// RFC793 sec 3.1 and RFC1340 p.7.
	if port <= 0 || port > 2<<15-1 {
		return ``, 0, ``, fmt.Errorf("invalid port [%d]", intPort)
	}

	return url.Hostname(), uint16(intPort), scheme, nil
}

// stageConnect implements the Bearer Connect stage.
func (rt *RoundTripper) stageConnect(ctx context.Context, hostname string, port uint16, scheme string) error {
	_, err := rt.Dispatch(ctx, &ConnectEvent{
		Host:   hostname,
		Port:   port,
		Scheme: scheme,
	})
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (rt *RoundTripper) stageRequest(request *http.Request) error {
	ctx := request.Context()
	_, err := rt.Dispatch(ctx, &RequestEvent{Request: request})
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}

	return nil
}

func (rt *RoundTripper) stageResponse(ctx context.Context, response *http.Response) error {
	_, err := rt.Dispatch(ctx, &ResponseEvent{Response: response})
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}

	return nil
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error
	ctx := request.Context()

	host, port, scheme, err := rt.stageValidate(request.URL)
	if err != nil {
		return nil, err
	}

	err = rt.stageConnect(ctx, host, port, scheme)
	if err != nil {
		return nil, err
	}

	err = rt.stageRequest(request)
	if err != nil {
		return nil, err
	}

	response, err := rt.underlying.RoundTrip(request)
	if err != nil {
		return response, err
	}

	err = rt.stageResponse(ctx, response)
	return response, err
}
