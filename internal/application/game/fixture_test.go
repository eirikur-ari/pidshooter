package game

import (
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

func newTargetFixture() *game.Target {
	return game.NewTarget(newInfoFixture(), newBoundsFixture())
}

func newTargetFixtureFor(pid int, name string) *game.Target {
	return game.NewTarget(process.NewInfo(pid, name, 4096, 0), newBoundsFixture())
}

func newClickableSessionFixture() (session *game.Session, x, y int) {
	session = newSessionFixture(game.Config{Speed: 1.0})
	x, y = session.Targets()[0].Motion.Position.Rounded()
	return session, x, y
}

func newSessionFixture(config game.Config) *game.Session {
	session := game.NewSession([]process.Info{newInfoFixture()}, config)
	session.Start(newBoundsFixture())
	return session
}
