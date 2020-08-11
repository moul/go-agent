package interception

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// RequestBodyParser is an events.Listener performing eager resBody loading on API
// requests, to perform sanitization and bandwidth reduction.
func (BodyParsingProvider) RequestBodyParser(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}
	request := e.Request()
	body := request.Body
	if body == nil {
		be.RequestBody = ``
		return nil
	}
	bodyReader, ok := body.(*BodyReadCloser)
	if !ok {
		be.RequestBody = BodyUndecodable
		return errors.New(`expected Body to be a BodyReadCloser`)
	}

	bodyBytes, err := bodyReader.Peek()
	if err != nil && err != io.EOF {
		be.RequestBody = BodyUndecodable
		return fmt.Errorf("error peeking body: %w", err)
	}
	reader := bytes.NewReader(bodyBytes)
	if reader.Len() == 0 {
		be.RequestBody = ``
		return nil
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
		err := d.Decode(&be.RequestBody)
		if err != nil {
			be.RequestBody = BodyUndecodable
			return fmt.Errorf("decoding JSON request reqBody: %w", err)
		}
		be.RequestSha = ToSha(be.RequestBody)
	case FormContentType.MatchString(ct):
		be.RequestBody, err = ParseFormData(reader)
		if err != nil {
			be.RequestBody = BodyUndecodable
			return fmt.Errorf("decoding HTML form request reqBody: %w", err)
		}
		be.RequestSha = `N/A`
		return nil
	default:
		be.RequestBody = string(bodyBytes)
	}

	return nil
}
