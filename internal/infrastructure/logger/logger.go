package logger

import (
	"fmt"
	"os"
)

// Logger writes leveled diagnostic messages to the process's standard streams.
type Logger struct{}

// NewLogger returns a Logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Warn reports msg as a warning.
func (l *Logger) Warn(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "warning: %s\n", msg)
}

// Error reports msg as an error.
func (l *Logger) Error(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", msg)
}
