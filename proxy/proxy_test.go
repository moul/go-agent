package proxy_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/proxy"
)

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

func TestSender_Stop(t *testing.T) {
	tests := []struct {
		name     string
		Done     chan bool
		FanIn    chan proxy.ReportLog
		wantDone bool
	}{
		{`happy`, make(chan bool, 1), make(chan proxy.ReportLog, proxy.FanInBacklog), true},
		{`sad`, func() chan bool { ch := make(chan bool, 2); ch <- false; return ch }(),
			make(chan proxy.ReportLog, proxy.FanInBacklog), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &proxy.Sender{
				Done:  tt.Done,
				FanIn: tt.FanIn,
			}
			s.Stop()
			done := <-s.Done
			if done != tt.wantDone {
				t.Errorf(`done: got %t, expected %t`, done, tt.wantDone)
			}
			if s.FanIn != nil {
				t.Errorf(`fanIn: got %d channel, want nil`, cap(s.FanIn))
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
	tests := []struct {
		name  string
		FanIn chan proxy.ReportLog
		log   proxy.ReportLog
	}{
		{`happy`, make(chan proxy.ReportLog, 1), proxy.ReportLog{LogLevel: `foo`}},
		{`sad`, nil, proxy.ReportLog{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			z := zerolog.New(sb)
			s := &proxy.Sender{
				FanIn:  tt.FanIn,
				Logger: &z,
			}
			s.Send(tt.log)
			if tt.FanIn == nil {
				if len(sb.String()) == 0 {
					t.Error(`send: expected warning, got nothing`)
				}
				return
			}
			log := <-s.FanIn
			if log.LogLevel != tt.log.LogLevel {
				t.Errorf(`reportLogLevel: got %s, expected %s`, log.LogLevel, tt.log.LogLevel)
			}
		})
	}
}
