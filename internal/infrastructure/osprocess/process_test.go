package osprocess

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestDiscoverReturnsResults(t *testing.T) {
	processes, err := newTestProcess(t).Discover()
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
	tests := parseProcessesTestCase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProcesses([]byte(tt.output))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseProcessesRejectsMissingHeader(t *testing.T) {
	_, err := parseProcesses([]byte(" 1000   123   4096 S    sleep\n"))
	require.Error(t, err)
}

func TestIsZombie(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  bool
	}{
		{"running", "R", false},
		{"running in foreground", "R+", false},
		{"sleeping", "S", false},
		{"zombie", "Z", true},
		{"zombie with trailing modifier flags", "ZN", true},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isZombie(tt.state))
		})
	}
}

func TestLookupNameReportsNotFoundForZombie(t *testing.T) {
	p := &process{psPath: writeZombiePS(t, 4242), timeout: 5 * time.Second}

	_, err := p.LookupName(4242)

	require.Error(t, err)
	assert.ErrorAs(t, err, &outbound.NotFoundError{}, "a zombie's PID slot still exists but can't be usefully signaled again, so LookupName should report NotFoundError rather than its stale live name")
	assert.ErrorContains(t, err, "ps lookup for PID 4242 failed: not found", "the error should include the PID and NotFoundError in its message")
}

func TestLookupNameReportsNotFoundWhenPIDIsMissing(t *testing.T) {
	p := &process{psPath: writeEmptyResultPS(t), timeout: 5 * time.Second}

	_, err := p.LookupName(4242)

	require.Error(t, err)
	assert.ErrorAs(t, err, &outbound.NotFoundError{}, "a PID absent from ps output entirely should report the same NotFoundError as a zombie, not a bespoke error")
	assert.ErrorContains(t, err, "ps lookup for PID 4242 failed: not found", "the error should include the PID and NotFoundError in its message")
}

func TestDiscoverTimesOutWhenPsHangs(t *testing.T) {
	p := &process{psPath: writeHangingPS(t), timeout: 50 * time.Millisecond}

	start := time.Now()
	_, err := p.Discover()
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second, "Discover should return once its timeout elapses, not hang for the full ps runtime")
}

func TestLookupNameTimesOutWhenPsHangs(t *testing.T) {
	p := &process{psPath: writeHangingPS(t), timeout: 50 * time.Millisecond}

	start := time.Now()
	_, err := p.LookupName(1)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second, "LookupName should return once its timeout elapses, not hang for the full ps runtime")
}

func TestDiscoverErrorIncludesPsStderr(t *testing.T) {
	p := &process{psPath: writeFailingPS(t), timeout: 5 * time.Second}

	_, err := p.Discover()

	require.Error(t, err)
	assert.ErrorContains(t, err, "permission denied", "the error should surface ps's own stderr text, not just an opaque exit status")
}

func newTestProcess(t *testing.T) *process {
	t.Helper()
	path, err := exec.LookPath("ps")
	require.NoError(t, err, "ps not found")
	return &process{psPath: path, timeout: 5 * time.Second}
}

func writeHangingPS(t *testing.T) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "ps")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexec sleep 10\n"), 0o755))
	return scriptPath
}

func writeFailingPS(t *testing.T) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "ps")
	script := "#!/bin/sh\necho 'ps: unrecognized option, permission denied' >&2\nexit 1\n"
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o755))
	return scriptPath
}

func writeZombiePS(t *testing.T, pid int) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "ps")
	script := fmt.Sprintf("#!/bin/sh\nprintf '  UID   PID    RSS STAT COMM\\n 1000   %d   4096 ZN   sleep\\n'\n", pid)
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o755))
	return scriptPath
}

func writeEmptyResultPS(t *testing.T) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "ps")
	script := "#!/bin/sh\nprintf '  UID   PID    RSS STAT COMM\\n'\n"
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o755))
	return scriptPath
}

func parseProcessesTestCase() []struct {
	name   string
	output string
	want   []outbound.ProcessInfo
} {
	tests := []struct {
		name   string
		output string
		want   []outbound.ProcessInfo
	}{
		{
			name:   "single process",
			output: "  UID   PID    RSS STAT COMM\n 1000   123   4096 S    sleep\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "S"}},
		},
		{
			name:   "collapses internal whitespace runs in the name",
			output: "  UID   PID    RSS STAT COMM\n 1000   123   4096 S    weird  double   spaces\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "weird double spaces", Rss: 4096 * 1024, UID: 1000, State: "S"}},
		},
		{
			name:   "reports a zombie's multi-letter state verbatim",
			output: "  UID   PID    RSS STAT COMM\n 1000   123   4096 ZN   sleep\n",
			want:   []outbound.ProcessInfo{{PID: 123, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "ZN"}},
		},
		{
			name:   "skips rows with too few fields",
			output: "  UID   PID    RSS STAT COMM\n  123\n 1000   456   4096 S    sleep\n",
			want:   []outbound.ProcessInfo{{PID: 456, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "S"}},
		},
		{
			name:   "skips rows with a non-numeric uid",
			output: "  UID   PID    RSS STAT COMM\n   ab   123   4096 S    sleep\n",
			want:   nil,
		},
		{
			name:   "skips rows with a non-numeric pid",
			output: "  UID   PID    RSS STAT COMM\n 1000    ab   4096 S    sleep\n",
			want:   nil,
		},
		{
			name:   "skips rows with a non-numeric rss",
			output: "  UID   PID    RSS STAT COMM\n 1000   123     ab S    sleep\n",
			want:   nil,
		},
		{
			name:   "no data rows",
			output: "  UID   PID    RSS STAT COMM\n",
			want:   nil,
		},
	}
	return tests
}
