package score

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestService_LoadScoreBoard_ReturnsEntriesAndHighScore(t *testing.T) {
	// Given
	store := scoreStoreMockSetup(outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
		{Kills: 8, Date: time.Now()},
		{Kills: 3, Date: time.Now()},
	}}, nil, nil)
	service := NewService(store, &testutil.FakeScoreReporter{})

	// When
	result, err := service.LoadScoreBoard()

	// Then
	require.NoError(t, err)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, 8, result.HighScore)
}

func TestService_LoadScoreBoard_ReturnsWarningAndEmptyBoardWhenBoardCannotBeLoaded(t *testing.T) {
	tests := newLoadFailureTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := scoreStoreMockSetup(outbound.ScoreBoard{}, test.loadErr, nil)
			service := NewService(store, &testutil.FakeScoreReporter{})
			var appErr *apperror.Error

			// When
			result, err := service.LoadScoreBoard()

			// Then
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
			assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
			assert.ErrorIs(t, err, test.loadErr)
			assert.ErrorContains(t, err, "score board not loaded")
			assert.Empty(t, result.Entries)
			assert.Equal(t, 0, result.HighScore)
		})
	}
}

func TestService_RecordScore_SavesBoardAndReturnsSessionBoard(t *testing.T) {
	tests := newRecordOutcomeTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := &MockScoreStore{}
			store.On("Load").Return(test.persisted, nil)
			store.On("Save", boardWithScores(test.expectedSavedScores)).Return(nil)
			service := NewService(store, &testutil.FakeScoreReporter{})

			// When
			result, err := service.RecordScore(test.request)

			// Then
			require.NoError(t, err)
			assert.Len(t, result.Entries, test.expectedResultEntries)
			store.AssertExpectations(t)
		})
	}
}

func TestService_RecordScore_SavesBoardOnlyWhenPersistedBoardIsMissingOrCorrupted(t *testing.T) {
	tests := newSaveAfterLoadFailureTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := scoreStoreMockSetup(outbound.ScoreBoard{}, test.cause, nil)
			service := NewService(store, &testutil.FakeScoreReporter{})

			// When
			result, err := service.RecordScore(RecordRequest{Kills: 1, Duration: 1.0})

			// Then
			store.AssertNumberOfCalls(t, "Save", test.expectedSaveCalls)
			assert.Equal(t, test.expectedErrorText, errorText(err))
			assert.Len(t, result.Entries, 1, "the result is returned even when the save is skipped")
		})
	}
}

func TestService_RecordScore_ReturnsWarningWhenSaveFails(t *testing.T) {
	// Given
	cause := errors.New("disk full")
	store := scoreStoreMockSetup(outbound.ScoreBoard{}, nil, cause)
	service := NewService(store, &testutil.FakeScoreReporter{})
	var appErr *apperror.Error

	// When
	result, err := service.RecordScore(RecordRequest{Kills: 1, Duration: 1.0})

	// Then
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeStoreFailed, appErr.Code)
	assert.Equal(t, apperror.SeverityWarning, appErr.Severity)
	assert.ErrorIs(t, err, cause)
	assert.ErrorContains(t, err, "score board not saved: disk full")
	assert.Len(t, result.Entries, 1, "the result is returned even though the save failed")
}

func TestService_RecordScore_ReportsNewHighScoreOnlyWhenKillsBeatEveryEntry(t *testing.T) {
	tests := newNewHighScoreTestCase()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			store := scoreStoreMockSetup(outbound.ScoreBoard{}, nil, nil)
			service := NewService(store, &testutil.FakeScoreReporter{})
			request := RecordRequest{Entries: []BoardEntry{{Kills: 3}}, Kills: test.kills}

			// When
			result, err := service.RecordScore(request)

			// Then
			require.NoError(t, err)
			assert.Equal(t, test.expected, result.NewHighScore)
		})
	}
}

func TestService_ReportResults_ReportsSummaryViaReporter(t *testing.T) {
	// Given
	reporter := &testutil.FakeScoreReporter{}
	service := NewService(&MockScoreStore{}, reporter)
	request := ReportRequest{Duration: 7.5, Kills: 3, Duds: 1, FreedMem: 4096, Entries: []BoardEntry{{Kills: 2}}, NewHighScore: true}

	// When
	service.ReportResults(request)

	// Then
	require.NotNil(t, reporter.Reported)
	assert.Equal(t, toScoreSummary(request), *reporter.Reported)
}

func newLoadFailureTestCase() []struct {
	name    string
	loadErr error
} {
	return []struct {
		name    string
		loadErr error
	}{
		{"board is not persisted yet", outbound.NotFoundError{}},
		{"persisted data is corrupted", outbound.CorruptedDataError{}},
		{"store fails", errors.New("disk error")},
	}
}

func newRecordOutcomeTestCase() []struct {
	name                  string
	persisted             outbound.ScoreBoard
	request               RecordRequest
	expectedResultEntries int
	expectedSavedScores   int
} {
	return []struct {
		name                  string
		persisted             outbound.ScoreBoard
		request               RecordRequest
		expectedResultEntries int
		expectedSavedScores   int
	}{
		{
			name:                  "records the session",
			persisted:             outbound.ScoreBoard{},
			request:               RecordRequest{Kills: 4, Duds: 1, FreedMem: 2048, LowestSpeed: 2.5, TimeLimit: 30, Duration: 12.5},
			expectedResultEntries: 1,
			expectedSavedScores:   1,
		},
		{
			name:                  "does not record a session without kills",
			persisted:             outbound.ScoreBoard{},
			request:               RecordRequest{Kills: 0, Duration: 1.0},
			expectedResultEntries: 0,
			expectedSavedScores:   0,
		},
		{
			name: "keeps an entry persisted after the board was loaded",
			persisted: outbound.ScoreBoard{Scores: []outbound.ScoreEntry{
				{Kills: 20, Date: time.Now()},
			}},
			request:               RecordRequest{Kills: 5, Duration: 1.0},
			expectedResultEntries: 1,
			expectedSavedScores:   2,
		},
	}
}

func newSaveAfterLoadFailureTestCase() []struct {
	name              string
	cause             error
	expectedSaveCalls int
	expectedErrorText string
} {
	return []struct {
		name              string
		cause             error
		expectedSaveCalls int
		expectedErrorText string
	}{
		{"board is not persisted yet", outbound.NotFoundError{}, 1, ""},
		{"persisted data is corrupted", outbound.CorruptedDataError{}, 1, ""},
		{"store fails", errors.New("disk error"), 0, "score board not saved: disk error"},
	}
}

func newNewHighScoreTestCase() []struct {
	name     string
	kills    int
	expected bool
} {
	return []struct {
		name     string
		kills    int
		expected bool
	}{
		{"more kills than the best entry", 5, true},
		{"as many kills as the best entry", 3, false},
		{"fewer kills than the best entry", 2, false},
		{"no kills", 0, false},
	}
}
