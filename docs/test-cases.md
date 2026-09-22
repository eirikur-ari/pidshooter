# Test Cases

Total: 49 test files across 21 packages — 432 top-level test functions, 489 total test cases (including table-driven subtests).

---

## Core Domain Layer

### `internal/core/process/info_test.go` (24 tests)

| Test | Description |
|------|-------------|
| `TestValidatePatternsNoPatterns` | `nil` and empty pattern slices are both rejected with an "at least one search pattern is required" message |
| `TestValidatePatternsTooShort` | Patterns of 0–2 characters (below `minPatternLength`) are rejected |
| `TestValidatePatternsExactMinLength` | Pattern at exactly `minPatternLength` is accepted |
| `TestValidatePatternsValid` | A normal pattern (`"firefox"`) is accepted |
| `TestValidatePatternsExactMaxLength` | Pattern at exactly `maxPatternLength` is accepted |
| `TestValidatePatternsTooLong` | Pattern exceeding `maxPatternLength` is rejected |
| `TestValidateProcessesEmpty` | An empty process slice is rejected with a "no processes found" message |
| `TestValidateProcessesNonEmpty` | A non-empty process slice passes validation |
| `TestValidateNameMismatch` | `ValidateName` reports both the expected and actual name when they differ |
| `TestValidateNameMatch` | `ValidateName` succeeds when names match |
| `TestInfoIsProtected` | PID 0 and PID 1 are protected regardless of name; any other PID is not |
| `TestInfoIsKillableBy` | `IsKillableBy` grants access to a process's owner and to root, and — only with `includeRoot=true` — lets a non-root caller target a root-owned process; `includeRoot` never grants access to another *non-root* user's process |
| `TestFindIncludeRootIncludesRootOwnedProcesses` | With `includeRoot=true`, `Find` surfaces a root-owned match alongside the caller's own |
| `TestFindWithoutIncludeRootExcludesRootOwnedProcesses` | With `includeRoot=false`, `Find` excludes a root-owned match even if its name matches |
| `TestFindExcludesProcessesNotOwnedByCaller` | `Find` excludes a match owned by a different, non-root user |
| `TestFindAsRootIncludesProcessesOwnedByAnyUser` | When the caller's own UID is 0 (root), `Find` includes matches regardless of owner |
| `TestFindMatchesByName` | Exact name match returns only the matching process |
| `TestFindSubstringMatch` | Partial name match is supported |
| `TestFindCaseInsensitive` | Name matching is case-insensitive |
| `TestFindMultipleTerms` | OR-semantics: multiple patterns each match independently |
| `TestFindNoMatch` | No matching name returns an empty slice |
| `TestFindExcludesPID1` | PID 1 (init) is always excluded from results |
| `TestFindExcludesOwnPID` | The caller's own PID is excluded from results |
| `TestFindExcludesPID0` | PID 0 (kernel swapper) is excluded alongside PID 1 |

### `internal/core/movement/bounds_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestBoundsBounceLeft` | Position/velocity crossing the left wall clamps to 0 and reverses X velocity |
| `TestBoundsBounceRight` | Position/velocity crossing the right wall (accounting for label width) clamps and reverses X velocity |
| `TestBoundsBounceTop` | Position/velocity crossing the top wall clamps to the top chrome row and reverses Y velocity |
| `TestBoundsBounceBottomDoesNotUndoTopReservation` | On a degenerate (2-row) terminal, bouncing off the bottom does not undo the top row's chrome reservation |
| `TestBoundsBounceBottom` | Position/velocity crossing the bottom wall clamps above the status bar row and reverses Y velocity |
| `TestBoundsBounceTopRespectsGivenTopValue` | A non-default top chrome size (3 rows) is respected as the clamp boundary |
| `TestBoundsBounceBottomRespectsGivenBottomValue` | A non-default bottom chrome size (4 rows) is respected as the clamp boundary |

### `internal/core/movement/motion_test.go` (5 tests)

| Test | Description |
|------|-------------|
| `TestMotionMoveAdvancesPosition` | `Move` advances position by velocity scaled by speed, with no wall interaction |
| `TestMotionMoveBouncesAtLeftWall` | Motion starting at the left wall bounces: position stays in bounds, velocity flips positive |
| `TestMotionMoveBouncesAtRightWall` | Motion approaching the right wall bounces: position clamped, velocity flips negative |
| `TestMotionMoveSpeedScalesVelocity` | Position advances by `velocity × speed` per `Move` call |
| `TestNewRandomPositionRespectsGivenTopAndBottom` | Across 200 random samples, spawn Y position always falls within the given top/bottom chrome reservation |

### `internal/core/movement/speed_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestNewSpeed` | `newSpeed` initializes both `Current` and `Lowest` to the given starting value |
| `TestSpeedSetTracksLowest` | `Set` updates `Current` and lowers `Lowest` when the new value is smaller |
| `TestSpeedSetAboveLowestDoesNotRaiseLowest` | Setting a higher value updates `Current` but leaves `Lowest` at the previously recorded minimum |
| `TestSpeedSetAboveStartNeverLowersLowest` | Setting a value above the initial value updates `Current` but `Lowest` never rises above the starting value |

### `internal/core/movement/throttle_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestValidateSpeedTooLow` | A speed just below `MinSpeed` is rejected |
| `TestValidateSpeedTooHigh` | A speed just above `MaxSpeed` is rejected |
| `TestValidateSpeedNaN` | `NaN` is rejected as an invalid speed |
| `TestValidateSpeedPositiveInf` | Positive infinity is rejected as an invalid speed |
| `TestValidateSpeedNegativeInf` | Negative infinity is rejected as an invalid speed |
| `TestValidateSpeedMinBoundary` | Exactly `MinSpeed` is accepted |
| `TestValidateSpeedMaxBoundary` | Exactly `MaxSpeed` is accepted |
| `TestNewThrottle` | `NewThrottle` stores the initial speed |
| `TestThrottleIncrease` | `Increase` raises speed by 0.5 |
| `TestThrottleDecrease` | `Decrease` lowers speed by 0.5 |
| `TestThrottleIncreaseCapsAtMax` | Speed is capped at `MaxSpeed` and does not exceed it on repeated increases |
| `TestThrottleDecreaseFloorsAtMin` | Speed is floored at `MinSpeed` and does not go below it on repeated decreases |
| `TestThrottleLowestSpeedStartsAtInitialSpeed` | `LowestSpeed` starts equal to the throttle's initial speed |
| `TestThrottleLowestSpeedTracksDecreases` | `LowestSpeed` remains the smallest value ever reached, even after subsequent increases |
| `TestThrottleLowestSpeedUnaffectedByIncreaseOnly` | `LowestSpeed` stays at the initial speed when only `Increase` is ever called |

### `internal/core/game/confirmation_test.go` (5 tests)

| Test | Description |
|------|-------------|
| `TestConfirmationPendingFalseWhenEmpty` | A freshly constructed `confirmation` has no pending target |
| `TestConfirmationRequestConfirmModeHoldsTarget` | In confirm mode, `Request` holds the target and returns `nil` |
| `TestConfirmationRequestPassthroughModeReturnsTarget` | Outside confirm mode, `Request` returns the target immediately without holding it |
| `TestConfirmationAcceptReturnsAndClearsTarget` | `Accept` returns the held target and clears pending state |
| `TestConfirmationCancelClearsPending` | `Cancel` clears the pending target without returning it |

### `internal/core/game/timer_test.go` (18 tests)

| Test | Description |
|------|-------------|
| `TestValidateTimeLimitNegative` | A negative time limit is rejected |
| `TestValidateTimeLimitZeroIsUnlimited` | Zero is accepted and means unlimited |
| `TestValidateTimeLimitPositive` | An ordinary positive value (30) is accepted |
| `TestValidateTimeLimitMaxBoundary` | The maximum allowed value (`maxTimeLimitSeconds`) is accepted |
| `TestValidateTimeLimitExceedsMax` | One second past the maximum is rejected |
| `TestTimerExpiredFalseBeforeLimit` | A freshly started 60s timer, using the real clock, has not expired |
| `TestTimerExpiredTrueWhenLimitReached` | With a fake clock advanced 2s past a 1s limit, `Expired` is true |
| `TestTimerExpiredFalseWhenUnlimited` | A zero-limit timer is never expired, even without calling `Start` |
| `TestTimerRemainingWithinLimit` | `Remaining` on a freshly started 60s timer is positive and ≤ 60s |
| `TestTimerRemainingZeroWhenExpired` | `Remaining` is exactly 0 once the fake clock has passed the limit |
| `TestTimerRemainingZeroWhenUnlimited` | `Remaining` is 0 for a zero-limit timer |
| `TestTimerSecondsLeft` | `SecondsLeft` on a fresh 60s timer is between 1 and 60 |
| `TestTimerSecondsLeftRoundsUpPartialSecond` | `SecondsLeft` rounds up: 0.9s actually remaining (4.1s elapsed of a 5s limit) still reports `1`, not `0` |
| `TestTimerSecondsLeftExactWholeSecondIsUnaffected` | An exact 3s remaining (2s elapsed of a 5s limit) reports `3`, confirming the rounding-up fix doesn't over-round whole seconds |
| `TestTimerSecondsLeftZeroWhenExpired` | `SecondsLeft` is `0` once the fake clock has passed the limit |
| `TestTimerLimitSeconds` | `LimitSeconds` reflects the configured limit, including 0 for unlimited |
| `TestTimerStartTimeZeroBeforeStart` | `StartTime` is the zero `time.Time` before `Start` is called |
| `TestTimerStartTimeSetAfterStart` | `StartTime` is set to at-or-after the moment `Start` was called |

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
| `TestGameAvailableTargetsExcludesDeadTargets` | `AvailableTargets` omits a target once its full kill-animation duration has elapsed and it reaches `Dead`, reporting 0 alive |
| `TestGameAvailableTargetsCountsAlive` | `AvailableTargets` counts only `Alive`-state targets, while a `Killing` target stays in the returned slice but uncounted |
| `TestUpdateStopsWhenTimeLimitExpired` | `Update` stops the session once a controllable fake clock reports elapsed time past the configured time limit |
| `TestUpdateStopsWhenAllTargetsDead` | `Update` stops the session once every target is in the `Dead` state |

