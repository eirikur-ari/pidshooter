package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// defaultKillGracePeriod is how long a new Service waits for kills still in
// progress once the session has ended.
const defaultKillGracePeriod = 5 * time.Second

// Service plays game sessions.
type Service struct {
	killer   processKiller
	renderer outbound.Renderer
	events   outbound.InputEventProvider
	// killGracePeriod is how long to wait, once the session has ended, for
	// kills still in progress to finish before giving up on them.
	killGracePeriod time.Duration
}

// PlayRequest holds the parameters of a play session.
type PlayRequest struct {
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// PlayResult is the outcome of a completed play session.
type PlayResult struct {
	Duration     float64
	LowestSpeed  float64
	Kills        int
	FreedMem     int64
	KillFailures []KillFailure
	Duds         []KillDud
}

// NewService returns a Service that kills through killer, draws with
// renderer, and reads input from events.
func NewService(
	killer processKiller,
	renderer outbound.Renderer,
	events outbound.InputEventProvider,
) *Service {
	return &Service{
		killer:          killer,
		renderer:        renderer,
		events:          events,
		killGracePeriod: defaultKillGracePeriod,
	}
}

// Play runs a game session over processes. highScore is the best score so
// far, and is raised during the session as kills exceed it.
func (s *Service) Play(req PlayRequest, processes []process.Info, highScore int) (PlayResult, error) {
	if err := validateGameConfig(req); err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}
	if err := process.ValidateProcesses(processes); err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	session := game.NewSession(processes, toGameConfig(req))
	tracker := newKillTracker(highScore)

	result, err := s.runPlaySession(session, tracker)
	if err != nil {
		return PlayResult{}, apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", err)
	}

	return result, nil
}

// runPlaySession runs session, recording its progress in tracker, without
// classifying its errors.
func (s *Service) runPlaySession(session *game.Session, tracker *killTracker) (PlayResult, error) {
	play, err := newPlaySession(s.renderer, s.events, s.killer, s.killGracePeriod, session, tracker)
	if err != nil {
		return PlayResult{}, err
	}

	endTime, err := play.run()
	if err != nil {
		return PlayResult{}, err
	}

	return play.result(endTime), nil
}

// validateGameConfig returns an error if the request's speed or time limit is invalid.
func validateGameConfig(req PlayRequest) error {
	if err := movement.ValidateSpeed(req.Speed); err != nil {
		return err
	}
	if err := game.ValidateTimeLimit(req.TimeLimit); err != nil {
		return err
	}
	return nil
}
