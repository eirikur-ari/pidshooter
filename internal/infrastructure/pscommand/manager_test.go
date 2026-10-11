package pscommand

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

func TestNewManager_ReturnsErrorWhenPsIsUnavailable(t *testing.T) {
	// Given
	t.Setenv("PATH", "")

	// When
	processManager, err := NewManager()

	// Then
	assert.ErrorContains(t, err, "ps not found")
	assert.Nil(t, processManager)
}

func TestManager_Discover_ReturnsProcessesFromPsOutput(t *testing.T) {
	// Given
	host := newHostFixture("linux")
	host.On("run", newDiscoverArgumentsFixture()).Return(newPsOutputFixture(), nil)
	processManager := &manager{host: host}
	expected := []outbound.ProcessInfo{newProcessInfoFixture()}

	// When
	processes, err := processManager.Discover()

	// Then
	require.NoError(t, err)
	assert.Equal(t, expected, processes)
	host.AssertExpectations(t)
}

func TestManager_Discover_QueriesColumnsForOperatingSystem(t *testing.T) {
	tests := newDiscoverColumnsTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			host := newHostFixture(test.goos)
			host.On("run", []string{"-eo", test.expected}).Return(newPsOutputFixture(), nil)
			processManager := &manager{host: host}

			// When
			_, err := processManager.Discover()

			// Then
			require.NoError(t, err)
			host.AssertExpectations(t)
		})
	}
}

func TestManager_Discover_ReturnsErrorWhenPsFails(t *testing.T) {
	// Given
	psErr := errors.New("boom")
	host := newHostFixture("linux")
	host.On("run", newDiscoverArgumentsFixture()).Return(nil, psErr)
	processManager := &manager{host: host}

	// When
	_, err := processManager.Discover()

	// Then
	assert.EqualError(t, err, "failed to collect processes: boom")
	assert.ErrorIs(t, err, psErr)
}

func TestManager_Discover_ReturnsErrorWhenOutputCannotBeParsed(t *testing.T) {
	// Given
	host := newHostFixture("linux")
	host.On("run", newDiscoverArgumentsFixture()).Return([]byte("garbage"), nil)
	processManager := &manager{host: host}

	// When
	_, err := processManager.Discover()

	// Then
	assert.ErrorContains(t, err, "failed to collect processes")
}

func TestManager_OwnPID_ReturnsPIDOfHost(t *testing.T) {
	// Given
	host := &MockHost{}
	host.On("pid").Return(4242)
	processManager := &manager{host: host}

	// When
	actual := processManager.OwnPID()

	// Then
	assert.Equal(t, 4242, actual)
}

func TestManager_OwnUID_ReturnsUIDOfHost(t *testing.T) {
	// Given
	host := &MockHost{}
	host.On("uid").Return(501)
	processManager := &manager{host: host}

	// When
	actual := processManager.OwnUID()

	// Then
	assert.Equal(t, 501, actual)
}

func TestManager_LookupName_ReturnsNameOfProcess(t *testing.T) {
	// Given
	host := newHostFixture("linux")
	host.On("run", newLookupArgumentsFixture(123)).Return(newPsOutputFixture(), nil)
	processManager := &manager{host: host}

	// When
	name, err := processManager.LookupName(123)

	// Then
	require.NoError(t, err)
	assert.Equal(t, "sleep", name)
	host.AssertExpectations(t)
}

func TestManager_LookupName_ReturnsNotFoundErrorWhenProcessIsGone(t *testing.T) {
	tests := newProcessGoneTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			host := newHostFixture("linux")
			host.On("run", newLookupArgumentsFixture(123)).Return(test.output, test.psErr)
			processManager := &manager{host: host}

			// When
			_, err := processManager.LookupName(123)

			// Then
			assert.ErrorAs(t, err, &outbound.NotFoundError{})
			assert.EqualError(t, err, "ps lookup for PID 123 failed: not found")
		})
	}
}

func TestManager_LookupName_ReturnsErrorWhenPsFails(t *testing.T) {
	// Given
	psErr := errors.New("boom")
	host := newHostFixture("linux")
	host.On("run", newLookupArgumentsFixture(123)).Return(nil, psErr)
	processManager := &manager{host: host}

	// When
	_, err := processManager.LookupName(123)

	// Then
	assert.EqualError(t, err, "ps lookup for PID 123 failed: boom")
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
}

func TestManager_LookupName_ReturnsErrorWhenOutputCannotBeParsed(t *testing.T) {
	// Given
	host := newHostFixture("linux")
	host.On("run", newLookupArgumentsFixture(123)).Return([]byte("garbage"), nil)
	processManager := &manager{host: host}

	// When
	_, err := processManager.LookupName(123)

	// Then
	require.Error(t, err)
	assert.NotErrorAs(t, err, &outbound.NotFoundError{})
}

