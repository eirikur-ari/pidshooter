package config

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// GameOptions holds the caller's explicitly-provided game mode values. A
// nil field means the caller did not provide that value.
type GameOptions struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessOptions holds the caller's explicitly-provided process-discovery
// values. A nil field means the caller did not explicitly provide that
// value.
type ProcessOptions struct {
	IncludeRoot *bool
	// AllowRoot permits running pidshooter itself as root; never read from
	// the stored config, only ever set per invocation.
	AllowRoot *bool
}

// Options holds the caller's explicitly-provided values.
type Options struct {
	Game    GameOptions
	Process ProcessOptions
}

// Service resolves run config via the injected outbound.ConfigStore.
type Service struct {
	store outbound.ConfigStore
	opts  Options
}

// NewService constructs a Service around the given outbound.ConfigStore, resolving the caller's options when Load is called.
func NewService(store outbound.ConfigStore, opts Options) *Service {
	return &Service{store: store, opts: opts}
}

// Load returns a Result by validating the options it was constructed with,
// then merging the stored config and those options onto domain defaults,
// the options taking precedence. Invalid options are rejected before the
// stored config is read. Any other problem loading or applying the stored
// config is non-fatal — Load still returns a usable Result.
func (s *Service) Load() (Result, error) {
	if err := s.validate(); err != nil {
		return Result{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}

	config, err := s.store.Load()
	switch {
	case err == nil:
		return s.applyConfig(config)
	case errors.As(err, &outbound.NotFoundError{}):
		// no persisted config yet.
		return newResult().fromOptions(s.opts), nil
	default:
		return newResult().fromOptions(s.opts), apperror.NewError(apperror.CodeStoreLoadFailed, apperror.SeverityWarning, "config store not loaded", err)
	}
}

// applyConfig merges the persisted config onto domain defaults and the
// service's options, classifying any rejected persisted field as a
// SeverityWarning apperror.Error.
func (s *Service) applyConfig(config outbound.Config) (Result, error) {
	result, err := newResult().apply(config, s.opts)
	if err != nil {
		return result, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityWarning, "", err)
	}
	return result, nil
}

// validate returns an error if any explicitly-provided option fails
// validation; unset fields are skipped.
func (s *Service) validate() error {
	if s.opts.Game.Speed != nil {
		if err := movement.ValidateSpeed(*s.opts.Game.Speed); err != nil {
			return err
		}
	}
	if s.opts.Game.TimeLimit != nil {
		if err := game.ValidateTimeLimit(*s.opts.Game.TimeLimit); err != nil {
			return err
		}
	}
	return nil
}
