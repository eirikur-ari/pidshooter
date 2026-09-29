package helper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Ptr returns a pointer to a copy of v, for constructing struct literals
// with optional (pointer-typed) fields inline.
func Ptr[T any](v T) *T {
	return &v
}

// UnsetEnv removes key from the environment for the duration of the test,
// restoring its prior value (or leaving it unset) afterward. Unlike
// t.Setenv(key, ""), this exercises the genuinely-absent case rather than
// relying on os.Getenv treating "" and unset identically.
func UnsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, wasSet := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, prev)
		}
	})
}
