package agent

import (
	"github.com/rs/zerolog"
)

// Agent is the type of the Bearer entry point for your programs.
type Agent struct {
	logger *zerolog.Logger
}
