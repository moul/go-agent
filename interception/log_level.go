package interception

//go:generate stringer -type=LogLevel -output log_level_names.go

import "strings"

// LogLevelKey is the key in contexts where the current LogLevel may be found.
const LogLevelKey = `BearerLogLevel`

// LogLevel represents the log levels defined by the Bearer platform.
type LogLevel int

const (
	// Detected specifies that the agent should send log level and connection data only.
	Detected LogLevel = iota - 1

	// Restricted specifies that the agent should send common data and all available stage data,
	// excluding request and response headers and bodies.
	Restricted

	// All specifies that the agent should send common data and all available stage data.
	All
)

// LogLevelFromInt converts any int to a valid LogLevel, adjusting non-valid values to
// the default LogLevel: Restricted.
func LogLevelFromInt(n int) LogLevel {
	l := LogLevel(n)
	if l < Detected || l > All {
		l = Restricted
	}
	return l
}

// LogLevelFromString builds a LogLevel from a string, defaulting to Restricted
// for all invalid strings.
func LogLevelFromString(s string) LogLevel {
	switch {
	case strings.EqualFold(s, Detected.String()):
		return Detected
	case strings.EqualFold(s, All.String()):
		return All
	default:
		return Restricted
	}
}

