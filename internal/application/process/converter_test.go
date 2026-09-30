package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// --- toInfo ---

func TestToInfoMapsFields(t *testing.T) {
	info := toInfo(outbound.ProcessInfo{PID: 42, Name: "suspect", Rss: 1024, UID: 1000})

	assert.Equal(t, 42, info.PID)
	assert.Equal(t, "suspect", info.Name)
	assert.Equal(t, int64(1024), info.Rss)
	assert.Equal(t, 1000, info.UID)
}

// --- toInfos ---

func TestToInfosMapsAll(t *testing.T) {
	infos := toInfos([]outbound.ProcessInfo{
		{PID: 1, Name: "a", Rss: 100},
		{PID: 2, Name: "b", Rss: 200},
	})

	require.Len(t, infos, 2)
	assert.Equal(t, 1, infos[0].PID)
	assert.Equal(t, "a", infos[0].Name)
	assert.Equal(t, int64(100), infos[0].Rss)
	assert.Equal(t, 2, infos[1].PID)
	assert.Equal(t, "b", infos[1].Name)
	assert.Equal(t, int64(200), infos[1].Rss)
}

func TestToInfosEmptyInput(t *testing.T) {
	infos := toInfos(nil)

	assert.Empty(t, infos)
}

// --- toProcessInfo ---

func TestToProcessInfoMapsFields(t *testing.T) {
	info := toProcessInfo(process.NewInfo(42, "suspect", 1024, 1000))

	assert.Equal(t, 42, info.PID)
	assert.Equal(t, "suspect", info.Name)
	assert.Equal(t, int64(1024), info.Rss)
	assert.Equal(t, 1000, info.UID)
}

// --- toProcessInfos ---

func TestToProcessInfosMapsAll(t *testing.T) {
	infos := toProcessInfos([]process.Info{
		process.NewInfo(1, "a", 100, 0),
		process.NewInfo(2, "b", 200, 0),
	})

	require.Len(t, infos, 2)
	assert.Equal(t, 1, infos[0].PID)
	assert.Equal(t, "a", infos[0].Name)
	assert.Equal(t, int64(100), infos[0].Rss)
	assert.Equal(t, 2, infos[1].PID)
	assert.Equal(t, "b", infos[1].Name)
	assert.Equal(t, int64(200), infos[1].Rss)
}

func TestToProcessInfosEmptyInput(t *testing.T) {
	infos := toProcessInfos(nil)

	assert.Empty(t, infos)
}
