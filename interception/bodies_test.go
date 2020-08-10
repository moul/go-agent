package interception

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestBodyReadCloser(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{`small`, `small`},
		{`large`, `0123456789`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brc := NewBodyReadCloser(ioutil.NopCloser(strings.NewReader(tt.data)), 11)
			length := len(tt.data)

			buffer := make([]byte, length)
			n, err := brc.Read(buffer)
			if n != length || err != io.EOF {
				t.Errorf(`Read() returned length: %d, error %s`, n, err)
			}
			actual := string(buffer)
			if actual != tt.data {
				t.Errorf(`Read() expected: %v, actual: %v`, tt.data, actual)
			}

			buffer, _ = brc.Peek()
			actual = string(buffer)
			peekLength := 10
			if length < peekLength {
				peekLength = length
			}
			if actual != tt.data[:peekLength] {
				t.Errorf(`Peek() expected: %v, actual: %v`, tt.data[:10], actual)
			}
		})
	}
}

func TestParseFormData(t *testing.T) {
	tests := []struct {
		name     string
		data     io.Reader
		expected map[string][]string
		wantErr  bool
	}{
		{`happy`, strings.NewReader(`x=1&y=2&y=3`), map[string][]string{
			`x`: []string{`1`},
			`y`: []string{`2`, `3`},
		}, false},
		{`sad`, strings.NewReader(`%INVALID`), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseFormData(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error %v but error %v", tt.wantErr, err != nil)
			}

			if err == nil && !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("expected: %#v, actual: %#v", tt.expected, actual)
			}
		})
	}
}
