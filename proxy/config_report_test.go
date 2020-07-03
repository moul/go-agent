package proxy_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/proxy"
)

func TestMakeConfigReport(t *testing.T) {
	const devEnv = `development`
	type args struct {
		version         string
		environmentType string
		expectedEnv string
		secretKey       string
	}
	tests := []struct {
		name    string
		args    args
	}{
		{`happy`, args{`0.0.1`, "", "", agent.ExampleWellFormedInvalidKey}},
		{`env case`, args{
			`0.0.1`,
			strings.Title(devEnv),
			base64.URLEncoding.EncodeToString([]byte(strings.ToLower(devEnv))),
			agent.ExampleWellFormedInvalidKey},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proxy.MakeConfigReport(tt.args.version, tt.args.environmentType, tt.args.secretKey); got.Application.Environment != tt.args.expectedEnv || got.AppEnvironment != tt.args.expectedEnv {
				t.Errorf("MakeConfigReport(environment) = %s and %s, want %s", got.AppEnvironment, got.Application.Environment, tt.args.expectedEnv)
			}
		})
	}
}
