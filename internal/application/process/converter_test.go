package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// --- toProcessInfo ---

func TestToProcessInfoMapsFields(t *testing.T) {
	info := toProcessInfo(outbound.ProcessInfo{Pid: 42, Name: "suspect", Rss: 1024})

	assert.Equal(t, 42, info.Pid)
	assert.Equal(t, "suspect", info.Name)
	assert.Equal(t, int64(1024), info.Rss)
}

// --- toProcessInfos ---

func TestToProcessInfosMapsAll(t *testing.T) {
	infos := toProcessInfos([]outbound.ProcessInfo{
		{Pid: 1, Name: "a", Rss: 100},
		{Pid: 2, Name: "b", Rss: 200},
	})

	require.Len(t, infos, 2)
	assert.Equal(t, 1, infos[0].Pid)
	assert.Equal(t, "a", infos[0].Name)
	assert.Equal(t, int64(100), infos[0].Rss)
	assert.Equal(t, 2, infos[1].Pid)
	assert.Equal(t, "b", infos[1].Name)
	assert.Equal(t, int64(200), infos[1].Rss)
}

func TestToProcessInfosEmptyInput(t *testing.T) {
	infos := toProcessInfos(nil)

	assert.Empty(t, infos)
}
