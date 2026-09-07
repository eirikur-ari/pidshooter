package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- FindProcesses ---

func TestFindProcessesReturnsErrorWhenProcessDiscoveryFails(t *testing.T) {
	svc := NewService(&fake.Process{ListErr: errors.New("ps failed")})

	_, err := svc.FindProcesses([]string{"foo"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "ps failed")
}

func TestFindProcessesReturnsErrorWhenNoProcessIsFound(t *testing.T) {
	svc := NewService(&fake.Process{})

	processes, err := svc.FindProcesses([]string{"nonexistent"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNoProcessesFound, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.Empty(t, processes)
	assert.Equal(t, "no processes found", err.Error(), "an empty wrapper Message should not change the displayed text")
	require.Error(t, appErr.Unwrap(), "the underlying cause should still be reachable, not discarded")
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
	svc := NewService(fp)

	processes, err := svc.FindProcesses([]string{"target"})

	require.NoError(t, err)
	assert.Equal(t, []process.Info{
		process.NewInfo(100, "target", 4096),
		process.NewInfo(101, "another-target", 2048),
	}, processes)
}

// --- Kill ---

func TestKillReturnsErrorIfPIDIsProtected(t *testing.T) {
	fp := &fake.Process{}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(1, "init", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.False(t, shouldReap)
	require.Error(t, err)
	assert.ErrorContains(t, err, "refusing to kill PID 1", "expected the service to refuse to kill protected PIDs")
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillReturnsErrorAndShouldReapIfProcessLookupByNameFails(t *testing.T) {
	fp := &fake.Process{LookupNameErr: errors.New("ps lookup failed")}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.True(t, shouldReap)
	require.Error(t, err)
	assert.ErrorContains(t, err, "could not verify PID 100: ps lookup failed", "expected the underlying ps error to still be visible")
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillReturnsErrorAndShouldReapIfNameValidationFails(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "somethingElse"}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.True(t, shouldReap)
	assert.ErrorContains(t, err, "pid name mismatch: expected \"target\", got \"somethingElse\"", "expected the service to report a name mismatch")
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillReturnsShouldReapWhenProcessAlreadyExited(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target", KillErr: outbound.NotFoundError{}}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.True(t, shouldReap)
	require.Error(t, err)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
}

func TestKillReturnsKillingProcessWasASuccess(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target"}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.True(t, killed)
	assert.False(t, shouldReap)
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, fp.KilledPIDs)
}
