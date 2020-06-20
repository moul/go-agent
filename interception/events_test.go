package interception

import (
	"net/url"
	"reflect"
	"testing"

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
	type fields struct {
		DCRs []*DataCollectionRule
	}
	type args struct {
		e events.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []events.Listener
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DCRProvider{
				DCRs: tt.fields.DCRs,
			}
			if got := p.Listeners(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Listeners() = %v, want %v", got, tt.want)
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
	type fields struct {
		Sender *proxy.Sender
	}
	type args struct {
		e events.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []events.Listener
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProxyProvider{
				Sender: tt.fields.Sender,
			}
			if got := p.Listeners(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Listeners() = %v, want %v", got, tt.want)
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
	type fields struct {
		apiEvent apiEvent
	}
	tests := []struct {
		name   string
		fields fields
		want   events.Topic
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := RequestEvent{
				apiEvent: tt.fields.apiEvent,
			}
			if got := re.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseEvent_Topic(t *testing.T) {
	type fields struct {
		apiEvent apiEvent
	}
	tests := []struct {
		name   string
		fields fields
		want   events.Topic
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ResponseEvent{
				apiEvent: tt.fields.apiEvent,
			}
			if got := re.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apiEvent_LogLevel(t *testing.T) {
	type fields struct {
		EventBase events.EventBase
		logLevel  LogLevel
	}
	tests := []struct {
		name   string
		fields fields
		want   LogLevel
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := &apiEvent{
				EventBase: tt.fields.EventBase,
				logLevel:  tt.fields.logLevel,
			}
			if got := ae.LogLevel(); got != tt.want {
				t.Errorf("LogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apiEvent_SetLogLevel(t *testing.T) {
	type fields struct {
		EventBase events.EventBase
		logLevel  LogLevel
	}
	type args struct {
		l LogLevel
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIEvent
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := &apiEvent{
				EventBase: tt.fields.EventBase,
				logLevel:  tt.fields.logLevel,
			}
			if got := ae.SetLogLevel(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
