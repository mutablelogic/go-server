package version

import "runtime"

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GitHash     string
	GoBuildTime string
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Version() string {
	if GitTag != "" {
		return GitTag
	}
	if GitBranch != "" {
		return GitBranch
	}
	if GitHash != "" {
		return GitHash
	}
	return "dev"
}

func Map() map[string]any {
	vars := map[string]any{}
	if GitSource != "" {
		vars["source"] = GitSource
	}
	if GitTag != "" || GitBranch != "" {
		vars["version"] = GitTag
		vars["branch"] = GitBranch
		vars["hash"] = GitHash
	}
	if GoBuildTime != "" {
		vars["build"] = GoBuildTime
	}
	vars["go"] = runtime.Version()
	vars["os"] = runtime.GOOS
	vars["arch"] = runtime.GOARCH
	return vars
}
