package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_IsProtected_IsTrueOnlyForPIDsAtOrBelowOne(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		expected bool
	}{
		{name: "negative PID", pid: -1, expected: true},
		{name: "PID 0", pid: 0, expected: true},
		{name: "PID 1", pid: 1, expected: true},
		{name: "PID 2", pid: 2, expected: false},
		{name: "ordinary PID", pid: 100, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			info := newInfoFixture()
			info.PID = tt.pid

			// When
			protected := info.IsProtected()

			// Then
			assert.Equal(t, tt.expected, protected)
		})
	}
}

func TestInfo_IsKillableBy_FollowsOwnershipAndRootRules(t *testing.T) {
	tests := []struct {
		name        string
		processUID  int
		ownUID      int
		includeRoot bool
		expected    bool
	}{
		{name: "caller owns the process", processUID: 1000, ownUID: 1000, expected: true},
		{name: "another user owns the process", processUID: 1000, ownUID: 2000, expected: false},
		{name: "root caller may kill any process", processUID: 1000, ownUID: 0, expected: true},
		{name: "root caller may kill root-owned process", processUID: 0, ownUID: 0, expected: true},
		{name: "non-root caller may not kill root-owned process", processUID: 0, ownUID: 1000, expected: false},
		{name: "includeRoot lets non-root caller kill root-owned process", processUID: 0, ownUID: 1000, includeRoot: true, expected: true},
		{name: "includeRoot does not grant another non-root user's process", processUID: 1000, ownUID: 2000, includeRoot: true, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			info := newInfoFixture()
			info.UID = tt.processUID

			// When
			killable := info.IsKillableBy(tt.ownUID, tt.includeRoot)

			// Then
			assert.Equal(t, tt.expected, killable)
		})
	}
}

func TestFind_MatchesProcessNamesByPattern(t *testing.T) {
	tests := []struct {
		name      string
		processes []Info
		patterns  []string
		expected  []int
	}{
		{
			name:      "exact name",
			processes: []Info{newInfoFixtureFor(100, "myapp"), newInfoFixtureFor(200, "worker")},
			patterns:  []string{"myapp"},
			expected:  []int{100},
		},
		{
			name:      "substring of the name",
			processes: []Info{newInfoFixtureFor(100, "myapp-worker")},
			patterns:  []string{"app"},
			expected:  []int{100},
		},
		{
			name:      "uppercase process name",
			processes: []Info{newInfoFixtureFor(100, "MyApp")},
			patterns:  []string{"myapp"},
			expected:  []int{100},
		},
		{
			name:      "uppercase pattern",
			processes: []Info{newInfoFixtureFor(100, "myapp")},
			patterns:  []string{"MYAPP"},
			expected:  []int{100},
		},
		{
			name:      "multiple patterns keep input order",
			processes: []Info{newInfoFixtureFor(100, "myapp"), newInfoFixtureFor(200, "worker"), newInfoFixtureFor(300, "other")},
			patterns:  []string{"worker", "myapp"},
			expected:  []int{100, 200},
		},
		{
			name:      "no match",
			processes: []Info{newInfoFixtureFor(100, "myapp")},
			patterns:  []string{"worker"},
		},
		{
			name:      "no patterns",
			processes: []Info{newInfoFixtureFor(100, "myapp")},
			patterns:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := Find(tt.processes, tt.patterns, 999, 1000, false)

			// Then
			assert.Equal(t, tt.expected, newPidsFixtureOf(result))
		})
	}
}

func TestFind_DoesNotListSameProcessTwiceWhenItMatchesSeveralPatterns(t *testing.T) {
	// Given
	processes := []Info{newInfoFixtureFor(100, "myapp")}

	// When
	result := Find(processes, []string{"app", "my"}, 999, 1000, false)

	// Then
	assert.Equal(t, []int{100}, newPidsFixtureOf(result))
}

func TestFind_ExcludesOwnPID(t *testing.T) {
	// Given
	const ownPID = 999
	processes := []Info{newInfoFixtureFor(ownPID, "testprocess"), newInfoFixtureFor(100, "testprocess")}

	// When
	result := Find(processes, []string{"testprocess"}, ownPID, 1000, false)

	// Then
	assert.Equal(t, []int{100}, newPidsFixtureOf(result))
}

