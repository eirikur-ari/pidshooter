package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

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

func TestToConfirmViewState_ReturnsNilWhenTargetIsNil(t *testing.T) {
	assert.Nil(t, toConfirmViewState(nil))
}

func TestToConfirmViewState_MapsPIDAndName(t *testing.T) {
	// Given
	target := newTargetFixtureFor(42, "dummy")

	// When
	actual := toConfirmViewState(target)

	// Then
	assert.Equal(t, &outbound.ConfirmViewState{PID: 42, Name: "dummy"}, actual)
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
	// Given
	targets := []*game.Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: game.Alive},
		{Info: process.NewInfo(2, "b", 0, 0), State: game.Alive},
		{Info: process.NewInfo(3, "c", 0, 0), State: game.Alive},
	}

	// When
	actual := toTargetViewStates(targets)

	// Then
	require.Len(t, actual, 3)
	assert.Equal(t, "[1 a]", actual[0].Tag)
	assert.Equal(t, "[2 b]", actual[1].Tag)
	assert.Equal(t, "[3 c]", actual[2].Tag)
}

func TestToTargetViewStates_ReturnsEmptySliceWhenTargetsAreNil(t *testing.T) {
	assert.Empty(t, toTargetViewStates(nil))
}

func TestToTargetViewStates_MapsNilTargetToEmptyView(t *testing.T) {
	// Given
	targets := []*game.Target{newTargetFixtureFor(1, "a"), nil}

	// When
	actual := toTargetViewStates(targets)

	// Then
	require.Len(t, actual, 2)
	assert.Equal(t, "[1 a]", actual[0].Tag)
	assert.Equal(t, outbound.TargetViewState{}, actual[1])
}

func TestToHUDViewState_ReturnsEmptyViewWhenTrackerIsNil(t *testing.T) {
	assert.Equal(t, outbound.HUDViewState{}, toHUDViewState(nil))
}

func TestToHUDViewState_MapsTrackerProgress(t *testing.T) {
	// Given
	tracker := newKillTracker(5)
	tracker.recordKill(2048)
	tracker.recordKill(2048)

	// When
	actual := toHUDViewState(tracker)

	// Then
	assert.Equal(t, outbound.HUDViewState{FreedMem: 4096, Kills: 2, HighScore: 5}, actual)
}

func TestToStatusViewState_ReturnsEmptyViewWhenSessionIsNil(t *testing.T) {
	assert.Equal(t, outbound.StatusViewState{}, toStatusViewState(nil, 3))
}

func TestToStatusViewState_MapsSessionState(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Speed: 2.0, TimeLimit: 30})

	// When
	actual := toStatusViewState(session, 3)

	// Then
	require.NotNil(t, actual.TimeLeft)
	assert.Equal(t, 3, actual.Alive)
	assert.Equal(t, 2.0, actual.Speed)
	assert.Equal(t, 30, *actual.TimeLeft)
	assert.Nil(t, actual.Confirming)
}

func TestToStatusViewState_LeavesTimeLeftNilWhenUntimed(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Speed: 2.0})

	// When
	actual := toStatusViewState(session, 1)

	// Then
	assert.Nil(t, actual.TimeLeft)
}

func TestToStatusViewState_IncludesPendingConfirmation(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Confirm: true, Speed: 1.0})
	session.RequestConfirm(session.Targets()[0])

	// When
	actual := toStatusViewState(session, 1)

	// Then
	assert.Equal(t, &outbound.ConfirmViewState{PID: 100, Name: "target"}, actual.Confirming)
}

func TestToFrameViewState_CombinesTargetsHUDAndStatusBar(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Speed: 1.0})
	tracker := newKillTracker(5)
	tracker.recordKill(2048)

	// When
	actual := toFrameViewState(session, tracker)

	// Then
	assert.Len(t, actual.Targets, 1)
	assert.Equal(t, outbound.HUDViewState{FreedMem: 2048, Kills: 1, HighScore: 5}, actual.HUD)
	assert.Equal(t, 1, actual.StatusBar.Alive)
}

func TestToFrameViewState_KeepsHUDWhenSessionIsNil(t *testing.T) {
	// Given
	tracker := newKillTracker(5)
	tracker.recordKill(2048)

	// When
	actual := toFrameViewState(nil, tracker)

	// Then
	assert.Empty(t, actual.Targets)
	assert.Equal(t, outbound.HUDViewState{FreedMem: 2048, Kills: 1, HighScore: 5}, actual.HUD)
	assert.Equal(t, outbound.StatusViewState{}, actual.StatusBar)
}

func TestToFrameViewState_KeepsTargetsAndStatusBarWhenTrackerIsNil(t *testing.T) {
	// Given
	session := newSessionFixture(game.Config{Speed: 1.0})

	// When
	actual := toFrameViewState(session, nil)

	// Then
	assert.Len(t, actual.Targets, 1)
	assert.Equal(t, outbound.HUDViewState{}, actual.HUD)
	assert.Equal(t, 1, actual.StatusBar.Alive)
}

func TestToFrameViewState_LeavesPartsEmptyWhenSessionAndTrackerAreNil(t *testing.T) {
	assert.Equal(t, outbound.FrameViewState{Targets: []outbound.TargetViewState{}}, toFrameViewState(nil, nil))
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
