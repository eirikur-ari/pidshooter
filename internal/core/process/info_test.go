package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_IsProtected_IsTrueOnlyForPIDsAtOrBelowOne(t *testing.T) {
	tests := newProtectedPIDTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			info := newInfoFixture()
			info.PID = test.pid

			// When
			protected := info.IsProtected()

			// Then
			assert.Equal(t, test.expected, protected)
		})
	}
}

func TestInfo_IsKillableBy_FollowsOwnershipAndRootRules(t *testing.T) {
	tests := newKillableByTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			info := newInfoFixture()
			info.UID = test.processUID

			// When
			killable := info.IsKillableBy(test.ownUID, test.includeRoot)

			// Then
			assert.Equal(t, test.expected, killable)
		})
	}
}

func TestFind_MatchesProcessNamesByPattern(t *testing.T) {
	tests := newPatternMatchTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			result := Find(test.processes, test.patterns, 999, 1000, false)
			pids := pidsOf(result)

			// Then
			assert.Equal(t, test.expected, pids)
		})
	}
}

func TestFind_DoesNotListSameProcessTwiceWhenItMatchesSeveralPatterns(t *testing.T) {
	// Given
	processes := []Info{newInfoFixtureFor(100, "myapp")}
	expectedPIDs := []int{100}

	// When
	result := Find(processes, []string{"app", "my"}, 999, 1000, false)
	pids := pidsOf(result)

	// Then
	assert.Equal(t, expectedPIDs, pids)
}

func TestFind_ExcludesOwnPID(t *testing.T) {
	// Given
	const ownPID = 999
	processes := []Info{newInfoFixtureFor(ownPID, "testprocess"), newInfoFixtureFor(100, "testprocess")}
	expectedPIDs := []int{100}

	// When
	result := Find(processes, []string{"testprocess"}, ownPID, 1000, false)
	pids := pidsOf(result)

	// Then
	assert.Equal(t, expectedPIDs, pids)
}

func TestFind_ExcludesProtectedProcesses(t *testing.T) {
	// Given
	processes := []Info{newInfoFixtureFor(0, "swapper"), newInfoFixtureFor(1, "init"), newInfoFixtureFor(100, "myapp")}
	expectedPIDs := []int{100}

	// When
	result := Find(processes, []string{"swapper", "init", "myapp"}, 999, 1000, false)
	pids := pidsOf(result)

	// Then
	assert.Equal(t, expectedPIDs, pids)
}

func TestFind_ListsRootOwnedProcessesOnlyWhenIncludeRootIsSet(t *testing.T) {
	tests := newIncludeRootTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			rootOwned := newInfoFixtureFor(200, "sshd")
			rootOwned.UID = 0
			processes := []Info{newInfoFixtureFor(100, "myapp"), rootOwned}

			// When
			result := Find(processes, []string{"app", "sshd"}, 999, 1000, test.includeRoot)
			pids := pidsOf(result)

			// Then
			assert.Equal(t, test.expected, pids)
		})
	}
}

func TestValidateProcesses_RejectsEmptyList(t *testing.T) {
	tests := newEmptyProcessListTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateProcesses(test.processes)

			// Then
			assert.ErrorContains(t, err, "no processes found")
		})
	}
}

func TestValidateProcesses_AcceptsNonEmptyList(t *testing.T) {
	// Given
	processes := []Info{newInfoFixture()}

	// When
	err := ValidateProcesses(processes)

	// Then
	assert.NoError(t, err)
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

func TestValidatePatterns_RejectsEmptyList(t *testing.T) {
	tests := newEmptyPatternListTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidatePatterns(test.patterns)

			// Then
			assert.ErrorContains(t, err, "at least one search pattern is required")
		})
	}
}

func TestValidatePatterns_RejectsPatternsOutsideLengthRange(t *testing.T) {
	tests := newOutOfRangePatternTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidatePatterns(test.patterns)

			// Then
			assert.ErrorContains(t, err, test.expected)
		})
	}
}

func TestValidatePatterns_AcceptsPatternsWithinLengthRange(t *testing.T) {
	tests := newInRangePatternTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidatePatterns(test.patterns)

			// Then
			assert.NoError(t, err)
		})
	}
}

func TestValidateRoot_RejectsRootWithoutOverride(t *testing.T) {
	tests := newRejectedRootTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateRoot(test.ownUID, test.allowRoot)

			// Then
			assert.ErrorContains(t, err, "refusing to run as root")
		})
	}
}

func TestValidateRoot_AcceptsEveryOtherCombination(t *testing.T) {
	tests := newAcceptedRootTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := ValidateRoot(test.ownUID, test.allowRoot)

			// Then
			assert.NoError(t, err)
		})
	}
}

