package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
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

func TestService_FindProcesses_ReturnsProcessesMatchingRequest(t *testing.T) {
	tests := newFindProcessesMatchTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, []string{"target"})

			// When
			processes, err := service.FindProcesses(test.request)

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expectedProcesses, processes)
		})
	}
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

func TestService_Kill_KillsOnlyVerifiedProcessAndReleasesWhatItPinned(t *testing.T) {
	tests := newKillSideEffectTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			service := NewService(test.manager, &testutil.FakeProcessReporter{}, nil)

			// When
			_ = service.Kill(100, "target", false)

			// Then
			assert.Equal(t, test.expectedKilled, test.manager.KilledPIDs)
			assert.Equal(t, test.expectedReleased, test.manager.ReleasedPIDs)
		})
	}
}

func TestService_Kill_ReturnsNoErrorWhenProcessIsKilled(t *testing.T) {
	// Given
	manager := &testutil.FakeProcessManager{LookupNameValue: "target"}
	service := NewService(manager, &testutil.FakeProcessReporter{}, nil)

	// When
	err := service.Kill(100, "target", false)

	// Then
	assert.NoError(t, err)
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

func newFindProcessesMatchTestCases() []struct {
	name              string
	manager           *testutil.FakeProcessManager
	request           FindRequest
	expectedProcesses []FindResult
} {
	const ownPID, ownUID = 100, 1000

	return []struct {
		name              string
		manager           *testutil.FakeProcessManager
		request           FindRequest
		expectedProcesses []FindResult
	}{
		{
			name: "only processes whose name contains a pattern",
			manager: &testutil.FakeProcessManager{
				Infos: []outbound.ProcessInfo{
					{PID: 200, Name: "target", Rss: 4096, UID: ownUID},
					{PID: 201, Name: "another-target", Rss: 2048, UID: ownUID},
					{PID: 202, Name: "unrelated", Rss: 1024, UID: ownUID},
				},
				OwnPIDValue: ownPID,
				OwnUIDValue: ownUID,
			},
			expectedProcesses: []FindResult{
				{PID: 200, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 201, Name: "another-target", Rss: 2048, UID: ownUID},
			},
		},
		{
			name: "not the program's own process",
			manager: &testutil.FakeProcessManager{
				Infos: []outbound.ProcessInfo{
					{PID: ownPID, Name: "target", Rss: 4096, UID: ownUID},
					{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
				},
				OwnPIDValue: ownPID,
				OwnUIDValue: ownUID,
			},
			expectedProcesses: []FindResult{{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID}},
		},
		{
			name: "not processes owned by another user",
			manager: &testutil.FakeProcessManager{
				Infos: []outbound.ProcessInfo{
					{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
					{PID: 102, Name: "target-other-owner", Rss: 2048, UID: 2000},
				},
				OwnPIDValue: ownPID,
				OwnUIDValue: ownUID,
			},
			expectedProcesses: []FindResult{{PID: 101, Name: "target", Rss: 4096, UID: ownUID}},
		},
		{
			name: "root-owned processes when requested",
			manager: &testutil.FakeProcessManager{
				Infos: []outbound.ProcessInfo{
					{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
					{PID: 102, Name: "target-root", Rss: 2048, UID: 0},
				},
				OwnPIDValue: ownPID,
				OwnUIDValue: ownUID,
			},
			request: FindRequest{IncludeRoot: true},
			expectedProcesses: []FindResult{
				{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 102, Name: "target-root", Rss: 2048},
			},
		},
		{
			name: "everything for a root user who allows running as root",
			manager: &testutil.FakeProcessManager{
				Infos:       []outbound.ProcessInfo{{PID: 101, Name: "target", Rss: 4096, UID: 0}},
				OwnPIDValue: ownPID,
				OwnUIDValue: 0,
			},
			request:           FindRequest{AllowRoot: true},
			expectedProcesses: []FindResult{{PID: 101, Name: "target", Rss: 4096}},
		},
	}
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

func newKillSideEffectTestCases() []struct {
	name             string
	manager          *testutil.FakeProcessManager
	expectedKilled   []int
	expectedReleased []int
} {
	return []struct {
		name             string
		manager          *testutil.FakeProcessManager
		expectedKilled   []int
		expectedReleased []int
	}{
		{name: "pin reports the process gone", manager: &testutil.FakeProcessManager{PinErr: outbound.NotFoundError{}}},
		{name: "pin fails", manager: &testutil.FakeProcessManager{PinErr: errors.New("operation not permitted")}},
		{
			name:             "name lookup fails",
			manager:          &testutil.FakeProcessManager{LookupNameErr: errors.New("ps lookup failed")},
			expectedReleased: []int{100},
		},
		{
			name:             "name lookup reports the process gone",
			manager:          &testutil.FakeProcessManager{LookupNameErr: outbound.NotFoundError{}},
			expectedReleased: []int{100},
		},
		{
			name:             "name no longer matches",
			manager:          &testutil.FakeProcessManager{LookupNameValue: "somethingElse"},
			expectedReleased: []int{100},
		},
		{
			name:             "kill reports the process gone",
			manager:          &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: outbound.NotFoundError{}},
			expectedKilled:   []int{100},
			expectedReleased: []int{100},
		},
		{
			name:             "kill fails",
			manager:          &testutil.FakeProcessManager{LookupNameValue: "target", KillErr: errors.New("permission denied")},
			expectedKilled:   []int{100},
			expectedReleased: []int{100},
		},
		{
			name:             "kill succeeds",
			manager:          &testutil.FakeProcessManager{LookupNameValue: "target"},
			expectedKilled:   []int{100},
			expectedReleased: []int{100},
		},
	}
}
