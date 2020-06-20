package interception

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/bearer/go-agent/events"
)

// MeasuredReader is a bytes.Reader, but with a Len method on the value in
// addition to the Len() value on the pointer, to allow type assertions on bodies,
// and with a built-in Nop Close.
type MeasuredReader bytes.Reader

// Clone returns a new MeasuredReader based on the same data, but with an independent
// underlying []byte, ensuring interception does not interfere with HTTP processing.
func (b *MeasuredReader) Clone() (*MeasuredReader, error) {
	pos, _ := b.Seek(0, io.SeekCurrent)
	_, _ = b.Seek(0, io.SeekStart)
	buffer := &bytes.Buffer{}
	// In this specific case, io.Copy cannot fail.
	_, _ = io.Copy(buffer, b)
	// io.Copy moved within the reader, return to its original position.
	_, _ = b.Seek(pos, io.SeekStart)
	r := (*MeasuredReader)(bytes.NewReader(buffer.Bytes()))
	return r, nil
}

// Read is part of io.Reader and io.ReadCloser interfaces, and defers to bytes.Reader
// for actual implementation.
func (b *MeasuredReader) Read(p []byte) (n int, err error) {
	buffer := (*bytes.Reader)(b)
	return buffer.Read(p)
}

// Close is part of io.ReadCloser, and does nothing since MeasurerReader is just
// a wrapper around a []byte.
func (*MeasuredReader) Close() error {
	return nil
}

// Len is the same as bytes.Buffer Len method, but with a value received.
func (b MeasuredReader) Len() int {
	buffer := bytes.Reader(b)
	return buffer.Len()
}

// Seek implements io.Seeker, deferring to bytes.Reader for implementation.
func (b *MeasuredReader) Seek(offset int64, whence int) (int64, error) {
	buffer := (*bytes.Reader)(b)
	return buffer.Seek(offset, whence)
}

// BodyParsingProvider is an events.Listener provider returning listeners
// performing data collection, hashing, and sanitization on request/reponse
// bodies.
type BodyParsingProvider struct{}

// Listeners implements events.ListenerProvider.
func (p BodyParsingProvider) Listeners(e events.Event) (l []events.Listener) {
	switch e.Topic() {
	case TopicRequest:
		l = []events.Listener{
			p.RequestBodyLoader,
		}
	case TopicBodies:
		l = []events.Listener{
			p.RequestBodyLoader,
			p.ResponseBodyLoader,
			p.RequestBodyParser,
			p.ResponseBodyParser,
		}
	}
	return
}

// Force body reading.
func (p BodyParsingProvider) loadBody(body io.ReadCloser) (io.ReadCloser, error) {
	if body == nil {
		return body, nil
	}
	defer body.Close()
	mr, ok := body.(*MeasuredReader)
	// Take shortcut if available.
	if ok {
		return mr.Clone()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	buf := (*MeasuredReader)(bytes.NewReader(b))
	return buf, nil
}
