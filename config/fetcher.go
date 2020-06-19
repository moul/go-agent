package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/filters"
)

const (
	// EOL is ASCII LF as a string.
	EOL = "\n"

	// ContentTypeHeader is the canonical content type header name.
	ContentTypeHeader = `Content-Type`
	// ContentTypeJSON is the canonical content type header value for JSON.
	ContentTypeJSON = `application/json; charset=utf-8`

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

// Description is a serialization-friendly description of the parts of Config
// which may come from the config server.
type Description struct {
	DataCollectionRules []DataCollectionRuleDescription
	Filters             map[string]filters.FilterDescription
	Rules               []struct {
		FilterHash   string
		ID           int
		Remediations []interface{}
		RuleType     string
	}
}

func (d Description) String() string {
	b := strings.Builder{}
	b.WriteString("Data Collection Rules\n")
	for _, dcr := range d.DataCollectionRules {
		b.WriteString(dcr.String())
	}
	b.WriteString("Filters\n")
	for k, f := range d.Filters {
		b.WriteString(fmt.Sprintf("%s: %s", k, f))
	}
	return b.String()
}

// filterHashes walks the Description.Filters to collate the filter descriptions.
func (d Description) filterDescriptions() (map[string]*filters.FilterDescription, error) {
	// All referenced hashes are supposed to have a matching filter description.
	//   - nil: not found
	//   - not nil: the description. Actual type depends on Filter type.
	hashes := make(map[string]*filters.FilterDescription, len(d.Filters))

	for hash, desc := range d.Filters {
		// Allocated a new instance every time.
		d2 := desc
		// Whether or not it was already seen, we now have a reference.
		hashes[hash] = &d2

		switch desc.TypeName {
		case filters.NotFilterType.Name():
			ch := desc.ChildHash
			if ch == `` {
				return nil, fmt.Errorf("a NotFilter(%s) has no child hash", hash)
			}
			if _, seen := hashes[ch]; !seen {
				hashes[ch] = nil
			}

		case filters.FilterSetFilterType.Name():
			chs := desc.ChildHashes
			// XXX Should we reject FilterSet filters without children ?
			for _, ch := range chs {
				if _, seen := hashes[ch]; !seen {
					hashes[ch] = nil
				}
			}
		}
	}

	var undefined []string
	for hash, description := range hashes {
		if description == nil {
			undefined = append(undefined, hash)
			break
		}
	}
	if len(undefined) > 0 {
		return nil, fmt.Errorf("undefined hashes referenced: %v", undefined)
	}

	return hashes, nil
}

// resolveHashes builds a filters.FilterMap from the filter descriptions it
// receives, resolving dependencies to allow instantiation.
// The function detects cyclic dependencies, and returns errors accordingly.
//
// Algorithm inspired by https://www.electricmonk.nl/docs/dependency_resolving_algorithm/dependency_resolving_algorithm.html
// Published under a permissive license
func (d Description) resolveHashes(descriptions map[string]*filters.FilterDescription) (filters.FilterMap, error) {
	type filterSlice []struct{hash string; filters.Filter}

	resolved := make(filterSlice, 0, len(descriptions))
	unresolved := make(map[string]*filters.FilterDescription)

	resolvedIndexOf := func(sl filterSlice, hash string) int {
		pos := -1
		for i := 0; i < len(sl); i++ {
			if hash == sl[i].hash {
				pos = i
				break
			}
		}
		return pos
	}

	var resolve func(string, *filters.FilterDescription) error
	resolve = func(hash string, desc *filters.FilterDescription) error {
		var dependencyHashes []string
		unresolved[hash] = desc
		switch desc.TypeName {
		case filters.NotFilterType.Name():
			dependencyHashes = []string{desc.ChildHash}
		case filters.FilterSetFilterType.Name():
			dependencyHashes = desc.ChildHashes
		}
		for _, dependencyHash := range dependencyHashes {
			if resolvedIndexOf(resolved, dependencyHash) == -1 {
				if _, ok := unresolved[dependencyHash]; ok {
					return fmt.Errorf("circular hash dependency: %s <-> %s", hash, dependencyHash)
				}
				err := resolve(dependencyHash, descriptions[dependencyHash])
				if err != nil {
					return nil
				}
			}
		}
		if resolvedIndexOf(resolved, hash) == -1 {
			resolved = append(resolved, struct{hash string; filters.Filter}{hash, nil})
		}
		delete(unresolved, hash)
		return nil
	}

	for hash, desc := range descriptions {
		err := resolve(hash, desc)
		if err != nil {
			return nil, err
		}
	}

	res := make(filters.FilterMap, len(resolved))
	for _, info := range resolved {
		res[info.hash] = info.Filter
	}
	return res, nil
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
func (f *Fetcher) Fetch() {
	report := &bytes.Buffer{}
	err := json.NewEncoder(report).Encode(makeConfigReport(f.version))
	if err != nil {
		f.logger.Warn().Msgf("building Bearer config report: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, f.config.configEndpoint, report)
	if err != nil {
		f.logger.Warn().Msgf("building Bearer remote config request: %v", err)
		return
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", f.config.secretKey)
	req.Header.Set(ContentTypeHeader, ContentTypeJSON)

	client := http.Client{Transport: f.transport}
	res, err := client.Do(req)
	if err != nil {
		f.logger.Warn().Msgf("failed remote config from Bearer: %v", err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		f.logger.Warn().Msgf("reading remote config received from Bearer: %v", err)
		return
	}
	remoteConf := Description{}
	if err := json.Unmarshal(body, &remoteConf); err != nil {
		f.logger.Warn().Msgf("decoding remote config received from Bearer: %v", err)
		return
	}
	fmt.Println(remoteConf)
	fmt.Printf("%s\n", body)
	fmt.Println(remoteConf.filterDescriptions())
	f.config.UpdateFromDescription(remoteConf)
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