### `internal/core/game/roster_test.go` (11 tests)

| Test | Description |
|------|-------------|
| `TestNewRosterEmptyBeforeSpawn` | `newRoster` holds no targets until `spawn` is called |
| `TestRosterSpawnCreatesTargetForEachProcess` | `spawn` creates one target per process in the roster |
| `TestRosterMoveAdvancesTargets` | `move` advances a target's position by one step |
| `TestRosterAllDeadFalseWhenEmpty` | `allDead` is `false` for an empty roster — no targets means nothing was ever killed |
| `TestRosterAllDeadTrueWhenAllDead` | `allDead` is `true` once every target has reached the `Dead` state |
| `TestRosterAllDeadFalseWhenSomeAlive` | `allDead` is `false` when at least one target is not `Dead` |
| `TestRosterHitAtReturnsTargetOnHit` | `hitAt` returns the target whose position covers the given coordinates |
| `TestRosterHitAtReturnsNilWhenShotAlreadyFired` | `hitAt` ignores a repeat hit on a target that already has a shot in flight (`FireShot` called) |
| `TestRosterHitAtReturnsNilOnMiss` | `hitAt` returns `nil` when no target covers the given coordinates |
| `TestRosterAvailableExcludesDeadTargets` | `available` excludes `Dead` targets and reports 0 alive |
| `TestRosterAvailableCountsAlive` | `available` counts only `Alive`-state targets, while a `Killing` target remains in the returned slice but uncounted |

### `internal/core/game/target_test.go` (37 tests)

| Test | Description |
|------|-------------|
| `TestNewTargetWithinBounds` | Newly created target spawns within terminal bounds with correct PID/name/RSS and `Alive` state |
| `TestNewTargetSmallTerminal` | Small terminal (5×5) still returns a non-nil target |
| `TestTargetTagAlive` | Alive target's `Tag()` is `"[PID name]"` |
| `TestTargetTagDead` | Dead target's `Tag()` is empty |
| `TestTargetTagKilling` | Killing target's `Tag()` is empty — the renderer draws its own animation frame instead |
| `TestTargetTagFleeing` | Fleeing target's `Tag()` is empty — same reasoning as Killing |
| `TestTargetAnimationProgressZeroWhenAlive` | `AnimationProgress` is 0 for an `Alive` target regardless of `AnimationTick` |
| `TestTargetAnimationProgressZeroWhenDead` | `AnimationProgress` is 0 for a `Dead` target regardless of `AnimationTick` |
| `TestTargetAnimationProgressReflectsTickWhenKilling` | `AnimationProgress` is `AnimationTick / AnimationDuration` while `Killing` |
| `TestTargetAnimationProgressReflectsTickWhenFleeing` | `AnimationProgress` is `AnimationTick / AnimationDuration` while `Fleeing` |
| `TestTargetUpdateKillingState` | A `Killing` target transitions to `Dead` once `AnimationTick` reaches `AnimationDuration` |
| `TestTargetUpdateFleeingState` | A `Fleeing` target transitions to `Dead` once `AnimationTick` reaches `AnimationDuration` |
| `TestTargetUpdateDeadNoOp` | A `Dead` target does not move on `Update` |
| `TestTargetUpdateBounceLeft` | Target moving left past x=0 reverses X velocity and stays in bounds |
| `TestTargetUpdateBounceRight` | Target moving right past the right bound (accounting for tag width) reverses X velocity and stays in bounds |
| `TestTargetUpdateBounceTop` | Target moving up past y=0 reverses Y velocity and stays in bounds |
| `TestTargetUpdateBounceBottom` | Target moving down past the bottom bound reverses Y velocity and stays in bounds |
| `TestTargetUpdateSpeedMultiplier` | Position advances by `velocity × speed` per `Update` call with no wall interaction |
| `TestTargetUpdateMultiByteRightWall` | Right-wall bound uses rune count, not byte count — regression for a multi-byte process-name (`"café"`) premature-bounce bug |
| `TestTargetMoveNoOpWhenKilling` | Calling the private `move` directly on a `Killing` target is a no-op — regression proving the `State`-gate holds even for a caller that bypasses `Update`'s dispatch |
| `TestTargetMoveNoOpWhenFleeing` | Calling `move` directly on a `Fleeing` target is a no-op, same guard as above |
| `TestTargetMoveNoOpWhenDead` | Calling `move` directly on a `Dead` target is a no-op, same guard as above |
| `TestTargetMoveAdvancesPositionWhenAlive` | Calling `move` directly on an `Alive` target does advance its position (control case for the three no-op tests above) |
| `TestTargetHitAtIsTrueWhenWithinTagBounds` | `isHitAt` is true across the full width of the rendered tag |
| `TestTargetHitAtIsFalseWhenOutsideBounds` | `isHitAt` is false one column before the tag, one column past it, and one row above it |
| `TestTargetHitAtMatchesRoundedRenderPositionNotTruncated` | `isHitAt` matches the *rounded* render position, not the truncated one — regression for a click/label misalignment bug at fractional positions |
| `TestTargetContainsMultiByteProcessName` | `isHitAt` uses rune count, not byte count, for tag width — regression for a multi-byte name (`"café"`) over-counting hit columns |
| `TestTargetHitAtWillReturnFalseWhenInKillingState` | Non-`Alive` (`Killing`) targets are never hit by `isHitAt` |
| `TestTargetHitAtWillReturnFalseWhenShotAlreadyFired` | A target with a shot already in flight (`FireShot` called) is not hit again |
| `TestTargetHitAtWillReturnTrueAfterCeaseFire` | A target becomes hittable again once `CeaseFire` clears the in-flight marker |
| `TestTargetIsAlive` | `isAlive` is true only in the `Alive` state (false for `Killing`, `Fleeing`, `Dead`) |
| `TestTargetIsDead` | `isDead` is true only in the `Dead` state (false for `Alive`, `Killing`, `Fleeing`) |
| `TestTargetKill` | `Kill` on an alive target returns true, sets state to `Killing`, resets the animation tick |
| `TestTargetKillNoOpWhenNotAlive` | `Kill` on a dead target returns false and leaves state unchanged |
| `TestTargetReap` | `Reap` on an alive target returns true and sets state to `Fleeing`, resetting the animation tick |
| `TestTargetReapNoOpWhenNotAlive` | `Reap` on a target that isn't alive (e.g. `Killing`) returns false and leaves state unchanged |
| `TestTargetKillNoOpWhenFleeing` | `Kill` on a `Fleeing` target returns false and leaves state unchanged |

### `internal/core/game/input_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestInputOnQuitStopsGame` | `OnQuit` stops the session when no confirmation is pending |
| `TestInputOnQuitCancelsConfirmWhenPending` | `OnQuit` cancels a pending confirmation instead of stopping the session |
| `TestInputOnYesReturnsConfirmedTarget` | `OnYes` returns the confirmed target and clears the pending confirmation |
| `TestInputOnNoCancelsPending` | `OnNo` clears the pending confirmation without stopping the session |
| `TestInputOnSpeedUpIncreasesSpeed` | `OnSpeedUp` raises the session's throttle speed by 0.5 (2.0 → 2.5) |
| `TestInputOnSpeedDownDecreasesSpeed` | `OnSpeedDown` lowers the session's throttle speed by 0.5 (2.0 → 1.5) |
| `TestInputOnClickAtReturnsTarget` | Clicking at a spawned target's rounded position returns that target |
| `TestInputOnClickAtMissReturnsNil` | Clicking empty space (no targets) returns `nil` |
| `TestInputOnClickAtNoOpWhenAlreadyConfirming` | A second click while a confirmation is already pending returns `nil` and leaves the pending target unchanged |

### `internal/core/score/entry_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestEntryBeatsByKills` | Higher kill count wins regardless of freed memory |
| `TestEntryBeatsTiebreakByDuds` | Equal kills — more duds wins, regardless of speed or freed memory |
| `TestEntryBeatsTiebreakBySpeed` | Equal kills and duds — higher speed wins, regardless of freed memory |
| `TestEntryBeatsTiebreakByDuration` | Equal kills, duds, and speed — shorter duration wins |
| `TestEntryBeatsTiebreakByFreedMem` | Equal kills, duds, speed, and duration — higher freed memory wins |
| `TestEntryIsScoreTrueForPositiveKillsAndFreedMem` | An entry with kills and non-negative freed memory is a genuine score |
| `TestEntryIsScoreFalseForZeroKills` | An entry with zero kills is not a genuine score |
| `TestEntryIsScoreFalseForNegativeKills` | An entry with negative kills is not a genuine score |
| `TestEntryIsScoreFalseForNegativeFreedMem` | An entry with negative freed memory (impossible from a real session) is not a genuine score |

### `internal/core/score/board_test.go` (18 tests)

