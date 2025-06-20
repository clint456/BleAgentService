package messagebus

import (
	"fmt"
	"runtime"
)

// Version information for the EdgeX MessageBus Client
const (
	// Version is the current version of the EdgeX MessageBus Client
	Version = "1.1.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "unknown"

	// BuildDate is the build date (set during build)
	BuildDate = "unknown"
)

// VersionInfo contains version and build information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// GetVersion returns the current version information
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (v VersionInfo) String() string {
	return fmt.Sprintf("EdgeX MessageBus Client v%s (commit: %s, built: %s, go: %s, platform: %s)",
		v.Version, v.GitCommit, v.BuildDate, v.GoVersion, v.Platform)
}

// GetVersionString returns a simple version string
func GetVersionString() string {
	return fmt.Sprintf("v%s", Version)
}
