package filters

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// FilterMap binds Filter hashes in a config.Description to the actual Filter instances.
type FilterMap map[string]Filter

// FilterType allows Filter types to have "static" properties.
type FilterType interface {
	// Create creates an instance of the described type. The actual type of the
	// description depends on the FilterType and will be asserted.
	Create(fm FilterMap, description *FilterDescription) Filter
	Name() string
	WantsRequest() bool
	WantsResponse() bool
	fmt.Stringer
}

type filterType struct {
	name                        string
	creator                     func(FilterMap, *FilterDescription) Filter
	wantsRequest, wantsResponse bool
}

func (ft filterType) Name() string {
	return ft.name
}

func (ft filterType) Create(fm FilterMap, description *FilterDescription) Filter {
	return ft.creator(fm, description)
}

func (ft filterType) WantsRequest() bool {
	return ft.wantsRequest
}

func (ft filterType) WantsResponse() bool {
	return ft.wantsResponse
}

// String implements fmt.Stringer.
func (ft filterType) String() string {
	b := strings.Builder{}
	b.WriteString(ft.name + `:`)
	b.WriteString(fmt.Sprintf("%016x:", reflect.ValueOf(ft.creator).Pointer()))
	b.WriteString(fmt.Sprintf(`%t:`, ft.wantsRequest))
	b.WriteString(fmt.Sprintf(`%t`, ft.wantsResponse))
	return b.String()
}

// Filter defines the behaviour common to all filters
type Filter interface {
	Type() FilterType
	// MatchesCall checks whether the filter, with its configuration, matches the
	// request and response passed to it, if any: some filters may not need a
	// request nor a response, in which case nil is a valid value to pass them.
	MatchesCall(*http.Request, *http.Response) bool
	// SetMatcher assigns a specific Matcher instance to the filter.
	// Passing a nil matcher will assign a filter-specific default Matcher.
	SetMatcher(Matcher) error
}

var (
	// NotFilterType describes NotFilter.
	NotFilterType FilterType = filterType{"NotFilter", notFilterFromDescription, true, true}
	// FilterSetFilterType describes the FilterSet.
	FilterSetFilterType FilterType = filterType{"FilterSet", setFilterFromDescription, true, true}

	// DomainFilterType describes DomainFilter.
	DomainFilterType FilterType = filterType{"DomainFilter", domainFilterFromDescription, true, false}

	// HTTPMethodFilterType describes HTTPMethodFilter.
	HTTPMethodFilterType FilterType = filterType{"HttpMethodFilter", methodFilterFromDescription, true, false}
	// ParamFilterType describes ParamFilter.
	ParamFilterType FilterType = filterType{"ParamFilter", paramFilterFromDescription, true, false}
	// PathFilterType describes PathFilter.
	PathFilterType FilterType = filterType{"PathFilter", pathFilterFromDescription, true, false}
	// RequestHeadersFilterType describes RequestHeadersFilter.
	RequestHeadersFilterType FilterType = filterType{"RequestHeadersFilter", requestFilterHeadersFromDescription, true, false}
	// ResponseHeadersFilterType describes ResponseHeadersFilter.
	ResponseHeadersFilterType FilterType = filterType{"ResponseHeadersFilter", responseHeadersFilterFromDescription, false, true}
	// StatusCodeFilterType describes StatusCodeFilter.
	StatusCodeFilterType FilterType = filterType{"StatusCodeFilter", statusCodeFilterFromDescription, false, true}

	//RequestBodiesFilterType  FilterType = filterType{"RequestBodiesFilter", requestBodiesFilterFromDescription, true, false}
	//ResponseBodiesFilterType FilterType = filterType{"ResponseBodiesFilter", responseBodiesFilterFromDescription, false, true}

	//ConnectionErrorFilterType FilterType = filterType{"ConnectionErrorFilter", connectionErrorFilterFromDescription, false, false}
	yesInternalFilter FilterType = filterType{"YesFilter", nil, false, false}
)

// FilterTypeByName returns a FilterType instance for the passed name, or nil if
// the name does not match an existing FilterType.
func FilterTypeByName(name string) FilterType {
	switch name {
	case NotFilterType.Name():
		return NotFilterType
	case FilterSetFilterType.Name():
		return FilterSetFilterType
	case DomainFilterType.Name():
		return DomainFilterType
	case HTTPMethodFilterType.Name():
		return HTTPMethodFilterType
	case ParamFilterType.Name():
		return ParamFilterType
	case PathFilterType.Name():
		return PathFilterType
	case RequestHeadersFilterType.Name():
		return RequestHeadersFilterType
	case ResponseHeadersFilterType.Name():
		return ResponseHeadersFilterType
	case StatusCodeFilterType.Name():
		return StatusCodeFilterType
	case yesInternalFilter.Name():
		return yesInternalFilter
	default:
		return nil
	}
}

// NewFilterFromDescription creates a Filter instance from a FilterDescription.
func NewFilterFromDescription(filterMap FilterMap, fd *FilterDescription) Filter {
	ft := FilterTypeByName(fd.TypeName)
	if ft == nil {
		return nil
	}
	f := ft.Create(filterMap, fd)
	return f
}

// FilterDescription is a kind of "union type" describing all possible
// fields returned by the config server, with TypeName acting as the discriminator.
type FilterDescription struct {
	// ChildHash is set on filters.NotFilter
	ChildHash string

	// Value is set on filters using filters.StringMatcher, like filters.HTTPMethodFilter.
	Value string

	// Pattern is set on filters using filters.RegexpMatcher, like filters.DomainFilter.
	// XXX Its fields are not portable across regexp implementations.
	Pattern RegexpMatcherDescription

	// FilterSetDescription carries the fields set on filters.FilterSet filters.
	FilterSetDescription

	// KeyValueDescription carries the fields set on filters using filters.KeyValueMatcher.
	// XXX Its fields are not portable across regexp implementations.
	KeyValueDescription

	// Range is set on filters using filters.RangeMatcher like filters.StatusCodeFilter.
	Range RangeMatcherDescription

	// StageType is one of the 4 API call stages.
	StageType string

	// TypeName is the name of the filter type, used to select which fields
	// from the config to parse.
	TypeName string
}

func (d FilterDescription) String() string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%-22s - %-20s - ", d.TypeName, d.StageType))
	l1 := len(b.String())
	if d.ChildHash != `` {
		b.WriteString(`H: ` + d.ChildHash + "\n")
	}
	if d.Value != `` {
		b.WriteString(`Value: ` + d.Value + "\n")
	}
	b.WriteString(d.Pattern.String())
	b.WriteString(d.FilterSetDescription.String())
	b.WriteString(d.KeyValueDescription.String())
	b.WriteString(d.Range.String())
	s := b.String()
	if len(s) == l1 {
		s += "\n"
	}
	return s
}
