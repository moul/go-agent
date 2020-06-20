package interception

import (
	"bytes"
	"io"
	"testing"
)

func TestMeasuredReader_Clone(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{`happy`, `hello`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := (*MeasuredReader)(bytes.NewReader([]byte(tt.message)))
			clone, err := reader.Clone()
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Seek on original testReader to ensure clone is not affected.
			_, _ = reader.Seek(int64(reader.Len() / 2), io.SeekStart)

			sl := make([]byte, len(tt.message))
			n, err := clone.Read(sl)
			if err != nil {
				t.Fatalf(`after seek, failed reading from clone: %v`, err)
			}
			if n != len(tt.message) {
				t.Fatalf(`after seek, failed reading original length from clone: clone %d, topic %d`, n, len(tt.message))
			}
			actual := string(sl)
			if actual != tt.message {
				t.Fatalf(`after seek, clone = %s, topic %s`, actual, tt.message)
			}

			// Close original testReader to ensure clone is not affected.
			_, _ = clone.Seek(0, io.SeekStart)
			_ = reader.Close()
			sl = make([]byte, len(tt.message))
			n, err = clone.Read(sl)
			if err != nil {
				t.Fatalf(`after close, failed reading from clone: %v`, err)
			}
			if n != len(tt.message) {
				t.Fatalf(`after close, failed reading original length from clone: clone %d, topic %d`, n, len(tt.message))
			}
			actual = string(sl)
			if actual != tt.message {
				t.Fatalf(`after close, clone = %s, topic %s`, actual, tt.message)
			}
		})
	}
}
