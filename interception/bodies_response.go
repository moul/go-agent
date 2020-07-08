package interception

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// ResponseBodyLoader is an events.Listener performing eager resBody loading on API
// responses, to ensure data collection by the agent.
func (p BodyParsingProvider) ResponseBodyLoader(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}
	response := be.Response()
	response.Body, be.Error = p.loadBody(response.Body)
	be.SetResponse(response)
	be.readTimestamp = time.Now()
	return nil
}

// ResponseBodyParser is an events.Listener performing eager resBody loading on API
// responses, to perform sanitization and bandwidth reduction.
func (p BodyParsingProvider) ResponseBodyParser(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}
	response := e.Response()
	var body io.Reader = response.Body
	if body == nil {
		be.ResponseBody = ``
		return nil
	}
	reader, ok := body.(*MeasuredReader)
	if !ok {
		return fmt.Errorf(`topic Body to have a Len(), got %T`, body)
	}
	if reader.Len() == 0 {
		be.ResponseBody = ``
		return nil
	}
	if reader.Len() >= MaximumBodySize {
		be.ResponseBody = BodyTooLong
		return nil
	}
	ct := response.Header.Get(proxy.ContentTypeHeader)
	if !ParsableContentType.MatchString(ct) {
		be.ResponseBody = BodyIsBinary
		return nil
	}
	switch {
	case JSONContentType.MatchString(ct):
		d := json.NewDecoder(reader)
		err := d.Decode(&be.ResponseBody)
		if err != nil {
			be.ResponseBody = BodyUndecodable
			return fmt.Errorf("decoding JSON response resBody: %w", err)
		}
		_, _ = reader.Seek(0, io.SeekStart)
		be.ResponseSha = ToSha(reader)
		_, _ = reader.Seek(0, io.SeekStart)
	case FormContentType.MatchString(ct):
		// Forms are not supported on http.Response so build a placeholder http.Request
		// to hold the data and apply standard parsing.
		pos, _ := reader.Seek(0, io.SeekCurrent)
		request := &http.Request{Body: reader, Header: make(http.Header)}
		request.Header.Set(proxy.ContentTypeHeader, proxy.ContentTypeSimpleForm)
		_, _ = reader.Seek(pos, io.SeekStart)

		err := request.ParseForm()
		if err != nil {
			be.ResponseBody = BodyUndecodable
			return fmt.Errorf("decoding HTML form response body: %w", err)
		}
		be.ResponseBody = request.Form
		be.ResponseSha = ToSha(request.Form)
		return nil
	default:
		body, err := ioutil.ReadAll(reader)
		if err != nil {
			be.ResponseBody = BodyUndecodable
			return nil
		}
		be.ResponseBody = body
	}

	return nil
}