| Test | Description |
|------|-------------|
| `TestBoardAddSortsDescending` | Entries are kept sorted by kills descending after each `Add` |
| `TestBoardAddTiebreakByMemory` | Equal kills are broken by freed memory descending |
| `TestBoardAddIgnoresZeroKillEntry` | A zero-kill session is not a score and is not recorded |
| `TestBoardAddIgnoresNegativeKillEntry` | A negative-kill entry is not recorded |
| `TestBoardAddIgnoresNegativeFreedMemEntry` | A negative freed-memory entry (impossible from a real session) is not recorded |
| `TestBoardAddFullTieKeepsInsertionOrder` | Entries fully tied across every ranked dimension (kills/duds/speed/duration/freed mem) deterministically preserve insertion order rather than depending on `sort.Slice`'s unspecified tie-breaking |
| `TestBoardAddCapsAtMax` | Board never exceeds `maxScores` entries; the lowest surviving kill count is asserted exactly after evicting the oldest, lowest entries |
| `TestBoardHighScoreEmpty` | `killScore` returns 0 for an empty board |
| `TestBoardHighScoreWithEntries` | `killScore` returns the maximum kills across all entries |
| `TestNewBoardPopulatesEntries` | `NewBoard` sets `Scores` to the given entries and `HighScore` reflects them |
| `TestNewBoardEmptyEntriesSeedsZeroHighScore` | `NewBoard(nil)` yields an empty board with `HighScore` 0 |
| `TestNewBoardDropsEntriesThatAreNotGenuineScores` | `NewBoard` filters out zero-kill, negative-kill, and negative-freed-mem entries from a persisted/hand-edited file, keeping only genuine scores |
| `TestNewBoardSortsAndCapsOversizedUnsortedInput` | `NewBoard` re-sorts and truncates a persisted file that's larger than `maxScores` and not already in ranked order |
| `TestBoardIsNewHighScoreTrueWhenBeatsRecord` | `IsNewHighScore` is true when kills exceed the board's prior high score |
| `TestBoardIsNewHighScoreFalseWhenDoesNotBeatRecord` | `IsNewHighScore` is false when kills fall short of the prior high score |
| `TestBoardIsNewHighScoreFalseWhenZeroKills` | `IsNewHighScore` is false for a zero-kill session even against an empty board |
| `TestBoardIsNewHighScoreTrueWhenTiesRecord` | `IsNewHighScore` is true when kills exactly tie the prior high score |
| `TestBoardIsNewHighScoreTrueForFirstEntry` | `IsNewHighScore` is true for the very first scored entry (prior high score is 0) |

---

## Application Layer

### `internal/application/apperror/apperror_test.go` (11 tests, 4 subtests)

| Test | Description |
|------|-------------|
| `TestErrorUnwrapReturnsErrorCause` | `Unwrap` returns the exact cause error passed to `NewError`, satisfying `errors.Is` |
| `TestErrorUnwrapReturnsNilWithNoErrorCause` | `Unwrap` returns `nil` when `NewError` was given a nil cause |
| `TestErrorNewError/with_both_message_and_cause` | `Error()` renders as `"message: cause"` when both are set |
| `TestErrorNewError/with_message_only` | `Error()` renders as just the message when there's no cause |
| `TestErrorNewError/with_cause_only` | `Error()` renders as just the cause's text when the message is empty |
| `TestErrorNewError/with_neither_message_nor_cause` | `Error()` renders as an empty string when both message and cause are empty/nil |
| `TestHandleAbsorbsWarningSeverity` | `Handle` returns `nil` for a `SeverityWarning` error (logged and swallowed) |
| `TestHandleAbsorbsErrorSeverity` | `Handle` returns `nil` for a `SeverityError` error |
| `TestHandleReturnsFatalSeverityError` | `Handle` returns the exact same error instance for `SeverityFatal`, not swallowing it |
| `TestHandlePropagatesUnknownSeverity` | A plain (non-`apperror`) error is treated as unknown severity and returned as-is |
| `TestHandleLogsFatalErrorOnlyOnce` | A `Fatal` error's internal `logged` flag is set after the first `Handle` call; a second `Handle` call on the same error still returns it but doesn't re-log |
| `TestHandleReturnsWrappedFatalError` | `Handle` returns the outer wrapped error unchanged (not re-derived) when a `Fatal` `apperror.Error` is wrapped via `fmt.Errorf("...: %w", ...)` |
| `TestHandleAbsorbsWrappedWarning` | A `Warning`-severity error wrapped via `fmt.Errorf` is still absorbed (returns `nil`) — severity detection unwraps |
| `TestHandleReturnsNilErrorWhenGivenNil` | `Handle(nil)` returns `nil` |

### `internal/application/config/config_test.go` (8 tests)

| Test | Description |
|------|-------------|
| `TestValidateNoPatterns` | `Validate` rejects a `Config` with no patterns |
| `TestValidateSpeedTooLow` | `Validate` rejects speed below `movement.MinSpeed` |
| `TestValidateSpeedTooHigh` | `Validate` rejects speed above `movement.MaxSpeed` |
| `TestValidateSpeedMinBoundary` | Speed exactly at `MinSpeed` is accepted |
| `TestValidateSpeedMaxBoundary` | Speed exactly at `MaxSpeed` is accepted |
| `TestValidateNegativeTimeLimit` | `Validate` rejects a negative `TimeLimit` |
| `TestValidateZeroTimeLimitIsUnlimited` | `TimeLimit: 0` is accepted (means unlimited) |
| `TestValidateValid` | A fully valid `Config` passes `Validate` and its fields are unchanged |

### `internal/application/contract/outbound/error_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestCorruptedDataErrorWithoutMessage` | `CorruptedDataError{}.Error()` is the literal string `"corrupted data"` when `Message` is empty |
| `TestCorruptedDataErrorWithMessagePrefixesIt` | A non-empty `Message` is prefixed onto `"corrupted data"` in `Error()`'s output |
| `TestCorruptedDataErrorMatchesRegardlessOfMessage` | `errors.As` matches `CorruptedDataError` regardless of the `Message` field's value (matching is by type, not content) |

### `internal/application/game/converter_test.go` (11 tests)

| Test | Description |
|------|-------------|
| `TestToConfirmViewStateReturnNilWhenInputIsNil` | `toConfirmViewState(nil)` returns `nil` |
| `TestToConfirmViewStateReturnsMappedFields` | Maps a target's PID and name onto the confirm view state |
| `TestToTargetViewStateReturnsMappedFields` | Maps a target's rounded position, tag, and `Killing=false`/zero `AnimationProgress` for an alive target |
| `TestToTargetViewStateIncludesAnimationProgressWhenKilling` | A target mid-kill (`AnimationTick` at half of `AnimationDuration`) maps to `Killing=true` and `AnimationProgress=0.5` |
| `TestToTargetViewStatesReturnsASliceOfMappedFields` | Maps a mixed-state slice (`Alive`/`Killing`/`Fleeing`) to view states with independent `Killing`/`Fleeing` flags each |
| `TestToTargetViewStatesReturnsEmptySliceWhenInputIsNil` | `toTargetViewStates(nil)` returns an empty slice |
| `TestToHUDViewStateReturnsMappedFields` | Maps a `killTracker`'s freed memory, kill count, and high score onto the HUD view |
| `TestToStatusViewStateReturnsMappedFieldsWithoutConfirmViewState` | Maps alive count, speed, and a non-nil `TimeLeft` pointer for a timed session with no pending confirmation |
| `TestToStatusViewStateTimeLeftNilWhenUntimed` | `TimeLeft` is `nil` for a session with no configured time limit |
| `TestToStatusViewStateIncludesConfirmViewState` | `Confirming` carries the pending target's PID once `RequestConfirm` is called |
| `TestToFrameViewStateReturnsMappedFields` | Builds a full frame view state with the session's targets included |

### `internal/application/game/kill_tracker_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestKillTrackerCreatesNewKillTracker` | `newKillTracker(highScore)` seeds the score tracker's high score and zeroes kills/freed memory, with an initialized `pids` map |
| `TestKillTrackerRecordKillAccumulates` | Repeated `recordKill` calls accumulate both kill count and freed memory |
| `TestKillTrackerRecordKillRaisesHighScore` | High score only rises once kill count actually exceeds the previously seeded/recorded value |
| `TestKillTrackerRecordFailureAddsFailedKillWithNoCause` | `recordFailure` with a nil cause still records a failure entry with the target's name and PID |
| `TestKillTrackerRecordFailureAddsFailedKillWithCause` | `recordFailure` with a non-nil cause preserves that error on the recorded failure |
| `TestKillTrackerRecordFailureAccumulatesDistinctPIDs` | Failures for different PIDs each add their own entry |
| `TestKillTrackerRecordFailureDeduplicatesSamePID` | A second failure for the same PID is a no-op — it does not add a duplicate entry or overwrite the first recorded error |

### `internal/application/game/service_test.go` (12 tests)

