package interception

import (
	"context"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/events"
)

const (
	testURL = `https://example.com`
	topic   = `test_topic`
)

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
		{`happy default query+path`, fields{nil, nil},
			`https://example.com/card/fake370057577167325card?client_id=secretname&email=john.doe@example.com&foo=bar`,
			``,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{`happy custom response request query+path`, fields{nil, nil},
			`https://example.com/card/fake370057577167325amex?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake373058337477712visa?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Damex?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dvisa?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{`sad bad URL`, fields{nil, nil},
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
				t.Errorf(`sanitizeQueryAndPaths error = %v, wantErr %v`, err, tt.wantErr)
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
		{`happy`, string(TopicReport), 5},
		{`sad`, topic, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := SanitizationProvider{}
			e := events.NewEvent(tt.topic)
			if got := len(p.Listeners(e)); got != tt.wantLen {
				t.Errorf(`Listeners() = %v, want %v`, got, tt.wantLen)
			}
		})
	}
}

func TestSanitizationProvider_SanitizeRequestHeaders(t *testing.T) {
	tests := []struct {
		name                 string
		reqName              string
		reqValues            []string
		resReqName           string
		resReqValues         []string
		expectedReqValues    []string
		expectedResReqValues []string
	}{
		{`happy`, `foo`, []string{`bar`}, `baz`, []string{`qux`}, []string{`bar`}, []string{`qux`}},
		{`happy same`, `foo`, []string{`bar`}, ``, nil, []string{`bar`}, []string{`bar`}},
		{`sad name`, `authorization`, []string{`Basic Dartmouth`}, ``, nil, []string{Filtered}, []string{Filtered}},
		{`sad single value`,
			`bankkarte`, []string{`fake370057577167325card`}, ``, nil,
			[]string{`fake` + Filtered + `card`},
			[]string{`fake` + Filtered + `card`},
		},
		{`mixed`, `tarjeta-de-credito`,
			[]string{`not a card`, `fake370057577167325card`, `nor that one`}, ``, nil,
			[]string{`not a card`, `fake` + Filtered + `card`, `nor that one`},
			[]string{`not a card`, `fake` + Filtered + `card`, `nor that one`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keysREs := []*regexp.Regexp{DefaultSensitiveKeys}
			valueREs := []*regexp.Regexp{DefaultSensitiveData}

			req, _ := http.NewRequest(``, testURL, nil)
			for _, v := range tt.reqValues {
				req.Header.Add(tt.reqName, v)
			}

			res := &http.Response{Request: req}
			resReq := req
			if tt.resReqName != `` || tt.resReqValues != nil {
				resReq, _ = http.NewRequest(``, testURL, nil)
				for _, v := range tt.resReqValues {
					resReq.Header.Add(tt.resReqName, v)
				}
			}
			res.Request = resReq

			p := SanitizationProvider{keysREs, valueREs}
			e := events.NewEvent(topic).SetRequest(req).SetResponse(res)
			err := p.SanitizeRequestHeaders(context.Background(), e)
			if err != nil {
				t.Fatalf(`sanitizeRequestHeaders unexpected error = %v`, err)
				return
			}
			actualReqValues := e.Request().Header.Values(tt.reqName)
			if !reflect.DeepEqual(actualReqValues, tt.expectedReqValues) {
				t.Errorf(`sanitizeRequestHeaders for %s expected %v, got %v`, tt.reqName, tt.expectedReqValues, actualReqValues)
			}

		})
	}
}

func TestSanitizationProvider_SanitizeResponseHeaders(t *testing.T) {
	tests := []struct {
		name           string
		Name           string
		Values         []string
		expectedValues []string
	}{
		{`happy`, `foo`, []string{`bar`}, []string{`bar`}},
		{`sad name`, `authorization`, []string{`Basic Dartmouth`}, []string{Filtered}},
		{`sad single value`, `bankkarte`,
			[]string{`fake370057577167325card`},
			[]string{`fake` + Filtered + `card`},
		},
		{`sad mix`, `tarjeta-de-credito`,
			[]string{`not a card`, `fake370057577167325card`, `nor that one`},
			[]string{`not a card`, `fake` + Filtered + `card`, `nor that one`},
		},
	}
	for _, tt := range tests {
		keysREs := []*regexp.Regexp{DefaultSensitiveKeys}
		valueREs := []*regexp.Regexp{DefaultSensitiveData}
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{Header: make(http.Header, 1)}
			for _, v := range tt.Values {
				res.Header.Add(tt.Name, v)
			}

			p := SanitizationProvider{keysREs, valueREs}
			e := events.NewEvent(topic).SetResponse(res)
			err := p.SanitizeResponseHeaders(context.Background(), e)
			if err != nil {
				t.Fatalf(`sanitizeResponseHeaders unexpected error = %v`, err)
				return
			}
			actualValues := e.Response().Header.Values(tt.Name)
			if !reflect.DeepEqual(actualValues, tt.expectedValues) {
				t.Errorf(`sanitizeResponseHeaders for %s expected %v, got %v`, tt.Name, tt.expectedValues, actualValues)
			}

		})
	}
}

func TestSanitizationProvider_sanitize(t *testing.T) {
	const email = `john.doe@example.com`
	const card = `fake` + Filtered + `card`

	tests := []struct {
		name     string
		x        interface{}
		expected interface{}
		wantErr  bool
	}{
		//{`untouched map[string]string`, map[string]string{`foo`: `bar`}, map[string]string{`foo`: `bar`}, false},
		//{`filtered key`, map[string]interface{}{`secret`: `bar`}, map[string]interface{}{`secret`: Filtered}, false},
		{`fully filtered map value`, map[string]interface{}{`foo`: email}, map[string]interface{}{`foo`: Filtered}, false},
		//{`partially filtered map value`, map[string]interface{}{`foo`: card}, map[string]interface{}{`foo`: `fake` + Filtered + `card`}, false},
		//{`[]string, filtered`, []string{email}, []string{Filtered}, false},
	}
	for _, tt := range tests {
		keysREs := []*regexp.Regexp{DefaultSensitiveKeys}
		valueREs := []*regexp.Regexp{DefaultSensitiveData}
		t.Run(tt.name, func(t *testing.T) {
			p := SanitizationProvider{keysREs, valueREs}
			if err := p.sanitize(tt.x, nil); (err != nil) != tt.wantErr {
				t.Errorf("sanitize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.x, tt.expected) {
				t.Errorf("sanitize got %v, wanted %v", tt.x, tt.expected)
			}
		})
	}
}
