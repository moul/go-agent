package interception

//go:generate stringer -type=LogLevel -output log_level_names.go

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/bearer/go-agent/proxy"
)

// ContextKey is the type to use for any key added to the context by this package.
type ContextKey string

const (
	// BodyTooLong is the replacement string for bodies beyond MaximumBodySize.
	BodyTooLong = `(omitted due to size)`

	// BodyIsBinary is the replacement string for unparseable bodies.
	BodyIsBinary = `(not showing binary data)`

	// BodyUndecodable is the replacement string for bodies which were expected to be parsable but failed decoding.
	BodyUndecodable = `(could not decode data)`

	// MaximumBodySize is the largest resBody size to store whole.
	MaximumBodySize = 1 << 20
)

// ParsableContentType is a regexp defining the types to attempt to parse.
var ParsableContentType = regexp.MustCompile(`(?i)(json|text|xml|x-www-form-urlencoded)`)

// StringContentType is a regexp defininig the types to return as plain strings.
var StringContentType = regexp.MustCompile(`(?i)(text|xml)`)

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
	triggeredRules := PrepareTriggeredRulesForReport(re.TriggeredDataCollectionRules())
	u := request.URL

	err := re.Error
	var errorCode, errorMessage string
	if err != nil {
		errorCode = err.Error()
		errorMessage = errorCode
	}

	rl.StartedAt = int(re.T0.UnixNano() / 1E6)
	rl.EndedAt = int(re.T1.UnixNano() / 1E6)
	rl.Stage = string(re.Stage)
	rl.ActiveDataCollectionRules = &triggeredRules
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
	request, response := re.Request(), re.Response()

	rl.RequestHeaders = request.Header
	rl.RequestBodyPayloadSHA = re.RequestSha
	rl.RequestBody = serializeBody(rl.RequestHeaders, re.RequestBody)
	if re.RequestBody != nil && rl.RequestBody == `` {
		rl.RequestBody = `(no body)`
	}

	if response == nil {
		return
	}

	rl.ResponseHeaders = response.Header
	rl.ResponseBodyPayloadSHA = re.ResponseSha
	rl.ResponseBody = serializeBody(rl.ResponseHeaders, re.ResponseBody)
	if re.ResponseBody != nil && rl.ResponseBody == `` {
		rl.ResponseBody = `(no body)`
	}
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

func serializeBody(headers http.Header, body interface{}) string {
	if body == nil {
		return ``
	}

	ct := headers.Get(proxy.ContentTypeHeader)
	// It will be a string if it's a text body or if we failed to parse it
	if s, ok := body.(string); ok {
		return s
	} else if ct == proxy.ContentTypeSimpleForm {
		if values, ok := body.(map[string][]string); ok {
			return url.Values(values).Encode()
		}
	} else { // Everything else must be JSON
		if json, err := json.Marshal(body); err == nil {
			return string(json)
		}
	}

	return BodyUndecodable
}
