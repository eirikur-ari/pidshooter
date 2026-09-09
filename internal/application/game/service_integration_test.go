//go:build integration

package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

// TestIntegrationServiceFrameLoopAppliesAsyncKillToResult drives a click through the real
// dispatch -> killOrReap goroutine -> applyKillSignals pipeline that Play relies on,
// riding out the full real-time kill animation (~600ms, 12 frame ticks) over the
// actual frame ticker, and verifies a confirmed kill is reflected in the tracked
// result. Bounded by a generous timeout so a regression that stalls the loop fails
// fast with a clear message instead of hanging until the test runner's own timeout.
func TestIntegrationServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096, 0)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	killer := fake.ProcessKiller(func(int, string, bool) (bool, error) {
		return false, nil
	})
	svc := NewService(killer, &fake.Renderer{}, events)
	dispatcher := event.NewDispatcher(game.NewInput(session))

	done := make(chan struct{})
	defer close(done)

	tracker := newKillTracker(0)
	finished := make(chan struct{})
	go func() {
		svc.frameLoop(session, tracker, dispatcher, nil, done)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(6 * time.Second):
		t.Fatal("frameLoop did not finish the kill animation within 6s")
	}

	assert.Equal(t, 1, tracker.score.kills)
	assert.Equal(t, info.Rss, tracker.score.freedMem)
	assert.Empty(t, tracker.failure.failures)
}
