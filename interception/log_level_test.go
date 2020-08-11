package interception

import (
	"io"
	"net/http"
	"testing"

	"github.com/bearer/go-agent/proxy"
)

func TestLogLevelFromString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want LogLevel
	}{
		{`happy detected`, `dEtECTed`, Detected},
		{`happy restricted`, `rESTRICTEd`, Restricted},
		{`happy all`, `all`, All},
		{`sad`, `sicksadworld`, Restricted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LogLevelFromString(tt.s); got != tt.want {
				t.Errorf("LogLevelFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogLevel_addRestrictedInfo(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType string
	}{
		{`happy`, nil, proxy.End},
		{`sad error`, io.EOF, proxy.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewReportEvent(proxy.StageBodies, tt.err)
			req, _ := http.NewRequest(http.MethodGet, defaultTestURL, nil)
			e.SetRequest(req)
			res := http.Response{
				Body: testReader(``),
			}
			e.SetResponse(&res)
			e.ResponseBody = ``
			rl := proxy.ReportLog{Type: proxy.End}
			level := All
			level.addRestrictedInfo(&rl, e)

			if rl.Type != tt.wantType {
				t.Fatalf(`addRestrictedInfo type: %s, want %s`, rl.Type, tt.wantType)
			}
		})
	}
}

func TestLogLevel_addAllInfo(t *testing.T) {
	jsonHeaders := http.Header{proxy.ContentTypeHeader: {proxy.ContentTypeJSON}}
	formHeaders := http.Header{proxy.ContentTypeHeader: {proxy.ContentTypeSimpleForm}}
	textHeaders := http.Header{proxy.ContentTypeHeader: {`text/plain`}}

	tests := []struct {
		name        string
		reqHeader   http.Header
		resHeader   http.Header
		reqBody     interface{}
		resBody     interface{}
		wantReqBody string
		wantResBody string
	}{
		{`request JSON`, jsonHeaders, nil, map[string]int{`x`: 21}, ``, `{"x":21}`, `(no body)`},
		{`response JSON`, nil, jsonHeaders, ``, []interface{}{`y`, 42}, `(no body)`, `["y",42]`},
		{`invalid JSON`, jsonHeaders, jsonHeaders, `oh`, `no`, `oh`, `no`},
		{`request FORM`, formHeaders, nil, map[string][]string{`x`: {`21`}}, ``, `x=21`, `(no body)`},
		{`response FORM`, nil, formHeaders, ``, map[string][]string{`y`: {`4`, `2`}}, `(no body)`, `y=4&y=2`},
		{`invalid FORM`, formHeaders, formHeaders, `oh`, `no`, `oh`, `no`},
		{`request TEXT`, textHeaders, nil, `hello`, ``, `hello`, `(no body)`},
		{`request TEXT`, nil, textHeaders, ``, `world`, `(no body)`, `world`},
		{`invalid TEXT`, formHeaders, formHeaders, `oh`, `no`, `oh`, `no`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewReportEvent(proxy.StageBodies, nil)
			req, _ := http.NewRequest(http.MethodGet, defaultTestURL, nil)
			req.Header = tt.reqHeader
			e.SetRequest(req)
			e.RequestBody = tt.reqBody
			res := http.Response{
				Header: tt.resHeader,
				Body:   testReader(``),
			}
			e.SetResponse(&res)
			e.ResponseBody = tt.resBody
			rl := proxy.ReportLog{Type: proxy.End}
			level := All
			level.addAllInfo(&rl, e)

			if rl.RequestBody != tt.wantReqBody {
				t.Fatalf(`RequestBody actual: %s, expected: %s`, rl.RequestBody, tt.wantReqBody)
			}
			if rl.ResponseBody != tt.wantResBody {
				t.Fatalf(`ResponseBody actual: %s, expected: %s`, rl.ResponseBody, tt.wantResBody)
			}
		})
	}
}