| Test | Description |
|------|-------------|
| `TestServicePlayNoProcessesReturnsClassifiedError` | `Play` with no processes returns an `apperror.Error` (`CodeProcessNotFound`/`SeverityFatal`) whose message is exactly `"no processes found"` and whose cause is still reachable via `Unwrap` |
| `TestServicePlayRendererInitFailureReturnsClassifiedError` | A `Renderer.Init` failure surfaces as `CodeGameFailed`/`SeverityFatal`, wrapping the underlying error text |
| `TestServicePlaySuccessReturnsPlayResult` | Quitting immediately returns a `PlayResult` with the starting speed unchanged, zero kills, zero freed memory, and no kill failures |
| `TestDrainEventQueueReturnsErrorWhenEventChannelCloses` | `drainEventQueue` returns an explicit `"input event channel closed"` error when the underlying events channel is closed |
| `TestServiceApplyKillsCompletesPendingKill` | A `killSignal` with no error transitions the target to `Killing` and increments the tracker's kill count |
| `TestServiceApplyKillsReapsAlreadyKilledTarget` | A `killSignal{shouldReap: true}` transitions the target to `Fleeing` (not `Dead` directly), awards no kill, and records one dud with the target's name/PID |
| `TestServiceApplyKillsEmptyChannelNoOps` | `applyKillSignals` on an empty channel returns immediately without blocking, leaving kill count at 0 |
| `TestServiceApplyKillsRecordsFailureWithoutMutatingTarget` | A `killSignal` carrying an error leaves the target `Alive`, awards no kill, and records exactly one failure with the target's name/PID/error text |
| `TestServiceApplyKillsDeduplicatesRepeatedFailuresForSamePID` | Two failure signals for the same target only produce one recorded failure entry |
| `TestKillOrReapReportsFailureWithoutReaping` | `killOrReap` reports `shouldReap=false` with the killer's error text when the process killer refuses (e.g. a protected PID) — the target is not silently reaped |
| `TestKillOrReapReapWithoutErrorStaysSilent` | `killOrReap` reports `shouldReap=true` and **no** `err` when the killer indicates the process is already gone — a reap must not also read as a reported failure |
| `TestRegisterTermSignalWatcherTermSignalClosesOnStop` | The channel returned by `registerTermSignalWatcher` closes once its `stop` function is called |

### `internal/application/game/service_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationServiceFrameLoopAppliesAsyncKillToResult` | A real click on a target drives a full async kill through `frameLoop`; the tracker ends up with 1 kill, the freed memory equals the process's RSS, and no failures |
| `TestIntegrationServiceFrameLoopWaitsForKillInFlightWhenSessionStops` | `frameLoop` blocks until an in-flight kill (started right before a `QuitEvent`) actually completes, and that kill still counts toward the tracker even though the session already stopped |
| `TestIntegrationAwaitOutstandingKillsReturnsOnceGracePeriodElapses` | `awaitOutstandingKills` gives up and returns once `killGracePeriod` elapses for a permanently wedged (never-`Done()`) `WaitGroup`, rather than hanging indefinitely |
| `TestIntegrationServiceDrainEventQueueIgnoresDuplicateClicksOnSameTarget` | Two clicks on the same target while its first kill attempt is still in flight dispatch only one actual kill call, confirmed via an atomic call counter |

### `internal/application/input/dispatcher_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestDispatcherClickReturnsHitTarget` | A `ClickEvent` at a target's rounded position is forwarded to `OnClickAt` and the hit target is returned |
| `TestDispatcherClickMissReturnsNil` | A `ClickEvent` with no targets present returns `nil` |
| `TestDispatcherQuitEventStopsGame` | Dispatching `QuitEvent{}` stops the session |
| `TestDispatcherConfirmEventAcceptReturnsConfirmedTarget` | `ConfirmEvent{Accept: true}` returns the confirmed target and clears the pending confirmation |
| `TestDispatcherConfirmEventDeclineCancelsConfirmation` | `ConfirmEvent{Accept: false}` clears the pending confirmation without returning a target |
| `TestDispatcherSpeedEventFasterIncreasesSpeed` | `SpeedEvent{Faster: true}` increases session speed by 0.5 |
| `TestDispatcherSpeedEventSlowerDecreasesSpeed` | `SpeedEvent{Faster: false}` decreases session speed by 0.5 |

### `internal/application/process/converter_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestToProcessInfoMapsFields` | `toProcessInfo` maps PID/Name/Rss/UID from an `outbound.ProcessInfo` |
| `TestToProcessInfosMapsAll` | `toProcessInfos` maps every element of a slice in order, preserving each field |
| `TestToProcessInfosEmptyInput` | `toProcessInfos(nil)` returns an empty result |

### `internal/application/process/service_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestFindProcessesReturnsErrorWhenProcessDiscoveryFails` | A discovery error from the process adapter is classified as `CodeProcessDiscoveryFailed`/`SeverityFatal` and its message is surfaced |
| `TestFindProcessesReturnsErrorWhenNoProcessIsFound` | No matches yields `CodeProcessNotFound`/`SeverityFatal`, an unwrappable cause, and the exact message `"no processes found"` |
| `TestFindProcessesDoesNotReportWhenDiscoveryFails` | A failed discovery never calls the process reporter |
| `TestFindProcessesDoesNotReportWhenNoProcessIsFound` | An empty match set (failing validation) never calls the process reporter |
| `TestFindProcessesReportsMatchCountAndPatterns` | On success, the process reporter is called with the exact match count and the patterns searched |
| `TestFindProcessesReturnsMatchingProcesses` | Only processes matching a pattern are returned, mapped to `core/process.Info` |
| `TestFindProcessesExcludesProcessesNotOwnedByCaller` | A process owned by a different UID is excluded from results by default |
| `TestFindProcessesIncludeRootIncludesRootOwnedProcesses` | With `includeRoot=true`, a root-owned (UID 0) process is included alongside the caller's own |
| `TestKillReturnsErrorIfPIDIsProtected` | `Kill` on a protected PID refuses immediately (`CodeProcessDiscoveryFailed`/`SeverityWarning`), never pinning the process |
| `TestKillReturnsErrorAndShouldReapIfPinFails` | A `Pin` failure returns `shouldReap=true` and surfaces the underlying error, with no handle to release |
| `TestKillReturnsErrorWithoutReapingIfProcessLookupByNameFailsTransiently` | A transient `LookupName` error returns `shouldReap=false` (not treated as "already exited") but still releases the pinned handle |
| `TestKillReturnsShouldReapIfProcessLookupByNameReportsProcessNotFound` | `LookupName` returning `outbound.NotFoundError` sets `shouldReap=true` with `CodeProcessNotFound` |
| `TestKillReturnsErrorAndShouldReapIfNameValidationFails` | A live name mismatch (PID likely recycled) sets `shouldReap=true` and reports the mismatch |
| `TestKillReturnsShouldReapWhenProcessAlreadyExited` | The pinned handle's `Kill` returning `outbound.NotFoundError` sets `shouldReap=true` and releases the handle |
| `TestKillReturnsErrorWhenKillFails` | A non-not-found `Kill` failure returns `shouldReap=false` with `CodeKillFailed`, releasing the handle |
| `TestKillReturnsKillingProcessWasASuccess` | A successful kill returns no error, records the killed PID, and releases the pinned handle |

### `internal/application/score/converter_test.go` (8 tests)

| Test | Description |
|------|-------------|
| `TestToEntryMapsFields` | `ToEntry` maps a `game.PlayResult`'s Kills/FreedMem/Speed/Duration and the dud count (`len(Duds)`) plus the configured time limit into a `score.Entry` |
| `TestToBoardMapsFields` | `toBoard` maps a persisted `outbound.ScoreBoard`'s entries into `score.Entry` values and seeds the board's high score |
| `TestToBoardEmptyInput` | `toBoard` on an empty `ScoreBoard` returns an empty board with high score 0 |
| `TestToScoreBoardMapsFields` | `toScoreBoard` maps a `score.Board`'s entries back into `outbound.ScoreEntry` values |
| `TestToScoreBoardEmptyInput` | `toScoreBoard` on an empty `score.Board` returns an empty `outbound.ScoreBoard` |
| `TestToScoreSummaryMapsFields` | `toScoreSummary` maps duration/kills/duds/freedMem plus the board's entries into an `outbound.ScoreSummary` |
| `TestToScoreSummaryNewHighScoreTrueWhenBeatsRecord` | `toScoreSummary`'s `NewHighScore` is true when the given kill count beats the board's current record |
| `TestToScoreSummaryNewHighScoreFalseWhenDoesNotBeatRecord` | `NewHighScore` is false when the given kill count falls short of the board's record |

### `internal/application/score/service_test.go` (11 tests)

| Test | Description |
|------|-------------|
| `TestLoadScoreBoardReturnsMappedEntries` | `LoadScoreBoard` maps stored entries and returns the correct high score with no error |
| `TestLoadScoreBoardReturnsNotFoundErrorWhenScoreBoardIsNotFound` | A `NotFoundError` from the store yields `CodeScoreLoadFailed`/`SeverityWarning`, an empty board, and high score 0 |
| `TestLoadScoreBoardReturnsUnderlyingError` | A generic store load error is wrapped as `CodeScoreLoadFailed` with the cause still reachable via `errors.Is` |
| `TestRecordScoreReturnsNilWhenBoardIsSavedSuccessfully` | `RecordScore` appends the given entry to the board and saves it, with all fields preserved |
| `TestRecordScoreDoesNotAddZeroKillEntryToBoardOrSave` | A zero-kill entry (instant quit) is not added to the board and the persisted board stays empty — regression guarding against zero-kill sessions polluting the score board |
| `TestRecordScoreMergesWithConcurrentlyPersistedEntries` | `RecordScore` re-reads the store before saving, so an entry saved by another process after this session's `Load` is not discarded |
| `TestRecordScoreSavesWhenScoreBoardWasNotFound` | A `NotFoundError` load failure (fresh/never-persisted board) still allows the new entry to be saved |
| `TestRecordScoreSavesWhenScoreBoardWasCorrupted` | A `CorruptedDataError` load failure (unrecoverable board) still allows the new entry to be saved |
| `TestRecordScoreSkipsSaveAndReturnsWarningWhenLoadFailed` | A generic (non-not-found, non-corrupted) load failure skips the save entirely and returns `CodeScoreSaveFailed` wrapping the original cause |
| `TestRecordScoreSaveErrorReturnsUnderlyingError` | A store save error is wrapped as `CodeScoreSaveFailed` with the cause reachable via `errors.Is` |
| `TestReportResultsReportsSummaryViaReporter` | `ReportResults` calls the score reporter with kills/duds/freedMem/duration, high-score status, and board entries |

