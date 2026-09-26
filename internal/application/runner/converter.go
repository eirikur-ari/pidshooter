package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/util"
)

func toConfigRequest(req inbound.RunRequest) config.Request {
	return config.Request{
		Game: config.GameRequest{
			ConfirmMode: util.ClonePtr(req.Game.ConfirmMode),
			Speed:       util.ClonePtr(req.Game.Speed),
			TimeLimit:   util.ClonePtr(req.Game.TimeLimit),
		},
		Process: config.ProcessRequest{
			IncludeRoot: util.ClonePtr(req.Process.IncludeRoot),
		},
	}
}

func toPlayRequest(result config.Result) game.PlayRequest {
	return game.PlayRequest{
		ConfirmMode: result.Game.ConfirmMode,
		Speed:       result.Game.Speed,
		TimeLimit:   result.Game.TimeLimit,
	}
}
