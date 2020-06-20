package interception

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

func TestBodyParsingProvider_ResponseBodyLoader(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{`happy`, `hello`, false},
		{`sad`, `goodbye`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := ioutil.NopCloser(strings.NewReader(tt.message))
			eb := events.EventBase{}
			eb.SetResponse(&http.Response{Body: body})
			var be events.Event
			if strings.Contains(`sad`, tt.name) {
				be = &eb
			} else {
				be = &BodiesEvent{apiEvent: apiEvent{EventBase: eb}}
			}
			var p BodyParsingProvider

			if err := p.ResponseBodyLoader(context.Background(), be); (err != nil) != tt.wantErr {
				t.Errorf("RequestBodyLoader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			loadedBody := be.Response().Body
			mb, ok := loadedBody.(*MeasuredReader)
			if !ok {
				t.Fatalf(`ResponseBodyLoader: got %T, topic %T`, loadedBody, mb)
			}

			_ = body.Close()
			sl := make([]byte, len(tt.message))
			n, err := mb.Read(sl)
			if n != len(tt.message) || err != nil || string(sl) != tt.message {
				t.Fatalf(`ResponseBodyLoader, failed re-reading after resBody close: %d, %v, %v`,
					n, err, sl)
			}
		})
	}
}

func TestBodyParsingProvider_ResponseBodyParser(t *testing.T) {
	reader := func(s string) *MeasuredReader {
		r := bytes.NewReader([]byte(s))
		return (*MeasuredReader)(r)
	}
	tests := []struct {
		name    string
		body    io.ReadCloser
		ct      string
		wantErr bool
	}{
		{`sad bad event`, nil, ``, true},
		{`happy nil`, nil, ``, false},
		{`sad basic`, ioutil.NopCloser(strings.NewReader(``)), ``, true},
		{`happy extra long`, reader(strings.Repeat(`a`, MaximumBodySize+1)), ``, false},
		{`happy non-parsable`, reader(`hello`), `application/binary`, false},
		{`sad bad JSON`, reader(`]`), proxy.ContentTypeJSON, true},
		{`happy JSON`, reader(`[]`), proxy.ContentTypeJSON, false},
		// No bad forms on responses: they are ignored.
		{`happy form`, reader(`bearer=api`), proxy.ContentTypeSimpleForm, false},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e events.Event
			if tt.name == `sad bad event` {
				e = &events.EventBase{}
			} else {
				e = &BodiesEvent{}
				res := &http.Response{Body: tt.body, Header: make(http.Header)}
				if tt.ct != `` {
					res.Header.Set(proxy.ContentTypeHeader, tt.ct)
				}
				e.SetResponse(res)
			}
			bo := BodyParsingProvider{}
			if err := bo.ResponseBodyParser(ctx, e); (err != nil) != tt.wantErr {
				t.Errorf("ResponseBodyParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
