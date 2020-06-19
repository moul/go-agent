package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/bearer/go-agent/filters"
	"github.com/bearer/go-agent/interception"
	"github.com/bearer/go-agent/proxy"
)

// Description is a serialization-friendly description of the parts of Config
// which may come from the config server.
type Description struct {
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

// FilterDescriptions walks the Description.Filters to collate the filter descriptions.
func (d Description) FilterDescriptions() (map[string]*filters.FilterDescription, error) {
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

// ResolveHashes builds a filters.FilterMap from the filter descriptions it
// receives, resolving dependencies to allow instantiation.
// The function detects cyclic dependencies, and returns errors accordingly.
//
// Algorithm inspired by https://www.electricmonk.nl/docs/dependency_resolving_algorithm/dependency_resolving_algorithm.html
// Published under a permissive license
func (d Description) ResolveHashes(descriptions map[string]*filters.FilterDescription) (filters.FilterMap, error) {
	// TODO simplify type: filter is always nil.
	type filterSlice []struct {
		hash string
		filters.Filter
	}

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
			resolved = append(resolved, struct {
				hash string
				filters.Filter
			}{hash, nil})
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
		res[info.hash] = filters.NewFilterFromDescription(res, descriptions[info.hash])
	}
	return res, nil
}

// ResolveDCRs creates a slice of DataCollectionRule values from a resolved filters.FilterMap.
func (d *Description) ResolveDCRs(filterMap filters.FilterMap) ([]*interception.DataCollectionRule, error) {
	dcrs := make([]*interception.DataCollectionRule, 0, len(d.DataCollectionRules))
	for _, desc := range d.DataCollectionRules {
		dcr := interception.NewDCRFromDescription(filterMap, desc)
		if dcr == nil {
			continue
		}
		dcrs = append(dcrs, dcr)
	}
	return dcrs, nil
}

// Fetcher describes the data used to perform the background configuration refresh.
type Fetcher struct {
	done            chan bool
	endpoint        string
	environmentType string
	logger          *zerolog.Logger
	secretKey       string
	ticker          *time.Ticker
	transport       http.RoundTripper
	version         string
}

// NewFetcher builds an un-started Fetcher.
func NewFetcher(transport http.RoundTripper, logger *zerolog.Logger, version string, fetchEndpoint string, fetchInterval time.Duration, environmentType string, secretKey string) *Fetcher {
	return &Fetcher{
		done:            make(chan bool),
		endpoint:        fetchEndpoint,
		environmentType: environmentType,
		logger:          logger,
		secretKey:       secretKey,
		ticker:          time.NewTicker(fetchInterval),
		transport:       transport,
		version:         version,
	}
}

// Fetch fetches a fresh configuration from the Bearer platform and assigns it
// to the current config. As per Agent spec, all config fetch errors are logged
// and ignored.
func (f *Fetcher) Fetch() (*Description, error) {
	report := &bytes.Buffer{}
	err := json.NewEncoder(report).Encode(proxy.MakeConfigReport(f.version, f.environmentType, f.secretKey))
	if err != nil {
		f.logger.Warn().Msgf("building Bearer config report: %v", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, f.endpoint, report)
	if err != nil {
		f.logger.Warn().Msgf("building Bearer remote config request: %v", err)
		return nil, err
	}
	req.Header.Add(proxy.AcceptHeader, "application/json")
	req.Header.Add("Authorization", f.secretKey)
	req.Header.Set(proxy.ContentTypeHeader, proxy.FullContentTypeJSON)

	client := http.Client{Transport: f.transport}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if err == nil {
			err = errors.New("the Bearer platform rejected the config fetch")
		}
		f.logger.Warn().Msgf("failed remote config from Bearer: %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		f.logger.Warn().Msgf("reading remote config received from Bearer: %v", err)
		return nil, err
	}
	remoteConf := Description{}
	if err := json.Unmarshal(body, &remoteConf); err != nil {
		f.logger.Warn().Msgf("decoding remote config received from Bearer: %v", err)
		return nil, err
	}
	if len(remoteConf.Error) > 0 {
		message := "the Bearer platform rejected the config request"
		errMessage, ok := remoteConf.Error[`message`]
		if ok {
			f.logger.Warn().Str(`config error message`, errMessage).Msg(message)
		} else {
			j, err := json.Marshal(remoteConf.Error)
			if err != nil {
				f.logger.Warn().RawJSON(`config response`, body).Msg(message)
			}
			f.logger.Warn().RawJSON(`config error`, j).Msg(message)
		}
		return nil, errors.New(message)
	}
	return &remoteConf, nil
}

// Stop deactivates the fetcher background operation.
func (f *Fetcher) Stop() {
	f.ticker.Stop()
	f.done <- true
}

// Start activates the fetcher background operation.
func (f *Fetcher) Start(configSetter func(*Description)) {
	if f.done == nil {
		f.done = make(chan bool)
	}
	if f.ticker == nil {
		f.ticker = time.NewTicker(DefaultFetchInterval)
	}
	go func() {
		defer f.ticker.Stop()
		for {
			select {
			case <-f.done:
				return
			case <-f.ticker.C:
				f.logger.Trace().Msgf(`Background config fetch`)
				d, err := f.Fetch()
				if err != nil {
					configSetter(d)
				}
			}
		}
	}()
}
