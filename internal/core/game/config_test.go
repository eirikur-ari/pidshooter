package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigReturnsDomainDefaults(t *testing.T) {
	assert.Equal(t, Config{Confirm: false, Speed: 2.0, TimeLimit: 30}, DefaultConfig())
}
