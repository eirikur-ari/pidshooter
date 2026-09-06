//go:build integration

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/application/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestIntegrationServiceFrameLoopAppliesAsyncKillToResult(t *testing.T) {
	info := process.NewInfo(100, "target", 4096)
	session := game.NewSession([]process.Info{info}, game.Config{Speed: 1.0})
	session.Start(80, 24)

	target := session.Targets()[0]
	x, y := int(target.Position.X), int(target.Position.Y)

	events := fake.NewInputSource()
	events.Ch <- outbound.ClickEvent{X: x, Y: y}

	killer := fake.Killer(func(*game.Target) (bool, bool, error) {
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
	assert.Empty(t, tracker.failure.failures)
}
