package version

import (
	"runtime/debug"
)

// Set via -ldflags at build time (e.g. by goreleaser).
// When left at their defaults, Info() falls back to debug.ReadBuildInfo().
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info returns version, commit, and date strings.
// It prefers values injected via ldflags; if those are still at their
// defaults it reads the embedded build info that Go populates automatically.
func Info() (version, commit, date string) {
	version = Version
	commit = Commit
	date = Date

	if version != "dev" && commit != "none" && date != "unknown" {
		return // ldflags were set — use them as-is
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = bi.Main.Version
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if commit == "none" && len(s.Value) > 0 {
				commit = s.Value
				if len(commit) > 12 {
					commit = commit[:12]
				}
			}
		case "vcs.time":
			if date == "unknown" && s.Value != "" {
				date = s.Value
			}
		case "vcs.modified":
			if s.Value == "true" && commit != "none" {
				commit += "-dirty"
			}
		}
	}

	return
}
