package interception

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/bearer/go-agent/events"
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
				t.Fatalf(`ResponseBodyLoader: got %T, expected %T`, loadedBody, mb)
			}

			_ = body.Close()
			sl := make([]byte, len(tt.message))
			n, err := mb.Read(sl)
			if n != len(tt.message) || err != nil || string(sl) != tt.message {
				t.Fatalf(`ResponseBodyLoader, failed re-reading after body close: %d, %v, %v`,
					n, err, sl)
			}
		})
	}
}
