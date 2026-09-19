package fixture

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Game returns a Session constructed with the given processes and config,
// started on an 80x24 board with a 1-row top/bottom chrome reservation,
// matching tcellui.TUI.ChromeSize.
func Game(processes []process.Info, cfg game.Config) *game.Session {
	s := game.NewSession(processes, cfg)
	s.Start(movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1}))
	return s
}

// ConfirmGameSession returns a Session with n processes and confirm mode enabled, the standard setup for confirm-flow tests.
func ConfirmGameSession(n int) *game.Session {
	return Game(Processes(n), game.Config{Confirm: true, Speed: 1.0})
}

// PendingConfirmGameSession returns a ConfirmGameSession with n processes and a pending
// confirmation already requested on its first target.
func PendingConfirmGameSession(n int) *game.Session {
	s := ConfirmGameSession(n)
	s.RequestConfirm(s.Targets()[0])
	return s
}
