package interception

//go:generate stringer -type=LogLevel -output log_level_names.go

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/bearer/go-agent/proxy"
)

// ContextKey is the type to use for any key added to the context by this package.
type ContextKey string

const (
	// BodyTooLong is the replacement string for bodies beyond MaximumBodySize.
	BodyTooLong  = `(omitted due to size)`

	// BodyIsBinary is the replacement string for unparseable bodies.
	BodyIsBinary = `(not showing binary data)`

	// LogLevelKey is the key in contexts where the current LogLevel may be found.
	LogLevelKey ContextKey = `BearerLogLevel`

	// MaximumBodySize is the largest body size to store whole.
	MaximumBodySize = 1 << 20
)

// ParsableContentType is a regexp defining the types to attempt to parse.
var ParsableContentType = regexp.MustCompile(`(?i)(json|text|xml|x-www-form-urlencoded)`)

// JSONContentType is a regexp defining the content types to handle as JSON.
var JSONContentType = regexp.MustCompile(`(?i)json`)

// FormContentType is a regexp definint the content types to handle as traditional web forms.
var FormContentType = regexp.MustCompile(`(?i)x-www-form-urlencoded`)

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

// addDetectedInfo adds to the report the info reported at the "DETECTED" log level.
func (ll *LogLevel) addDetectedInfo(rl *proxy.ReportLog, re *ReportEvent) {
	request := re.Request()
	u := request.URL

	// Cf. Go runtime: src/net/http/transport.go
	PortMap := map[string]uint16{
		"http":   80,
		"https":  443,
		"socks5": 1080,
	}
	port := PortMap[u.Scheme] // Having 0 in case of errors is expected.

	// The Agent spec specifies errors are not part of the minimal Detected level report.
	rl.Hostname = u.Hostname()
	rl.LogLevel = strings.ToUpper(ll.String())
	rl.Port = port
	rl.Protocol = u.Scheme
}

// addRestrictedInfo adds to the report the info reported at the "RESTRICTED" log level.
func (ll *LogLevel) addRestrictedInfo(rl *proxy.ReportLog, re *ReportEvent) {
	request := re.Request()
	response := re.Response()
	u := request.URL

	err := re.error
	var errorCode, errorMessage string
	if err != nil {
		errorCode = err.Error()
		errorMessage = errorCode
	}

	rl.StartedAt = int(re.T0.UnixNano() / 1E6)
	rl.EndedAt = int(re.T1.UnixNano() / 1E6)
	rl.Stage = string(re.Stage)
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

// addAllInfo adds to the report the info reported at the "ALL" log level.
func (ll *LogLevel) addAllInfo(rl *proxy.ReportLog, re *ReportEvent) {
	request, response, err := re.Request(), re.Response(), re.error

	rl.RequestHeaders = request.Header
	rl.ResponseHeaders = response.Header

	var body, sha string
	if request.Body != nil {
		body, sha, err = ll.parseBody(request.Header.Get(proxy.ContentTypeHeader), request.Body)
	}
	if err != nil {
		rl.Type = proxy.Error
		rl.RequestBody = ``
	} else {
		rl.RequestBody = hex.EncodeToString([]byte(body))
	}
	reqSha := sha256.Sum256([]byte(rl.RequestBody))
	rl.RequestBodyPayloadSHA = hex.EncodeToString(reqSha[:])

	body, sha, err = ll.parseBody(response.Header.Get(proxy.ContentTypeHeader), response.Body)
	if err != nil {
		rl.Type = proxy.Error
		rl.ResponseBody = ``
	} else {
		defer response.Body.Close()
		rl.ResponseBody = hex.EncodeToString([]byte(sha))
	}
	resSha := sha256.Sum256([]byte(rl.ResponseBody))
	rl.ResponseBodyPayloadSHA = hex.EncodeToString(resSha[:])
}

func (ll *LogLevel) parseBody(ct string, in io.ReadCloser) (out, sha string, err error) {
	defer in.Close()
	return
}

// Prepare extract the ReportLog information from the API call, depending on the LogLevel.
func (ll *LogLevel) Prepare(re *ReportEvent) proxy.ReportLog {
	if request := re.Request(); request == nil {
		request, _ = http.NewRequest(``, ``, nil)
		re.SetRequest(request)
	}

	rl := proxy.ReportLog{}
	ll.addDetectedInfo(&rl, re)

	if *ll >= Restricted {
		ll.addRestrictedInfo(&rl, re)
	}

	if *ll >= All {
		ll.addAllInfo(&rl, re)
	}
	return rl
}
