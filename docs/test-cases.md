# Test Cases

Total: 30 test files across 14 packages — 244 test functions (including 20 subtests).

---

## Core Domain Layer

### `internal/core/process/process_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestValidateNoPatterns` | `Validate` returns `ErrNoPatterns` for both `nil` and empty pattern slices |
| `TestValidateTooShort` | Patterns of 0–2 characters (below `MinPatternLength`) are rejected |
| `TestValidateExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestValidateValid` | A normal pattern (`"firefox"`) is accepted |
| `TestValidateExactMaxLength` | Pattern at exactly `MaxPatternLength` is accepted |
| `TestValidateTooLong` | Pattern exceeding `MaxPatternLength` is rejected |
| `TestErrNoPatternsIsSentinel` | `ErrNoPatterns` satisfies `errors.Is` against itself |
| `TestInfoIsProtected` | PID 0 and PID 1 are protected regardless of name; any other PID is not |
| `TestFindMatchesByName` | Exact name match returns only the matching process |
| `TestFindSubstringMatch` | Partial name match is supported |
| `TestFindCaseInsensitive` | Name matching is case-insensitive |
| `TestFindMultipleTerms` | OR-semantics: multiple patterns each match independently |
| `TestFindNoMatch` | No matching name returns an empty slice |
| `TestFindExcludesPID1` | PID 1 (init) is always excluded from results |
| `TestFindExcludesOwnPID` | The caller's own PID is excluded from results |
| `TestFindExcludesPID0` | PID 0 (kernel swapper) is excluded alongside PID 1 |

### `internal/core/movement/throttle_test.go` (8 tests)

| Test | Description |
|------|-------------|
| `TestNewThrottle` | `NewThrottle` stores the initial speed |
| `TestThrottleIncrease` | `Increase` raises speed by 0.5 |
| `TestThrottleDecrease` | `Decrease` lowers speed by 0.5 |
| `TestThrottleIncreaseCapsAtMax` | Speed is capped at `MaxSpeed` and does not exceed it on repeated increases |
| `TestThrottleDecreaseFloorsAtMin` | Speed is floored at `MinSpeed` and does not go below it on repeated decreases |
| `TestThrottleLowestSpeedStartsAtInitialSpeed` | `LowestSpeed` starts equal to the throttle's initial speed |
| `TestThrottleLowestSpeedTracksDecreases` | `LowestSpeed` remains the smallest value ever reached, even after subsequent increases |
| `TestThrottleLowestSpeedUnaffectedByIncreaseOnly` | `LowestSpeed` stays at the initial speed when only `Increase` is ever called |

### `internal/core/movement/bounds_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestBoundsBounceLeft` | Position/velocity crossing the left wall clamps to 0 and reverses X velocity |
| `TestBoundsBounceRight` | Position/velocity crossing the right wall (accounting for label width) clamps and reverses X velocity |
| `TestBoundsBounceTop` | Position/velocity crossing the top wall clamps to 0 and reverses Y velocity |
| `TestBoundsBounceBottom` | Position/velocity crossing the bottom wall clamps and reverses Y velocity |

### `internal/core/movement/motion_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestMotionMoveAdvancesPosition` | `Move` advances position by velocity scaled by speed, with no wall interaction |
| `TestMotionMoveBouncesAtLeftWall` | Motion starting at the left wall bounces: position stays in bounds, velocity flips positive |
| `TestMotionMoveBouncesAtRightWall` | Motion approaching the right wall bounces: position clamped, velocity flips negative |
| `TestMotionMoveSpeedScalesVelocity` | Position advances by `velocity × speed` per `Move` call |

### `internal/core/movement/speed_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestNewSpeed` | `newSpeed` initialises both `Current` and `Lowest` to the given starting value |
| `TestSpeedSetTracksLowest` | `Set` updates `Current` and lowers `Lowest` when the new value is smaller |
| `TestSpeedSetAboveLowestDoesNotRaiseLowest` | Setting a higher value updates `Current` but leaves `Lowest` at the previously recorded minimum |
| `TestSpeedSetAboveStartNeverLowersLowest` | Setting a value above the initial value updates `Current` but `Lowest` never rises above the starting value |

### `internal/core/game/lifecycle_test.go` (2 tests)

| Test | Description |
|------|-------------|
| `TestAtomicLifecycleDefaultIsPending` | The zero-value `atomicLifecycle` reads as `pending` |
| `TestAtomicLifecycleStoreLoad` | `Store` followed by `Load` round-trips `running` and `stopped` |

### `internal/core/game/confirmation_test.go` (5 tests)

| Test | Description |
|------|-------------|
| `TestConfirmationPendingFalseWhenEmpty` | A freshly constructed `confirmation` has no pending target |
| `TestConfirmationRequestConfirmModeHoldsTarget` | In confirm mode, `Request` holds the target and returns `nil` |
| `TestConfirmationRequestPassthroughModeReturnsTarget` | Outside confirm mode, `Request` returns the target immediately without holding it |
| `TestConfirmationAcceptReturnsAndClearsTarget` | `Accept` returns the held target and clears pending state |
| `TestConfirmationCancelClearsPending` | `Cancel` clears the pending target without returning it |

