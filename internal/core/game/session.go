package game

import (
	"time"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// lifecycle represents the progression of a game session from creation to completion.
type lifecycle int32

const (
	pending lifecycle = iota // not yet started
	running                  // game loop active
	stopped                  // game over
)

// Config holds the gameplay parameters for a session.
type Config struct {
	Confirm   bool
	Speed     float64
	TimeLimit int
}

// Session manages targets and lifecycle state for a single play-through, as
// a pure state machine.
type Session struct {
	cfg      Config
	state    lifecycle
	timer    timer
	confirm  confirmation
	throttle *movement.Throttle
	roster   roster
	bounds   movement.Bounds
}

// NewSession creates a new Session with the given configuration. Call Start before the first Step.
func NewSession(processes []process.Info, cfg Config) *Session {
	return &Session{
		cfg:      cfg,
		roster:   newRoster(processes),
		timer:    newTimer(cfg.TimeLimit),
		throttle: movement.NewThrottle(cfg.Speed),
		confirm:  newConfirmation(cfg.Confirm),
	}
}

// Start transitions the session from pending to running. bounds' chrome
// rows are remembered for the rest of the session, reused by every later
// Update call. Panics if called on a session that is already running or
// stopped.
func (s *Session) Start(bounds movement.Bounds) {
	switch s.state {
	case pending:
		s.bounds = bounds
		s.initialize()
	case running:
		panic("Start called on a running game session")
	case stopped:
		panic("Start called on a stopped game session")
	}
}

// Update advances the session state by one frame. window is the current
// display dimensions. It moves every target and stops the session if the
// time limit has expired or all targets are dead.
func (s *Session) Update(window movement.WindowSize) {
	if s.timer.Expired() {
		s.Stop()
		return
	}

	s.bounds = s.bounds.Update(window)
	s.roster.move(s.bounds, s.throttle.Speed())

	if s.roster.allDead() {
		s.Stop()
	}
}

// IsRunning reports whether the session loop should continue.
func (s *Session) IsRunning() bool { return s.state == running }

// Stop transitions the session to stopped, signaling the loop to exit.
func (s *Session) Stop() { s.state = stopped }

// StartTime returns when the session was started.
func (s *Session) StartTime() time.Time { return s.timer.StartTime() }

// Throttle returns the session's throttle.
func (s *Session) Throttle() *movement.Throttle { return s.throttle }

// Targets returns the live target slice for frame assembly. Callers must not modify it.
func (s *Session) Targets() []*Target { return s.roster.targets }

// AvailableTargets returns the targets still in play — dead targets excluded —
// along with how many are currently alive.
func (s *Session) AvailableTargets() ([]*Target, int) { return s.roster.available() }

// TimeLimit returns the configured time limit in seconds (0 = unlimited).
func (s *Session) TimeLimit() int { return s.timer.LimitSeconds() }

// TimeLeft returns the remaining time in seconds.
func (s *Session) TimeLeft() int { return s.timer.SecondsLeft() }

// PendingConfirm returns the pending kill target, or nil if none is pending.
func (s *Session) PendingConfirm() *Target {
	if !s.confirm.Pending() {
		return nil
	}
	return s.confirm.target
}

// RequestConfirm registers target as pending confirmation (confirm mode) or returns
// it immediately for killing (passthrough mode).
func (s *Session) RequestConfirm(t *Target) *Target { return s.confirm.Request(t) }

func (s *Session) initialize() {
	s.timer.Start()
	s.roster.spawn(s.bounds)
	s.state = running
}

// currentState returns the current lifecycle of the session.
func (s *Session) currentState() lifecycle { return s.state }
