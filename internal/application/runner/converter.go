package runner

import (
	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/application/game"
	"github.com/eirikur-ari/pidshooter/internal/application/process"
	"github.com/eirikur-ari/pidshooter/internal/application/score"
)

func toPlayRequest(cfg config.GameResult, found []process.FindResult, highScore int) game.PlayRequest {
	processes := make([]game.ProcessRequest, len(found))
	for i, f := range found {
		processes[i] = game.ProcessRequest{PID: f.PID, Name: f.Name, Rss: f.Rss, UID: f.UID}
	}

	return game.PlayRequest{
		ConfirmMode: cfg.ConfirmMode,
		Speed:       cfg.Speed,
		TimeLimit:   cfg.TimeLimit,
		Processes:   processes,
		HighScore:   highScore,
	}
}

func toFindRequest(cfg config.ProcessResult) process.FindRequest {
	return process.FindRequest{
		IncludeRoot: cfg.IncludeRoot,
		AllowRoot:   cfg.AllowRoot,
	}
}

func toRecordRequest(loaded score.LoadResult, result game.PlayResult, timeLimit int) score.RecordRequest {
	return score.RecordRequest{
		Entries:     loaded.Entries,
		Kills:       result.Kills,
		Duds:        result.Duds,
		FreedMem:    result.FreedMem,
		LowestSpeed: result.LowestSpeed,
		TimeLimit:   timeLimit,
		Duration:    result.Duration,
	}
}

func toReportRequest(result game.PlayResult, recorded score.RecordResult) score.ReportRequest {
	return score.ReportRequest{
		Duration:     result.Duration,
		Kills:        result.Kills,
		Duds:         result.Duds,
		FreedMem:     result.FreedMem,
		Entries:      recorded.Entries,
		NewHighScore: recorded.NewHighScore,
	}
}
