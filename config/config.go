package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	// TraceLoggingName is the environment variable used to enable tracing in logs,
	// when the application uses the default logger instead of injecting its own.
	TraceLoggingName = `BEARER_TRACE`

	// DefaultRuntimeEnvironmentType is the default environment type.
	DefaultRuntimeEnvironmentType = "development" // "default"

	// DefaultConfigEndpoint is the default configuration endpoint for Bearer.
	DefaultConfigEndpoint = "https://config.bearer.sh/config"

	// DefaultFetchInterval is the default rate at which the Fetcher will
	// asynchronously fetch configuration refreshes from Bearer.
	DefaultFetchInterval = 5 * time.Second

	// DefaultReportEndpoint is the default reporting endpoint for Bearer.
	DefaultReportEndpoint = "https://agent.bearer.sh/logs"

	// DefaultReportOutstanding it the default maximum number of pending data
	// collection writes in flight at any given time. When that limit is
	// exceeded, records are no longer sent to Bearer to avoid saturating the
	// client.
	DefaultReportOutstanding = 1000
)

// TraceLogging is set in init() and enabled the default logger for Trace level.
var TraceLogging = false

// IsSecretKeyWellFormed verifies whether the secret key matches the expected format.
func IsSecretKeyWellFormed(secretKey string) bool {
	// SecretKeyRegex is the format of Bearer secret keys.
	// It is used to verify the shape of submitted secret keys, before they are
	// sent over to Bearer for value validation.
	SecretKeyRegex := regexp.MustCompile(`^app_[[:xdigit:]]{50}$`)

	return SecretKeyRegex.MatchString(secretKey)
}

// DefaultLogger builds a logger to os.Stderr that won't log Trace information.
func DefaultLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stderr)
	logger = logger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		if TraceLogging || level != zerolog.TraceLevel {
			return
		}
		e.Discard()
	}))
	return &logger
}

func init() {
	t := strings.ToUpper(strings.Trim(os.Getenv(TraceLoggingName), " \r\n\t"))
	if t == `TRUE` || t == `T` || t == `YES` || t == `Y` || t == `ON` {
		TraceLogging = true
		return
	}
	if n, err := strconv.Atoi(t); err != nil && n != 0 {
		TraceLogging = true
		return
	}
}
