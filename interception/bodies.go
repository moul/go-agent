package interception

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

// BodyReadCloser wraps a io.ReadCloser to give access to the first peekSize
// bytes without interfering with the normal behaviour
type BodyReadCloser struct {
	peekSize   int
	peekBuffer []byte
	peekError  error
	pos        int
	readCloser io.ReadCloser
}

// NewBodyReadCloser constructs a BodyReadCloser wrapper
func NewBodyReadCloser(readCloser io.ReadCloser, peekSize int) *BodyReadCloser {
	return &BodyReadCloser{
		readCloser: readCloser,
		pos:        0,
		peekSize:   peekSize,
	}
}

// Read gives the usual io.Reader behaviour
func (r *BodyReadCloser) Read(p []byte) (int, error) {
	if r.pos < r.peekSize && (r.peekBuffer == nil || r.pos < len(r.peekBuffer)) {
		r.ensurePeekBuffer()
		peekN := copy(p, r.peekBuffer)
		if r.peekError != nil {
			// Use the normal read
			r.pos = r.peekSize + 1

			return peekN, r.peekError
		}

		n, err := r.readCloser.Read(p[peekN:])
		r.pos += peekN + n
		return peekN + n, err
	}

	return r.readCloser.Read(p)
}

// Peek returns the result of reading the first peek bytes block
func (r *BodyReadCloser) Peek() ([]byte, error) {
	r.ensurePeekBuffer()
	return r.peekBuffer, r.peekError
}

func (r *BodyReadCloser) ensurePeekBuffer() {
	if r.peekBuffer != nil {
		return
	}

	buffer := make([]byte, r.peekSize)
	n, err := io.ReadFull(r.readCloser, buffer)
	r.peekBuffer = buffer[:n]

	r.peekError = err
	if err == io.ErrUnexpectedEOF {
		r.peekError = io.EOF
	}
}

// Close closes the underlying io.ReadCloser
func (r *BodyReadCloser) Close() error {
	return r.readCloser.Close()
}

// BodyParsingProvider is an events.Listener provider returning listeners
// performing data collection, hashing, and sanitization on request/reponse
// bodies.
type BodyParsingProvider struct{}

// Listeners implements events.ListenerProvider.
func (p BodyParsingProvider) Listeners(e events.Event) (l []events.Listener) {
	switch e.Topic() {
	case TopicBodies:
		l = []events.Listener{
			p.RequestBodyParser,
			p.ResponseBodyParser,
		}
	}
	return
}

// ParseFormData parses form data
func ParseFormData(reader io.Reader) (map[string][]string, error) {
	request := &http.Request{Method: `POST`, Body: ioutil.NopCloser(reader), Header: make(http.Header)}
	request.Header.Set(proxy.ContentTypeHeader, proxy.ContentTypeSimpleForm)

	err := request.ParseForm()
	return request.Form, err
}
