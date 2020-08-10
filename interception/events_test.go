package interception

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/bearer/go-agent/filters"

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

func TestDCRProvider_onActiveTopics(t *testing.T) {
	req, _ := http.NewRequest(`POST`, `http://test.example.com`, nil)
	baseEvent := events.EventBase{}
	baseEvent.SetRequest(req)

	restricted := Restricted
	all := All
	falseVal := false
	rule1 := &DataCollectionRule{
		Filter:   &filters.HTTPMethodFilter{StringMatcher: filters.NewStringMatcher(`POST`, false)},
		LogLevel: &restricted,
		IsActive: &falseVal,
	}
	rule2 := &DataCollectionRule{
		// Non-matching filter
		Filter: &filters.HTTPMethodFilter{StringMatcher: filters.NewStringMatcher(`GET`, false)},
	}
	rule3 := &DataCollectionRule{Filter: nil, LogLevel: &all}

	tests := []struct {
		name                   string
		dcrs                   []*DataCollectionRule
		e                      events.Event
		expectedTriggeredRules []*DataCollectionRule
		expectedLogLevel       LogLevel
		expectedIsActive       bool
		wantErr                bool
	}{
		{`defaults`, []*DataCollectionRule{}, &apiEvent{EventBase: baseEvent},
			[]*DataCollectionRule{}, Detected, true, false},
		{`matching rules`, []*DataCollectionRule{rule1, rule2, rule3}, &apiEvent{EventBase: baseEvent},
			[]*DataCollectionRule{rule1, rule3}, All, false, false},
		{`sad bad event`, []*DataCollectionRule{}, &events.EventBase{}, nil, Detected, false, true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DCRProvider{
				DCRs: tt.dcrs,
			}
			if err := p.onActiveTopics(ctx, tt.e); (err != nil) != tt.wantErr {
				t.Errorf("onActiveTopics() error = %v, wantErr %v", err, tt.wantErr)
			}

			if ae, ok := tt.e.(APIEvent); ok && !tt.wantErr {
				if !reflect.DeepEqual(ae.TriggeredDataCollectionRules(), tt.expectedTriggeredRules) {
					t.Errorf(
						"TriggeredDataCollectionRules do not match. Expected:\n%#v\n\nActual:\n%#v\n",
						tt.expectedTriggeredRules,
						ae.TriggeredDataCollectionRules(),
					)
				}

				if ae.Config().LogLevel != tt.expectedLogLevel {
					t.Errorf(
						"LogLevel does not match. Expected: %#v Actual: %#v\n",
						tt.expectedLogLevel,
						ae.Config().LogLevel,
					)
				}

				if ae.Config().IsActive != tt.expectedIsActive {
					t.Errorf(
						"IsActive does not match. Expected: %#v Actual: %#v\n",
						tt.expectedIsActive,
						ae.Config().IsActive,
					)
				}
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
			re := NewReportEvent(proxy.StageUndefined, nil)
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

func Test_apiEvent_SetConfig(t *testing.T) {
	config := &APIEventConfig{
		LogLevel: All,
		IsActive: false,
	}

	ae := &apiEvent{}
	ae.SetConfig(config)

	actual := ae.Config()
	if actual != config {
		t.Errorf("SetLogLevel() = %v, want %v", actual, config)
	}
}

func Test_apiEvent_SetTriggeredDataCollectionRules(t *testing.T) {
	expected := []*DataCollectionRule{
		&DataCollectionRule{
			Signature: "test",
		},
	}

	ae := &apiEvent{}
	ae.SetTriggeredDataCollectionRules(expected)
	actual := ae.TriggeredDataCollectionRules()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("SetTriggeredDataCollectionRules() = %v, want %v", actual, expected)
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
		{`happy`, &stubSender, NewReportEvent(proxy.StageConnect, nil), false},
		{`sad bad event`, &stubSender, &events.EventBase{}, true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProxyProvider{
				Sender: tt.Sender,
			}
			if re, ok := tt.e.(*ReportEvent); ok {
				res := http.Response{Body: testReader(``)}
				re.SetResponse(&res)
				re.SetConfig(&APIEventConfig{})
			}
			if err := p.onReport(ctx, tt.e); (err != nil) != tt.wantErr {
				t.Errorf("onReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
