package fixture

import (
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Game returns a Game constructed with the given processes and config, started on an 80x24 board.
func Game(processes []process.Info, cfg game.Config) *game.Game {
	g := game.New(processes, cfg)
	g.Start(80, 24)
	return g
}

// ConfirmGame returns a Game with n processes and confirm mode enabled, the standard setup for confirm-flow tests.
func ConfirmGame(n int) *game.Game {
	return Game(Processes(n), game.Config{Confirm: true, Speed: 1.0})
}

// PendingConfirmGame returns a ConfirmGame with n processes and a pending
// confirmation already requested on its first target.
func PendingConfirmGame(n int) *game.Game {
	g := ConfirmGame(n)
	g.RequestConfirm(g.Targets()[0])
	return g
}
