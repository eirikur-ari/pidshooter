package osprocess

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func newTestProcess(t *testing.T) *process {
	t.Helper()
	path, err := exec.LookPath("ps")
	require.NoError(t, err, "ps not found")
	return &process{psPath: path, timeout: 5 * time.Second}
}

func TestListReturnsResults(t *testing.T) {
	processes, err := newTestProcess(t).List()
	require.NoError(t, err)
	assert.NotEmpty(t, processes)
}

func TestOwnPIDMatchesOSGetpid(t *testing.T) {
	p := newTestProcess(t)
	assert.Equal(t, os.Getpid(), p.OwnPID())
}

func TestOwnUIDMatchesOSGeteuid(t *testing.T) {
	p := newTestProcess(t)
	assert.Equal(t, os.Geteuid(), p.OwnUID())
}

func TestLookupNameReturnsOwnName(t *testing.T) {
	p := newTestProcess(t)
	name, err := p.LookupName(os.Getpid())
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}

func TestParseProcesses(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []outbound.ProcessInfo
	}{
		{
			name:   "single process",
			output: "  UID   PID    RSS COMM\n 1000   123   4096 sleep\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "sleep", Rss: 4096 * 1024, UID: 1000}},
		},
		{
			name:   "collapses internal whitespace runs in the name",
			output: "  UID   PID    RSS COMM\n 1000   123   4096 weird  double   spaces\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "weird double spaces", Rss: 4096 * 1024, UID: 1000}},
		},
		{
			name:   "skips rows with too few fields",
			output: "  UID   PID    RSS COMM\n  123\n 1000   456   4096 sleep\n",
			want:   []outbound.ProcessInfo{{PID: 456, Name: "sleep", Rss: 4096 * 1024, UID: 1000}},
		},
		{
			name:   "skips rows with a non-numeric uid",
			output: "  UID   PID    RSS COMM\n   ab   123   4096 sleep\n",
			want:   nil,
		},
		{
			name:   "skips rows with a non-numeric pid",
			output: "  UID   PID    RSS COMM\n 1000    ab   4096 sleep\n",
			want:   nil,
		},
		{
			name:   "defaults rss to zero on parse failure",
			output: "  UID   PID    RSS COMM\n 1000   123     ab sleep\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "sleep", Rss: 0, UID: 1000}},
		},
		{
			name:   "no data rows",
			output: "  UID   PID    RSS COMM\n",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProcesses([]byte(tt.output))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func writeHangingPS(t *testing.T) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "ps")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexec sleep 10\n"), 0o755))
	return scriptPath
}

func TestListTimesOutWhenPsHangs(t *testing.T) {
	p := &process{psPath: writeHangingPS(t), timeout: 50 * time.Millisecond}

	start := time.Now()
	_, err := p.List()
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second, "List should return once its timeout elapses, not hang for the full ps runtime")
}

func TestLookupNameTimesOutWhenPsHangs(t *testing.T) {
	p := &process{psPath: writeHangingPS(t), timeout: 50 * time.Millisecond}

	start := time.Now()
	_, err := p.LookupName(1)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second, "LookupName should return once its timeout elapses, not hang for the full ps runtime")
}

