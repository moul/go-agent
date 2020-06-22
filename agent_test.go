package agent

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/interception"
	"github.com/bearer/go-agent/proxy"
)

// Perform decoration tests without going to the network.
func TestNewAgent(t *testing.T) {
	const expected = `test handler`

	tests := []struct {
		name               string
		scheme, host, port string
		wantErr            bool
	}{
		{`happy`, `https`, `example.com`, `443`, false},
		{`bad scheme`, `bea@rer`, `example.com`, `1023`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAgent(ExampleWellFormedInvalidKey)
			if a == nil && !tt.wantErr {
				t.Fatal("got unexpected nil agent")
			}

			// Set up test server.
			ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte(expected))
			}))
			defer ts.Close()

			c := ts.Client()
			c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			a.DecorateClientTransports(c)

			// Don't use NewRequest or url.Parse, as they validate URLs,
			// catching errors before we do, and we still want to intercept
			// users manufacturing invalid Requests on their own.
			var u *url.URL
			if regexp.MustCompile(`^happy`).MatchString(tt.name) {
				u, _ = url.Parse(ts.URL)
			} else {
				u = &url.URL{
					Scheme: tt.scheme,
					Host:   strings.Trim(strings.Join([]string{tt.host, tt.port}, `:`), `:`),
				}
			}
			r := &http.Request{URL: u}
			res, err := c.Do(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			defer res.Body.Close()
			if actual, err := ioutil.ReadAll(res.Body); err != nil || string(actual) != expected {
				t.Errorf("Got incorrect response: %s instead of %s", actual, expected)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{"well-formed invalid key", ExampleWellFormedInvalidKey, true},
		{"ill-formed key", "foo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Init(tt.secretKey)
			if err := got(); (err != nil) != tt.wantErr {
				t.Errorf("Init()() = %v", err)
			}
		})
	}
}

func TestAgent_Provider(t *testing.T) {
	tests := []struct {
		name  string
		topic events.Topic
		want  int
	}{
		{`happy`, interception.TopicConnect, 1},
		{`sad`, interception.TopicReport, 0},
	}
	var a Agent
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := (&events.EventBase{}).SetTopic(string(tt.topic))
			if got := a.Provider(e); len(got) != tt.want {
				t.Errorf("Provider() len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func Test_close(t *testing.T) {
	sb := &strings.Builder{}
	z := zerolog.New(sb)
	a := Agent{config: &Config{Logger: &z}, Sender: &proxy.Sender{}}
	closer := close(&a)
	if closer == nil {
		t.Fatalf(`non-callable close result`)
	}
	err := closer()
	if err != nil {
		t.Fatalf("closer returned error: %v", err)
	}
	logLines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	if len(logLines) != 1 {
		t.Fatalf("closer returned more than one event: %d", len(logLines))
	}
}
