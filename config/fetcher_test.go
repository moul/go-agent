package config

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/filters"
	"github.com/bearer/go-agent/interception"
	"github.com/bearer/go-agent/proxy"
)

func TestDescription_String(t *testing.T) {
	dcr := interception.DataCollectionRuleDescription{
		FilterHash: "hash",
		Params: struct {
			AggregationFilterHash string
			Buid                  string
			IsErrorTriggerfilter  bool
			TypeName              string
		}{`af hash`, `buid`, false, ``},
		Config: interception.DynamicConfigDescription{},
	}
	yes := filters.FilterDescription{
		StageType: proxy.StageUndefined,
		TypeName:  `yes`,
	}
	d := Description{
		DataCollectionRules: []interception.DataCollectionRuleDescription{dcr},
		Filters:             map[string]filters.FilterDescription{`yes`: yes},
	}
	got := strings.Split(d.String(), "\n")
	if len(got) != 5 {
		t.Fatalf(`description has %d lines, expected %d`, len(got), 5)
	}
	row, expectedHeading := got[0], `Data Collection Rules`
	if row != expectedHeading {
		t.Fatalf("String() heading = %v, want %v", row, expectedHeading)
	}
	row, expectedDCR := got[1], strings.TrimRight(dcr.String(), "\n")
	if row != expectedDCR {
		t.Fatalf("String() DCR row = %v, want %v", row, expectedDCR)
	}
	row, expectedHeading = got[2], `Filters`
	if row != expectedHeading {
		t.Fatalf("String() filters heading = %v, want %v", row, expectedHeading)
	}
	row, expectedFilter := got[3], `yes: `+strings.TrimRight(yes.String(), "\n")
	if row != expectedFilter {
		t.Fatalf("String() filter row = %v, want %v", row, expectedFilter)
	}
}

func TestDescription_filterDescriptions(t *testing.T) {
	type fields struct {
		DataCollectionRules []interception.DataCollectionRuleDescription
		Filters             map[string]filters.FilterDescription
		Rules               []struct {
			FilterHash   string
			ID           int
			Remediations []interface{}
			RuleType     string
		}
		Error map[string]string
	}
	tests := []struct {
		name     string
		fields   fields
		expected []string
		wantErr  bool
	}{
		{`happy not`, fields{Filters: map[string]filters.FilterDescription{
			`two`: {TypeName: filters.YesInternalFilter.Name()},
			`one`: {TypeName: filters.NotFilterType.Name(), ChildHash: `two`},
		}}, []string{`one`, `two`}, false},
		{`happy set`, fields{Filters: map[string]filters.FilterDescription{
			`two`:   {TypeName: filters.YesInternalFilter.Name()},
			`three`: {TypeName: filters.YesInternalFilter.Name()},
			`one`: {
				FilterSetDescription: filters.FilterSetDescription{
					ChildHashes: []string{`two`, `three`},
				},
				TypeName: filters.FilterSetFilterType.Name(),
			},
		}}, []string{`one`, `three`, `two`}, false},
		{`sad undefined not hash`, fields{
			Filters: map[string]filters.FilterDescription{
				`one`: {ChildHash: `two`, TypeName: filters.NotFilterType.Name()}},
		}, []string{}, true},
		{`sad undefined set hash`, fields{
			Filters: map[string]filters.FilterDescription{
				`one`: {FilterSetDescription: filters.FilterSetDescription{
					ChildHashes: []string{`two`},
				}, TypeName: filters.FilterSetFilterType.Name()},
			},
		}, []string{}, true},
		{`sad childless not`, fields{Filters: map[string]filters.FilterDescription{
			`one`: {TypeName: filters.NotFilterType.Name()}},
		}, []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Description{
				Filters: tt.fields.Filters,
				Rules:   tt.fields.Rules,
				Error:   tt.fields.Error,
			}
			got, err := d.FilterDescriptions()
			if (err != nil) != tt.wantErr {
				t.Errorf("filterDescriptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			actual := make([]string, 0, len(got))
			for k := range got {
				actual = append(actual, k)
			}
			sort.Strings(actual)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("filterDescriptions() got = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestFetcher_Start(t *testing.T) {
	sb := &strings.Builder{}
	z := zerolog.New(sb)
	tests := []struct {
		name string
		done chan bool
		tick time.Duration
	}{
		{`no channel set`, nil, 0},
		{`done set`, make(chan bool), 0},
		{`ticking`, nil, 1 * time.Microsecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				done:     tt.done,
				endpoint: "_://",
				logger:   &z,
			}
			if tt.tick != 0 {
				f.ticker = time.NewTicker(tt.tick)
			}
			f.Start(func(*Description) {})
			if tt.tick != 0 {
				// Ensure enough time for at least a tick to be emitted.
				time.Sleep(100 * tt.tick)
			}
			f.done <- true
			if len(f.done) != 0 {
				t.Fatalf(`Fetcher select did not catch done signal`)
			}
			if tt.tick == 0 {
				return
			}
			sl := strings.Split(sb.String(), "\n")
			if len(sl) <= 1 {
				t.Fatalf(`Fetcher select did not catch ticks`)
			}
		})
	}
}

func TestFetcher_Fetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		var config string
		switch path {
		case `/401`:
			writer.WriteHeader(http.StatusUnauthorized)
			return
		case `/bad-config`:
			config = `]`
		case `/error/message`:
			config = `{	"error": { "message": "an error message" }}`
		case `/error/plain`:
			config = `{	"error": { "exception": "uncaught" }}`
		default:
			config = `
{
	"blockRules":[],
	"blockedDomains":[],
	"dataCollectionRules":[],
	"filters":{},
	"generatedAt":"2020-05-27T10:09:40.935730Z",
	"retryingRules":[],
	"rules":[],
	"timeoutRules":[]
}`
		}
		writer.Header().Set(proxy.ContentTypeHeader, proxy.FullContentTypeJSON)
		_, _ = writer.Write([]byte(config))
	}))
	defer ts.Close()
	sb := &strings.Builder{}
	z := zerolog.New(sb)
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{`happy empty`, ts.URL + `/config`, false},
		{`sad bad method`, `_://`, true},
		{`sad bad config host`, ``, true},
		{`sad rejected`, ts.URL + `/401`, true},
		{`sad bad config`, ts.URL + `/bad-config`, true},
		{`sad error message`, ts.URL + `/error/message`, true},
		{`sad error no message`, ts.URL + `/error/plain`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{endpoint: tt.endpoint, logger: &z}
			if _, err := f.Fetch(); (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
