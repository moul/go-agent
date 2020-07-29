package agent

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// Logger returns a valid zerolog.Logger instance for the agent.
func (a *Agent) Logger() *zerolog.Logger {
	if a.config == nil || a.config.Logger == nil {
		a.SetLogger(os.Stderr)
	}
	return a.config.Logger
}

// SetLogger changes the logger with a specific zerolog.Logger.
//
// If the writer is a zerolog.Writer, it is used as such, otherwise a new
// zerolog.Logger is used to wrap it. If the agent has no current config, an
// empty config will be added to it to carry the logger.
func (a *Agent) SetLogger(w io.Writer) *Agent {
	zl, ok := w.(*zerolog.Logger)
	if !ok {
		l := zerolog.New(w)
		zl = &l
	}
	if a.config == nil {
		a.config = &Config{}
	}
	_ = WithLogger(zl)(a.config)
	return a
}

// LogTrace logs a trace-level debug event with the specified message and fields.
func (a *Agent) LogTrace(msg string, fields map[string]interface{}) {
	a.Logger().Trace().Fields(fields).Msg(msg)
}

// LogWarn logs a warning with the specified message and fields.
func (a *Agent) LogWarn(msg string, fields map[string]interface{}) {
	a.Logger().Warn().Fields(fields).Msg(msg)
}

// LogError logs an error with the specified message and fields.
func (a *Agent) LogError(msg string, fields map[string]interface{}) {
	a.Logger().Error().Fields(fields).Msg(msg)
}
