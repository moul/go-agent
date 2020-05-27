package filters

import (
	"net/http"
)

// Stage represents the stage an API call is in.
type Stage string

// Next provides the stage following the current stage.
func (s Stage) Next() Stage {
	switch s {
	case StageConnect:
		return StageRequest
	case StageRequest:
		return StageResponse
	case StageResponse:
		return StageBodies
	case StageUndefined:
	case StageInvalid:
	default:
		return StageInvalid
	}
	return StageInvalid
}

const (
	// StageUndefined represents a lack of requirement for any specific stage.
	// It is not used as one of the actual API stage.
	StageUndefined = "UNDEFINED"
	// StageConnect is the initial API call stage.
	StageConnect Stage = "CONNECT"
	// StageRequest is the stage at which the request is being built.
	StageRequest Stage = "REQUEST"
	// StageResponse is the stage at which the response has started to return.
	StageResponse Stage = "RESPONSE"
	// StageBodies is the stage at which request and response bodies are available.
	StageBodies Stage = "BODIES"
	// StageInvalid is an invalid stage a request should never reach.
	StageInvalid Stage = "INVALID"
)

// FilterType allows Filter types to have "static" properties.
type FilterType interface {
	Name() string
	WantsRequest() bool
	WantsResponse() bool
}

type filterType struct {
	name                        string
	wantsRequest, wantsResponse bool
}

func (ft filterType) Name() string {
	return ft.name
}

func (ft filterType) WantsRequest() bool {
	return ft.wantsRequest
}

func (ft filterType) WantsResponse() bool {
	return ft.wantsResponse
}

// Filter defines the behaviour common to all filters
type Filter interface {
	Type() FilterType
	// Matches checks whether the filter, with its configuration, matches the
	// request and response passed to it, if any: some filters may not need a
	// request nor a response, in which case nil is a valid value to pass them.
	Matches(*http.Request, *http.Response) bool
}

var (
	notFilter       FilterType = filterType{"NotFilter", true, true}
	filterSetFilter FilterType = filterType{"FilterSet", true, true}

	domainFilter FilterType = filterType{"DomainFilter", true, false}

	httpMethodFilter     FilterType = filterType{"HttpMethodFilter", true, false}
	paramFilter          FilterType = filterType{"ParamFilter", true, false}
	pathFilter           FilterType = filterType{"PathFilter", true, false}
	requestHeadersFilter FilterType = filterType{"RequestHeadersFilter", true, false}

	responseHeadersFilter FilterType = filterType{"ResponseHadersFilter", false, true}
	statusCodeFilter      FilterType = filterType{"StatusCodeFilter", false, true}

	requestBodiesFilter  FilterType = filterType{"RequestBodiesFilter", true, false}
	responseBodiesFilter FilterType = filterType{"ResponseBodiesFilter", false, true}

	connectionErrorFilter FilterType = filterType{"ConnectionErrorFilter", false, false}
	yesInternalFilter     FilterType = filterType{"yesFilter", false, false}
)