### `internal/core/game/timer_test.go` (10 tests)

| Test | Description |
|------|-------------|
| `TestTimerExpiredFalseBeforeLimit` | A freshly started timer with a time limit has not expired |
| `TestTimerExpiredTrueWhenLimitReached` | Timer reports expired once elapsed time passes the limit |
| `TestTimerExpiredFalseWhenUnlimited` | A zero-limit ("unlimited") timer never expires, however long it has run |
| `TestTimerRemainingWithinLimit` | `Remaining` returns a positive duration no greater than the configured limit |
| `TestTimerRemainingZeroWhenExpired` | `Remaining` returns 0 once the limit has passed |
| `TestTimerRemainingZeroWhenUnlimited` | `Remaining` returns 0 for an unlimited timer |
| `TestTimerSecondsLeft` | `SecondsLeft` returns a value between 1 and the configured limit |
| `TestTimerLimitSeconds` | `LimitSeconds` reflects the configured limit, including 0 for unlimited |
| `TestTimerStartTimeZeroBeforeStart` | `StartTime` is zero before `Start` is called |
| `TestTimerStartTimeSetAfterStart` | `StartTime` is set to (at or after) the moment `Start` was called |

### `internal/core/game/session_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestNewSession` | `NewSession` sets `Confirm`, throttle speed, and `TimeLimit`; the session starts in the `pending` state |
| `TestStartTransitionsToRunning` | `Start` makes `IsRunning` true and spawns one target per process |
| `TestStopTransitionsToStopped` | `Stop` transitions the internal lifecycle to `stopped` |
| `TestStartPanicsWhenRunning` | Calling `Start` twice (already running) panics |
| `TestStartPanicsWhenStopped` | Calling `Start` after `Stop` (already stopped) panics |
| `TestGameTimeLimit` | `TimeLimit` reflects the configured value |
| `TestGameTargetsEmptyBeforeStart` | `Targets` is empty before `Start` is called |
| `TestGameTargetsPopulatedAfterStart` | `Targets` has one entry per process after `Start` |
| `TestGamePendingConfirmNilWhenNoPending` | `PendingConfirm` returns `nil` when no confirmation is pending |
| `TestGamePendingConfirmReturnsPendingTarget` | `PendingConfirm` returns the target held by the internal `confirmation` |
| `TestGameConfirmPendingFalseInitially` | `confirm.Pending()` is false on a new session |
| `TestGameThrottleMutationAffectsSpeed` | Mutating the throttle returned by `Throttle()` changes `Speed()` |
| `TestGameAvailableTargetsExcludesDeadTargets` | `AvailableTargets` omits targets once their kill animation completes and they go `Dead`, reporting 0 alive |
| `TestGameAvailableTargetsCountsAlive` | `AvailableTargets` counts only `Alive`-state targets, while a `Killing` target stays in the returned slice but uncounted |
| `TestUpdateStopsWhenTimeLimitExpired` | `Update` stops the session once elapsed time exceeds the configured time limit |
| `TestUpdateStopsWhenAllTargetsDead` | `Update` stops the session once every target is in the `Dead` state |

### `internal/core/game/roster_test.go` (10 tests)

| Test | Description |
|------|-------------|
| `TestNewRosterEmptyBeforeSpawn` | `newRoster` holds no targets until `spawn` is called |
| `TestRosterSpawnCreatesTargetForEachProcess` | `spawn` creates one target per process in the roster |
| `TestRosterMoveAdvancesTargets` | `move` advances every target's position by one step |
| `TestRosterAllDeadFalseWhenEmpty` | `allDead` is `false` for an empty roster — no targets means nothing was ever killed |
| `TestRosterAllDeadTrueWhenAllDead` | `allDead` is `true` once every target has reached the `Dead` state |
| `TestRosterAllDeadFalseWhenSomeAlive` | `allDead` is `false` when at least one target is not `Dead` |
| `TestRosterHitAtReturnsTargetAtCoordinates` | `hitAt` returns the target whose tag covers the given coordinates |
| `TestRosterHitAtMissReturnsNil` | `hitAt` returns `nil` when no target covers the given coordinates |
| `TestRosterAvailableExcludesDeadTargets` | `available` excludes `Dead` targets and reports 0 alive |
| `TestRosterAvailableCountsAlive` | `available` counts only `Alive`-state targets, while `Killing` targets remain in the returned slice |

### `internal/core/game/target_test.go` (22 tests)

