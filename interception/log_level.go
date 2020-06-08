package interception

//go:generate stringer -type=LogLevel -output log_level_names.go

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/bearer/go-agent/proxy"
)

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

// Prepare extract the ReportLog information from the API call, depending on the LogLevel.
func (ll *LogLevel) Prepare(re *ReportEvent) proxy.ReportLog {
	request := re.Request()
	response := re.Response()
	err := re.Error

	if re.Request() == nil {
		request, _ = http.NewRequest(``, ``, nil)
	}
	u := request.URL
	// Cf. Go runtime: src/net/http/transport.go
	PortMap := map[string]uint16{
		"http":   80,
		"https":  443,
		"socks5": 1080,
	}
	port := PortMap[u.Scheme] // Having 0 in case of errors is expected.

	var errorCode, errorMessage string
	if err != nil {
		errorCode = err.Error()
		errorMessage = errorCode
	}

	rl := proxy.ReportLog{
		LogLevel: strings.ToUpper(ll.String()),
		Port:     port,
		Protocol: u.Scheme,
		Hostname: u.Hostname(),
	}

	if *ll >= Restricted {
		rl.StartedAt = int(re.T0.UnixNano() / 1E6)
		rl.EndedAt = int(re.T1.UnixNano() / 1E6)
		rl.Stage = re.Stage
		rl.ActiveDataCollectionRules = []string{}
		rl.Path = u.Path
		rl.Method = request.Method
		rl.URL = u.String()
		if response != nil {
			rl.StatusCode = response.StatusCode
		}
		rl.ErrorCode = errorCode
		rl.ErrorFullMessage = errorMessage

		if err != nil {
			rl.Type = proxy.Error
		} else {
			rl.Type = proxy.End
		}

	}
	if *ll >= All {
		rl.RequestHeaders = request.Header
		rl.ResponseHeaders = response.Header
		rl.RequestBody = ``
		rl.ResponseBody = ``

		reqSha := sha256.Sum256([]byte(rl.RequestBody))
		rl.RequestBodyPayloadSHA = hex.EncodeToString(reqSha[:])

		resSha := sha256.Sum256([]byte(rl.ResponseBody))
		rl.ResponseBodyPayloadSHA = hex.EncodeToString(resSha[:])
	}
	return rl
}
