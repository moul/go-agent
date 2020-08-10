package interception

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

func testReader(s string) io.ReadCloser {
	return NewBodyReadCloser(ioutil.NopCloser(strings.NewReader(s)), MaximumBodySize+1)
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
