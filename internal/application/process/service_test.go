package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_FindProcesses_ReturnsFatalErrorAndReportsNothing(t *testing.T) {
	tests := newFindProcessesErrorTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			reporter := &testutil.FakeProcessReporter{}
			service := NewService(test.manager, reporter, test.patterns)
			var appError *apperror.Error

			// When
			processes, err := service.FindProcesses(FindRequest{})

			// Then
			require.ErrorAs(t, err, &appError)
			assert.Equal(t, test.expectedCode, appError.Code)
			assert.Equal(t, apperror.SeverityFatal, appError.Severity)
			assert.ErrorContains(t, err, test.expectedText)
			assert.Nil(t, processes)
			assert.Nil(t, reporter.Reported)
		})
	}
}

func TestService_FindProcesses_AllowsRootWhenRequested(t *testing.T) {
	// Given
	manager := &testutil.FakeProcessManager{
		Infos:       []outbound.ProcessInfo{{PID: 100, Name: "target", Rss: 4096}},
		OwnUIDValue: 0,
	}
	service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

	// When
	processes, err := service.FindProcesses(FindRequest{AllowRoot: true})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []process.Info{process.NewInfo(100, "target", 4096, 0)}, processes)
}

func TestService_FindProcesses_ReturnsMatchingProcesses(t *testing.T) {
	// Given
	manager := &testutil.FakeProcessManager{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096},
			{PID: 101, Name: "another-target", Rss: 2048},
			{PID: 200, Name: "unrelated", Rss: 1024},
		},
	}
	service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

	// When
	processes, err := service.FindProcesses(FindRequest{AllowRoot: true})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096, 0),
		process.NewInfo(101, "another-target", 2048, 0),
	}, processes)
}

func TestService_FindProcesses_ExcludesProgramsOwnProcess(t *testing.T) {
	// Given
	const ownPID, ownUID = 100, 1000
	manager := &testutil.FakeProcessManager{
		Infos: []outbound.ProcessInfo{
			{PID: ownPID, Name: "target", Rss: 4096, UID: ownUID},
			{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
		},
		OwnPIDValue: ownPID,
		OwnUIDValue: ownUID,
	}
	service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

	// When
	processes, err := service.FindProcesses(FindRequest{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []process.Info{process.NewInfo(101, "another-target", 2048, ownUID)}, processes)
}

func TestService_FindProcesses_ExcludesProcessesNotOwnedByCurrentUser(t *testing.T) {
	// Given
	const ownUID = 1000
	manager := &testutil.FakeProcessManager{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096, UID: ownUID},
			{PID: 101, Name: "target-other-owner", Rss: 2048, UID: 2000},
		},
		OwnUIDValue: ownUID,
	}
	service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

	// When
	processes, err := service.FindProcesses(FindRequest{})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []process.Info{process.NewInfo(100, "target", 4096, ownUID)}, processes)
}

func TestService_FindProcesses_IncludesRootOwnedProcessesWhenRequested(t *testing.T) {
	// Given
	const ownUID = 1000
	manager := &testutil.FakeProcessManager{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096, UID: ownUID},
			{PID: 101, Name: "target-root", Rss: 2048, UID: 0},
		},
		OwnUIDValue: ownUID,
	}
	service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

	// When
	processes, err := service.FindProcesses(FindRequest{IncludeRoot: true})

	// Then
	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096, ownUID),
		process.NewInfo(101, "target-root", 2048, 0),
	}, processes)
}

func TestService_FindProcesses_ReportsMatchesAndPatterns(t *testing.T) {
	// Given
	const ownUID = 1000
	manager := &testutil.FakeProcessManager{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096, UID: ownUID},
			{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
		},
		OwnUIDValue: ownUID,
	}
	reporter := &testutil.FakeProcessReporter{}
	service := NewService(manager, reporter, []string{"target"})

	// When
	_, err := service.FindProcesses(FindRequest{})

	// Then
	require.NoError(t, err)
	require.NotNil(t, reporter.Reported)
	assert.Equal(t, []outbound.ProcessInfo{
		{PID: 100, Name: "target", Rss: 4096, UID: ownUID},
		{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
	}, reporter.Reported.Matches)
	assert.Equal(t, []string{"target"}, reporter.Reported.Patterns)
}

