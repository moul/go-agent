package interception_test

import (
	"context"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/interception"
)

const (
	testURL = `https://example.com`
	topic   = `test_topic`
	card    = `fake370057577167325card`
	mail    = `john.doe@example.com`
)

func newSanitizationProvider() *interception.SanitizationProvider {
	keysREs := []*regexp.Regexp{interception.DefaultSensitiveKeys}
	valueREs := []*regexp.Regexp{interception.DefaultSensitiveData}
	p := &interception.SanitizationProvider{keysREs, valueREs}
	return p
}

func TestSanitizationProvider_SanitizeQueryAndPaths(t *testing.T) {
	tests := []struct {
		name              string
		requestURL        string
		resReqURL         string // `` means reuse the request in response.
		expectedReqURL    string
		expectedResReqURL string
		wantErr           bool
	}{
		{`happy default query+path`,
			`https://example.com/card/` + card + `?client_id=secretname&email=john.doe@example.com&foo=bar`,
			``,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dcard?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{`happy custom response request query+path`,
			`https://example.com/card/fake370057577167325amex?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake373058337477712visa?client_id=secretname&email=john.doe@example.com&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Damex?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			`https://example.com/card/fake%5BFILTERED%5Dvisa?client_id=%5BFILTERED%5D&email=%5BFILTERED%5D&foo=bar`,
			false},
		{`sad bad URL`,
			// Invalid IPv6 address and missing scheme separator.
			`http//2020:609:241:98::80:10/authentication/login/`, ``, ``, ``, true},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
		{`happy`, string(interception.TopicReport), 5},
		{`sad`, topic, 0},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		{`sad name`, `authorization`, []string{`Basic Dartmouth`}, ``, nil, []string{interception.Filtered}, []string{interception.Filtered}},
		{`sad single value`,
			`bankkarte`, []string{`fake370057577167325card`}, ``, nil,
			[]string{`fake` + interception.Filtered + `card`},
			[]string{`fake` + interception.Filtered + `card`},
		},
		{`mixed`, `tarjeta-de-credito`,
			[]string{`not a card`, `fake370057577167325card`, `nor that one`}, ``, nil,
			[]string{`not a card`, `fake` + interception.Filtered + `card`, `nor that one`},
			[]string{`not a card`, `fake` + interception.Filtered + `card`, `nor that one`},
		},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		{`sad name`, `authorization`, []string{`Basic Dartmouth`}, []string{interception.Filtered}},
		{`sad single value`, `bankkarte`,
			[]string{`fake370057577167325card`},
			[]string{`fake` + interception.Filtered + `card`},
		},
		{`sad mix`, `tarjeta-de-credito`,
			[]string{`not a card`, `fake370057577167325card`, `nor that one`},
			[]string{`not a card`, `fake` + interception.Filtered + `card`, `nor that one`},
		},
	}
	for _, tt := range tests {
		p := newSanitizationProvider()
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{Header: make(http.Header, 1)}
			for _, v := range tt.Values {
				res.Header.Add(tt.Name, v)
			}

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
	const card = `fake` + interception.Filtered + `card`

	tests := []struct {
		name     string
		x        interface{}
		expected interface{}
		wantErr  bool
	}{
		{`untouched map[string]string`, map[string]string{`foo`: `bar`}, map[string]string{`foo`: `bar`}, false},
		{`filtered key`, map[string]interface{}{`secret`: `bar`}, map[string]interface{}{`secret`: interception.Filtered}, false},
		{`fully filtered map value`, map[string]interface{}{`foo`: mail}, map[string]interface{}{`foo`: interception.Filtered}, false},
		{`partially filtered map value`, map[string]interface{}{`foo`: card}, map[string]interface{}{`foo`: `fake` + interception.Filtered + `card`}, false},
		{`[]string, filtered`, []string{mail}, []string{interception.Filtered}, false},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := interception.NewWalker(tt.x)
			var accu interface{}

			if err := w.Walk(&accu, p.BodySanitizer); (err != nil) != tt.wantErr {
				t.Errorf("sanitize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.x, tt.expected) {
				t.Errorf("sanitize got %v, wanted %v", tt.x, tt.expected)
			}
		})
	}
}

func TestSanitizationProvider_SanitizeRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		body     interface{}
		expected interface{}
		wantErr  bool
	}{
		{`untouched map[string]string`, map[string]string{`foo`: `bar`}, map[string]string{`foo`: `bar`}, false},
		{`filtered key`, map[string]interface{}{`secret`: `bar`}, map[string]interface{}{`secret`: interception.Filtered}, false},
		{`fully filtered map value`, map[string]interface{}{`foo`: mail}, map[string]interface{}{`foo`: interception.Filtered}, false},
		{`partially filtered map value`, map[string]interface{}{`foo`: card}, map[string]interface{}{`foo`: `fake` + interception.Filtered + `card`}, false},
		{`[]string, filtered`, []string{mail}, []string{interception.Filtered}, false},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &interception.ReportEvent{
				BodiesEvent: interception.BodiesEvent{RequestBody: tt.body},
			}
			if err := p.SanitizeRequestBody(context.Background(), e); (err != nil) != tt.wantErr {
				t.Errorf("SanitizeRequestBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			actual, expected := e.RequestBody, tt.expected
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("SanitizeRequestBody got %v expected %v", e.RequestBody, tt.expected)
			}
		})
	}
}

func TestSanitizationProvider_SanitizeResponseBody(t *testing.T) {
	tests := []struct {
		name     string
		body     interface{}
		expected interface{}
		wantErr  bool
	}{
		{`untouched map[string]string`, map[string]string{`foo`: `bar`}, map[string]string{`foo`: `bar`}, false},
		{`filtered key`, map[string]interface{}{`secret`: `bar`}, map[string]interface{}{`secret`: interception.Filtered}, false},
		{`fully filtered map value`, map[string]interface{}{`foo`: mail}, map[string]interface{}{`foo`: interception.Filtered}, false},
		{`partially filtered map value`, map[string]interface{}{`foo`: card}, map[string]interface{}{`foo`: `fake` + interception.Filtered + `card`}, false},
		{`[]string, filtered`, []string{mail}, []string{interception.Filtered}, false},
	}
	p := newSanitizationProvider()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &interception.ReportEvent{
				BodiesEvent: interception.BodiesEvent{ResponseBody: tt.body},
			}
			if err := p.SanitizeResponseBody(context.Background(), e); (err != nil) != tt.wantErr {
				t.Errorf("SanitizeResponseBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			actual, expected := e.ResponseBody, tt.expected
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("SanitizeResponseBody got %v expected %v", e.ResponseBody, tt.expected)
			}
		})
	}
}