| Test | Description |
|------|-------------|
| `TestNewTargetWithinBounds` | Newly created target spawns within terminal bounds with correct PID/name/RSS and `Alive` state |
| `TestNewTargetSmallTerminal` | Small terminal (5×5) still returns a non-nil target |
| `TestTargetTagAlive` | Alive target's `Tag()` is `"[PID name]"` |
| `TestTargetTagDead` | Dead target's `Tag()` is empty |
| `TestTargetTagKilling` | Killing target's `Tag()` is a non-empty animation frame |
| `TestTargetUpdateKillingState` | Target advances to `Dead` after the final kill-animation tick |
| `TestTargetUpdateDeadNoOp` | Dead target does not move on `Update` |
| `TestTargetUpdateBounceLeft` | Target moving left past x=0 reverses X velocity and stays in bounds |
| `TestTargetUpdateBounceRight` | Target moving right past the right bound reverses X velocity and stays in bounds |
| `TestTargetUpdateBounceTop` | Target moving up past y=0 reverses Y velocity and stays in bounds |
| `TestTargetUpdateBounceBottom` | Target moving down past the bottom bound reverses Y velocity and stays in bounds |
| `TestTargetUpdateSpeedMultiplier` | Position advances by `velocity × speed` per `Update` call with no wall interaction |
| `TestTargetUpdateMultiByteRightWall` | Right-wall bound uses rune count, not byte count — regression for multi-byte process names (`"café"`) |
| `TestTargetContains` | `isHitAt` is true across the label rectangle and false just outside it |
| `TestTargetContainsMultiByteProcessName` | `isHitAt` uses rune count for width so multi-byte names don't over-count — regression for byte-count bug |
| `TestTargetContainsNotAlive` | Non-alive target returns `false` from `isHitAt` for all positions |
| `TestTargetIsAlive` | `isAlive` is true only in the `Alive` state |
| `TestTargetIsDead` | `isDead` is true only in the `Dead` state |
| `TestTargetKill` | `Kill` on an alive target returns true, sets state to `Killing`, resets the animation tick |
| `TestTargetKillNoOpWhenNotAlive` | `Kill` on a dead target returns false and leaves state unchanged |
| `TestTargetReap` | `Reap` on an alive target returns true and sets state straight to `Dead`, skipping the kill animation |
| `TestTargetReapNoOpWhenNotAlive` | `Reap` on a target that isn't alive (e.g. `Killing`) returns false and leaves state unchanged |

### `internal/core/game/input_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestInputOnQuitStopsGame` | `OnQuit` stops the session when no confirmation is pending |
| `TestInputOnQuitCancelsConfirmWhenPending` | `OnQuit` cancels a pending confirmation instead of stopping the session |
| `TestInputOnYesReturnsConfirmedTarget` | `OnYes` returns the confirmed target and clears the pending confirmation |
| `TestInputOnNoCancelsPending` | `OnNo` clears the pending confirmation without stopping the session |
| `TestInputOnSpeedUpIncreasesSpeed` | `OnSpeedUp` increases session speed by 0.5 |
| `TestInputOnSpeedDownDecreasesSpeed` | `OnSpeedDown` decreases session speed by 0.5 |
| `TestInputOnClickAtReturnsTarget` | Clicking on a target's position returns that target |
| `TestInputOnClickAtMissReturnsNil` | Clicking empty space returns `nil` |
| `TestInputOnClickAtNoOpWhenAlreadyConfirming` | A second click while a confirmation is already pending returns `nil` and leaves the pending target unchanged |

### `internal/core/score/entry_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestEntryBeatsByKills` | Higher kill count wins regardless of freed memory |
| `TestEntryBeatsTiebreakBySpeed` | Equal kills — higher speed wins, regardless of freed memory |
| `TestEntryBeatsTiebreakByDuration` | Equal kills and speed — shorter duration wins, regardless of freed memory |
| `TestEntryBeatsTiebreakByFreedMem` | Equal kills, speed, and duration — higher freed memory wins |

### `internal/core/score/board_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestBoardAddSortsDescending` | Entries are kept sorted by kills descending after each `Add` |
| `TestBoardAddTiebreakByMemory` | Equal kills are broken by freed memory descending |
| `TestBoardAddCapsAtMax` | Board never exceeds `maxScores` entries; the lowest surviving kill count is asserted after evicting the oldest, lowest entries |
| `TestBoardHighScoreEmpty` | `killScore` returns 0 for an empty board |
| `TestBoardHighScoreWithEntries` | `killScore` returns the maximum kills across all entries |
| `TestNewBoardPopulatesEntries` | `NewBoard` sets `Scores` to the given entries and seeds the returned tracker's `HighScore` from them |
| `TestNewBoardEmptyEntriesSeedsZeroHighScore` | `NewBoard(nil)` yields an empty board and a tracker with `HighScore` 0 |
| `TestBoardPrintScoresPrintsTrophyWhenBeatsRecord` | Trophy message printed when the tracker's kills exceed the board's pre-`Add` high score |
| `TestBoardPrintScoresNoTrophyWhenDoesNotBeatRecord` | No trophy message when the tracker's kills fall short of the existing record |
| `TestBoardPrintScoresNoTrophyWhenZeroKills` | No trophy message when the tracker has zero kills, even on an empty board |
| `TestBoardPrintScoresPrintsTrophyWhenTiesRecord` | Trophy message printed when the tracker's kills tie the board's prior high score |
| `TestBoardPrintScoresPrintsTrophyForFirstEntry` | Trophy message printed for the very first scored entry (previous high score was 0) |
| `TestBoardPrintScoresPrintsNoScoresMessageWhenEmpty` | `"No high scores yet!"` is printed when the board has no entries |
| `TestBoardPrintScoresPrintsTable` | Output includes `"Kills"` and `"Freed"` column headers |
| `TestBoardPrintScoresPrintsDuration` | Output includes a `"Time"` column and formats duration as e.g. `"12.3s"` |

