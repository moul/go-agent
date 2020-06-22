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

func testReader(s string) *MeasuredReader {
	r := bytes.NewReader([]byte(s))
	return (*MeasuredReader)(r)
}

func TestBodyParsingProvider_RequestBodyLoader(t *testing.T) {
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
			eb.SetRequest(&http.Request{Body: body})
			var be events.Event
			if strings.Contains(`sad`, tt.name) {
				be = &eb
			} else {
				be = &BodiesEvent{apiEvent: apiEvent{EventBase: eb}}
			}
			var p BodyParsingProvider

			if err := p.RequestBodyLoader(context.Background(), be); (err != nil) != tt.wantErr {
				t.Errorf("RequestBodyLoader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			loadedBody := be.Request().Body
			mb, ok := loadedBody.(*MeasuredReader)
			if !ok {
				t.Fatalf(`RequestBodyLoader: got %T, expected %T`, loadedBody, mb)
			}

			_ = body.Close()
			sl := make([]byte, len(tt.message))
			n, err := mb.Read(sl)
			if n != len(tt.message) || err != nil || string(sl) != tt.message {
				t.Fatalf(`RequestBodyLoader, failed re-reading after resBody close: %d, %v, %v`,
					n, err, sl)
			}
		})
	}
}

func TestBodyParsingProvider_RequestBodyParser(t *testing.T) {
	tests := []struct {
		name    string
		body    io.ReadCloser
		ct      string
		wantErr bool
	}{
		{`sad bad event`, nil, ``, true},
		{`happy nil`, nil, ``, false},
		{`sad basic`, ioutil.NopCloser(strings.NewReader(``)), ``, true},
		{`happy extra long`, testReader(strings.Repeat(`a`, MaximumBodySize+1)), ``, false},
		{`happy non-parsable`, testReader(`hello`), `application/binary`, false},
		{`sad bad JSON`, testReader(`]`), proxy.ContentTypeJSON, true},
		{`happy JSON`, testReader(`[]`), proxy.ContentTypeJSON, false},
		{`sad bad form`, testReader(`%=1`), proxy.ContentTypeSimpleForm, true},
		{`happy form`, testReader(`bearer=api`), proxy.ContentTypeSimpleForm, false},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e events.Event
			if tt.name == `sad bad event` {
				e = &events.EventBase{}
			} else {
				e = &BodiesEvent{}
				req, _ := http.NewRequest(http.MethodPost, defaultTestURL, tt.body)
				if tt.ct != `` {
					req.Header.Set(proxy.ContentTypeHeader, tt.ct)
				}
				e.SetRequest(req)
			}
			bo := BodyParsingProvider{}
			if err := bo.RequestBodyParser(ctx, e); (err != nil) != tt.wantErr {
				t.Errorf("RequestBodyParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