func TestManager_Pin_ReturnsHandleForExistingProcess(t *testing.T) {
	// Given
	process := &MockOsProcess{}
	process.On("Signal", syscall.Signal(0)).Return(nil)
	host := &MockHost{}
	host.On("find", 123).Return(process, nil)
	processManager := &manager{host: host}

	// When
	processHandle, err := processManager.Pin(123)

	// Then
	require.NoError(t, err)
	assert.NotNil(t, processHandle)
	process.AssertExpectations(t)
}

func TestManager_Pin_ReturnsNotFoundErrorAndReleasesProcessWhenItIsGone(t *testing.T) {
	// Given
	process := &MockOsProcess{}
	process.On("Signal", syscall.Signal(0)).Return(os.ErrProcessDone)
	process.On("Release").Return(nil)
	host := &MockHost{}
	host.On("find", 123).Return(process, nil)
	processManager := &manager{host: host}

	// When
	processHandle, err := processManager.Pin(123)

	// Then
	assert.Nil(t, processHandle)
	assert.ErrorAs(t, err, &outbound.NotFoundError{})
	process.AssertExpectations(t)
}

func TestManager_Pin_ReturnsErrorWhenProcessCannotBeFound(t *testing.T) {
	// Given
	findErr := errors.New("boom")
	host := &MockHost{}
	host.On("find", 123).Return(nil, findErr)
	processManager := &manager{host: host}

	// When
	processHandle, err := processManager.Pin(123)

	// Then
	assert.Nil(t, processHandle)
	assert.ErrorIs(t, err, findErr)
}

func TestIsZombie_ReportsZombieStates(t *testing.T) {
	tests := newIsZombieTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := isZombie(test.state)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseProcesses_ParsesRowsAndSkipsMalformedOnes(t *testing.T) {
	tests := newParseProcessesTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual, err := parseProcesses([]byte(test.output))

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseProcesses_RejectsMissingHeader(t *testing.T) {
	// Given
	output := " 1000   123   4096 S    sleep\n"

	// When
	_, err := parseProcesses([]byte(output))

	// Then
	assert.Error(t, err)
}

func newDiscoverColumnsTestCase() []struct {
	name     string
	goos     string
	expected string
} {
	return []struct {
		name     string
		goos     string
		expected string
	}{
		{"darwin uses the short command name", "darwin", "uid,pid,rss,stat,ucomm"},
		{"linux uses the command name", "linux", "uid,pid,rss,stat,comm"},
	}
}

func newProcessGoneTestCase() []struct {
	name   string
	output []byte
	psErr  error
} {
	return []struct {
		name   string
		output []byte
		psErr  error
	}{
		{"zombie", newPsOutputWithStateFixture("ZN"), nil},
		{"absent from the output", []byte(psHeader), nil},
		{"ps matched no process", nil, outbound.NotFoundError{}},
	}
}

func newIsZombieTestCase() []struct {
	name     string
	state    string
	expected bool
} {
	return []struct {
		name     string
		state    string
		expected bool
	}{
		{"running", "R", false},
		{"running in foreground", "R+", false},
		{"sleeping", "S", false},
		{"zombie", "Z", true},
		{"zombie with trailing modifier flags", "ZN", true},
		{"empty", "", false},
	}
}

func newParseProcessesTestCase() []struct {
	name     string
	output   string
	expected []outbound.ProcessInfo
} {
	return []struct {
		name     string
		output   string
		expected []outbound.ProcessInfo
	}{
		{
			"single process",
			psHeader + " 1000   123   4096 S    sleep\n",
			[]outbound.ProcessInfo{newProcessInfoFixture()},
		},
		{
			"collapses internal whitespace runs in the name",
			psHeader + " 1000   123   4096 S    weird  double   spaces\n",
			[]outbound.ProcessInfo{{PID: 123, Name: "weird double spaces", Rss: 4096 * 1024, UID: 1000, State: "S"}},
		},
		{
			"reports a zombie's multi-letter state verbatim",
			psHeader + " 1000   123   4096 ZN   sleep\n",
			[]outbound.ProcessInfo{{PID: 123, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "ZN"}},
		},
		{
			"skips rows with too few fields",
			psHeader + "  123\n 1000   456   4096 S    sleep\n",
			[]outbound.ProcessInfo{{PID: 456, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "S"}},
		},
		{
			"skips rows with a non-numeric uid",
			psHeader + "   ab   123   4096 S    sleep\n",
			nil,
		},
		{
			"skips rows with a non-numeric pid",
			psHeader + " 1000    ab   4096 S    sleep\n",
			nil,
		},
		{
			"skips rows with a non-numeric rss",
			psHeader + " 1000   123     ab S    sleep\n",
			nil,
		},
		{
			"no data rows",
			psHeader,
			nil,
		},
	}
}