### `internal/core/score/tracker_test.go` (2 tests)

| Test | Description |
|------|-------------|
| `TestTrackerRecordKill` | Each `RecordKill` increments `Kills` by 1 and accumulates `FreedMem` |
| `TestTrackerRecordKillUpdatesHighScore` | `HighScore` only advances once `Kills` exceeds the previously set record |

---

## Application Layer

### `internal/application/event/dispatcher_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestDispatcherClickReturnsHitTarget` | A `ClickEvent` on a target's position is forwarded to `OnClickAt` and the hit target is returned |
| `TestDispatcherClickMissReturnsNil` | A `ClickEvent` with no targets present returns `nil` |
| `TestDispatcherEscapeStopsGame` | `KeyEscape` calls `OnQuit` and stops the session |
| `TestDispatcherCtrlCStopsGame` | `KeyCtrlC` calls `OnQuit` and stops the session |
| `TestDispatcherCtrlZStopsGame` | `KeyCtrlZ` calls `OnQuit` and stops the session |
| `TestDispatcherQStopsGame` | Rune `q` calls `OnQuit` and stops the session |
| `TestDispatcherQUppercaseStopsGame` | Rune `Q` also calls `OnQuit` and stops the session |
| `TestDispatcherYReturnsConfirmedTarget` | Rune `y` accepts a pending confirmation and returns the target, clearing the pending state |
| `TestDispatcherYUppercaseAcceptsConfirmation` | Rune `Y` also accepts a pending confirmation |
| `TestDispatcherNCancelsConfirmation` | Rune `n` clears a pending confirmation without returning a target |
| `TestDispatcherNUppercaseCancelsConfirmation` | Rune `N` also clears a pending confirmation |
| `TestDispatcherPlusIncreasesSpeed` | Rune `+` calls `OnSpeedUp` |
| `TestDispatcherEqualsIncreasesSpeed` | Rune `=` is an alias for `+` and also calls `OnSpeedUp` |
| `TestDispatcherMinusDecreasesSpeed` | Rune `-` calls `OnSpeedDown` |
| `TestDispatcherUnderscoreDecreasesSpeed` | Rune `_` is an alias for `-` and also calls `OnSpeedDown` |
| `TestDispatcherUnknownRuneNoOp` | An unmapped rune (`z`) returns `nil` and leaves running state and speed unchanged |

### `internal/application/game/converter_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestToConfirmViewStateNilInput` | `toConfirmViewState(nil)` returns `nil` |
| `TestToConfirmViewStateMapsFields` | `toConfirmViewState` maps a target's PID and name onto the view state |
| `TestToTargetViewStateRoundsPosition` | `toTargetViewState` rounds position to integer X/Y, carries the tag, and reports `Killing=false` for an alive target |
| `TestToTargetViewStateMarksKilling` | `toTargetViewState` reports `Killing=true` for a target mid kill-animation |
| `TestToTargetViewStatesMapsAll` | `toTargetViewStates` maps every target in order, preserving each one's `Killing` flag |
| `TestToTargetViewStatesEmptyInput` | `toTargetViewStates(nil)` returns an empty result |
| `TestToHUDStateMapsFields` | `toHUDState` maps a tracker's `FreedMem`/`Kills`/`HighScore` onto the HUD view |
| `TestToStatusStateMapsFields` | `toStatusState` maps alive count, speed, and time limit, with `Confirming` nil when none pending |
| `TestToStatusStateIncludesConfirming` | `toStatusState`'s `Confirming` field carries the pending target's PID once `RequestConfirm` is called |

### `internal/application/game/service_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestBuildFrameAliveTargetIncluded` | `buildFrame` includes an alive target with `Killing=false` |
| `TestBuildFrameKillingTargetMarked` | `buildFrame` marks a target mid-kill with `Killing=true` |
| `TestBuildFrameDeadTargetExcluded` | `buildFrame` omits a target once its kill animation finishes and it goes `Dead` |
| `TestBuildFrameHUDReflectsStats` | `buildFrame`'s HUD carries the tracker's current kills and freed memory |
| `TestBuildFrameStatusBarAliveCount` | `buildFrame`'s `StatusBar.Alive` counts only `Alive`-state targets |
| `TestBuildFrameNoConfirmPending` | `buildFrame`'s `StatusBar.Confirming` is `nil` when no confirmation is pending |
| `TestServiceApplyKillsCompletesPendingKill` | `applyKills` transitions a queued target to `Killing` and increments the tracker's kill count |
| `TestServiceApplyKillsReapsAlreadyKilledTarget` | `applyKills` with `shouldReap` set transitions the target straight to `Dead` without awarding a kill |
| `TestServiceApplyKillsEmptyChannelNoOps` | `applyKills` on an empty kills channel returns immediately without blocking and leaves kills at 0 |