### `internal/application/runner_test.go` (6 tests)

| Test | Description |
|------|-------------|
| `TestRunnerRunReturnsErrorWhenSpeedTooLow` | `Run` rejects a speed below `movement.MinSpeed` as a fatal config error |
| `TestRunnerRunReturnsErrorWhenSpeedTooHigh` | `Run` rejects a speed above `movement.MaxSpeed` as a fatal config error |
| `TestRunnerRunReturnsErrorWhenTimeLimitIsNegative` | `Run` rejects a negative `TimeLimit` as a fatal config error |
| `TestRunnerRunReturnsErrorWhenNoPatternsAreProvided` | `Run` with no patterns fails fatally with "at least one search pattern is required" |
| `TestRunnerRunReturnsErrorWhenPatternTooShort` | Patterns of 1 or 2 characters each fail fatally |
| `TestRunnerRunReturnsErrorWhenProcessDiscoveryFails` | A process-discovery error from `FindProcesses` is classified as `CodeProcessDiscoveryFailed`/`SeverityFatal` |

### `internal/application/runner_integration_test.go` (10 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationRunnerRunIsSuccessful` | Immediately quitting a session with zero kills runs to completion and does not persist any score entry |
| `TestIntegrationRunnerIncludeRootFalseExcludesRootOwnedProcess` | Without `--include-root`, a root-owned-only match set is excluded and `Run` fails with `CodeProcessNotFound` |
| `TestIntegrationRunnerIncludeRootTrueIncludesRootOwnedProcess` | With `IncludeRoot: true`, a root-owned process is included and the session completes without error |
| `TestIntegrationRunnerRunReturnsErrorWhenRendererInitFails` | A renderer `Init` failure surfaces as `CodeGameFailed`/`SeverityFatal` |
| `TestIntegrationRunnerRunSkipsSaveWhenLoadingScoreBoardFails` | A generic score-load failure causes `Run` to still succeed but skip saving the board |
| `TestIntegrationRunnerRunDoesNotFailWhenSavingScoreFails` | A score-save failure does not fail `Run` overall |
| `TestIntegrationRunnerRunWillSaveScoreWhenScoreBoardWasNotFound` | A fresh (never-persisted) score board is still saved after the session |
| `TestIntegrationRunnerRunWillQuitOnQuitEvent` | An asynchronously-sent quit event causes `Run` to return promptly (under 1s), not run indefinitely |
| `TestIntegrationRunnerRunWillQuitWhenTimeLimitExpires` | A 1-second time limit ends the session on its own within 3 seconds |
| `TestIntegrationRunnerRunSignalGoroutineDoesNotAccumulate` | Warms up `os/signal`'s background goroutine, runs 3 sequential games, and asserts goroutine count doesn't grow — the signal-handling goroutine inside `runLoop` exits cleanly every time |

---

## Composition Layer

### `internal/composition/runner_creator_test.go` (2 tests)

| Test | Description |
|------|-------------|
| `TestCreateReturnsErrorWhenPSIsNotOnPath` | `RunnerCreator.Create()` fails with a `"ps not found"` error when `PATH` is empty |
| `TestCreateReturnsErrorWhenTerminalIsUnavailable` | `Create()` fails with a `"failed to create screen"` error when `TERM` is unset |

---

## Entrypoint Layer

### `internal/entrypoint/cli/error_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestArgumentErrorReturnsWrappedMessage` | `ArgumentError.Error()` returns the wrapped cause's message directly |
| `TestArgumentErrorMatchesRegardlessOfCause` | `errors.As` matches an `ArgumentError` regardless of what `Cause` it wraps |
| `TestArgumentErrorUnwrapsToCause` | `Unwrap()` returns the exact same `Cause` error instance |

### `internal/entrypoint/cli/flag_splitter_test.go` (13 tests)

| Test | Description |
|------|-------------|
| `TestSplitAllPatterns` | All-positional args are returned entirely as patterns with no flag args |
| `TestSplitPatternThenBoolFlag` | A pattern followed by a bool flag correctly separates the two |
| `TestSplitBoolFlagThenPattern` | A bool flag followed by a pattern correctly separates the two, regardless of order |
| `TestSplitValueFlagWithInlineValue` | A `--flag=value` value flag is kept as one flag-arg token, distinct from patterns |
| `TestSplitValueFlagWithSeparateValue` | A value flag given as two separate tokens (`--speed 3.5`) is kept together as flag args |
| `TestSplitPatternsAndFlagsFullyInterspersed` | Patterns and flags (bool and value, inline and separate) can be freely interspersed and are still correctly separated |
| `TestSplitUnknownFlag` | An unrecognized flag returns an error matching `flag`'s own "flag provided but not defined" message |
| `TestSplitValueFlagMissingItsValue` | A value flag given with no following value returns a "flag needs an argument" error |
| `TestSplitLoneDashIsAPattern` | A single `-` token is treated as a pattern, not a flag |
| `TestSplitValueFlagFollowedByFlagLikeTokenReportsMissingValue` | A value flag immediately followed by another flag-like token is treated as a missing value, not as consuming that token |
| `TestSplitDoubleDashTreatsRemainingArgsAsPatterns` | `--` ends flag processing; everything after it (even flag-shaped strings) is treated as a pattern |
| `TestSplitDoubleDashAsLastArg` | A trailing `--` with nothing after it is a no-op |
| `TestSplitEmptyArgs` | `nil` args produce empty patterns and empty flag args with no error |

### `internal/entrypoint/cli/program_test.go` (24 tests, 21 subtests)

| Test | Description |
|------|-------------|
| `TestRunNoArgsPrintsUsageAndReturnsNil` | No arguments at all prints usage and returns no error |
| `TestRunHelpFlagPrintsUsageAndReturnsNil/--help` | `--help` prints usage and returns no error |
| `TestRunHelpFlagPrintsUsageAndReturnsNil/-h` | `-h` prints usage and returns no error |
| `TestRunHelpWinsRegardlessOfPositionOrOtherErrors/--help --unknown` | Help short-circuits even when combined with an unknown flag, help first |
| `TestRunHelpWinsRegardlessOfPositionOrOtherErrors/--unknown --help` | Help short-circuits even when combined with an unknown flag, help last |
| `TestRunHelpWinsRegardlessOfPositionOrOtherErrors/proc --speed --help` | Help wins even after a pattern and a malformed value flag |
| `TestRunHelpWinsRegardlessOfPositionOrOtherErrors/--help --speed` | Help wins regardless of what follows it |
| `TestRunMalformedFlagReturnsArgumentError/proc --help=true` | `--help=true` (not an exact `-h`/`--help` token) falls through as an unrecognized flag and errors |
| `TestRunMalformedFlagReturnsArgumentError/proc --help=false` | Same as above for `--help=false` |
| `TestRunMalformedFlagReturnsArgumentError/proc --unknown` | An unrecognized flag returns an `ArgumentError` |
| `TestRunMalformedFlagReturnsArgumentError/proc --speed` | A value flag missing its value returns an `ArgumentError` |
| `TestRunMalformedFlagReturnsArgumentError/proc --speed=abc` | A non-numeric value for a float flag returns an `ArgumentError` |
| `TestRunMalformedFlagReturnsArgumentError/proc --time=abc` | A non-numeric value for an int flag returns an `ArgumentError` |
| `TestRunBasicPattern` | A single pattern with no flags produces the expected pattern list and default `ConfirmMode=false`, `Speed=2.0`, `TimeLimit=30` |
| `TestRunMultiplePatterns` | Multiple positional patterns are captured in order |
| `TestRunConfirmFlag` | `--confirm` sets `ConfirmMode=true` |
| `TestRunIncludeRootFlag` | `--include-root` sets `IncludeRoot=true` |
| `TestRunIncludeRootDefaultsToFalse` | `IncludeRoot` defaults to `false` when the flag is omitted |
| `TestRunSpeedFlag/valid speed` | `--speed=3.5` is parsed into `Config.Speed` |
| `TestRunSpeedFlag/min speed` | `--speed=0.5` (the minimum) is accepted |
| `TestRunSpeedFlag/max speed` | `--speed=5.0` (the maximum) is accepted |
| `TestRunSpeedFlag/invalid` | A non-numeric speed value errors |
| `TestRunTimeFlag/valid time` | `--time=60` is parsed into `Config.TimeLimit` |
| `TestRunTimeFlag/no limit` | `--time=0` is accepted and means no limit |
| `TestRunTimeFlag/invalid` | A non-numeric time value errors |
| `TestRunPatternAfterFlag` | A pattern appearing after a flag is still captured alongside one appearing before it |
| `TestRunFlagsSurroundingPatterns` | Patterns and flags (value and bool) freely interspersed all parse into the correct `Config` |
| `TestRunFlagsWithoutPatternsIsRejected` | Flags with zero patterns is rejected as an error |
| `TestRunNoArgsDoesNotConstructRunner` | With no arguments, the `RunnerCreator` is never invoked |
| `TestRunHelpDoesNotConstructRunner` | `--help` never invokes the `RunnerCreator` |
| `TestRunInvalidFlagsDoesNotConstructRunner` | An unknown flag never invokes the `RunnerCreator` |
| `TestRunInvalidConfigDoesNotConstructRunner` | A config that fails validation (e.g. flags with no patterns) never invokes the `RunnerCreator` |
| `TestRunReturnsErrorWhenRunnerCreatorFails` | If `RunnerCreator.Create` errors, `Run` surfaces that error and the creator was called exactly once |
| `TestRunNoArgsPrintsUsageToStdout` | With no arguments, usage text is written to stdout and nothing to stderr |
| `TestRunHelpPrintsUsageToStdout/--help` | `--help` writes usage to stdout, nothing to stderr |
| `TestRunHelpPrintsUsageToStdout/-h` | `-h` writes usage to stdout, nothing to stderr |
| `TestRunUnknownFlagPrintsUsageToStderr` | An unknown flag writes usage to stderr, nothing to stdout |
| `TestRunInvalidConfigPrintsUsageToStderr` | An invalid config writes usage to stderr, nothing to stdout |
| `TestRunReturnsErrorOnFatal` | A `Fatal`-severity `apperror.Error` returned by the constructed `Runner` propagates back out of `Run` with its message intact |

