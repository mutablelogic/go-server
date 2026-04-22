package version

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitTag    string
	GitBranch string
	BuildDate string
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Version() string {
	// Return ldflags values if set
	if GitTag != "" {
		return GitTag
	}
	if GitBranch != "" {
		return GitBranch
	}

	// Fall back to vcs.revision from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				return s.Value[:12]
			}
		}
	}

	// Fall back to "dev" if no version information is available
	return "dev"
}

func BuildTime() time.Time {
	if BuildDate != "" {
		if t, err := time.Parse(time.RFC3339, BuildDate); err == nil {
			return t.UTC()
		}
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.time" && s.Value != "" {
				if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
					return t.UTC()
				}
			}
		}
	}

	return time.Time{}
}

func JSON(execName string) json.RawMessage {
	metadata := map[string]string{
		"name":     execName,
		"version":  Version(),
		"compiler": runtime.Version(),
	}

	// Add ldflags values if set
	if GitTag != "" {
		metadata["tag"] = GitTag
	}
	if GitBranch != "" {
		metadata["branch"] = GitBranch
	}
	if buildTime := BuildTime(); !buildTime.IsZero() {
		metadata["build_time"] = buildTime.Format(time.RFC3339)
	}

	// Add build info from runtime/debug
	var goos, goarch string
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path != "" {
			metadata["source"] = info.Main.Path
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if s.Value != "" {
					metadata["hash"] = s.Value
				}
			case "vcs.modified":
				if s.Value == "true" {
					metadata["modified"] = s.Value
				}
			case "GOOS":
				goos = s.Value
			case "GOARCH":
				goarch = s.Value
			}
		}
	}
	if goos != "" && goarch != "" {
		metadata["platform"] = goos + "/" + goarch
	}

	// Encode JSON metadata
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	}

	// Return JSON metadata
	return data
}
