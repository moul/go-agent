package agent

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// Logger returns a valid zerolog.Logger instance for the agent.
func (a *Agent) Logger() *zerolog.Logger {
	if a.logger == nil {
		a.SetLogger(os.Stderr)
	}
	return a.logger
}

// SetLogger initializes the writer with a specific writer.
//
// If the writer is a zerolog.Writer, it is used as such, otherwise a new
// zerolog.Logger is used to wrap it.
func (a *Agent) SetLogger(w io.Writer) *Agent {
	zl, ok := w.(*zerolog.Logger)
	if !ok {
		l := zerolog.New(w)
		zl = &l
	}
	a.logger = zl
	return a
}

// Warn logs a warning with the specified message and fields.
func (a *Agent) Warn(msg string, fields map[string]interface{}) {
	a.Logger().Warn().Fields(fields).Msg(msg)
}

// Error logs an error with the specified message and fields.
func (a *Agent) Error(msg string, fields map[string]interface{}) {
	a.Logger().Error().Fields(fields).Msg(msg)
}

