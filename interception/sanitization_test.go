package interception

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/events"
)

const topic = "test_topic"

func TestSanitizationProvider_SanitizeQueryAndPaths(t *testing.T) {
	type fields struct {
		SensitiveKeys    []string
		SensitiveRegexps []string
	}
	tests := []struct {
		name              string
		fields            fields
		requestURL        string
		resReqURL         string // `` means reuse the request in response.
		expectedReqURL    string
		expectedResReqURL string
		wantErr           bool
	}{
		{"happy default query+path", fields{nil, nil},
			`https://example.com/card/fake370057577167325card?client_id=secretname&email=john.doe@example.com&foo=bar`,
			``,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{"happy custom response request query+path", fields{nil, nil},
			`https://example.com/card/fake370057577167325amex?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake373058337477712visa?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Damex?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dvisa?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{"sad bad URL", fields{nil, nil},
			// Invalid IPv6 address and missing scheme separator.
			`http//2020:609:241:98::80:10/authentication/login/`, ``, ``, ``, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keysREs := []*regexp.Regexp{DefaultSensitiveKeys}
			for _, re := range tt.fields.SensitiveKeys {
				keysREs = append(keysREs, regexp.MustCompile(re))
			}
			valueREs := []*regexp.Regexp{DefaultSensitiveData}
			for _, re := range tt.fields.SensitiveRegexps {
				valueREs = append(valueREs, regexp.MustCompile(re))
			}

			req, err := http.NewRequest(``, tt.requestURL, nil)
			if err != nil {
				t.Fatalf(`unexpected error building request: %v`, err)
			}
			res := &http.Response{Request: req}
			resReq := req
			if tt.resReqURL != `` {
				resReq, err = http.NewRequest(``, tt.resReqURL, nil)
				if err != nil {
					t.Fatalf(`unexpected error building response request: %v`, err)
				}
			}
			res.Request = resReq

			p := SanitizationProvider{keysREs, valueREs}
			e := events.NewEvent(topic).SetRequest(req).SetResponse(res)
			err = p.SanitizeQueryAndPaths(context.Background(), e)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizeQueryAndPaths error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			actualReqURL := e.Request().URL.String()
			if actualReqURL != tt.expectedReqURL {
				t.Errorf(`sanitizeQueryAndPaths URL: got %s, expected %s`, actualReqURL, tt.expectedReqURL)
			}

			actualResReqURL := e.Response().Request.URL.String()
			if actualResReqURL != tt.expectedResReqURL {
				t.Errorf(`sanitizeQueryAndPaths response URL: got %s, expected %s`, actualResReqURL, tt.expectedResReqURL)
			}
		})
	}
}

func TestSanitizationProvider_Listeners(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		wantLen int
	}{
		{`happy`, string(TopicReport), 3},
		{`sad`, topic, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := SanitizationProvider{}
			e := events.NewEvent(tt.topic)
			if got := len(p.Listeners(e)); got != tt.wantLen {
				t.Errorf("Listeners() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
