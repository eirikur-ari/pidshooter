package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// newSessionFixture returns a Session constructed with the given processes and config,
// started on a 80x24 board with a 1-row top/bottom chrome reservation.
func newSessionFixture(processes []process.Info, cfg Config) *Session {
	session := NewSession(processes, cfg)
	bounds := newBoundsFixture()
	session.Start(bounds)
	return session
}

// pendingConfirmSessionFixture returns a Session with one process, confirm mode
// enabled, and a pending confirmation already requested on its target.
func pendingConfirmSessionFixture() *Session {
	info := newInfoFixture()
	info.Name = "a"
	session := newSessionFixture([]process.Info{info}, Config{Confirm: true, Speed: 1.0})
	session.RequestConfirm(session.Targets()[0])
	return session
}

func newBoundsFixture() movement.Bounds {
	return movement.NewBounds(
		windowSizeFixture(),
		chromeSizeFixture(),
	)
}

func newInfoFixture() process.Info {
	return process.NewInfo(1234, "test", 1024, 0)
}

func windowSizeFixture() movement.WindowSize {
	return movement.WindowSize{Width: 80, Height: 24}
}

func chromeSizeFixture() movement.ChromeSize {
	return movement.ChromeSize{Top: 1, Bottom: 1}
}

func aliveTargetAt(pid int, name string, x, y float64) *Target {
	return &Target{
		Info:   process.NewInfo(pid, name, 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: x, Y: y}},
		State:  Alive,
	}
}

func newStartedTimer(limitSeconds int) (*timer, *FakeClock) {
	clock := &FakeClock{T: time.Date(2026, time.October, 4, 12, 0, 0, 0, time.UTC)}
	tr := newTimer(limitSeconds, clock.Now)
	tr.Start()
	return &tr, clock
}