### `internal/application/process/converter_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestToProcessInfoMapsFields` | `toProcessInfo` maps Pid/Name/Rss from an `outbound.ProcessInfo` |
| `TestToProcessInfosMapsAll` | `toProcessInfos` maps every element of a slice in order |
| `TestToProcessInfosEmptyInput` | `toProcessInfos(nil)` returns an empty result |

### `internal/application/process/service_test.go` (6 tests)

| Test | Description |
|------|-------------|
| `TestFindProcessesListError` | `FindProcesses` surfaces an error returned by the process adapter's `List` |
| `TestFindProcessesNoMatches` | `FindProcesses` returns an empty slice, no error, when no pattern matches |
| `TestKillProtectedPIDReturnsError` | `Kill` refuses a protected PID (e.g. PID 1, `"init"`) — returns `killed=false`, `shouldReap=false`, and an error — without calling the process adapter's `Kill` |
| `TestKillLookupErrorReturnsShouldReap` | `Kill` returns `shouldReap=true` and surfaces the underlying `LookupName` error when the live-name lookup fails, without calling `Kill` |
| `TestKillNameMismatchReturnsShouldReap` | `Kill` returns `shouldReap=true` when the live process name no longer matches the target's (PID likely recycled), without calling `Kill` |
| `TestKillNameMatchInvokesKill` | `Kill` calls the process adapter's `Kill` once the live name is verified to match, returning `killed=true` |

### `internal/application/process/validation_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestValidateSearchPatternsEmptySlice` | Empty pattern slice is rejected, wrapping `process.ErrNoPatterns` |
| `TestValidateSearchPatternsEmptyTerm` | A pattern slice containing an empty string is rejected |
| `TestValidateSearchPatternsTooShort` | Patterns of 1 or 2 characters (below `MinPatternLength`) are rejected |
| `TestValidateSearchPatternsExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestValidateSearchPatternsTooLong` | Pattern exceeding `MaxPatternLength` is rejected |
| `TestValidateSearchPatternsValid` | Multiple valid patterns are accepted |
| `TestValidateSearchPatternsExactMaxLength` | Pattern at exactly `MaxPatternLength` is accepted |

### `internal/application/score/converter_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestToBoardMapsFields` | `toBoard` maps a stored `outbound.ScoreBoard`'s entries into `score.Entry` values and seeds the tracker's `HighScore` |
| `TestToBoardEmptyInput` | `toBoard` on an empty `ScoreBoard` returns an empty board and `HighScore` 0 |
| `TestToScoreBoardMapsFields` | `toScoreBoard` maps a `score.Board`'s entries back into `outbound.ScoreEntry` values |
| `TestToScoreBoardEmptyInput` | `toScoreBoard` on an empty `score.Board` returns an empty `outbound.ScoreBoard` |

### `internal/application/score/service_test.go` (8 tests)

| Test | Description |
|------|-------------|
| `TestLoadScoreBoardMapsStoredEntries` | `LoadScoreBoard` maps stored entries and returns `success=true` with the correct `HighScore` |
| `TestLoadScoreBoardErrorFallsBackToEmptyBoard` | A store load error yields an empty board, `HighScore=0`, `success=false`, and a `"could not load scores"` stderr warning |
| `TestRecordScoreAddsEntryFromTracker` | `RecordScore` appends a board entry built from the tracker's kills/freed memory plus the given speed/time/duration |
| `TestRecordScoreSavesWhenPersistTrue` | `RecordScore` saves the board to the store when `persist` is true |
| `TestRecordScoreSkipsSaveWhenPersistFalse` | `RecordScore` does not save to the store when `persist` is false |
| `TestRecordScoreSaveErrorPrintsWarning` | A store save error during `RecordScore` prints a `"score not saved"` warning containing the underlying error |
| `TestPrintResultsPrintsGameOverSummary` | `PrintResults` prints `"Game Over!"` and the correct kill count |
| `TestPrintResultsDelegatesToBoardPrintScores` | `PrintResults` delegates to the board's score printing, showing `"No high scores yet!"` for an empty board |

### `internal/application/runner_integration_test.go` (9 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationRunnerHappyPath` | Pre-queued `q` key runs `Runner.Run` to completion; the store's `Save` is called; the saved entry has `kills=0` and `duration>0` |
| `TestIntegrationRunnerLoadErrorPrintsWarningAndSkipsSave` | A `Load` failure prints `"could not load scores"` to stderr; `Run` still returns `nil`; `Save` is skipped |
| `TestIntegrationRunnerSaveErrorPrintsWarning` | A `Save` failure prints `"score not saved"` and the underlying error to stderr; `Run` still returns `nil` |
| `TestIntegrationRunnerQuitOnQ` | `q` key sent asynchronously during a live session causes `Run` to return without error |
| `TestIntegrationRunnerQuitOnEscape` | Escape key sent asynchronously causes `Run` to return without error |
| `TestIntegrationRunnerQuitOnCtrlC` | Ctrl+C sent asynchronously causes `Run` to return without error |
| `TestIntegrationRunnerQuitOnCtrlZ` | Ctrl+Z sent asynchronously causes `Run` to return without error |
| `TestIntegrationRunnerTimeLimitExpires` | A session with a 1-second time limit exits on its own within 3 seconds |
| `TestIntegrationRunnerSignalGoroutineDoesNotAccumulate` | The signal-handling goroutine started inside `runLoop` exits when `Run` returns; three sequential games (after an `os/signal` warm-up run) leave no extra goroutines behind |

