package testutil

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

// Pointer returns a pointer to a copy of the given value, for constructing
// struct literals with optional (pointer-typed) fields inline.
func Pointer[T any](v T) *T {
	return new(v)
}

// UnsetEnv removes key from the environment for the duration of the test,
// restoring its prior value (or leaving it unset) afterward.
func UnsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, wasSet := os.LookupEnv(key)
	// Unsetenv, not Setenv(key, ""), to exercise the genuinely-absent case
	// rather than relying on os.Getenv treating "" and unset identically.
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, prev)
		}
	})
}

// WarmUpSignalPackage starts the signal package's process-wide goroutine, so
// that a testing/synctest bubble created afterward does not own it. Call it
// before synctest.Test when the code under test watches for signals.
func WarmUpSignalPackage() {
	_, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	stop()
}