func TestFind_ExcludesProtectedProcesses(t *testing.T) {
	// Given
	processes := []Info{newInfoFixtureFor(0, "swapper"), newInfoFixtureFor(1, "init"), newInfoFixtureFor(100, "myapp")}

	// When
	result := Find(processes, []string{"swapper", "init", "myapp"}, 999, 1000, false)

	// Then
	assert.Equal(t, []int{100}, newPidsFixtureOf(result))
}

func TestFind_IncludesRootOwnedProcessesWhenIncludeRootIsSet(t *testing.T) {
	// Given
	rootOwned := newInfoFixtureFor(200, "sshd")
	rootOwned.UID = 0
	processes := []Info{newInfoFixtureFor(100, "myapp"), rootOwned}

	// When
	result := Find(processes, []string{"app", "sshd"}, 999, 1000, true)

	// Then
	assert.Equal(t, []int{100, 200}, newPidsFixtureOf(result))
}

func TestFind_ExcludesRootOwnedProcessesWhenIncludeRootIsNotSet(t *testing.T) {
	// Given
	rootOwned := newInfoFixtureFor(200, "sshd")
	rootOwned.UID = 0
	processes := []Info{newInfoFixtureFor(100, "myapp"), rootOwned}

	// When
	result := Find(processes, []string{"app", "sshd"}, 999, 1000, false)

	// Then
	assert.Equal(t, []int{100}, newPidsFixtureOf(result))
}

func TestValidateProcesses_RejectsEmptyList(t *testing.T) {
	tests := map[string][]Info{
		"nil list":   nil,
		"empty list": {},
	}
	for name, processes := range tests {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidateProcesses(processes)

			// Then
			assert.ErrorContains(t, err, "no processes found")
		})
	}
}

func TestValidateProcesses_AcceptsNonEmptyList(t *testing.T) {
	// When
	err := ValidateProcesses([]Info{newInfoFixture()})

	// Then
	assert.NoError(t, err)
}

func TestValidatePatterns_RejectsEmptyList(t *testing.T) {
	tests := map[string][]string{
		"nil list":   nil,
		"empty list": {},
	}
	for name, patterns := range tests {
		t.Run(name, func(t *testing.T) {
			// When
			err := ValidatePatterns(patterns)

			// Then
			assert.ErrorContains(t, err, "at least one search pattern is required")
		})
	}
}

func TestValidatePatterns_RejectsPatternsOutsideLengthRange(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		expected string
	}{
		{name: "empty pattern", patterns: []string{""}, expected: "must be at least"},
		{name: "below min", patterns: []string{strings.Repeat("a", minPatternLength-1)}, expected: "must be at least"},
		{name: "above max", patterns: []string{strings.Repeat("a", maxPatternLength+1)}, expected: "exceeds maximum length"},
		{name: "one bad pattern among valid ones", patterns: []string{"firefox", "ab", "chrome"}, expected: `"ab"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := ValidatePatterns(tt.patterns)

			// Then
			assert.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestValidatePatterns_AcceptsPatternsWithinLengthRange(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{name: "min length", patterns: []string{strings.Repeat("a", minPatternLength)}},
		{name: "max length", patterns: []string{strings.Repeat("a", maxPatternLength)}},
		{name: "mid range", patterns: []string{"firefox"}},
		{name: "several valid patterns", patterns: []string{"firefox", "chrome"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := ValidatePatterns(tt.patterns)

			// Then
			assert.NoError(t, err)
		})
	}
}

func TestValidateName_RejectsMismatch(t *testing.T) {
	// When
	err := ValidateName("target", "somethingElse")

	// Then
	assert.ErrorContains(t, err, `expected "target"`)
	assert.ErrorContains(t, err, `got "somethingElse"`)
}

func TestValidateName_AcceptsMatch(t *testing.T) {
	// When
	err := ValidateName("target", "target")

	// Then
	assert.NoError(t, err)
}

func TestValidateRoot_RefusesOnlyRootWithoutOverride(t *testing.T) {
	tests := []struct {
		name      string
		ownUID    int
		allowRoot bool
		expected  bool
	}{
		{name: "root without override", ownUID: 0, allowRoot: false, expected: true},
		{name: "root with override", ownUID: 0, allowRoot: true},
		{name: "non-root without override", ownUID: 1000, allowRoot: false},
		{name: "non-root with override", ownUID: 1000, allowRoot: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := ValidateRoot(tt.ownUID, tt.allowRoot)

			// Then
			if tt.expected {
				assert.ErrorContains(t, err, "refusing to run as root")
				return
			}
			assert.NoError(t, err)
		})
	}
}
