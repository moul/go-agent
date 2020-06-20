package interception

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// RequestBodyLoader is an events.Listener performing eager body loading on API
// requests, to ensure data collection by the agent.
func (p BodyParsingProvider) RequestBodyLoader(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}
	request := be.Request()
	request.Body, be.Error = p.loadBody(request.Body)
	be.SetRequest(request)
	be.readTimestamp = time.Now()
	return nil
}

// RequestBodyParser is an events.Listener performing eager body loading on API
// requests, to perform sanitization and bandwidth reduction.
func (BodyParsingProvider) RequestBodyParser(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}
	request := e.Request()
	body := request.Body
	if body == nil {
		return nil
	}
	reader, ok := body.(*MeasuredReader)
	if !ok {
		return fmt.Errorf(`topic Body to have a Len(), got %T`, body)
	}
	if reader.Len() >= MaximumBodySize {
		be.RequestBody = BodyTooLong
		return nil
	}
	ct := request.Header.Get(proxy.ContentTypeHeader)
	if !ParsableContentType.MatchString(ct) {
		be.RequestBody = BodyIsBinary
		return nil
	}
	switch {
	case JSONContentType.MatchString(ct):
		d := json.NewDecoder(reader)
		err := d.Decode(be.RequestBody)
		if err != nil {
			be.RequestBody = BodyUndecodable
			return fmt.Errorf("decoding JSON request body: %w", err)
		}
		_, _ = reader.Seek(0, io.SeekStart)
		be.RequestSha = ToSha(reader)
		_, _ = reader.Seek(0, io.SeekStart)
	case FormContentType.MatchString(ct):
		err := request.ParseForm()
		if err != nil {
			be.RequestBody = BodyUndecodable
			return fmt.Errorf("decoding HTML form request body: %w", err)
		}
		be.RequestBody = request.Form
		be.RequestSha = ToSha(request.Form)
		return nil
	}

	return nil
}
