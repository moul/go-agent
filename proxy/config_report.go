package proxy

import (
	"encoding/base64"
	"os"
	"runtime"
	"strings"
)

// HostUnknown is a reserved host name used when the Agent cannot obtain the
// client host name from the operating system.
const HostUnknown = `unknown`

// MakeConfigReport creates a valid Report
func MakeConfigReport(version string, environmentType string) Report {
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
			Environment: base64.URLEncoding.EncodeToString([]byte(strings.ToLower(environmentType))),
		},
	}
}

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
	SecretKey   string            `json:"secretKey"`
	Application ApplicationReport `json:"appEnvironment"`
	Runtime     RuntimeReport     `json:"runtime"`
	Agent       AgentReport       `json:"agent"`
	Logs        []ReportLog       `json:"logs"`
}
