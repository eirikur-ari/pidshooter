package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestToInfos_MapsEveryProcessInOrder(t *testing.T) {
	tests := newToInfosTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toInfos(test.processes)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToGameConfig_MapsEveryRequestField(t *testing.T) {
	// When
	actual := toGameConfig(PlayRequest{ConfirmMode: true, Speed: 3.0, TimeLimit: 45})

	// Then
	assert.Equal(t, game.Config{Confirm: true, Speed: 3.0, TimeLimit: 45}, actual)
}

func TestToWindowSize_MapsWidthAndHeight(t *testing.T) {
	// When
	actual := toWindowSize(outboundWindowSizeFixture())

	// Then
	assert.Equal(t, windowSizeFixture(), actual)
}

func TestToBounds_MapsWindowAndChromeSizes(t *testing.T) {
	// Given
	expected := movement.NewBounds(windowSizeFixture(), movement.ChromeSize{Top: 2, Bottom: 3})

	// When
	actual := toBounds(outboundWindowSizeFixture(), outbound.ChromeSize{Top: 2, Bottom: 3})

	// Then
	assert.Equal(t, expected, actual)
}

func TestToConfirmViewState_MapsTargetToConfirmation(t *testing.T) {
	tests := newConfirmViewStateTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toConfirmViewState(test.target)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToTargetViewState_MapsTargetState(t *testing.T) {
	tests := newTargetViewStateTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toTargetViewState(test.target)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToTargetViewStates_MapsEveryTargetInOrder(t *testing.T) {
	tests := newTargetViewStatesTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toTargetViewStates(test.targets)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToHUDViewState_MapsTrackerProgress(t *testing.T) {
	tests := newHUDViewStateTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toHUDViewState(test.tracker)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToStatusViewState_MapsSessionState(t *testing.T) {
	tests := newStatusViewStateTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toStatusViewState(test.session, test.alive)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToFrameViewState_CombinesTargetsHUDAndStatusBar(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Speed: 1.0})
	tracker := newKillTracker(5)
	tracker.recordKill(2048)

	// When
	actual := toFrameViewState(session, tracker)

	// Then
	require.Len(t, actual.Targets, 1)
	assert.Equal(t, "[100 target]", actual.Targets[0].Tag)
	assert.Equal(t, outbound.HUDViewState{FreedMem: 2048, Kills: 1, HighScore: 5}, actual.HUD)
	assert.Equal(t, outbound.StatusViewState{Alive: 1, Speed: 1.0}, actual.StatusBar)
}

func TestToFrameViewState_KeepsHUDWhenSessionIsNil(t *testing.T) {
	// Given
	tracker := newKillTracker(5)
	tracker.recordKill(2048)

	// When
	actual := toFrameViewState(nil, tracker)

	// Then
	assert.Equal(t, outbound.FrameViewState{
		Targets: []outbound.TargetViewState{},
		HUD:     outbound.HUDViewState{FreedMem: 2048, Kills: 1, HighScore: 5},
	}, actual)
}

func newTargetViewStateTestCase() []struct {
	name     string
	target   *game.Target
	expected outbound.TargetViewState
} {
	return []struct {
		name     string
		target   *game.Target
		expected outbound.TargetViewState
	}{
		{
			"nil target maps to an empty view",
			nil,
			outbound.TargetViewState{},
		},
		{
			"alive target maps rounded position and tag",
			&game.Target{
				Info:   process.NewInfo(42, "dummy", 0, 0),
				Motion: movement.Motion{Position: movement.Vector{X: 10.6, Y: 5.4}},
				State:  game.Alive,
			},
			outbound.TargetViewState{X: 11, Y: 5, Tag: "[42 dummy]"},
		},
		{
			"killing target maps animation progress",
			&game.Target{
				Info:          process.NewInfo(42, "dummy", 0, 0),
				State:         game.Killing,
				AnimationTick: game.AnimationDuration / 2,
			},
			outbound.TargetViewState{Killing: true, AnimationProgress: 0.5},
		},
		{
			"fleeing target maps animation progress",
			&game.Target{
				Info:          process.NewInfo(42, "dummy", 0, 0),
				State:         game.Fleeing,
				AnimationTick: game.AnimationDuration / 2,
			},
			outbound.TargetViewState{Fleeing: true, AnimationProgress: 0.5},
		},
	}
}

func newToInfosTestCase() []struct {
	name      string
	processes []ProcessRequest
	expected  []process.Info
} {
	return []struct {
		name      string
		processes []ProcessRequest
		expected  []process.Info
	}{
		{
			name: "several processes",
			processes: []ProcessRequest{
				{PID: 1, Name: "a", Rss: 100, UID: 1000},
				{PID: 2, Name: "b", Rss: 200, UID: 2000},
			},
			expected: []process.Info{
				process.NewInfo(1, "a", 100, 1000),
				process.NewInfo(2, "b", 200, 2000),
			},
		},
		{name: "no processes", processes: nil, expected: []process.Info{}},
	}
}

func newConfirmViewStateTestCase() []struct {
	name     string
	target   *game.Target
	expected *outbound.ConfirmViewState
} {
	return []struct {
		name     string
		target   *game.Target
		expected *outbound.ConfirmViewState
	}{
		{"nil target", nil, nil},
		{"target", newTargetFixtureFor(42, "dummy"), &outbound.ConfirmViewState{PID: 42, Name: "dummy"}},
	}
}

func newTargetViewStatesTestCase() []struct {
	name     string
	targets  []*game.Target
	expected []outbound.TargetViewState
} {
	return []struct {
		name     string
		targets  []*game.Target
		expected []outbound.TargetViewState
	}{
		{"no targets", nil, []outbound.TargetViewState{}},
		{
			name: "several targets",
			targets: []*game.Target{
				{Info: process.NewInfo(1, "a", 0, 0), State: game.Alive},
				{Info: process.NewInfo(2, "b", 0, 0), State: game.Alive},
				{Info: process.NewInfo(3, "c", 0, 0), State: game.Alive},
			},
			expected: []outbound.TargetViewState{{Tag: "[1 a]"}, {Tag: "[2 b]"}, {Tag: "[3 c]"}},
		},
		{
			name:     "a nil target among targets",
			targets:  []*game.Target{{Info: process.NewInfo(1, "a", 0, 0), State: game.Alive}, nil},
			expected: []outbound.TargetViewState{{Tag: "[1 a]"}, {}},
		},
	}
}

func newHUDViewStateTestCase() []struct {
	name     string
	tracker  *killTracker
	expected outbound.HUDViewState
} {
	tracker := newKillTracker(5)
	tracker.recordKill(2048)
	tracker.recordKill(2048)

	return []struct {
		name     string
		tracker  *killTracker
		expected outbound.HUDViewState
	}{
		{"nil tracker", nil, outbound.HUDViewState{}},
		{"tracker with kills", tracker, outbound.HUDViewState{FreedMem: 4096, Kills: 2, HighScore: 5}},
	}
}

func newStatusViewStateTestCase() []struct {
	name     string
	session  *game.Session
	alive    int
	expected outbound.StatusViewState
} {
	confirming := newSessionFixture(game.Config{Confirm: true, Speed: 1.0})
	confirming.RequestConfirm(confirming.Targets()[0])

	return []struct {
		name     string
		session  *game.Session
		alive    int
		expected outbound.StatusViewState
	}{
		{"nil session", nil, 3, outbound.StatusViewState{}},
		{
			name:     "timed session",
			session:  newSessionFixture(game.Config{Speed: 2.0, TimeLimit: 30}),
			alive:    3,
			expected: outbound.StatusViewState{Alive: 3, Speed: 2.0, TimeLeft: testutil.Pointer(30)},
		},
		{
			name:     "untimed session",
			session:  newSessionFixture(game.Config{Speed: 2.0}),
			alive:    1,
			expected: outbound.StatusViewState{Alive: 1, Speed: 2.0},
		},
		{
			name:     "session with a pending confirmation",
			session:  confirming,
			alive:    1,
			expected: outbound.StatusViewState{Alive: 1, Speed: 1.0, Confirming: &outbound.ConfirmViewState{PID: 100, Name: "target"}},
		},
	}
}