---

## Infrastructure Layer

### `internal/infrastructure/osprocess/process_test.go` (12 tests, 14 subtests)

| Test | Description |
|------|-------------|
| `TestDiscoverReturnsResults` | A real `ps`-backed `Discover()` call returns a non-empty process list |
| `TestOwnPIDMatchesOSGetpid` | `OwnPID()` matches `os.Getpid()` |
| `TestOwnUIDMatchesOSGeteuid` | `OwnUID()` matches `os.Geteuid()` |
| `TestLookupNameReturnsOwnName` | `LookupName` on the running test binary's own PID returns a non-empty name |
| `TestParseProcesses/single_process` | A single `uid pid rss stat comm` row parses into one `ProcessInfo`, RSS converted from KB to bytes |
| `TestParseProcesses/collapses_internal_whitespace` | Multiple spaces inside a process name collapse to single spaces |
| `TestParseProcesses/reports_a_zombie's_state_verbatim` | A `ZN`-state row keeps the full state string (`"ZN"`), not just the leading `Z` |
| `TestParseProcesses/skips_rows_with_too_few_fields` | A malformed row with fewer than 5 fields is silently skipped, valid rows still parse |
| `TestParseProcesses/skips_rows_with_a_non-numeric_uid` | A row with a non-numeric UID column is dropped entirely |
| `TestParseProcesses/skips_rows_with_a_non-numeric_pid` | A row with a non-numeric PID column is dropped entirely |
| `TestParseProcesses/skips_rows_with_a_non-numeric_rss` | A row with a non-numeric RSS column is dropped entirely |
| `TestParseProcesses/no_data_rows` | Header-only output (no process rows) parses to an empty slice, no error |
| `TestParseProcessesRejectsMissingHeader` | Output missing the expected header line is rejected as an error |
| `TestIsZombie/running` | State `"R"` is not a zombie |
| `TestIsZombie/running_in_foreground` | State `"R+"` is not a zombie |
| `TestIsZombie/sleeping` | State `"S"` is not a zombie |
| `TestIsZombie/zombie` | State `"Z"` is a zombie |
| `TestIsZombie/zombie_with_trailing_modifier_flags` | State `"ZN"` (zombie plus a modifier flag) is still detected as a zombie |
| `TestIsZombie/empty` | An empty state string is not a zombie |
| `TestLookupNameReportsNotFoundForZombie` | `LookupName` on a PID whose `ps` row reports a zombie state returns `outbound.NotFoundError`, not the zombie's stale name |
| `TestLookupNameReportsNotFoundWhenPIDIsMissing` | `LookupName` on a PID absent from `ps` output entirely returns the same `NotFoundError` as a zombie |
| `TestDiscoverTimesOutWhenPsHangs` | `Discover` returns an error once its configured timeout elapses, rather than blocking for the full (hung) `ps` runtime |
| `TestLookupNameTimesOutWhenPsHangs` | `LookupName` likewise returns once its timeout elapses against a hung `ps` |
| `TestDiscoverErrorIncludesPsStderr` | A failing `ps` invocation's stderr text (e.g. `"permission denied"`) is surfaced in `Discover`'s returned error, not just an opaque exit status |

### `internal/infrastructure/osprocess/process_integration_test.go` (5 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationDiscoverReturnsResults` | A real, fully-constructed `osprocess.NewProcess()` adapter's `Discover()` returns a non-empty list |
| `TestIntegrationDiscoverValidFields` | Every discovered process has a positive PID, non-empty name, and non-negative RSS |
| `TestIntegrationDiscoverShortProcessNames` | Discovered process names are base executable names, never a full path |
| `TestIntegrationFindExcludesOwnAndInitPID` | `process.Find` run against real adapter output excludes the caller's own PID and PID 1, and (for a non-root caller) only returns processes it owns |
| `TestIntegrationLookupNameResistsArgv0Spoofing` | `LookupName` on a process started with a spoofed `argv[0]` still reports the real executable name (`"sleep"`), not the spoofed one |

### `internal/infrastructure/osprocess/process_handle_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestPinReturnsHandleForExistingPID` | `Pin` on the test binary's own live PID returns a non-nil handle without error |
| `TestReleaseSucceedsForExistingPID` | `Release` on a handle pinned to a live PID succeeds |
| `TestKillerNonexistentPID` | `Kill` on a handle pinned to a PID that doesn't exist returns `outbound.NotFoundError` |

### `internal/infrastructure/osprocess/process_handle_integration_test.go` (1 test) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegrationPinThenKillReportsNotFoundAfterProcessExits` | A handle pinned *before* its process exits still reports `NotFoundError` (not a silent success) when `Kill` is called after the process has already exited and been reaped — the PID-pinning race-closure behavior |

### `internal/infrastructure/tcellui/animation_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestAnimationFrameAtZeroProgress` | `animation.frame(0)` returns the first frame |
| `TestAnimationFrameAtMidProgress` | `animation.frame(0.5)` returns the middle frame of a 3-frame animation |
| `TestAnimationFrameClampsAtUpperBound` | `animation.frame(1.0)` returns the last frame |
| `TestAnimationFrameClampsAtNegativeProgress` | `animation.frame(-0.5)` clamps to the first frame rather than indexing out of range |

### `internal/infrastructure/tcellui/hud_test.go` (3 tests, 4 subtests)

| Test | Description |
|------|-------------|
| `TestDrawHUDNarrowTerminalSuppressesCenter` | On a 30-column terminal, the centered `HighScore` HUD element is omitted while `FREED`/`KILLS` still render |
| `TestDrawHUDWideTerminalDrawsAllThree` | On a standard-width terminal, all three HUD elements (`FREED`, `HighScore`, `KILLS`) render on row 0 |
| `TestDrawHUDFreedLabelBudgetedAgainstKillsColumn/w=10` | At 10 columns, only the `KILLS` segment fits and renders |
| `TestDrawHUDFreedLabelBudgetedAgainstKillsColumn/w=14` | At 14 columns, `FREED` is truncated so it doesn't overrun into the `KILLS` column |
| `TestDrawHUDFreedLabelBudgetedAgainstKillsColumn/w=18` | At 18 columns, `FREED:` renders with its value truncated away, still not overrunning `KILLS` |
| `TestDrawHUDFreedLabelBudgetedAgainstKillsColumn/w=24` | At 24 columns, the full `FREED: 0 B` renders alongside `KILLS: 7` with correct spacing |

### `internal/infrastructure/tcellui/poller_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestPollGoroutineExitsAfterCleanup` | The poll goroutine started by `Init` terminates within a second of `Cleanup()`, after 15 injected key events |
| `TestEventsChannelClosesAfterCleanup` | The `InputEvents().Events()` channel is closed after `Cleanup()` |
| `TestPollForwardsTranslatedEventToChannel` | An injected `'q'` keypress is translated and delivered on `InputEvents().Events()` as `input.QuitEvent{}` |

### `internal/infrastructure/tcellui/statusbar_test.go` (9 tests, 3 subtests)

| Test | Description |
|------|-------------|
| `TestDrawStatusBarNormal` | The status bar shows `"Targets:"`, `"Speed:"`, and `"Click to kill"` in the normal (non-confirming) state |
| `TestDrawStatusBarConfirming` | During a confirm prompt, the status bar shows the target's PID, name, `"(Y)es"`, and `"(N)o"` |
| `TestDrawStatusBarConfirmingMultiByteName` | A multi-byte process name (`"café-server"`) renders intact in the confirm prompt, with no byte-offset gap |
| `TestDrawStatusBarConfirmingLongNameStillShowsAllOptions/w=40` | At 40 columns, a long process name doesn't crowd out `(Y)es`/`(N)o`/`(Q)uit` |
| `TestDrawStatusBarConfirmingLongNameStillShowsAllOptions/w=44` | Same guarantee at 44 columns |
| `TestDrawStatusBarConfirmingLongNameStillShowsAllOptions/w=50` | Same guarantee at 50 columns |
| `TestDrawStatusBarConfirmingLongNameIsTruncatedWithEllipsis` | At 44 columns, a long name is truncated with a trailing `"…"` |
| `TestDrawStatusBarConfirmingNameBudgetOfOneIsJustEllipsis` | At a name budget of exactly one column, the name renders as just `"…"` |
| `TestDrawStatusBarConfirmingNameOmittedWhenNoBudgetLeft` | With no budget left for the name at all, it's dropped entirely rather than partially shown |
| `TestDrawStatusBarWithTimeLimit` | With `StatusViewState.TimeLeft` set (via the `*int` pointer), the bar shows `"Time:"` and the remaining seconds |
| `TestDrawStatusBarNoTimeLimit` | With `TimeLeft` left nil (untimed), the `"Time:"` segment is absent entirely |

