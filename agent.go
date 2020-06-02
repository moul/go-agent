package agent

import (
	"github.com/rs/zerolog"
	"fmt"
)

// Agent is the type of the Bearer entry point for your programs.
type Agent struct {
	logger *zerolog.Logger
	config        *Config
}

// NewAgent is the Agent constructor.
//
// In most usage scenarios, you will only use a single Agent in a given application.
func NewAgent(secretKey string, opts ...Option) (*Agent, error) {
	c, err := NewConfig(append(opts, WithSecretKey(secretKey))...)
	if err != nil {
		return nil, fmt.Errorf("configuring new agent: %w", err)
	}
	a := Agent{config: c}
	return &a, nil
}
