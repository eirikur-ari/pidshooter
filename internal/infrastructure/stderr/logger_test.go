package stderr

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarnWritesPrefixedMessage(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{w: &buf}

	l.Warn("could not load scores: disk full")

	assert.Equal(t, "warning: could not load scores: disk full\n", buf.String())
}

func TestErrorWritesPrefixedMessage(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{w: &buf}

	l.Error("game session failed: renderer initialization failed")

	assert.Equal(t, "error: game session failed: renderer initialization failed\n", buf.String())
}