func TestService_Kill_ReturnsWarning(t *testing.T) {
	tests := newKillErrorTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, nil)
			var appError *apperror.Error

			// When
			err := service.Kill(100, "target", false)

			// Then
			require.ErrorAs(t, err, &appError)
			assert.Equal(t, test.expectedCode, appError.Code)
			assert.Equal(t, apperror.SeverityWarning, appError.Severity)
			assert.ErrorContains(t, err, test.expectedText)
		})
	}
}

func TestService_Kill_RefusesProtectedProcessWithoutPinningIt(t *testing.T) {
	// Given
	manager := &testutil.FakeProcessManager{LookupNameValue: "init"}
	service := NewService(manager, &testutil.FakeProcessReporter{}, nil)
	var appError *apperror.Error

	// When
	err := service.Kill(1, "init", true)

	// Then
	require.ErrorAs(t, err, &appError)
	assert.Equal(t, apperror.CodeKillFailed, appError.Code)
	assert.Equal(t, apperror.SeverityWarning, appError.Severity)
	assert.ErrorContains(t, err, "refusing to kill PID 1")
	assert.Empty(t, manager.KilledPIDs)
	assert.Empty(t, manager.ReleasedPIDs)
}

func TestService_Kill_KeepsNotFoundErrorAsCauseWhenProcessIsGone(t *testing.T) {
	tests := newKillNotFoundTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, nil)

			// When
			err := service.Kill(100, "target", false)

			// Then
			assert.ErrorAs(t, err, &outbound.NotFoundError{})
		})
	}
}

func TestService_Kill_ReleasesPinnedProcessOnEveryOutcome(t *testing.T) {
	tests := newKillReleaseTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, nil)

			// When
			_ = service.Kill(100, "target", false)

			// Then
			assert.Equal(t, []int{100}, test.manager.ReleasedPIDs)
		})
	}
}

func TestService_Kill_DoesNotKillWhenProcessCannotBeVerified(t *testing.T) {
	tests := newKillUnverifiedTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, nil)

			// When
			_ = service.Kill(100, "target", false)

			// Then
			assert.Empty(t, test.manager.KilledPIDs)
		})
	}
}

func TestService_Kill_KillsAndReleasesPinnedProcess(t *testing.T) {
	// Given
	manager := &testutil.FakeProcessManager{LookupNameValue: "target"}
	service := NewService(manager, &testutil.FakeProcessReporter{}, nil)

	// When
	err := service.Kill(100, "target", false)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, manager.KilledPIDs)
	assert.Equal(t, []int{100}, manager.ReleasedPIDs)
}

func newFindProcessesErrorTestCases() []struct {
	name         string
	manager      *testutil.FakeProcessManager
	patterns     []string
	expectedCode apperror.Code
	expectedText string
} {
	tests := []struct {
		name         string
		manager      *testutil.FakeProcessManager
		patterns     []string
		expectedCode apperror.Code
		expectedText string
	}{
		{
			name:         "running as root without allow root is refused before the patterns are validated",
			manager:      &testutil.FakeProcessManager{OwnUIDValue: 0},
			patterns:     nil,
			expectedCode: apperror.CodeInvalidConfig,
			expectedText: "refusing to run as root",
		},
		{
			name:         "invalid patterns",
			manager:      &testutil.FakeProcessManager{OwnUIDValue: 1000},
			patterns:     nil,
			expectedCode: apperror.CodeInvalidConfig,
			expectedText: "at least one search pattern is required",
		},
		{
			name:         "process discovery fails",
			manager:      &testutil.FakeProcessManager{DiscoverErr: errors.New("ps failed"), OwnUIDValue: 1000},
			patterns:     []string{"foo"},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "ps failed",
		},
		{
			name:         "no process matches",
			manager:      &testutil.FakeProcessManager{OwnUIDValue: 1000},
			patterns:     []string{"nonexistent"},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "no processes found",
		},
	}
	return tests
}

