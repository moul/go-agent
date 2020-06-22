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

func TestLogLevel_addAllInfo(t *testing.T) {
	tests := []struct {
		name      string
		reqHeader http.Header
		resHeader http.Header
		reqBody   interface{}
		resBody   interface{}
		err       error
		wantType  string
	}{
		{`sad error`, nil, nil, ``, ``, io.EOF, proxy.Error},
		{`sad request JSON`, nil, nil, 2i, ``, nil, proxy.Error},
		{`happy request JSON`, nil, nil, 21, ``, nil, proxy.End},
		{`sad response JSON`, nil, nil, ``, 2i, nil, proxy.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewReportEvent(All, proxy.StageBodies, tt.err)
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

			if rl.Type != tt.wantType {
				t.Fatalf(`addAllInfo type: %s, want %s`, rl.Type, tt.wantType)
			}
		})
	}
}