---

## Entrypoint Layer

### `internal/entrypoint/cli/cli_test.go` (12 tests, 12 subtests)

| Test | Description |
|------|-------------|
| `TestRunNoArgsPrintsUsageAndReturnsNil` | `Run` with no arguments prints usage and returns `nil` |
| `TestRunHelpFlagPrintsUsageAndReturnsNil` | Table-driven, 2 cases: `--help` and `-h` each print usage and return `nil` |
| `TestRunBasicPattern` | Single pattern sets defaults: `ConfirmMode=false`, `Speed=2.0`, `TimeLimit=30` |
| `TestRunMultiplePatterns` | Three patterns are all captured in order |
| `TestRunConfirmFlag` | `--confirm` sets `ConfirmMode=true` |
| `TestRunSpeedFlag` | Table-driven, 6 cases: valid (3.5), min boundary (0.5), max boundary (5.0), too low (0.0, error), too high (6.0, error), non-numeric (error) |
| `TestRunTimeFlag` | Table-driven, 4 cases: valid (60), zero — no limit (0, no error), negative (-1, error), non-numeric (error) |
| `TestRunUnknownFlag` | Unrecognised flags return an error |
| `TestRunNoPatterns` | Flags-only args (no process patterns) return an error wrapping `process.ErrNoPatterns` |
| `TestRunPatternTooShort` | Patterns of 1 or 2 characters (below `MinPatternLength`) are rejected at CLI parse time |
| `TestRunPatternExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestRunPatternTooLong` | Pattern exceeding `MaxPatternLength` is rejected at CLI parse time |

---

## Infrastructure Layer

### `internal/infrastructure/osprocess/osprocess_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestListReturnsResults` | A real `ps` call via the adapter returns at least one process |
| `TestOwnPidMatchesOSGetpid` | `OwnPid` matches `os.Getpid()` |
| `TestLookupNameReturnsOwnName` | `LookupName` on the running test binary's own PID returns a non-empty name |
| `TestKillerNonexistentPID` | `Kill` on a PID that doesn't exist returns an error — `Kill` itself no longer guards PID or verifies name (that's `application/process.Service.Kill`'s job), so this only checks the one safe, deterministic case |

### `internal/infrastructure/osprocess/osprocess_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationListReturnsResults` | Real `ps` call returns at least one process |
| `TestIntegrationListValidFields` | Every listed process has a positive PID, non-empty name, and non-negative RSS |
| `TestIntegrationListShortProcessNames` | Process names are base names, not full paths |
| `TestIntegrationFindExcludesOwnAndInitPID` | `process.Find` — the domain policy, not the adapter — excludes the caller's own PID and PID 1 when run against real adapter output |

### `internal/infrastructure/scorefilestore/score_file_store_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestLoadFileNotExistReturnsEmptyBoard` | Missing score file returns an empty board with zero high score |
| `TestLoadInvalidJSONReturnsError` | Corrupted JSON file returns a parse error |
| `TestSaveCreatesFile` | Saving an empty board creates the file on disk |
| `TestSaveFilePermissions` | Saved file has `0600` permissions (not world-readable) |
| `TestSaveLoadRoundTrip` | Single entry survives a save/load cycle with correct field values |
| `TestSaveLoadMultipleEntries` | Three entries persist correctly; order and values are preserved |
| `TestSaveOverwritesPreviousFile` | A second Save replaces the first; only the latest data is visible |

### `internal/infrastructure/tcellui/tcellui_test.go` (14 tests)

