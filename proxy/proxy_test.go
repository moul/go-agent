package proxy_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/proxy"
)

type ConcurrentBuilder struct {
	strings.Builder
	sync.RWMutex
}

func (b *ConcurrentBuilder) String() string {
	b.RLock()
	defer b.RUnlock()
	return b.Builder.String()
}

func (b *ConcurrentBuilder) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Builder.Write(p)
}

func makeTestSender() (*proxy.Sender, *ConcurrentBuilder) {
	sb := &ConcurrentBuilder{}
	z := zerolog.New(sb)
	sender := proxy.NewSender(config.DefaultReportOutstanding, config.DefaultReportEndpoint,
		agent.Version, agent.ExampleWellFormedInvalidKey, `test`, nil, &z)
	return sender, sb
}

func TestMustParseURL(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		want      *url.URL
		wantPanic bool
	}{
		{`happy`, `https://account:secret@example.com/path`, &url.URL{
			Scheme: `https`,
			User:   url.UserPassword(`account`, `secret`),
			Host:   `example.com`,
			Path:   `/path`,
		}, false},
		{`sad`, "http:\n//example.com", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil != tt.wantPanic {
					t.Errorf(`panic: got %t expected %t`, r != nil, tt.wantPanic)
				}
			}()
			if got := proxy.MustParseURL(tt.rawURL); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustParseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSender(t *testing.T) {
	s := proxy.NewSender(proxy.AckBacklog, `http://localhost`, agent.Version,
		agent.ExampleWellFormedInvalidKey, `test`, nil, nil)
	if s == nil {
		t.Fatalf(`NewSender returned nil`)
	}
}

func TestSender_Send(t *testing.T) {
	log := proxy.ReportLog{LogLevel: `foo`}
	sender, builder := makeTestSender()

	// Normal operation

	sender.Send(log)
	if len(builder.String()) != 0 {
		t.Errorf(`unexpected warning: %s`, builder.String())
	}

	select {
	case sentLog := <-sender.FanIn:
		if !reflect.DeepEqual(sentLog, log) {
			t.Error(`received log is not the one that was sent`)
		}
	default:
		t.Error(`log was not sent`)
	}

	// Draining phase

	close(sender.Draining)
	sender.Send(log)
	if len(builder.String()) == 0 {
		t.Errorf(`expected warning but got none`)
	}

	select {
	case <-sender.FanIn:
		t.Error(`log was sent but should have been dropped`)
	default:
	}
}

func TestSender_StartHappyAck(t *testing.T) {
	sender, builder := makeTestSender()
	sender.InFlight = 1
	go sender.Start()
	sender.Acks <- 1
	// Ensure at least one loop iteration.
	time.Sleep(1 * proxy.QuietLoopPause)
	l := len(strings.Split(builder.String(), "\n"))
	if l != 2 {
		t.Errorf("sender did not log ack. Expected %d lines, got %d", 1, l-1)
	}
	sender.Stop()
	// Ensure at least two loop iterations.
	time.Sleep(2 * proxy.QuietLoopPause)
	s := strings.Split(builder.String(), "\n")
	l = len(s)
	if l != 3 {
		t.Errorf("sender did not log finishing. Expected %d lines, got %d: %v", 2, l-1, s)
	}
}

func TestNewReportLossReport(t *testing.T) {
	tests := []struct {
		name        string
		n           uint
		wantCode    string
		wantMessage string
	}{
		{`no loss`, 0, `0`, `0 report logs were lost`},
		{`losses`, 10, `10`, `10 report logs were lost`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proxy.NewReportLossReport(tt.n)
			if actualCode := fmt.Sprint(got.ErrorCode); actualCode != tt.wantCode {
				t.Errorf("NewReportLossRport() code = %s, want %s", actualCode, tt.wantCode)
			}
			if actualMsg := got.ErrorFullMessage; actualMsg != tt.wantMessage {
				t.Errorf("NewReportLossReport() = %v, want %v", got, tt.wantMessage)
			}
		})
	}
}

func TestSender_WriteLog(t *testing.T) {
	tests := []struct {
		name        string
		logEndpoint string
		method      string
		wantErr     bool
	}{
		{`happy`, ``, ``, false},
		{`sad bad endpoint`, `_://`, ``, true},
		{`sad mute endpoint`, `http://example.invalid`, ``, true},
		{`sad rejected response`, ``, http.MethodConnect, true},
	}
	// Set up test server.
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := ioutil.ReadAll(request.Body)
		lr := proxy.LogReport{}
		err := json.Unmarshal(body, &lr)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		method := lr.Logs[0].Method
		if method != `` && method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, cb := makeTestSender()
			s.Client = *ts.Client()
			if tt.logEndpoint != `` {
				s.LogEndpoint = tt.logEndpoint
			} else {
				s.LogEndpoint = ts.URL
			}

			s.WriteLog(proxy.ReportLog{Method: tt.method})
			log := struct {
				Level    string
				ReportId int
				Status   string
			}{}
			err := json.Unmarshal([]byte(cb.String()), &log)
			if err != nil {
				t.Fatalf(`unexpected log format: %v`, err)
			}
			wantLevel := `trace`
			if tt.wantErr {
				wantLevel = `warn`
			}
			if log.Level != wantLevel {
				t.Fatalf(`unexpected log level during report: %s want %s`, log.Level, wantLevel)
			}
		})
	}
}
