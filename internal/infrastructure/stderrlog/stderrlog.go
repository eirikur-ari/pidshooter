// Package stderrlog implements outbound.Logger backed by standard error.
package stderrlog

import (
	"fmt"
	"io"
	"os"
)

// Logger implements outbound.Logger by writing to an io.Writer.
type Logger struct {
	w io.Writer
}

// NewLogger returns a Logger that writes to os.Stderr.
func NewLogger() *Logger {
	return &Logger{w: os.Stderr}
}

// Warn writes msg to the underlying writer, prefixed 'warning: '.
func (l *Logger) Warn(msg string) {
	_, _ = fmt.Fprintf(l.w, "warning: %s\n", msg)
}

// Error writes msg to the underlying writer, prefixed 'error: '.
func (l *Logger) Error(msg string) {
	_, _ = fmt.Fprintf(l.w, "error: %s\n", msg)
}
