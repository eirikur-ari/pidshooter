package process

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestToInfo_MapsEveryField(t *testing.T) {
	// When
	actual := toInfo(outbound.ProcessInfo{PID: 42, Name: "suspect", Rss: 1024, UID: 1000})

	// Then
	assert.Equal(t, process.NewInfo(42, "suspect", 1024, 1000), actual)
}

func TestToInfos_MapsEveryElementInOrder(t *testing.T) {
	// Given
	expected := []process.Info{
		process.NewInfo(1, "a", 100, 1000),
		process.NewInfo(2, "b", 200, 2000),
	}

	// When
	actual := toInfos([]outbound.ProcessInfo{
		{PID: 1, Name: "a", Rss: 100, UID: 1000},
		{PID: 2, Name: "b", Rss: 200, UID: 2000},
	})

	// Then
	assert.Equal(t, expected, actual)
}

func TestToInfos_ReturnsEmptyForNoInfos(t *testing.T) {
	// When
	actual := toInfos(nil)

	// Then
	assert.Empty(t, actual)
}

func TestToProcessInfo_MapsEveryField(t *testing.T) {
	// When
	actual := toProcessInfo(process.NewInfo(42, "suspect", 1024, 1000))

	// Then
	assert.Equal(t, outbound.ProcessInfo{PID: 42, Name: "suspect", Rss: 1024, UID: 1000}, actual)
}

func TestToProcessInfos_MapsEveryElementInOrder(t *testing.T) {
	// Given
	expected := []outbound.ProcessInfo{
		{PID: 1, Name: "a", Rss: 100, UID: 1000},
		{PID: 2, Name: "b", Rss: 200, UID: 2000},
	}

	// When
	actual := toProcessInfos([]process.Info{
		process.NewInfo(1, "a", 100, 1000),
		process.NewInfo(2, "b", 200, 2000),
	})

	// Then
	assert.Equal(t, expected, actual)
}

func TestToProcessInfos_ReturnsEmptyForNoInfos(t *testing.T) {
	// When
	actual := toProcessInfos(nil)

	// Then
	assert.Empty(t, actual)
}
