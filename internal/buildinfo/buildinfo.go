// Package buildinfo holds the version details of the running binary.
package buildinfo

import (
	"runtime"
	"runtime/debug"
)

// These are meant to be set at build time, for example:
//
//	go build -ldflags "-X github.com/dewlonsystems/dewlon/internal/buildinfo.Version=v0.2.0 \
//	  -X github.com/dewlonsystems/dewlon/internal/buildinfo.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/dewlonsystems/dewlon/internal/buildinfo.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info is a snapshot of everything we know about this build.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the build details. Values injected through ldflags win; for
// anything left at its default we try the metadata Go embeds itself, which
// covers people who installed with `go install ...@version`.
func Get() Info {
	info := Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	if info.Version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		info.Version = bi.Main.Version
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if info.Commit == "none" && s.Value != "" {
				info.Commit = shorten(s.Value)
			}
		case "vcs.time":
			if info.Date == "unknown" && s.Value != "" {
				info.Date = s.Value
			}
		}
	}

	return info
}

func shorten(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}
	return rev
}
