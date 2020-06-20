package interception

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/bearer/go-agent/events"
)

func TestRFCListener(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{`happy normal`, &url.URL{Scheme: `http`}, false},
		{`sad bad event`, nil, true},
		{`sad no URL`, `not an URL`, true},
		{`sad bad scheme`, &url.URL{Scheme: `_`}, true},
		{`sad no port not HTTP`, &url.URL{Scheme: `ftp`, Host: `localhost`}, true},
		{`sad bad port for int`, &url.URL{Scheme: `ftp`, Host: `localhost:12345678901234567890`}, true},
		{`sad bad port for TCP`, &url.URL{Scheme: `ftp`, Host: `localhost:91140`}, true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e events.Event
			if tt.name == `sad bad event` {
				e = &events.EventBase{}
			} else {
				e = &ConnectEvent{}
				_ = e.SetData(tt.data)
			}
			if err := RFCListener(ctx, e); (err != nil) != tt.wantErr {
				t.Errorf("RFCListener() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type emptyReader struct{}

func (emptyReader) Read([]byte) (n int, err error) {
	return 0, nil
}

func (e emptyReader) Close() error {
	return nil
}

type testRoundTripper struct{}

func (t testRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	res := http.Response{}
	return &res, nil
}

func TestRoundTripper_RoundTrip(t *testing.T) {
	const defaultURL = `http://localhost:80`
	tests := []struct {
		name    string
		liveContext bool
		body    io.ReadCloser
		want    *http.Response
		wantErr bool
	}{
		{`happy empty`, true, emptyReader{}, &http.Response{}, false},
		{`sad context`, false, emptyReader{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if !tt.liveContext {
				canceled, fn := context.WithDeadline(ctx, time.Now().Add(-1*time.Second))
				ctx = canceled
				fn()
			}
			rt := &RoundTripper{
				Dispatcher: events.NewDispatcher(),
				Underlying: testRoundTripper{},
			}
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, defaultURL, tt.body)
			got, err := rt.RoundTrip(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoundTrip() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
