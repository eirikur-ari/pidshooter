package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_FindProcesses_ReturnsFatalErrorAndReportsNothing(t *testing.T) {
	tests := newFindProcessesErrorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			manager := discoveryMockSetup(nil, 0, test.ownUID, test.discoverErr)
			reporter := &testutil.FakeProcessReporter{}
			service := NewService(manager, reporter, test.patterns)
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
			manager.AssertNumberOfCalls(t, "Discover", test.expectedDiscoverCalls)
		})
	}
}

func TestService_FindProcesses_ReturnsProcessesMatchingRequest(t *testing.T) {
	tests := newFindProcessesMatchTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			manager := discoveryMockSetup(test.infos, test.ownPID, test.ownUID, nil)
			service := NewService(manager, &testutil.FakeProcessReporter{}, []string{"target"})

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
	manager := discoveryMockSetup([]outbound.ProcessInfo{
		{PID: 100, Name: "target", Rss: 4096, UID: ownUID},
		{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
	}, 0, ownUID, nil)
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
	tests := newKillErrorTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			manager, _ := pinnedProcessMockSetup(100, test.pinErr, test.lookupName, test.lookupErr, test.killErr)
			service := NewService(manager, &testutil.FakeProcessReporter{}, nil)
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
	manager := &MockProcessManager{}
	service := NewService(manager, &testutil.FakeProcessReporter{}, nil)
	var appError *apperror.Error

	// When
	err := service.Kill(1, "init", true)

	// Then
	require.ErrorAs(t, err, &appError)
	assert.Equal(t, apperror.CodeKillFailed, appError.Code)
	assert.Equal(t, apperror.SeverityWarning, appError.Severity)
	assert.ErrorContains(t, err, "refusing to kill PID 1")
	manager.AssertNotCalled(t, "Pin", mock.Anything)
}

func TestService_Kill_KeepsNotFoundErrorAsCauseWhenProcessIsGone(t *testing.T) {
	tests := newKillNotFoundTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			manager, _ := pinnedProcessMockSetup(100, test.pinErr, test.lookupName, test.lookupErr, test.killErr)
			service := NewService(manager, &testutil.FakeProcessReporter{}, nil)

			// When
			err := service.Kill(100, "target", false)

			// Then
			assert.ErrorAs(t, err, &outbound.NotFoundError{})
		})
	}
}

func TestService_Kill_KillsOnlyVerifiedProcessAndReleasesWhatItPinned(t *testing.T) {
	tests := newKillSideEffectTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			manager, handle := pinnedProcessMockSetup(100, test.pinErr, test.lookupName, test.lookupErr, test.killErr)
			service := NewService(manager, &testutil.FakeProcessReporter{}, nil)

			// When
			_ = service.Kill(100, "target", false)

			// Then
			handle.AssertNumberOfCalls(t, "Kill", test.expectedKillCalls)
			handle.AssertNumberOfCalls(t, "Release", test.expectedReleaseCalls)
		})
	}
}

func TestService_Kill_ReturnsNoErrorWhenProcessIsKilled(t *testing.T) {
	// Given
	manager, _ := pinnedProcessMockSetup(100, nil, "target", nil, nil)
	service := NewService(manager, &testutil.FakeProcessReporter{}, nil)

	// When
	err := service.Kill(100, "target", false)

	// Then
	assert.NoError(t, err)
}

func newFindProcessesErrorTestCase() []struct {
	name                  string
	ownUID                int
	discoverErr           error
	patterns              []string
	expectedCode          apperror.Code
	expectedText          string
	expectedDiscoverCalls int
} {
	return []struct {
		name                  string
		ownUID                int
		discoverErr           error
		patterns              []string
		expectedCode          apperror.Code
		expectedText          string
		expectedDiscoverCalls int
	}{
		{
			name:                  "running as root without allow root is refused before the patterns are validated",
			ownUID:                0,
			patterns:              nil,
			expectedCode:          apperror.CodeInvalidConfig,
			expectedText:          "refusing to run as root",
			expectedDiscoverCalls: 0,
		},
		{
			name:                  "invalid patterns",
			ownUID:                1000,
			patterns:              nil,
			expectedCode:          apperror.CodeInvalidConfig,
			expectedText:          "at least one search pattern is required",
			expectedDiscoverCalls: 0,
		},
		{
			name:                  "process discovery fails",
			ownUID:                1000,
			discoverErr:           errors.New("ps failed"),
			patterns:              []string{"foo"},
			expectedCode:          apperror.CodeProcessNotFound,
			expectedText:          "ps failed",
			expectedDiscoverCalls: 1,
		},
		{
			name:                  "no process matches",
			ownUID:                1000,
			patterns:              []string{"nonexistent"},
			expectedCode:          apperror.CodeProcessNotFound,
			expectedText:          "no processes found",
			expectedDiscoverCalls: 1,
		},
	}
}

