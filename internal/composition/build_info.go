package composition

import (
	"runtime/debug"
	"strings"
)

const (
	devVersion   = "dev"
	goDevVersion = "(devel)"
)

// buildInfo holds the build information embedded in the program, if any.
type buildInfo struct {
	info *debug.BuildInfo
}

// newBuildInfo returns a buildInfo holding the build information embedded in
// the running program.
func newBuildInfo() *buildInfo {
	info, _ := debug.ReadBuildInfo()
	return &buildInfo{info: info}
}

// Version returns the version the program was built from without the leading
// "v" of its tag, or "dev" when the build does not record one.
func (b *buildInfo) Version() string {
	if b.info == nil || b.info.Main.Version == "" || b.info.Main.Version == goDevVersion {
		return devVersion
	}

	return strings.TrimPrefix(b.info.Main.Version, "v")
}