### `internal/infrastructure/tcellui/target_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestRenderMultiByteLabelColumnLayout` | A tag starting with a multi-byte rune renders with correct column alignment for the runes that follow |
| `TestRenderKillingIgnoresTagAndShowsAnimationFrame` | A `Killing` target ignores its `Tag` and draws the kill-animation's first frame (`"💥"`) instead |
| `TestRenderKillingShowsTextFrameAtMidProgress` | At 30% kill-animation progress, the frame includes the text `"KILLED"` |
| `TestRenderFleeingIgnoresTagAndShowsAnimationFrame` | A `Fleeing` target ignores its `Tag` and draws the flee-animation's first frame (`"🏃"`) instead |
| `TestRenderWideRuneTagDoesNotDropCharacters` | A tag containing wide East-Asian runes renders every character without dropping any |
| `TestRenderDrawsMultipleTargets` | Two targets at different rows are both drawn correctly, not just the first |
| `TestRenderClipsTargetAtHUDRow` | A target positioned on row 0 (the HUD row) is not drawn there, so it can't visually collide with the HUD |

### `internal/infrastructure/tcellui/translator_test.go` (2 tests, 17 subtests)

| Test | Description |
|------|-------------|
| `TestTranslateKeyEvent/escape_quits` | `KeyEscape` translates to `input.QuitEvent{}` |
| `TestTranslateKeyEvent/ctrl+c_quits` | `KeyCtrlC` translates to `input.QuitEvent{}` |
| `TestTranslateKeyEvent/ctrl+z_quits` | `KeyCtrlZ` translates to `input.QuitEvent{}` |
| `TestTranslateKeyEvent/q_quits` | Rune `'q'` translates to `input.QuitEvent{}` |
| `TestTranslateKeyEvent/uppercase_Q_quits` | Rune `'Q'` also translates to `input.QuitEvent{}` |
| `TestTranslateKeyEvent/y_confirms_accept` | Rune `'y'` translates to `input.ConfirmEvent{Accept: true}` |
| `TestTranslateKeyEvent/uppercase_Y_confirms_accept` | Rune `'Y'` also translates to `ConfirmEvent{Accept: true}` |
| `TestTranslateKeyEvent/n_confirms_decline` | Rune `'n'` translates to `input.ConfirmEvent{Accept: false}` |
| `TestTranslateKeyEvent/uppercase_N_confirms_decline` | Rune `'N'` also translates to `ConfirmEvent{Accept: false}` |
| `TestTranslateKeyEvent/+_speeds_up` | Rune `'+'` translates to `input.SpeedEvent{Faster: true}` |
| `TestTranslateKeyEvent/=_speeds_up` | Rune `'='` is an alias for `+`, also translates to `SpeedEvent{Faster: true}` |
| `TestTranslateKeyEvent/-_speeds_down` | Rune `'-'` translates to `input.SpeedEvent{Faster: false}` |
| `TestTranslateKeyEvent/_speeds_down` | Rune `'_'` is an alias for `-`, also translates to `SpeedEvent{Faster: false}` |
| `TestTranslateKeyEvent/unrecognized_rune_dropped` | An unmapped rune (`'z'`) is dropped |
| `TestTranslateMouseEvent/button1_emits_click` | A `Button1` mouse event translates to `input.ClickEvent{X, Y}` |
| `TestTranslateMouseEvent/non-button1_dropped` | A `Button2` mouse event is dropped |
| `TestTranslateMouseEvent/click_on_row_0_emits_click` | A click on row 0 still translates to a `ClickEvent` at the translator level — row-0 filtering happens elsewhere (in `poll()`) |

### `internal/infrastructure/tcellui/tui_test.go` (6 tests)

| Test | Description |
|------|-------------|
| `TestNewTUIPanicsOnNilScreen` | `NewTUI(nil)` panics rather than constructing an unusable TUI |
| `TestCleanupBeforeInitDoesNotPanicOnRealScreen` | Calling `Cleanup()` before `Init()` on a real (non-simulation) screen does not panic |
| `TestInitSecondCallReturnsErrorAndDoesNotSpawnSecondPollGoroutine` | A second `Init()` call returns an error and does not spawn a second poll goroutine |
| `TestChromeSizeReservesOneRowTopAndBottom` | `ChromeSize()` reports one reserved row at the top and one at the bottom |
| `TestWindowSizeReturnsScreenDimensions` | `WindowSize()` reflects the screen's current dimensions |
| `TestRenderClearsStaleContentFromPreviousFrame` | A target drawn in one frame is cleared once a subsequent `Render` call no longer includes it |

### `internal/infrastructure/filescore/file_score_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestNewFileScoreResolvesDefaultPath` | `NewFileScore()` resolves its path to `~/.config/pidshooter/highscores.json` under `$HOME` |
| `TestNewFileScoreReturnsErrorWhenHomeUnset` | `NewFileScore()` errors when neither `$HOME` nor `$XDG_CONFIG_HOME` is set |
| `TestLoadFileNotExistReturnsNotFoundError` | Loading a nonexistent score file returns `outbound.NotFoundError` and an empty board |
| `TestLoadInvalidJSONReturnsCorruptedDataError` | Loading malformed JSON returns `outbound.CorruptedDataError` |
| `TestSaveCreatesFile` | `Save` creates the score file on disk if it doesn't already exist |
| `TestSaveFilePermissions` | The saved file has `0600` permissions |
| `TestSaveLoadRoundTrip` | A single saved entry (including the `Duds` field) survives a save/load cycle intact |
| `TestSaveLoadMultipleEntries` | Three saved entries survive a save/load round trip with order and values preserved |
| `TestSaveOverwritesPreviousFile` | A second `Save` fully replaces the first — only the latest data is visible on load |
| `TestSaveToNestedNonexistentDirectoryCreatesParentDirs` | `Save` creates all necessary parent directories for a deeply nested path |
| `TestSaveDoesNotLeaveTempFileAfterSuccess` | No leftover temp file remains in the directory after a successful save (atomic-write cleanup) |
| `TestSaveWritesJSONMatchingOnDiskSchema` | The raw on-disk JSON matches the documented schema exactly: a `version` field plus `scores` entries with `kills`, `duds`, `freed_mem`, `speed`, `time_limit`, `duration_secs`, and RFC3339Nano `date` keys |
| `TestLoadAcceptsFileWithoutVersionField` | A score file with no `version` field at all still loads successfully (treated as an old/default schema) |
| `TestLoadRejectsNewerSchemaVersion` | A file declaring a schema version newer than this build supports is rejected — an error that is neither `NotFoundError` nor `CorruptedDataError`, since it's valid data from a future version, not damage |
| `TestLoadRejectsFileOverMaxSize` | A file larger than the configured max size is rejected as `CorruptedDataError` mentioning the size limit |

### `internal/infrastructure/fsutil/fsutil_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestConfigDirJoinsHomeAndAppName` | `ConfigDir` joins `$HOME/.config` with the given app name when `$XDG_CONFIG_HOME` is unset |
| `TestConfigDirReturnsErrorWhenHomeUnset` | `ConfigDir` errors when neither `$HOME` nor `$XDG_CONFIG_HOME` is set |
| `TestConfigDirPrefersXDGConfigHomeWhenSet` | `ConfigDir` uses `$XDG_CONFIG_HOME` in preference to `$HOME/.config` when it's set |
| `TestConfigDirIgnoresRelativeXDGConfigHome` | A relative (non-absolute) `$XDG_CONFIG_HOME` is ignored, falling back to `$HOME/.config` |
| `TestWriteFileAtomicCreatesParentDirectory` | `WriteFileAtomic` creates any missing parent directories for the target path |
| `TestWriteFileAtomicSetsPermissions` | The written file has exactly the requested permission bits |
| `TestWriteFileAtomicOverwritesExistingFile` | A second atomic write fully replaces the first, with no leftover temp file |
| `TestWriteFileAtomicLeavesExistingFileUntouchedOnFailure` | If the write is forced to fail (parent directory made untraversable), the original file's contents are left completely untouched, with no leftover temp file |
| `TestWriteFileAtomicTightensExistingDirectoryPermissions` | `WriteFileAtomic` tightens an overly permissive existing config directory (`0777`) down to `0700` as a side effect of writing to it |

### `internal/infrastructure/console/process_reporter_test.go` (1 test)

| Test | Description |
|------|-------------|
| `TestProcessReporterReportPrintsCountAndPatterns` | `Report` prints the match count and the pattern list |

### `internal/infrastructure/console/score_reporter_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestReportPrintsGameOverSummary` | `Report` prints `"Game Over!"` along with the kills and duds counts |
| `TestReportPrintsTrophyWhenNewHighScore` | A `NewHighScore: true` summary prints a "New high score" line |
| `TestReportOmitsTrophyWhenNotNewHighScore` | A `NewHighScore: false` summary omits the "New high score" line |
| `TestReportPrintsNoScoresMessageWhenEmpty` | An empty summary prints `"No high scores yet!"` |
| `TestReportPrintsTable` | A summary with entries prints a table containing `"Kills"` and `"Freed"` column headers |
| `TestReportPrintsDudsColumn` | The score table includes a `"Duds"` column, with kills and duds values shown in the correct column order |
| `TestReportPrintsDuration` | The score table includes a `"Time"` column formatting duration like `"12.3s"` |

---

## Utility

### `internal/util/util_test.go` (1 test, 8 subtests)

