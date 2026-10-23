package composition

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuildInfo_HoldsTheBuildInformationOfTheRunningProgram(t *testing.T) {
	// When
	info := newBuildInfo()

	// Then
	assert.NotNil(t, info.info)
	assert.Equal(t, releaseVersion, info.release)
}

func TestBuildInfo_Version_PrefersTheReleaseVersion(t *testing.T) {
	// Given
	info := &buildInfo{release: "1.2.3", info: &debug.BuildInfo{Main: debug.Module{Version: "v0.3.0"}}}

	// When
	result := info.Version()

	// Then
	assert.Equal(t, "1.2.3", result)
}

func TestBuildInfo_Version_ReturnsRecordedVersionOrDev(t *testing.T) {
	tests := newBuildInfoVersionTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			info := &buildInfo{info: test.info}

			// When
			result := info.Version()

			// Then
			assert.Equal(t, test.expected, result)
		})
	}
}

func newBuildInfoVersionTestCase() []struct {
	name     string
	info     *debug.BuildInfo
	expected string
} {
	return []struct {
		name     string
		info     *debug.BuildInfo
		expected string
	}{
		{"tagged build", &debug.BuildInfo{Main: debug.Module{Version: "v0.3.0"}}, "0.3.0"},
		{"pseudo-version build", &debug.BuildInfo{Main: debug.Module{Version: "v0.2.1-0.20261022101500-abcdef123456+dirty"}}, "0.2.1-0.20261022101500-abcdef123456+dirty"},
		{"development build", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev"},
		{"empty version", &debug.BuildInfo{}, "dev"},
		{"no build information", nil, "dev"},
	}
}
