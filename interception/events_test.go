package interception

import (
	"context"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/events"
	"github.com/bearer/go-agent/proxy"
)

func TestConnectEvent_Request(t *testing.T) {
	type fields struct {
		Host   string
		Port   uint16
		Scheme string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{`happy`, fields{`localhost`, 80, `http`}, `http://localhost:80`, false},
		{`bad port`, fields{`localhost`, 0, `http`}, `http://localhost:0`, false},
		{`protocol-relative`, fields{`localhost`, 80, ``}, `//localhost:80`, false},
		{`sad bad host`, fields{`foo bar`, 443, `https`}, ``, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ConnectEvent{
				Host:   tt.fields.Host,
				Port:   tt.fields.Port,
				Scheme: tt.fields.Scheme,
			}
			request := re.Request()
			if (request != nil) == tt.wantErr {
				t.Fatalf(`Unexpected error`)
			}
			if tt.wantErr {
				return
			}
			if got := request.URL.String(); got != tt.want {
				t.Errorf("Request() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectEvent_Topic(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  events.Topic
	}{
		{`happy`, string(TopicConnect), TopicConnect},
		{`sad`, `whatever`, TopicConnect},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ConnectEvent{}
			re.SetTopic(tt.topic)
			if got := re.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDCRProvider_Listeners(t *testing.T) {
	tests := []struct {
		name    string
		topic   events.Topic
		wantLen int
	}{
		{`happy connect`, TopicConnect, 1},
		{`happy request`, TopicRequest, 1},
		{`happy response`, TopicResponse, 1},
		{`happy bodies`, TopicBodies, 1},
		{`sad report`, TopicReport, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DCRProvider{}
			if got := p.Listeners((&events.EventBase{}).SetTopic(string(tt.topic))); len(got) != tt.wantLen {
				t.Errorf("Listeners() = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestNewConnectEvent(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{`happy`, `http://localhost:80`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewConnectEvent(proxy.MustParseURL(tt.url))
			got, ok := e.Data().(*url.URL)
			if !ok {
				t.Errorf("NewConnectEvent() data type %T, want %T", e.Data(), got)
			}
			if got.String() != tt.url {
				t.Errorf("NewConnectEvent() data %s, want %s", got.String(), tt.url)
			}
		})
	}
}

func TestProxyProvider_Listeners(t *testing.T) {
	tests := []struct {
		name    string
		topic   events.Topic
		wantLen int
	}{
		{`happy`, TopicReport, 1},
		{`sad`, TopicConnect, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProxyProvider{}
			if got := p.Listeners((&events.EventBase{}).SetTopic(string(tt.topic))); len(got) != tt.wantLen {
				t.Errorf("Listeners() = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestReportEvent_Topic(t *testing.T) {
	tests := []struct {
		name  string
		topic string
	}{
		{`happy`, string(TopicReport)},
		{`sad`, `whatever`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := NewReportEvent(Restricted, proxy.StageUndefined, nil)
			re.SetTopic(tt.topic)
			if got := re.Topic(); got != TopicReport {
				t.Errorf("Topic() = %v, want %v", got, TopicReport)
			}
		})
	}
}

func TestRequestEvent_Topic(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  events.Topic
	}{
		{`happy`, string(TopicRequest), TopicRequest},
		{`sad`, `whatever`, TopicRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := RequestEvent{}
			re.SetTopic(tt.topic)
			if got := re.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseEvent_Topic(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  events.Topic
	}{
		{`happy`, string(TopicResponse), TopicResponse},
		{`sad`, `whatever`, TopicResponse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ResponseEvent{}
			re.SetTopic(tt.topic)
			if got := re.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apiEvent_SetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		want     LogLevel
	}{
		{`happy`, Restricted, Restricted},
		{`sad`, -2, Restricted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := &apiEvent{}
			ae.SetLogLevel(tt.logLevel)
			actual := ae.LogLevel()
			if actual != tt.want {
				t.Errorf("SetLogLevel() = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestProxyProvider_onReport(t *testing.T) {
	stubLogger := zerolog.New(ioutil.Discard)
	stubSender := proxy.Sender{
		FanIn:  nil,
		Logger: &stubLogger,
	}
	tests := []struct {
		name    string
		Sender  *proxy.Sender
		e       events.Event
		wantErr bool
	}{
		{`happy`, &stubSender, NewReportEvent(Restricted, proxy.StageConnect, nil), false},
		{`sad bad event`, &stubSender, &events.EventBase{}, true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProxyProvider{
				Sender: tt.Sender,
			}
			if err := p.onReport(ctx, tt.e); (err != nil) != tt.wantErr {
				t.Errorf("onReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
