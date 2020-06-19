package proxy

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
	default:
		return StageInvalid
	}
}

const (
	// StageUndefined represents a lack of requirement for any specific stage.
	// It is not used as one of the actual API stage.
	StageUndefined = "UndefinedStage"
	// StageConnect is the initial API call stage.
	StageConnect Stage = "ConnectStage"
	// StageRequest is the stage at which the request is being built.
	StageRequest Stage = "RequestStage"
	// StageResponse is the stage at which the response has started to return.
	StageResponse Stage = "ResponseStage"
	// StageBodies is the stage at which request and response bodies are available.
	StageBodies Stage = "BodiesStage"
	// StageInvalid is an invalid stage a request should never reach.
	StageInvalid Stage = "InvalidStage"
)
