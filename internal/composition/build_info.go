package composition

import (
	"runtime/debug"
	"strings"
)

const (
	devVersion   = "dev"
	goDevVersion = "(devel)"
)

// releaseVersion is the version the program reports when it was built for a
// release. It is empty otherwise.
var releaseVersion string

// buildInfo holds the release version of the program and the build information
// embedded in it, if any.
type buildInfo struct {
	release string
	info    *debug.BuildInfo
}

// newBuildInfo returns a buildInfo of the running program.
func newBuildInfo() *buildInfo {
	info, _ := debug.ReadBuildInfo()
	return &buildInfo{release: releaseVersion, info: info}
}

// Version returns the version of the program without a leading "v", or "dev"
// when it has none.
func (b *buildInfo) Version() string {
	if b.release != "" {
		return b.release
	}

	if b.info == nil || b.info.Main.Version == "" || b.info.Main.Version == goDevVersion {
		return devVersion
	}

	return strings.TrimPrefix(b.info.Main.Version, "v")
}