| Test | Description |
|------|-------------|
| `TestPollGoroutineExitsAfterCleanup` | Poll goroutine count returns to baseline after `Cleanup()` — regression for issue #6 |
| `TestDrawHUDNarrowTerminalSuppressesCenter` | On a narrow (30-col) terminal the centre `Highscore` element is suppressed — regression for issue #14 |
| `TestDrawHUDWideTerminalDrawsAllThree` | On a standard (80-col) terminal all three HUD elements (`FREED`, `Highscore`, `KILLS`) are visible |
| `TestRenderMultiByteLabelColumnLayout` | Multi-byte kill-animation rune (`✦`) occupies one column, not multiple — regression for issue #13 |
| `TestPollTranslatesEscape` | tcell `KeyEscape` → `outbound.KeyEvent{Key: KeyEscape}` |
| `TestPollTranslatesCtrlC` | tcell `KeyCtrlC` → `outbound.KeyEvent{Key: KeyCtrlC}` |
| `TestPollTranslatesCtrlZ` | tcell `KeyCtrlZ` → `outbound.KeyEvent{Key: KeyCtrlZ}` |
| `TestPollTranslatesRune` | Plain rune `'q'` → `outbound.KeyEvent{Key: KeyNone, Ch: 'q'}` |
| `TestPollMouseButton1EmitsClickEvent` | Button1 mouse click → `outbound.ClickEvent{X, Y}` |
| `TestPollNonButton1DropsEvent` | Button2 click is dropped; next event is the following key |
| `TestDrawStatusBarNormal` | Status bar contains `Targets:`, `Speed:`, and `Click to kill` |
| `TestDrawStatusBarConfirming` | Confirm bar shows PID, name, `(Y)es`, `(N)o` |
| `TestDrawStatusBarWithTimeLimit` | Status bar shows `Time:` and remaining seconds when `TimeLimit > 0` |
| `TestDrawStatusBarNoTimeLimit` | `Time:` segment is absent when `TimeLimit == 0` |

---

## Utility

### `internal/util/util_test.go` (1 test, 8 subtests)

| Test | Description |
|------|-------------|
| `TestFormatBytes/0 B` | 0 bytes formats as `"0 B"` |
| `TestFormatBytes/512 B` | 512 bytes formats as `"512 B"` |
| `TestFormatBytes/1.0 KB` | 1024 bytes formats as `"1.0 KB"` |
| `TestFormatBytes/1.5 KB` | 1536 bytes formats as `"1.5 KB"` |
| `TestFormatBytes/1.0 MB` | 1 048 576 bytes formats as `"1.0 MB"` |
| `TestFormatBytes/1.5 MB` | 1 572 864 bytes formats as `"1.5 MB"` |
| `TestFormatBytes/1.0 GB` | 1 073 741 824 bytes formats as `"1.0 GB"` |
| `TestFormatBytes/1.5 GB` | 1 610 612 736 bytes formats as `"1.5 GB"` |

---

## Highlights

**Domain logic is split into standalone, independently-tested packages** — `core/process` (pattern validation, `Info.IsProtected`, `Find` filtering), `core/movement` (`Bounds`, `Motion`, `Speed`, `Throttle`), and `core/score` (`Entry.beats` ranking, `Board` add/cap/print, `Tracker` kill accounting) each own their rules and tests independently of the game loop, and of each other.

**`core/game` is decomposed by responsibility, one file per concern** — `lifecycle_test.go` (atomic lifecycle: `pending`/`running`/`stopped`), `confirmation_test.go` (hold/accept/cancel a pending kill), `timer_test.go` (elapsed/remaining time, unlimited mode), `roster_test.go` (the internal target collection: spawn, move, `allDead`, `hitAt`, `available`), `session_test.go` (the `Session` type's lifecycle, `Update`, and `AvailableTargets` — this absorbed what used to be a separate `loop_test.go`'s per-tick `Update` tests when `Game` was renamed to `Session`), `target_test.go` (per-target physics, hit-detection, kill/reap lifecycle), and `input_test.go` (player intent → domain calls). `Session` no longer builds its own frame/snapshot — that moved to the application layer.

**The old `Game`/`Frame` API is gone** — `core/game.Session` (in `session.go`) replaces `Game`; there is no `Frame` or `Snapshot` type left in `core/game`. Frame assembly for rendering now happens in `application/game.buildFrame`, which returns `outbound.FrameState`, tested directly in `application/game/service_test.go` and `application/game/converter_test.go` against structured data rather than terminal output.

**The old single `application/service` package was split into three independently-tested services, plus a composing `application` package** — `application/game` (`Service.Play`, the game loop and frame-building; `converter_test.go` + `service_test.go`), `application/process` (`Service.FindProcesses`/`Service.Kill`, process discovery and kill verification; `converter_test.go` + `service_test.go` + `validation_test.go`), and `application/score` (`Service.LoadScoreBoard`/`RecordScore`, plus package-level `PrintResults`; `converter_test.go` + `service_test.go`). The new top-level `application` package (`runner.go`) wires all three together behind `Runner.Run`; its own `runner_integration_test.go` is what the old `service/game_integration_test.go` behavior now lives under.

**Kill safety lives in `application/process.Service.Kill` and returns its result directly, no sentinel error** — `Kill` now returns `(killed, shouldReap bool, err error)` instead of wrapping a sentinel `errAlreadyKilled`. It calls `process.Info.IsProtected()` and re-verifies the live process name via `LookupName` before ever invoking the adapter's `Kill`, setting `shouldReap=true` on either failure so the caller can `Reap()` the target instead of leaving it stuck `Alive` forever. Covered by `TestKillProtectedPIDReturnsError`, `TestKillLookupErrorReturnsShouldReap`, `TestKillNameMismatchReturnsShouldReap`, `TestKillNameMatchInvokesKill` (`application/process/service_test.go`), and the consuming side in `TestServiceApplyKillsReapsAlreadyKilledTarget` (`application/game/service_test.go`). The integration suite (`TestIntegrationFindExcludesOwnAndInitPID`) replicates the exclusion policy against a live process table, and `osprocess_test.go` now documents in a comment that PID/name guarding is deliberately not tested at the adapter level.