| Test | Description |
|------|-------------|
| `TestFormatBytes/0_B` | 0 bytes formats as `"0 B"` |
| `TestFormatBytes/512_B` | 512 bytes formats as `"512 B"` |
| `TestFormatBytes/1.0_KB` | 1024 bytes formats as `"1.0 KB"` |
| `TestFormatBytes/1.5_KB` | 1536 bytes formats as `"1.5 KB"` |
| `TestFormatBytes/1.0_MB` | 1,048,576 bytes formats as `"1.0 MB"` |
| `TestFormatBytes/1.5_MB` | 1,572,864 bytes formats as `"1.5 MB"` |
| `TestFormatBytes/1.0_GB` | 1,073,741,824 bytes formats as `"1.0 GB"` |
| `TestFormatBytes/1.5_GB` | 1,610,612,736 bytes formats as `"1.5 GB"` |

---

## Binary (`cmd/pidshooter`)

### `cmd/pidshooter/main_test.go` (6 tests)

| Test | Description |
|------|-------------|
| `TestExitCodeReturnsZeroForNil` | `exitCode(nil)` is 0 |
| `TestExitCodeReturnsTwoForArgumentError` | A `cli.ArgumentError` maps to exit code 2 |
| `TestExitCodeReturnsTwoForWrappedInvalidConfig` | An `apperror` with `CodeInvalidConfig` wrapped inside an `ArgumentError` still maps to exit code 2 |
| `TestExitCodeReturnsOneForUnwrappedInvalidConfig` | An `apperror` with `CodeInvalidConfig` *not* wrapped in an `ArgumentError` maps to exit code 1, not 2 — the wrapping, not just the code, determines the exit status |
| `TestExitCodeReturnsOneForOtherErrors` | An `apperror` with a different code (`CodeGameFailed`) maps to exit code 1 |
| `TestExitCodeReturnsOneForPlainError` | A plain, unclassified error maps to exit code 1 |

### `cmd/pidshooter/main_acceptance_test.go` (1 test) `// go:build acceptance`

| Test | Description |
|------|-------------|
| `TestAcceptanceHelpSucceedsWithoutTerm` | Builds the real `pidshooter` binary and runs `pidshooter --help` with `$TERM` stripped from the environment, asserting it exits successfully and prints usage — confirms `--help` never touches the terminal |

---

## Highlights

**Three build-tag categories, not two** — plain (unit, no tag, runs offline with fakes), `integration` (touches a real OS process table, a real terminal screen, or runs a full game loop through the composed `Runner`), and a third, `acceptance` (`cmd/pidshooter/main_acceptance_test.go`), which builds the actual compiled binary and runs it as a subprocess. Run all three with `go test -tags "integration acceptance" ./...`; plain `go test ./...` runs only the unit suite.

**A `Fleeing` state exists in parallel with `Killing` everywhere it matters** — `Tag`, `AnimationProgress`, `Update`, `isAlive`/`isDead`, the renderer's style/animation selection, and the score's tracking are all exercised for both. This is the "target's backing process already exited before the kill could land" path (a *dud*), distinct from a player-initiated kill, complete with its own animation (`🏃💨 → ↝ RAN AWAY ↝ → · · · → · → blank`) and its own orange render style.

**`Target` no longer embeds `process.Info`/`movement.Motion`** — `session_test.go` and `target_test.go` construct targets via `Info: process.NewInfo(...)` and `Motion: movement.Motion{...}` as named (not embedded) fields, and read them back as `e.Info.PID`/`e.Motion.Position.X`. `target_test.go` also carries a deliberate quartet of tests (`TestTargetMoveNoOpWhenKilling`/`Fleeing`/`Dead` plus the `...WhenAlive` control case) that call the unexported `move` method directly, bypassing `Update`'s normal dispatch, specifically to prove the `State` guard holds for any caller — not just the intended path.

**Kill safety is now a real pin-then-verify-then-kill pipeline** — `application/process.Service.Kill` pins the target PID (`outbound.Process.Pin` → `ProcessHandle`) before re-verifying its live name, so a PID recycled by the OS between discovery and kill can't be silently signaled in the wrong process's place. `osprocess`'s own `process_handle_test.go`/`process_handle_integration_test.go` verify this at the adapter level: a handle obtained before a process exits still correctly reports `NotFoundError` on `Kill`, never a false success. `shouldReap` semantics are carefully distinguished per failure point — a transient lookup error is *not* a reap, but a not-found lookup or a not-found kill *is*.

**Reporting moved behind real ports** — `application/process.Service.FindProcesses` and `application/score.Service` used to print directly; both now hold an injected `outbound.ProcessReporter`/`outbound.ScoreReporter`, implemented by the new `infrastructure/console` package. Every process/score test now injects a fake reporter alongside the store/process fakes, and reporting behavior (whether it's called, with what arguments) is asserted independently of the surrounding operation's success/failure.

**Structured, classified errors flow through one chokepoint** — `application/apperror.Handle` turns any `Code`+`Severity`-classified error into either `nil` (absorbed, after logging once, for `Warning`/`Severity`) or the original error returned unchanged (for `Fatal` or anything unclassified). It correctly unwraps through `fmt.Errorf("%w", ...)` wrapping to find severity while still returning the outer wrapped error object. `cmd/pidshooter/main_test.go`'s `exitCode` tests reveal a subtlety worth knowing: only an `apperror` wrapped inside a `cli.ArgumentError` maps to exit code 2 (a usage problem) — the same code left unwrapped maps to exit code 1 (a runtime failure).

**`cobra` is gone** — `entrypoint/cli`'s flag parsing is now hand-rolled: a `flagSplitter` walks raw argv splitting positional patterns from flag tokens (inline `=` values, separate-token values, and `--` as an explicit end-of-flags marker) before handing the remainder to the stdlib `flag.FlagSet`. `program_test.go` separately verifies that a `Runner` is never even constructed for the "shouldn't get that far" cases (no args, `--help`, invalid flags, invalid config) and that usage text routes to stdout on success/help vs. stderr on error.

**The composition root is now its own package, and is tested as pure wiring** — `internal/composition.RunnerCreator` is new; its two tests only cover its two adapter-construction failure paths (`ps` missing, terminal unavailable), both fatal.

**Score persistence gained real domain safety** — `core/score.Board.Add`/`NewBoard` now enforce that an entry must have `Kills > 0` and non-negative `FreedMem` to be recorded at all (`Entry.isScore()`), so a zero-kill instant-quit session, or a hand-edited/corrupted file with negative values, can no longer pollute the leaderboard. `NewBoard` re-validates, re-sorts, and re-caps whatever it's given rather than trusting it. `TestBoardAddFullTieKeepsInsertionOrder` locks in that a full tie across every ranked dimension deterministically preserves insertion order (`sort.SliceStable`), rather than relying on `sort.Slice`'s explicitly-unspecified tie behavior. `Entry.beats` also gained a `Duds` tiebreak dimension, ranking Kills → Duds → Speed → Duration → FreedMem.

**Score file storage gained schema evolution and atomicity guarantees** — `infrastructure/filescore` (renamed from `scorefilestore`) accepts files with no `version` field (treated as an old/default schema), explicitly rejects a *newer* schema version as neither "not found" nor "corrupted" (it's valid data this build just doesn't understand yet), rejects oversized files, and writes atomically via `infrastructure/fsutil.WriteFileAtomic`. `fsutil`'s own tests go further than a happy-path check: `TestWriteFileAtomicLeavesExistingFileUntouchedOnFailure` deliberately breaks a write and asserts the pre-existing file is byte-identical afterward with no leftover temp file — a genuine proof of the write-temp-then-rename design, not just a functional test.

**`tcellui` is decomposed by concern, one file (and one test file) per responsibility** — `tui.go`/`TUI` (the top-level `Renderer`), `input_events.go`/`inputEvents` (the `InputEventProvider`, obtained via `TUI.InputEvents()` and sharing `TUI`'s poller), `poller.go`, `renderer.go`, `hud.go`, `statusbar.go`, `target.go` (drawing), `animation.go`, and `translator.go` — replacing what used to be one monolithic file and type. Several of its tests are explicit regressions for named rendering bugs: multi-byte/wide-rune column layout, a HUD label overrunning the kills column at narrow widths, and targets/HUD rows clipping into each other.

**Timer determinism via explicit clock injection** — `core/game.timer` no longer defaults its clock lazily; `newTimer(limitSeconds, now func() time.Time)` requires it at construction. Tests use a small shared `fakeClock` (mutable `time.Time` with `now()`/`advance(d)`) rather than reaching into private fields or using real `time.Sleep` — no test in `timer_test.go` or `session_test.go` sleeps for real time.

**Ownership and root-targeting are tested end-to-end** — `core/process.Info.IsKillableBy`/`Find` (domain), `application/process.Service.FindProcesses` (service), and `application/runner_integration_test.go` (full `Runner.Run`) each independently verify that `--include-root`/`IncludeRoot` controls whether a root-owned process is a valid target, with the domain-level `IsKillableBy` test explicitly confirming `includeRoot` never leaks into granting access to another *non-root* user's process.

**Test infrastructure**: `internal/testutil/fake` now provides typed doubles for every outbound port (`Process`, `ProcessHandle`, `ProcessKiller`, `ProcessReporter`, `Renderer`, `ScoreReporter`, `Store`, `InputEventProvider`) plus `inbound.Runner` and `composition.RunnerCreator` themselves, so higher-level tests (like `entrypoint/cli/program_test.go`) can substitute the entire application without touching real adapters. `internal/testutil/fixture` builds `*core/game.Session` and `core/process.Info` values; `internal/testutil/helper` provides a small shared `UnsetEnv` helper. There is no `capture` package any more — output-capturing tests inject a `bytes.Buffer` directly instead of intercepting `os.Stdout`/`os.Stderr`.
