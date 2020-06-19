package proxy_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/proxy"
)

func TestMakeConfigReport(t *testing.T) {
	const defaultEnv = `default`
	expectedEnv := base64.URLEncoding.EncodeToString([]byte(strings.ToLower(defaultEnv)))
	type args struct {
		version         string
		environmentType string
		secretKey       string
	}
	tests := []struct {
		name    string
		args    args
	}{
		{`happy`, args{`0.0.1`, defaultEnv, agent.ExampleWellFormedInvalidKey}},
		{`env case`, args{`0.0.1`, strings.Title(defaultEnv), agent.ExampleWellFormedInvalidKey}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proxy.MakeConfigReport(tt.args.version, tt.args.environmentType, tt.args.secretKey); got.Application.Environment != expectedEnv || got.AppEnvironment != expectedEnv {
				t.Errorf("MakeConfigReport(environment) = %s and %s, want %s", got.AppEnvironment, got.Application.Environment, expectedEnv)
			}
		})
	}
}