**Input handling lives directly in `core/game`** — `Input` (in `core/game/input.go`) is the domain gateway for player intent (`OnQuit`, `OnYes`, `OnNo`, `OnSpeedUp`, `OnSpeedDown`, `OnClickAt`), operating on a `*Session`. `application/event.Dispatcher` translates raw `outbound.KeyEvent`/`ClickEvent` values into calls on it. The two halves are tested independently: `application/event/dispatcher_test.go` (renamed from `input_test.go`, same behavior) verifies event-to-method routing (including key aliases `=`/`+` and `_`/`-`), while `core/game/input_test.go` verifies `Input`'s actual effect on a real `Session`.

**Score persistence is a three-stage pipeline, each stage tested on its own** — `core/score` owns ranking/cap/print logic (`board_test.go`, `entry_test.go`, `tracker_test.go`); `application/score/converter_test.go` covers the pure mapping between the persisted `outbound.ScoreBoard` shape and the domain `score.Board`/`score.Entry`; `application/score/service_test.go` covers the orchestration (`LoadScoreBoard`, `RecordScore`, `PrintResults`) against a fake store; `infrastructure/scorefilestore` covers the actual file I/O.

**Validation is checked at every seam it's used** — `core/process.Validate` is the canonical boundary check (`MinPatternLength`/`MaxPatternLength`); `application/process.validateSearchPatterns` and `entrypoint/cli`'s tests re-assert the same min/max boundaries at the service and CLI layers using the shared constants, so a change to the limits is caught everywhere a pattern can enter the system.

**Integration vs unit split** — Two files carry `//go:build integration` tags: `internal/infrastructure/osprocess/osprocess_integration_test.go` and `internal/application/runner_integration_test.go`. These touch the real OS process table or run a full game loop through the composed `Runner`. All other tests are pure unit tests that run offline with fake/stub dependencies. Run with `go test -tags integration ./...`.

**Unicode correctness is enforced by regression tests** — `TestTargetUpdateMultiByteRightWall` and `TestTargetContainsMultiByteProcessName` (`core/game/target_test.go`) guard against a byte-count bug where multi-byte UTF-8 names caused targets to bounce too early or accept out-of-bounds click hits. `TestRenderMultiByteLabelColumnLayout` (`infrastructure/tcellui`) guards the same class of bug in the tcell renderer, and is explicitly marked as a regression test for issue #13. All three use multi-byte fixtures (`"café"`, `"✦"`).

**Named regressions are called out in comments and preserved in the doc** — `TestPollGoroutineExitsAfterCleanup` (issue #6), `TestDrawHUDNarrowTerminalSuppressesCenter` (issue #14), and `TestRenderMultiByteLabelColumnLayout` (issue #13) all carry doc comments identifying the GitHub issue they guard against, in `internal/infrastructure/tcellui/tcellui_test.go`.

**All quit signals are exercised** — `runner_integration_test.go`'s `TestIntegrationRunnerQuitOnQ`/`QuitOnEscape`/`QuitOnCtrlC`/`QuitOnCtrlZ` verify that `q`, Escape, Ctrl+C, and Ctrl+Z each independently terminate a live session end-to-end through `Runner.Run`.

**Signal goroutine lifecycle is verified under repeated runs** — `TestIntegrationRunnerSignalGoroutineDoesNotAccumulate` warms up `os/signal`'s lazily-created background goroutine with one throwaway game, then runs three more sequential sessions and asserts goroutine count does not grow, confirming the signal-handling goroutine started inside `application/game.Service.runLoop` exits cleanly every time `Play` (and therefore `Run`) returns. `TestPollGoroutineExitsAfterCleanup` verifies the analogous property for the tcell poll goroutine.

**Test infrastructure is organised as three shared sub-packages under `internal/testutil`** — `fake` provides typed test doubles: `Process` (test double for `outbound.ProcessManager`, with `ListErr`/`LookupNameErr`/`LookupNameValue`/`KillErr`/`KilledPIDs` knobs), `Store` (`outbound.ScoreStore` double capturing the last `Saved` board), `Renderer` (no-op `outbound.Renderer`), and `InputSource` (buffered-channel `outbound.InputSource` driven by pushing onto `Ch`). `capture` provides `Output`/`Stderr`, dependency-free stdout/stderr capture helpers built on `os.Pipe`. `fixture` is now two files: `game.go` provides session builders (`Game`, `ConfirmGameSession`, `PendingConfirmGameSession`) used throughout `core/game`'s own tests as well as `application/event` and `application/game` tests, and `process.go` provides `Process`/`Processes` builders for `process.Info` fixtures. All three packages exist to prevent duplicated fakes and avoid import cycles between test files and the packages they exercise.
