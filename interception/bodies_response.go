package interception

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// ResponseBodyParser is an events.Listener performing eager resBody loading on API
// responses, to perform sanitization and bandwidth reduction.
func (p BodyParsingProvider) ResponseBodyParser(_ context.Context, e events.Event) error {
	be, ok := e.(*BodiesEvent)
	if !ok {
		return fmt.Errorf(`topic BodiesEvent, got %T`, e)
	}

	response := e.Response()
	body := response.Body
	if body == nil {
		be.ResponseBody = ``
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
		be.ResponseSha = ToSha(be.ResponseBody)
	case FormContentType.MatchString(ct):
		be.ResponseBody, err = ParseFormData(reader)
		if err != nil {
			be.ResponseBody = BodyUndecodable
			return fmt.Errorf("decoding HTML form response body: %w", err)
		}
		be.ResponseSha = `N/A`
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
