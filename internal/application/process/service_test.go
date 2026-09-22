package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- FindProcesses ---

func TestFindProcessesReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	svc := NewService(&fake.Process{DiscoverErr: errors.New("ps failed")}, &fake.ProcessReporter{})

	_, err := svc.FindProcesses([]string{"foo"}, false)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "ps failed")
}

func TestFindProcessesReturnsErrorWhenNoProcessIsFound(t *testing.T) {
	svc := NewService(&fake.Process{}, &fake.ProcessReporter{})

	processes, err := svc.FindProcesses([]string{"nonexistent"}, false)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.Empty(t, processes)
	assert.Equal(t, "no processes found", err.Error(), "an empty wrapper Message should not change the displayed text")
	require.Error(t, appErr.Unwrap(), "the underlying cause should still be reachable, not discarded")
}

func TestFindProcessesDoesNotReportWhenDiscoveryFails(t *testing.T) {
	reporter := &fake.ProcessReporter{}
	svc := NewService(&fake.Process{DiscoverErr: errors.New("ps failed")}, reporter)

	_, _ = svc.FindProcesses([]string{"foo"}, false)

	assert.Nil(t, reporter.Reported, "a failed discovery must not be reported as a match count")
}

func TestFindProcessesDoesNotReportWhenNoProcessIsFound(t *testing.T) {
	reporter := &fake.ProcessReporter{}
	svc := NewService(&fake.Process{}, reporter)

	_, _ = svc.FindProcesses([]string{"nonexistent"}, false)

	assert.Nil(t, reporter.Reported, "an empty match set fails validation before ever reaching the report call")
}

func TestFindProcessesReportsMatchCountAndPatterns(t *testing.T) {
	fp := &fake.Process{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096},
			{PID: 101, Name: "another-target", Rss: 2048},
		},
	}
	reporter := &fake.ProcessReporter{}
	svc := NewService(fp, reporter)

	_, err := svc.FindProcesses([]string{"target"}, false)

	require.NoError(t, err)
	require.NotNil(t, reporter.Reported)
	assert.Equal(t, 2, reporter.Reported.Count)
	assert.Equal(t, []string{"target"}, reporter.Reported.Patterns)
}

func TestFindProcessesReturnsMatchingProcesses(t *testing.T) {
	fp := &fake.Process{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096},
			{PID: 101, Name: "another-target", Rss: 2048},
			{PID: 200, Name: "unrelated", Rss: 1024},
		},
		OwnPIDValue: 999,
	}
	svc := NewService(fp, &fake.ProcessReporter{})

	processes, err := svc.FindProcesses([]string{"target"}, false)

	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096, 0),
		process.NewInfo(101, "another-target", 2048, 0),
	}, processes)
}

func TestFindProcessesExcludesProcessesNotOwnedByCaller(t *testing.T) {
	fp := &fake.Process{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096, UID: 1000},
			{PID: 101, Name: "target-other-owner", Rss: 2048, UID: 2000},
		},
		OwnUIDValue: 1000,
	}
	svc := NewService(fp, &fake.ProcessReporter{})

	processes, err := svc.FindProcesses([]string{"target"}, false)

	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096, 1000),
	}, processes)
}

func TestFindProcessesIncludeRootIncludesRootOwnedProcesses(t *testing.T) {
	fp := &fake.Process{
		Infos: []outbound.ProcessInfo{
			{PID: 100, Name: "target", Rss: 4096, UID: 1000},
			{PID: 101, Name: "target-root", Rss: 2048, UID: 0},
		},
		OwnUIDValue: 1000,
	}
	svc := NewService(fp, &fake.ProcessReporter{})

	processes, err := svc.FindProcesses([]string{"target"}, true)

	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096, 1000),
		process.NewInfo(101, "target-root", 2048, 0),
	}, processes)
}

// --- Kill ---

func TestKillReturnsErrorIfPIDIsProtected(t *testing.T) {
	fp := &fake.Process{}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(1, "init", true)

	assert.False(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "refusing to kill PID 1", "expected the service to refuse to kill protected PIDs")
	assert.Empty(t, fp.KilledPIDs)
	assert.Empty(t, fp.ReleasedPIDs, "a protected PID is refused before Pin is ever called")
}

func TestKillReturnsErrorAndShouldReapIfPinFails(t *testing.T) {
	fp := &fake.Process{PinErr: errors.New("could not find process")}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.True(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "could not pin PID 100: could not find process")
	assert.Empty(t, fp.KilledPIDs)
	assert.Empty(t, fp.ReleasedPIDs, "there is no handle to release when Pin itself fails")
}

func TestKillReturnsErrorWithoutReapingIfProcessLookupByNameFailsTransiently(t *testing.T) {
	fp := &fake.Process{LookupNameErr: errors.New("ps lookup failed")}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.False(t, shouldReap, "a transient lookup failure must not be treated as the process already having exited")
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "could not verify PID 100: ps lookup failed", "expected the underlying ps error to still be visible")
	assert.Empty(t, fp.KilledPIDs)
	assert.Equal(t, []int{100}, fp.ReleasedPIDs, "the pinned handle must be released even when LookupName fails afterward")
}

func TestKillReturnsShouldReapIfProcessLookupByNameReportsProcessNotFound(t *testing.T) {
	fp := &fake.Process{LookupNameErr: outbound.NotFoundError{}}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.True(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Empty(t, fp.KilledPIDs)
	assert.Equal(t, []int{100}, fp.ReleasedPIDs, "the pinned handle must be released even when LookupName reports the process gone")
}

func TestKillReturnsErrorAndShouldReapIfNameValidationFails(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "somethingElse"}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.True(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "pid name mismatch: expected \"target\", got \"somethingElse\"", "expected the service to report a name mismatch")
	assert.Empty(t, fp.KilledPIDs)
	assert.Equal(t, []int{100}, fp.ReleasedPIDs, "the pinned handle must be released even when name validation fails afterward")
}

func TestKillReturnsShouldReapWhenProcessAlreadyExited(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target", KillErr: outbound.NotFoundError{}}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.True(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessNotFound, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	assert.Equal(t, []int{100}, fp.ReleasedPIDs)
}

func TestKillReturnsErrorWhenKillFails(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target", KillErr: errors.New("permission denied")}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.False(t, shouldReap)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeKillFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorContains(t, err, "failed to kill PID 100: permission denied")
	assert.Equal(t, []int{100}, fp.ReleasedPIDs)
}

func TestKillReturnsKillingProcessWasASuccess(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target"}
	svc := NewService(fp, &fake.ProcessReporter{})

	shouldReap, err := svc.Kill(100, "target", false)

	assert.False(t, shouldReap)
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, fp.KilledPIDs)
	assert.Equal(t, []int{100}, fp.ReleasedPIDs, "the pinned handle must be released after a successful kill too")
}
