package config

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
)

// GameOptions holds explicitly provided game mode values. A nil field means the
// value was not provided.
type GameOptions struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessOptions holds explicitly provided process-discovery values. A nil
// field means the value was not provided.
type ProcessOptions struct {
	IncludeRoot *bool
	// AllowRoot permits running pidshooter itself as root; never read from
	// the stored config, only ever set per invocation.
	AllowRoot *bool
}

// Options holds explicitly provided values.
type Options struct {
	Game    GameOptions
	Process ProcessOptions
}

// Service resolves the run configuration.
type Service struct {
	store   outbound.ConfigStore
	options Options
}

// NewService returns a Service that resolves the run configuration from the
// given store and options when Load is called.
func NewService(store outbound.ConfigStore, options Options) *Service {
	return &Service{store: store, options: options}
}

// Load returns the run configuration: it validates the options, then merges the
// stored config and the options onto the built-in defaults, the options taking
// precedence. Invalid options are rejected with a fatal error before the stored
// config is read. If no config is stored yet, the defaults are persisted for
// next time. Any other problem loading, saving or applying the stored config is
// non-fatal — Load still returns a usable Result.
func (s *Service) Load() (Result, error) {
	if err := s.validate(); err != nil {
		return Result{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}

	storedConfig, err := s.store.Load()
	switch {
	case err == nil:
		return s.applyConfig(storedConfig)
	case errors.As(err, &outbound.NotFoundError{}):
		return s.createConfigFile()
	default:
		return newResult().fromOptions(s.options), apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "config store not loaded", err)
	}
}

// createConfigFile persists the built-in defaults and returns the run
// configuration, reporting a failure to persist as a warning.
func (s *Service) createConfigFile() (Result, error) {
	defaults := newResult()
	result := defaults.fromOptions(s.options)
	if err := s.store.Save(defaults.toConfig()); err != nil {
		return result, apperror.NewError(apperror.CodeStoreFailed, apperror.SeverityWarning, "default config not saved", err)
	}
	return result, nil
}

// applyConfig merges the stored config and the options onto the built-in
// defaults, reporting any rejected stored field as a warning.
func (s *Service) applyConfig(storedConfig outbound.Config) (Result, error) {
	result, err := newResult().apply(storedConfig, s.options)
	if err != nil {
		return result, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityWarning, "", err)
	}
	return result, nil
}

// validate returns an error if any explicitly provided option fails validation;
// unset fields are skipped.
func (s *Service) validate() error {
	if s.options.Game.Speed != nil {
		if err := movement.ValidateSpeed(*s.options.Game.Speed); err != nil {
			return err
		}
	}
	if s.options.Game.TimeLimit != nil {
		if err := game.ValidateTimeLimit(*s.options.Game.TimeLimit); err != nil {
			return err
		}
	}
	return nil
}