func newProtectedPIDTestCase() []struct {
	name     string
	pid      int
	expected bool
} {
	return []struct {
		name     string
		pid      int
		expected bool
	}{
		{"negative PID", -1, true},
		{"PID 0", 0, true},
		{"PID 1", 1, true},
		{"PID 2", 2, false},
		{"ordinary PID", 100, false},
	}
}

func newKillableByTestCase() []struct {
	name        string
	processUID  int
	ownUID      int
	includeRoot bool
	expected    bool
} {
	return []struct {
		name        string
		processUID  int
		ownUID      int
		includeRoot bool
		expected    bool
	}{
		{"caller owns the process", 1000, 1000, false, true},
		{"another user owns the process", 1000, 2000, false, false},
		{"root caller may kill any process", 1000, 0, false, true},
		{"root caller may kill root-owned process", 0, 0, false, true},
		{"non-root caller may not kill root-owned process", 0, 1000, false, false},
		{"includeRoot lets non-root caller kill root-owned process", 0, 1000, true, true},
		{"includeRoot does not grant another non-root user's process", 1000, 2000, true, false},
	}
}

func newPatternMatchTestCase() []struct {
	name      string
	processes []Info
	patterns  []string
	expected  []int
} {
	return []struct {
		name      string
		processes []Info
		patterns  []string
		expected  []int
	}{
		{
			"exact name",
			[]Info{newInfoFixtureFor(100, "myapp"), newInfoFixtureFor(200, "worker")},
			[]string{"myapp"},
			[]int{100},
		},
		{
			"substring of the name",
			[]Info{newInfoFixtureFor(100, "myapp-worker")},
			[]string{"app"},
			[]int{100},
		},
		{
			"uppercase process name",
			[]Info{newInfoFixtureFor(100, "MyApp")},
			[]string{"myapp"},
			[]int{100},
		},
		{
			"uppercase pattern",
			[]Info{newInfoFixtureFor(100, "myapp")},
			[]string{"MYAPP"},
			[]int{100},
		},
		{
			"multiple patterns keep input order",
			[]Info{newInfoFixtureFor(100, "myapp"), newInfoFixtureFor(200, "worker"), newInfoFixtureFor(300, "other")},
			[]string{"worker", "myapp"},
			[]int{100, 200},
		},
		{
			"no match",
			[]Info{newInfoFixtureFor(100, "myapp")},
			[]string{"worker"},
			nil,
		},
		{
			"no patterns",
			[]Info{newInfoFixtureFor(100, "myapp")},
			nil,
			nil,
		},
	}
}

func newIncludeRootTestCase() []struct {
	name        string
	includeRoot bool
	expected    []int
} {
	return []struct {
		name        string
		includeRoot bool
		expected    []int
	}{
		{"included when includeRoot is set", true, []int{100, 200}},
		{"excluded when includeRoot is not set", false, []int{100}},
	}
}

func newEmptyProcessListTestCase() []struct {
	name      string
	processes []Info
} {
	return []struct {
		name      string
		processes []Info
	}{
		{"nil list", nil},
		{"empty list", []Info{}},
	}
}

func newEmptyPatternListTestCase() []struct {
	name     string
	patterns []string
} {
	return []struct {
		name     string
		patterns []string
	}{
		{"nil list", nil},
		{"empty list", []string{}},
	}
}

func newOutOfRangePatternTestCase() []struct {
	name     string
	patterns []string
	expected string
} {
	return []struct {
		name     string
		patterns []string
		expected string
	}{
		{"empty pattern", []string{""}, "must be at least"},
		{"below min", []string{strings.Repeat("a", minPatternLength-1)}, "must be at least"},
		{"above max", []string{strings.Repeat("a", maxPatternLength+1)}, "exceeds maximum length"},
		{"one bad pattern among valid ones", []string{"firefox", "ab", "chrome"}, `"ab"`},
	}
}

func newInRangePatternTestCase() []struct {
	name     string
	patterns []string
} {
	return []struct {
		name     string
		patterns []string
	}{
		{"min length", []string{strings.Repeat("a", minPatternLength)}},
		{"max length", []string{strings.Repeat("a", maxPatternLength)}},
		{"mid range", []string{"firefox"}},
		{"several valid patterns", []string{"firefox", "chrome"}},
	}
}

func newRejectedRootTestCase() []struct {
	name      string
	ownUID    int
	allowRoot bool
} {
	return []struct {
		name      string
		ownUID    int
		allowRoot bool
	}{
		{"root without override", 0, false},
	}
}

func newAcceptedRootTestCase() []struct {
	name      string
	ownUID    int
	allowRoot bool
} {
	return []struct {
		name      string
		ownUID    int
		allowRoot bool
	}{
		{"root with override", 0, true},
		{"non-root without override", 1000, false},
		{"non-root with override", 1000, true},
	}
}
