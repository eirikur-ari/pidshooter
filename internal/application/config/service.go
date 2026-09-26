package config

import (
	"errors"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Service resolves run config via the injected outbound.ConfigStore.
type Service struct {
	store outbound.ConfigStore
}

// NewService constructs a Service wrapping the given outbound.ConfigStore.
func NewService(store outbound.ConfigStore) *Service {
	return &Service{store: store}
}

// Load returns a Result by validating req, then merging the persisted
// config file and req onto domain defaults, req taking precedence. An
// invalid req is rejected immediately, before the config file is even
// read. Any other problem loading or applying the persisted config is
// non-fatal — Load still returns a usable Result.
func (s *Service) Load(req Request) (Result, error) {
	if err := req.validate(); err != nil {
		return Result{}, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}

	config, err := s.store.Load()
	switch {
	case err == nil:
		return s.applyConfig(config, req)
	case errors.As(err, &outbound.NotFoundError{}):
		// no persisted config yet.
		return newResult().fromRequest(req), nil
	default:
		return newResult().fromRequest(req), apperror.NewError(apperror.CodeStoreLoadFailed, apperror.SeverityWarning, "config defaults not loaded", err)
	}
}

// applyConfig merges config onto domain defaults and req, classifying any
// rejected persisted field as a SeverityWarning apperror.Error.
func (s *Service) applyConfig(config outbound.ConfigStoreResult, req Request) (Result, error) {
	result, err := newResult().apply(config, req)
	if err != nil {
		return result, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityWarning, "", err)
	}
	return result, nil
}
