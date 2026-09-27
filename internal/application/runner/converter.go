package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
)

func toPlayRequest(result config.Result) game.PlayRequest {
	return game.PlayRequest{
		ConfirmMode: result.Game.ConfirmMode,
		Speed:       result.Game.Speed,
		TimeLimit:   result.Game.TimeLimit,
	}
}