func newKillErrorTestCases() []struct {
	name         string
	manager      *testutil.FakeProcessManager
	expectedCode apperror.Code
	expectedText string
} {
	tests := []struct {
		name         string
		manager      *testutil.FakeProcessManager
		expectedCode apperror.Code
		expectedText string
	}{
		{
			name:         "pin reports the process gone",
			manager:      &testutil.FakeProcessManager{PinErr: outbound.NotFoundError{}},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "pin fails",
			manager:      &testutil.FakeProcessManager{PinErr: errors.New("operation not permitted")},
			expectedCode: apperror.CodeKillFailed,
			expectedText: "could not pin PID 100: operation not permitted",
		},
		{
			name:         "name lookup fails transiently",
			manager:      &testutil.FakeProcessManager{LookupNameErr: errors.New("ps lookup failed")},
			expectedCode: apperror.CodeKillFailed,
			expectedText: "could not verify PID 100: ps lookup failed",
		},
		{
			name:         "name lookup reports the process gone",
			manager:      &testutil.FakeProcessManager{LookupNameErr: outbound.NotFoundError{}},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "name no longer matches",
			manager:      &testutil.FakeProcessManager{LookupNameValue: "somethingElse"},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: `PID 100 now belongs to another process: pid name mismatch: expected "target", got "somethingElse"`,
		},
		{
			name:         "kill reports the process gone",
			manager:      &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: outbound.NotFoundError{}},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "kill fails",
			manager:      &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: errors.New("permission denied")},
			expectedCode: apperror.CodeKillFailed,
			expectedText: "failed to kill PID 100: permission denied",
		},
	}
	return tests
}

func newKillNotFoundTestCases() []struct {
	name    string
	manager *testutil.FakeProcessManager
} {
	tests := []struct {
		name    string
		manager *testutil.FakeProcessManager
	}{
		{
			name:    "pin reports the process gone",
			manager: &testutil.FakeProcessManager{PinErr: outbound.NotFoundError{}},
		},
		{
			name:    "name lookup reports the process gone",
			manager: &testutil.FakeProcessManager{LookupNameErr: outbound.NotFoundError{}},
		},
		{
			name:    "kill reports the process gone",
			manager: &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: outbound.NotFoundError{}},
		},
	}
	return tests
}

func newKillReleaseTestCases() []struct {
	name    string
	manager *testutil.FakeProcessManager
} {
	tests := []struct {
		name    string
		manager *testutil.FakeProcessManager
	}{
		{
			name:    "name lookup fails",
			manager: &testutil.FakeProcessManager{LookupNameErr: errors.New("ps lookup failed")},
		},
		{
			name:    "name lookup reports the process gone",
			manager: &testutil.FakeProcessManager{LookupNameErr: outbound.NotFoundError{}},
		},
		{
			name:    "name no longer matches",
			manager: &testutil.FakeProcessManager{LookupNameValue: "somethingElse"},
		},
		{
			name:    "kill reports the process gone",
			manager: &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: outbound.NotFoundError{}},
		},
		{
			name:    "kill fails",
			manager: &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: errors.New("permission denied")},
		},
		{
			name:    "kill succeeds",
			manager: &testutil.FakeProcessManager{LookupNameValue: "target"},
		},
	}
	return tests
}

func newKillUnverifiedTestCases() []struct {
	name    string
	manager *testutil.FakeProcessManager
} {
	tests := []struct {
		name    string
		manager *testutil.FakeProcessManager
	}{
		{
			name:    "name lookup fails",
			manager: &testutil.FakeProcessManager{LookupNameErr: errors.New("ps lookup failed")},
		},
		{
			name:    "name lookup reports the process gone",
			manager: &testutil.FakeProcessManager{LookupNameErr: outbound.NotFoundError{}},
		},
		{
			name:    "name no longer matches",
			manager: &testutil.FakeProcessManager{LookupNameValue: "somethingElse"},
		},
	}
	return tests
}
