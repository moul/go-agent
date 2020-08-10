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

func TestBodyParsingProvider_ResponseBodyParser(t *testing.T) {
	reader := func(s string) io.ReadCloser {
		return NewBodyReadCloser(ioutil.NopCloser(strings.NewReader(s)), MaximumBodySize+1)
	}
	tests := []struct {
		name    string
		body    io.ReadCloser
		ct      string
		wantErr bool
	}{
		{`sad bad event`, nil, ``, true},
		{`happy nil`, nil, ``, false},
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
