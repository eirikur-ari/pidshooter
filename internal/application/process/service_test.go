package process

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- FindProcesses ---

func TestFindProcessesListError(t *testing.T) {
	svc := NewService(&fake.Process{ListErr: errors.New("ps failed")})

	_, err := svc.FindProcesses([]string{"foo"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeProcessDiscoveryFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.ErrorContains(t, err, "ps failed")
}

func TestFindProcessesNoMatches(t *testing.T) {
	svc := NewService(&fake.Process{})

	processes, err := svc.FindProcesses([]string{"nonexistent"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNoProcessesFound, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.Empty(t, processes)
}

// --- Kill ---

func TestKillProtectedPIDReturnsError(t *testing.T) {
	fp := &fake.Process{}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(1, "init", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.False(t, shouldReap)
	require.Error(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillLookupErrorReturnsShouldReap(t *testing.T) {
	fp := &fake.Process{LookupNameErr: errors.New("ps lookup failed")}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.True(t, shouldReap)
	require.Error(t, err)
	assert.ErrorContains(t, err, "ps lookup failed", "expected the underlying ps error to still be visible")
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillNameMismatchReturnsShouldReap(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "somethingElse"}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.False(t, killed)
	assert.True(t, shouldReap)
	assert.Error(t, err)
	assert.Empty(t, fp.KilledPIDs)
}

func TestKillNameMatchInvokesKill(t *testing.T) {
	fp := &fake.Process{LookupNameValue: "target"}
	svc := NewService(fp)
	target := game.NewTarget(process.NewInfo(100, "target", 0), movement.NewBounds(80, 24))

	killed, shouldReap, err := svc.Kill(target)

	assert.True(t, killed)
	assert.False(t, shouldReap)
	assert.NoError(t, err)
	assert.Equal(t, []int{100}, fp.KilledPIDs)
}
