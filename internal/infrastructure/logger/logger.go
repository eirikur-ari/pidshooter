package logger

import (
	"fmt"
	"io"
	"os"
)

// Logger logs warnings and errors to standard error.
type Logger struct {
	writer io.Writer
}

// NewLogger returns a Logger that writes to standard error.
func NewLogger() *Logger {
	return &Logger{writer: os.Stderr}
}

// Warn logs message as a warning.
func (l *Logger) Warn(message string) {
	_, _ = fmt.Fprintf(l.writer, "warning: %s\n", message)
}

// Error logs message as an error.
func (l *Logger) Error(message string) {
	_, _ = fmt.Fprintf(l.writer, "error: %s\n", message)
}
