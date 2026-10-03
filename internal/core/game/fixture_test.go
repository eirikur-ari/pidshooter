package game

import (
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// newStartedSession returns a Session constructed with the given processes and config,
// started on a 80x24 board with a 1-row top/bottom chrome reservation.
func newStartedSession(processes []process.Info, cfg Config) *Session {
	session := NewSession(processes, cfg)
	bounds := movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1})
	session.Start(bounds)
	return session
}

// pendingConfirmSession returns a Session with one process, confirm mode
// enabled, and a pending confirmation already requested on its target.
func pendingConfirmSession() *Session {
	session := newStartedSession([]process.Info{process.NewInfo(1, "a", 0, 0)}, Config{Confirm: true, Speed: 1.0})
	session.RequestConfirm(session.Targets()[0])
	return session
}
