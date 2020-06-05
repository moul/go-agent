package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	// ContentTypeHeader is the canonical content type header name.
	ContentTypeHeader = `Content-Type`
	// ContentTypeJSON is the canonical content type header value for JSON.
	ContentTypeJSON   = `application/json; charset=utf-8`

	// HostUnknown is a reserved host name used when the Agent cannot obtain the
	// client host name from the operating system.
	HostUnknown = `unknown`
)

// RuntimeReport is the part of the Report describing the client runtime environment.
type RuntimeReport struct {
	Version  string `json:"version"`
	Arch     string `json:"arch"`
	Platform string `json:"platform"`
	Type     string `json:"type"`
	Hostname string `json:"hostname,omitempty"`
}

// AgentReport is the part of the Report describing the Agent code.
type AgentReport struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

// ApplicationReport is the part of the Report describing the application
// execution environment, like "development", "staging", or "production".
type ApplicationReport struct {
	Environment string `json:"environment"`
}

// Report is the information sent to the Bearer configuration server, describing
// the current agent operating environment.
type Report struct {
	Runtime     RuntimeReport     `json:"runtime"`
	Agent       AgentReport       `json:"agent"`
	Application ApplicationReport `json:"application"`
}

func makeConfigReport(version string) Report {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = HostUnknown
	}
	return Report{
		Runtime: RuntimeReport{
			Version:  runtime.Version(),
			Arch:     runtime.GOARCH,
			Platform: runtime.GOOS,
			Type:     runtime.GOOS,
			Hostname: hostname,
		},
		Agent: AgentReport{
			Type:    "go",
			Version: version,
		},
		Application: ApplicationReport{
			Environment: base64.URLEncoding.EncodeToString([]byte(strings.ToLower("de-zoom-ing"))),
		},
	}
}

// Fetcher describes the data used to perform the background configuration refresh.
type Fetcher struct {
	config    *Config
	done      chan bool
	logger    *zerolog.Logger
	ticker    *time.Ticker
	transport http.RoundTripper
	version   string
}

// NewFetcher builds an un-started Fetcher.
func NewFetcher(transport http.RoundTripper, logger *zerolog.Logger, version string, config *Config) *Fetcher {
	return &Fetcher{
		config:    config,
		logger:    logger,
		ticker:    time.NewTicker(config.fetchInterval),
		transport: transport,
		version:   version,
	}
}

// Fetch fetches a fresh configuration from the Bearer platform and assigns it
// to the current config. As per Agent spec, all config fetch errors are logged
// and ignored.
func (f *Fetcher) Fetch() *Config {
	report := &bytes.Buffer{}
	err := json.NewEncoder(report).Encode(makeConfigReport(f.version))
	if err != nil {
		f.logger.Warn().Msgf("building Bearer config report: %v", err)
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, f.config.configEndpoint, report)
	if err != nil {
		f.logger.Warn().Msgf("building Bearer remote config request: %v", err)
		return nil
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", f.config.secretKey)
	req.Header.Set(ContentTypeHeader, ContentTypeJSON)

	client := http.Client{Transport: f.transport}
	res, err := client.Do(req)
	if err != nil {
		f.logger.Warn().Msgf("failed remote config from Bearer: %v", err)
		return nil
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		f.logger.Warn().Msgf("reading remote config received from Bearer: %v", err)
		return nil
	}
	remoteConf := make(map[string]interface{})
	if err := json.Unmarshal(body, &remoteConf); err != nil {
		f.logger.Warn().Msgf("decoding remote config received from Bearer: %v", err)
		return nil
	}
	f.logger.Debug().Msgf("Remote conf: %#v", remoteConf)

	return &Config{}
}

// Stop deactivates the fetcher background operation.
func (f *Fetcher) Stop() {
	f.ticker.Stop()
	f.done <- true
}

// Start activates the fetcher background operation.
func (f *Fetcher) Start() {
	// TODO implement.
}
