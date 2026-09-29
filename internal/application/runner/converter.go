package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
)

func toPlayRequest(cfg config.GameResult) game.PlayRequest {
	return game.PlayRequest{
		ConfirmMode: cfg.ConfirmMode,
		Speed:       cfg.Speed,
		TimeLimit:   cfg.TimeLimit,
	}
}

func toFindRequest(cfg config.ProcessResult) process.FindRequest {
	return process.FindRequest{
		IncludeRoot: cfg.IncludeRoot,
		AllowRoot:   cfg.AllowRoot,
	}
}
