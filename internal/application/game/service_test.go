package game

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// --- Play ---

func TestServicePlayNoProcessesReturnsClassifiedError(t *testing.T) {
	svc := NewService(nil, nil, nil)

	_, err := svc.Play(inbound.Config{Patterns: []string{"proc"}}, nil, 0)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNoProcessesFound, appErr.Code)
	assert.Equal(t, apperror.SeverityFatal, appErr.Severity)
	assert.Equal(t, "no processes found", err.Error(), "an empty wrapper Message should not change the displayed text")
	require.Error(t, appErr.Unwrap(), "the underlying cause should still be reachable, not discarded")
}

// killerFunc adapts a plain function to the killer interface for tests.
type killerFunc func(target *game.Target) (killed, shouldReap bool, err error)

func (f killerFunc) Kill(target *game.Target) (killed, shouldReap bool, err error) {
	return f(target)
}

// --- applyKillSignals ---

func TestServiceApplyKillsCompletesPendingKill(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Killing, target.State)
	assert.Equal(t, 1, tracker.score.kills)
}

func TestServiceApplyKillsReapsAlreadyKilledTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target, shouldReap: true}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Dead, target.State)
	assert.Equal(t, 0, tracker.score.kills, "reaping an already-gone target should not award a kill")
}

func TestServiceApplyKillsEmptyChannelNoOps(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills) // must not block

	assert.Equal(t, 0, tracker.score.kills)
}

func TestServiceApplyKillsRecordsFailureWithoutMutatingTarget(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 1)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Equal(t, game.Alive, target.State, "a kill failure must leave the target alive")
	assert.Equal(t, 0, tracker.score.kills)
	require.Len(t, tracker.failure.messages, 1)
	assert.Equal(t, "could not kill target (PID 100): operation not permitted", tracker.failure.messages[0])
}

func TestServiceApplyKillsDeduplicatesRepeatedFailuresForSamePID(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())
	kills := make(chan killSignal, 2)

	target := game.NewTarget(info, movement.NewBounds(80, 24))
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}
	kills <- killSignal{target: target, err: errors.New("operation not permitted")}

	tracker := newKillTracker(0)
	svc.applyKillSignals(tracker, kills)

	assert.Len(t, tracker.failure.messages, 1, "a repeat failure for the same PID should not duplicate the entry")
}

// --- frameLoop (async kill end-to-end) ---

// TestServiceFrameLoopAppliesAsyncKillToResult drives a click through the real
// dispatch -> killOrReap goroutine -> applyKillSignals pipeline that Play
// relies on, verifying a confirmed kill is reflected in the tracked result.
func TestServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	killer := killerFunc(func(*game.Target) (bool, bool, error) {
		return true, false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := event.NewDispatcher(game.NewInput(session))

	done := make(chan struct{})
	defer close(done)

	tracker := newKillTracker(0)
	svc.frameLoop(session, tracker, dispatcher, nil, done)

	assert.Equal(t, 1, tracker.score.kills)
	assert.Equal(t, info.Rss, tracker.score.freedMem)
	assert.Empty(t, tracker.failure.messages)
}

// --- killOrReap ---

// TestKillOrReapReportsFailureWithoutReaping covers the two cases that used
// to be silently dropped: a protected-PID refusal and a kill syscall failure
// (both shouldReap=false, err!=nil). The target must stay unreaped and the
// failure must be reported on killSignals rather than swallowed.
func TestKillOrReapReportsFailureWithoutReaping(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	target := game.NewTarget(info, movement.NewBounds(80, 24))
	killer := killerFunc(func(*game.Target) (bool, bool, error) {
		return false, false, errors.New("refusing to kill PID 100")
	})
	svc := NewService(killer, &fake.Renderer{}, fake.NewInputSource())
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	svc.killOrReap(target, killSignals, done)

	sig := <-killSignals
	assert.False(t, sig.shouldReap)
	assert.ErrorContains(t, sig.err, "refusing to kill PID 100")
}

// TestKillOrReapFalseKilledWithNilErrorReportsFallbackFailure covers a
// combination the killer interface permits but osprocess never actually
// returns (killed=false, shouldReap=false, err=nil): killOrReap must still
// report a failure rather than silently treating it as success.
func TestKillOrReapFalseKilledWithNilErrorReportsFallbackFailure(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	target := game.NewTarget(info, movement.NewBounds(80, 24))
	killer := killerFunc(func(*game.Target) (bool, bool, error) {
		return false, false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, fake.NewInputSource())
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	svc.killOrReap(target, killSignals, done)

	sig := <-killSignals
	assert.False(t, sig.shouldReap)
	assert.Error(t, sig.err)
}

// TestKillOrReapReapWithoutErrorStaysSilent verifies the existing silent-reap
// path (process already gone / PID recycled — shouldReap=true, err set) does
// not start reporting a failure now that killOrReap also surfaces them: the
// signal must carry shouldReap and no err, exactly as before this change.
func TestKillOrReapReapWithoutErrorStaysSilent(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	target := game.NewTarget(info, movement.NewBounds(80, 24))
	killer := killerFunc(func(*game.Target) (bool, bool, error) {
		return false, true, errors.New("could not verify PID 100: process not found")
	})
	svc := NewService(killer, &fake.Renderer{}, fake.NewInputSource())
	killSignals := make(chan killSignal, 1)
	done := make(chan struct{})
	defer close(done)

	svc.killOrReap(target, killSignals, done)

	sig := <-killSignals
	assert.True(t, sig.shouldReap)
	assert.NoError(t, sig.err, "a reap outcome must not be reported as a kill failure")
}

// --- registerTermSignalWatcher ---

func TestRegisterTermSignalWatcherTermSignalClosesOnStop(t *testing.T) {
	svc := NewService(nil, &fake.Renderer{}, fake.NewInputSource())

	termSignal, stop := svc.registerTermSignalWatcher()
	stop()

	select {
	case <-termSignal:
	case <-time.After(time.Second):
		t.Fatal("termSignal was not closed after stop()")
	}
}