func newFindProcessesMatchTestCase() []struct {
	name              string
	infos             []outbound.ProcessInfo
	ownPID            int
	ownUID            int
	request           FindRequest
	expectedProcesses []FindResult
} {
	const ownPID, ownUID = 100, 1000

	return []struct {
		name              string
		infos             []outbound.ProcessInfo
		ownPID            int
		ownUID            int
		request           FindRequest
		expectedProcesses []FindResult
	}{
		{
			name: "only processes whose name contains a pattern",
			infos: []outbound.ProcessInfo{
				{PID: 200, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 201, Name: "another-target", Rss: 2048, UID: ownUID},
				{PID: 202, Name: "unrelated", Rss: 1024, UID: ownUID},
			},
			ownPID: ownPID,
			ownUID: ownUID,
			expectedProcesses: []FindResult{
				{PID: 200, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 201, Name: "another-target", Rss: 2048, UID: ownUID},
			},
		},
		{
			name: "not the program's own process",
			infos: []outbound.ProcessInfo{
				{PID: ownPID, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID},
			},
			ownPID:            ownPID,
			ownUID:            ownUID,
			expectedProcesses: []FindResult{{PID: 101, Name: "another-target", Rss: 2048, UID: ownUID}},
		},
		{
			name: "not processes owned by another user",
			infos: []outbound.ProcessInfo{
				{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 102, Name: "target-other-owner", Rss: 2048, UID: 2000},
			},
			ownPID:            ownPID,
			ownUID:            ownUID,
			expectedProcesses: []FindResult{{PID: 101, Name: "target", Rss: 4096, UID: ownUID}},
		},
		{
			name: "root-owned processes when requested",
			infos: []outbound.ProcessInfo{
				{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 102, Name: "target-root", Rss: 2048, UID: 0},
			},
			ownPID:  ownPID,
			ownUID:  ownUID,
			request: FindRequest{IncludeRoot: true},
			expectedProcesses: []FindResult{
				{PID: 101, Name: "target", Rss: 4096, UID: ownUID},
				{PID: 102, Name: "target-root", Rss: 2048},
			},
		},
		{
			name:              "everything for a root user who allows running as root",
			infos:             []outbound.ProcessInfo{{PID: 101, Name: "target", Rss: 4096, UID: 0}},
			ownPID:            ownPID,
			ownUID:            0,
			request:           FindRequest{AllowRoot: true},
			expectedProcesses: []FindResult{{PID: 101, Name: "target", Rss: 4096}},
		},
	}
}

func newKillErrorTestCase() []struct {
	name         string
	pinErr       error
	lookupName   string
	lookupErr    error
	killErr      error
	expectedCode apperror.Code
	expectedText string
} {
	return []struct {
		name         string
		pinErr       error
		lookupName   string
		lookupErr    error
		killErr      error
		expectedCode apperror.Code
		expectedText string
	}{
		{
			name:         "pin reports the process gone",
			pinErr:       outbound.NotFoundError{},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "pin fails",
			pinErr:       errors.New("operation not permitted"),
			expectedCode: apperror.CodeKillFailed,
			expectedText: "could not pin PID 100: operation not permitted",
		},
		{
			name:         "name lookup fails transiently",
			lookupErr:    errors.New("ps lookup failed"),
			expectedCode: apperror.CodeKillFailed,
			expectedText: "could not verify PID 100: ps lookup failed",
		},
		{
			name:         "name lookup reports the process gone",
			lookupErr:    outbound.NotFoundError{},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "name no longer matches",
			lookupName:   "somethingElse",
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: `PID 100 now belongs to another process: pid name mismatch: expected "target", got "somethingElse"`,
		},
		{
			name:         "kill reports the process gone",
			lookupName:   "target",
			killErr:      outbound.NotFoundError{},
			expectedCode: apperror.CodeProcessNotFound,
			expectedText: "PID 100 already exited",
		},
		{
			name:         "kill fails",
			lookupName:   "target",
			killErr:      errors.New("permission denied"),
			expectedCode: apperror.CodeKillFailed,
			expectedText: "failed to kill PID 100: permission denied",
		},
	}
}

func newKillNotFoundTestCase() []struct {
	name       string
	pinErr     error
	lookupName string
	lookupErr  error
	killErr    error
} {
	return []struct {
		name       string
		pinErr     error
		lookupName string
		lookupErr  error
		killErr    error
	}{
		{name: "pin reports the process gone", pinErr: outbound.NotFoundError{}},
		{name: "name lookup reports the process gone", lookupErr: outbound.NotFoundError{}},
		{name: "kill reports the process gone", lookupName: "target", killErr: outbound.NotFoundError{}},
	}
}

func newKillSideEffectTestCase() []struct {
	name                 string
	pinErr               error
	lookupName           string
	lookupErr            error
	killErr              error
	expectedKillCalls    int
	expectedReleaseCalls int
} {
	return []struct {
		name                 string
		pinErr               error
		lookupName           string
		lookupErr            error
		killErr              error
		expectedKillCalls    int
		expectedReleaseCalls int
	}{
		{name: "pin reports the process gone", pinErr: outbound.NotFoundError{}},
		{name: "pin fails", pinErr: errors.New("operation not permitted")},
		{name: "name lookup fails", lookupErr: errors.New("ps lookup failed"), expectedReleaseCalls: 1},
		{name: "name lookup reports the process gone", lookupErr: outbound.NotFoundError{}, expectedReleaseCalls: 1},
		{name: "name no longer matches", lookupName: "somethingElse", expectedReleaseCalls: 1},
		{name: "kill reports the process gone", lookupName: "target", killErr: outbound.NotFoundError{}, expectedKillCalls: 1, expectedReleaseCalls: 1},
		{name: "kill fails", lookupName: "target", killErr: errors.New("permission denied"), expectedKillCalls: 1, expectedReleaseCalls: 1},
		{name: "kill succeeds", lookupName: "target", expectedKillCalls: 1, expectedReleaseCalls: 1},
	}
}
