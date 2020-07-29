package agent_test

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent"
)

type eventArgs struct {
	msg    string
	fields map[string]interface{}
}

var eventTests = []struct {
	name string
	args eventArgs
}{
	{"typical", eventArgs{"a message", map[string]interface{}{"k": "v"}}},
	{"no fields", eventArgs{"a message", nil}},
}

func TestAgent_Error(t *testing.T) {
	const ErrorLevel = "error"
	for _, tt := range eventTests {
		t.Run(tt.name, func(t *testing.T) {
			a := &agent.Agent{}
			w := &strings.Builder{}
			a.SetLogger(w)
			a.LogError(tt.args.msg, tt.args.fields)
			raw := []byte(w.String())
			actual := map[string]string{}
			err := json.Unmarshal(raw, &actual)
			if err != nil {
				t.Errorf("could not decode log message: %v", err)
			}
			level, ok := actual["level"]
			if !ok {
				t.Errorf("level not found in log")
			}
			if level != ErrorLevel {
				t.Errorf("Incorrect error level: %s", level)
			}
			if v, ok := tt.args.fields["k"]; ok {
				if actual["k"] != v {
					t.Errorf("incorrect value for context k: expected v, got %s", actual["k"])
				}
			}
		})
	}
}

func TestAgent_Warn(t *testing.T) {
	const ErrorLevel = "warn"
	for _, tt := range eventTests {
		t.Run(tt.name, func(t *testing.T) {
			a := &agent.Agent{}
			w := &strings.Builder{}
			a.SetLogger(w)
			a.LogWarn(tt.args.msg, tt.args.fields)
			raw := []byte(w.String())
			actual := map[string]string{}
			err := json.Unmarshal(raw, &actual)
			if err != nil {
				t.Errorf("could not decode log message: %v", err)
			}
			level, ok := actual["level"]
			if !ok {
				t.Errorf("level not found in log")
			}
			if level != ErrorLevel {
				t.Errorf("Incorrect error level: %s", level)
			}
			if v, ok := tt.args.fields["k"]; ok {
				if actual["k"] != v {
					t.Errorf("incorrect value for context k: expected v, got %s", actual["k"])
				}
			}
		})
	}
}

func TestAgent_Logger(t *testing.T) {
	eventTests := []struct {
		name   string
		writer io.Writer
	}{
		{"nil", nil},
		{"other", &strings.Builder{}},
	}
	for _, tt := range eventTests {
		t.Run(tt.name, func(t *testing.T) {
			var a agent.Agent
			if tt.writer != nil {
				a.SetLogger(tt.writer)
			}
			if a.Logger() == nil {
				t.Error("got nil from Logger()")
			}
		})
	}
}

func TestAgent_SetLogger(t *testing.T) {
	eventTests := []struct {
		name     string
		writer   io.Writer
		wantSame bool
	}{
		{"nil", nil, false},
		{"zerolog/logger", func() *zerolog.Logger { l := zerolog.Nop(); return &l }(), true},
		{"other", &strings.Builder{}, false},
	}
	for _, tt := range eventTests {
		t.Run(tt.name, func(t *testing.T) {
			a := agent.Agent{}
			if tt.writer != nil {
				a.SetLogger(tt.writer)
			}
			if tt.wantSame {
				if a.Logger() != tt.writer.(*zerolog.Logger) {
					t.Errorf("expected same logger but got different one")
				}
			}
			// wantSame == false is verified by the type system: nothing to test.
		})
	}
}
