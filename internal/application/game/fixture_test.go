package game

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newBoundsFixture() movement.Bounds {
	return movement.NewBounds(
		windowSizeFixture(),
		movement.ChromeSize{Top: 1, Bottom: 1},
	)
}

func windowSizeFixture() movement.WindowSize {
	return movement.WindowSize{Width: 80, Height: 24}
}

func outboundWindowSizeFixture() outbound.WindowSize {
	return outbound.WindowSize{Width: 80, Height: 24}
}

func newInfoFixture() process.Info {
	return process.NewInfo(100, "target", 4096, 0)
}

func newProcessRequestFixture() ProcessRequest {
	return ProcessRequest{PID: 100, Name: "target", Rss: 4096}
}

func newTargetFixture() *game.Target {
	return game.NewTarget(newInfoFixture(), newBoundsFixture())
}

func newTargetFixtureFor(pid int, name string) *game.Target {
	return game.NewTarget(process.NewInfo(pid, name, 4096, 0), newBoundsFixture())
}

func newPlaySessionFixture(t *testing.T, killer processKiller, renderer outbound.Renderer, events outbound.InputEventProvider, config game.Config) *playSession {
	t.Helper()
	play, err := newPlaySession(renderer, events, killer, defaultKillGracePeriod, newSessionFixture(config), newKillTracker(0))
	require.NoError(t, err)
	return play
}

func newSessionFixture(config game.Config) *game.Session {
	session := game.NewSession([]process.Info{newInfoFixture()}, config)
	session.Start(newBoundsFixture())
	return session
}
