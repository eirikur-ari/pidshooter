# Test Cases

Total: 23 test files across 12 packages — 204 test functions (including 20 subtests).

---

## Core Domain Layer

### `internal/core/process/process_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestValidate_TooShort` | Empty string and patterns of 1–2 characters (below `MinPatternLength`) are rejected |
| `TestValidate_ExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestValidate_Valid` | A normal pattern (`"firefox"`) is accepted |
| `TestValidate_ExactMaxLength` | Pattern at exactly `MaxPatternLength` is accepted |
| `TestValidate_TooLong` | Pattern exceeding `MaxPatternLength` is rejected |
| `TestErrNoPatterns_IsSentinel` | `ErrNoPatterns` satisfies `errors.Is` against itself |
| `TestInfo_IsProtected` | PID 0 and PID 1 are protected regardless of name; any other PID is not |
| `TestFind_MatchesByName` | Exact name match returns only the matching process |
| `TestFind_SubstringMatch` | Partial name match is supported |
| `TestFind_CaseInsensitive` | Name matching is case-insensitive |
| `TestFind_MultipleTerms` | OR-semantics: multiple terms each match independently |
| `TestFind_NoMatch` | No matching name returns an empty slice |
| `TestFind_ExcludesPID1` | PID 1 (init) is always excluded from results |
| `TestFind_ExcludesOwnPID` | The caller's own PID is excluded from results |
| `TestFind_ExcludesPID0` | PID 0 (kernel swapper) is excluded alongside PID 1 |

### `internal/core/movement/throttle_test.go` (5 tests)

| Test | Description |
|------|-------------|
| `TestNewThrottle` | `NewThrottle` stores the initial speed |
| `TestThrottle_Increase` | `Increase` raises speed by 0.5 |
| `TestThrottle_Decrease` | `Decrease` lowers speed by 0.5 |
| `TestThrottle_IncreaseCapsAtMax` | Speed is capped at `MaxSpeed` and does not exceed it on repeated increases |
| `TestThrottle_DecreaseFloorsAtMin` | Speed is floored at `MinSpeed` and does not go below it on repeated decreases |

### `internal/core/movement/bounds_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestBounds_Bounce_Left` | Position/velocity crossing the left wall clamps to 0 and reverses X velocity |
| `TestBounds_Bounce_Right` | Position/velocity crossing the right wall (accounting for label width) clamps and reverses X velocity |
| `TestBounds_Bounce_Top` | Position/velocity crossing the top wall clamps to 0 and reverses Y velocity |
| `TestBounds_Bounce_Bottom` | Position/velocity crossing the bottom wall clamps and reverses Y velocity |

### `internal/core/movement/motion_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestMotion_Move_AdvancesPosition` | `Move` advances position by velocity scaled by speed, with no wall interaction |
| `TestMotion_Move_BouncesAtLeftWall` | Motion starting at the left wall bounces: position stays in bounds, velocity flips positive |
| `TestMotion_Move_BouncesAtRightWall` | Motion approaching the right wall bounces: position clamped, velocity flips negative |
| `TestMotion_Move_SpeedScalesVelocity` | Position advances by `velocity × speed` per `Move` call |

### `internal/core/game/state_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestState_DefaultIsPending` | Zero-value `state` reads as `Pending` |
| `TestState_Store_Load` | `Store` followed by `Load` round-trips `Running` and `Stopped` |
| `TestState_CompareAndSwap_Success` | CAS succeeds and updates state when the old value matches |
| `TestState_CompareAndSwap_Failure` | CAS fails and leaves state unchanged when the old value does not match |

### `internal/core/game/confirmation_test.go` (5 tests)

| Test | Description |
|------|-------------|
| `TestConfirmation_Pending_FalseWhenEmpty` | A freshly constructed `Confirmation` has no pending target |
| `TestConfirmation_Request_ConfirmMode_HoldsTarget` | In confirm mode, `Request` holds the target and returns `nil` |
| `TestConfirmation_Request_PassthroughMode_ReturnsTarget` | Outside confirm mode, `Request` returns the target immediately without holding it |
| `TestConfirmation_Accept_ReturnsAndClearsTarget` | `Accept` returns the held target and clears pending state |
| `TestConfirmation_Cancel_ClearsPending` | `Cancel` clears the pending target without returning it |

### `internal/core/game/stats_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestStats_RecordKill` | Each `RecordKill` increments `Kills` by 1 and accumulates freed bytes |
| `TestStats_RecordKill_UpdatesHighScore` | `HighScore` only advances once kills exceed the previously set record |
| `TestStats_SetHighScore` | `SetHighScore` stores the supplied value |
| `TestStats_HighScore` | `HighScore` defaults to 0 and reflects the last value set |

### `internal/core/game/timer_test.go` (10 tests)

| Test | Description |
|------|-------------|
| `TestTimer_Expired_FalseBeforeLimit` | A freshly started timer with a time limit has not expired |
| `TestTimer_Expired_TrueWhenLimitReached` | Timer reports expired once elapsed time passes the limit |
| `TestTimer_Expired_FalseWhenUnlimited` | A zero-limit ("unlimited") timer never expires, however long it has run |
| `TestTimer_Remaining_WithinLimit` | `Remaining` returns a positive duration no greater than the configured limit |
| `TestTimer_Remaining_ZeroWhenExpired` | `Remaining` returns 0 once the limit has passed |
| `TestTimer_Remaining_ZeroWhenUnlimited` | `Remaining` returns 0 for an unlimited timer |
| `TestTimer_SecondsLeft` | `SecondsLeft` returns a value between 1 and the configured limit |
| `TestTimer_LimitSeconds` | `LimitSeconds` reflects the configured limit, including 0 for unlimited |
| `TestTimer_StartTime_ZeroBeforeStart` | `StartTime` is zero before `Start` is called |
| `TestTimer_StartTime_SetAfterStart` | `StartTime` is set to (at or after) the moment `Start` was called |

### `internal/core/game/loop_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestKill_TransitionsToKilling` | `Game.Kill` on an alive target sets state to `Killing`, increments kills and freed memory |
| `TestKill_NoOpWhenNotAlive` | `Game.Kill` on a dead target has no effect on the kill count |
| `TestUpdate_StopsWhenTimeLimitExpired` | `Update` stops the game once elapsed time exceeds the configured time limit |
| `TestUpdate_StopsWhenAllTargetsDead` | `Update` stops the game once every target is in the `Dead` state |

### `internal/core/game/game_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestNew` | `New` sets `Confirm`, speed, `TimeLimit`; state starts `Pending`; kills and freed memory at zero |
| `TestStart_TransitionsToRunning` | `Start` transitions state to `Running` and spawns one target per process |
| `TestStop_TransitionsToStopped` | `Stop` transitions state to `Stopped` |
| `TestStart_PanicsWhenRunning` | Calling `Start` twice (already `Running`) panics |
| `TestStart_PanicsWhenStopped` | Calling `Start` after `Stop` (already `Stopped`) panics |
| `TestGame_Speed` | `Speed` reflects the configured initial throttle speed |
| `TestGame_TimeLimit` | `TimeLimit` reflects the configured value |
| `TestGame_Targets_EmptyBeforeStart` | `Targets` is empty before `Start` is called |
| `TestGame_Targets_PopulatedAfterStart` | `Targets` has one entry per process after `Start` |
| `TestGame_ConfirmTarget_NilWhenNoPending` | `ConfirmTarget` returns `nil` when no confirmation is pending |
| `TestGame_ConfirmTarget_ReturnsPendingTarget` | `ConfirmTarget` returns the target held by the internal `Confirmation` |
| `TestGame_Confirm_PendingFalseInitially` | `Confirm().Pending()` is false on a new game |
| `TestGame_Throttle_MutationAffectsSpeed` | Mutating the throttle returned by `Throttle()` changes `Speed()` |
| `TestGame_Snapshot_ExcludesDeadTargets` | `Snapshot` omits targets once their kill animation completes and they go `Dead` |
| `TestGame_Snapshot_CountsAlive` | `Snapshot.Alive` counts only `Alive`-state targets, while `Killing` targets remain visible |

### `internal/core/game/target_test.go` (23 tests)

| Test | Description |
|------|-------------|
| `TestNewTarget_WithinBounds` | Newly created target spawns within terminal bounds with correct PID/name/RSS and `Alive` state |
| `TestNewTarget_SmallTerminal` | Small terminal (5×5) still returns a non-nil target |
| `TestTarget_Tag_Alive` | Alive target's `Tag()` is `"[PID name]"` |
| `TestTarget_Tag_Dead` | Dead target's `Tag()` is empty |
| `TestTarget_Tag_Killing` | Killing target's `Tag()` is a non-empty animation frame |
| `TestTarget_Update_KillingState` | Target advances to `Dead` after the final kill-animation tick |
| `TestTarget_Update_DeadNoOp` | Dead target does not move on `Update` |
| `TestTarget_Update_BounceLeft` | Target moving left past x=0 reverses X velocity and stays in bounds |
| `TestTarget_Update_BounceRight` | Target moving right past the right bound reverses X velocity and stays in bounds |
| `TestTarget_Update_BounceTop` | Target moving up past y=0 reverses Y velocity and stays in bounds |
| `TestTarget_Update_BounceBottom` | Target moving down past the bottom bound reverses Y velocity and stays in bounds |
| `TestTarget_Update_SpeedMultiplier` | Position advances by `velocity × speed` per `Update` call with no wall interaction |
| `TestTarget_Update_MultiByteRightWall` | Right-wall bound uses rune count, not byte count — regression for multi-byte process names (`"café"`) |
| `TestTarget_Contains` | `IsHitAt` is true across the label rectangle and false just outside it |
| `TestTarget_Contains_MultiByteProcessName` | `IsHitAt` uses rune count for width so multi-byte names don't over-count — regression for byte-count bug |
| `TestTarget_Contains_NotAlive` | Non-alive target returns `false` from `IsHitAt` for all positions |
| `TestTarget_IsAlive` | `IsAlive` is true only in the `Alive` state |
| `TestTarget_IsKilling` | `IsKilling` is true only in the `Killing` state |
| `TestTarget_IsDead` | `IsDead` is true only in the `Dead` state |
| `TestTarget_Kill` | `Kill` on an alive target returns true, sets state to `Killing`, resets the animation tick |
| `TestTarget_Kill_NoOpWhenNotAlive` | `Kill` on a dead target returns false and leaves state unchanged |
| `TestTarget_Snapshot_AliveTarget` | `Snapshot` rounds position to integer X/Y, carries the tag, and reports `Killing=false` |
| `TestTarget_Snapshot_KillingTarget` | `Snapshot` reports `Killing=true` for a target mid kill-animation |

### `internal/core/handler/input_test.go` (9 tests)

| Test | Description |
|------|-------------|
| `TestHandler_OnQuit_StopsGame` | `OnQuit` stops the game when no confirmation is pending |
| `TestHandler_OnQuit_CancelsConfirmWhenPending` | `OnQuit` cancels a pending confirmation instead of stopping the game |
| `TestHandler_OnYes_ReturnsConfirmedTarget` | `OnYes` returns the confirmed target and clears the pending confirmation |
| `TestHandler_OnNo_CancelsPending` | `OnNo` clears the pending confirmation without stopping the game |
| `TestHandler_OnSpeedUp_IncreasesSpeed` | `OnSpeedUp` increases game speed by 0.5 |
| `TestHandler_OnSpeedDown_DecreasesSpeed` | `OnSpeedDown` decreases game speed by 0.5 |
| `TestHandler_OnClickAt_ReturnsTarget` | Clicking on a target's position returns that target |
| `TestHandler_OnClickAt_MissReturnsNil` | Clicking empty space returns `nil` |
| `TestHandler_OnClickAt_NoOpWhenAlreadyConfirming` | A second click while a confirmation is already pending returns `nil` and leaves the pending target unchanged |

### `internal/core/score/score_test.go` (12 tests)

| Test | Description |
|------|-------------|
| `TestEntry_Beats_ByKills` | Higher kill count wins regardless of freed memory |
| `TestEntry_Beats_TiebreakByFreedMem` | Equal kills — higher freed memory wins |
| `TestBoard_Add_SortsDescending` | Entries are kept sorted by kills descending after each `Add` |
| `TestBoard_Add_TiebreakByMemory` | Equal kills are broken by freed memory descending |
| `TestBoard_Add_CapsAtMax` | Board never exceeds `maxScores` entries; exact lowest retained kill count is asserted |
| `TestBoard_HighScore_Empty` | Empty board returns 0 |
| `TestBoard_HighScore_WithEntries` | `HighScore` returns the maximum kills across all entries |
| `TestBoard_PrintHighScore_PrintsWhenBeatsRecord` | Trophy message printed when current kills exceed the pre-`Add` high score |
| `TestBoard_PrintHighScore_SilentWhenDoesNotBeatRecord` | No output when kills do not beat the existing record |
| `TestBoard_PrintHighScore_SilentWhenZeroKills` | No output when kills is zero, even on an empty board |
| `TestBoard_PrintHighScore_PrintsWhenTiesRecord` | Trophy message printed when kills equal the current high score (tie counts as a new record) |
| `TestBoard_PrintHighScore_PrintsForFirstEntry` | Trophy message printed for the very first entry (previous high score was 0) |

---

## Application Layer

### `internal/application/event/input_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestDispatcher_Click_CallsOnClickAt` | A `ClickEvent` is forwarded to `OnClickAt` with the correct X/Y |
| `TestDispatcher_Click_ReturnsTarget` | The target returned by `OnClickAt` is propagated back from `Dispatch` |
| `TestDispatcher_Escape_CallsOnQuit` | `KeyEscape` calls `OnQuit` |
| `TestDispatcher_CtrlC_CallsOnQuit` | `KeyCtrlC` calls `OnQuit` |
| `TestDispatcher_CtrlZ_CallsOnQuit` | `KeyCtrlZ` calls `OnQuit` |
| `TestDispatcher_Q_CallsOnQuit` | Rune `q` calls `OnQuit` |
| `TestDispatcher_QUppercase_CallsOnQuit` | Rune `Q` calls `OnQuit` |
| `TestDispatcher_Y_CallsOnYes` | Rune `y` calls `OnYes` and propagates its result |
| `TestDispatcher_YUppercase_CallsOnYes` | Rune `Y` calls `OnYes` |
| `TestDispatcher_N_CallsOnNo` | Rune `n` calls `OnNo` |
| `TestDispatcher_NUppercase_CallsOnNo` | Rune `N` calls `OnNo` |
| `TestDispatcher_Plus_CallsOnSpeedUp` | Rune `+` calls `OnSpeedUp` |
| `TestDispatcher_Equals_CallsOnSpeedUp` | Rune `=` is an alias for `+` and also calls `OnSpeedUp` |
| `TestDispatcher_Minus_CallsOnSpeedDown` | Rune `-` calls `OnSpeedDown` |
| `TestDispatcher_Underscore_CallsOnSpeedDown` | Rune `_` is an alias for `-` and also calls `OnSpeedDown` |
| `TestDispatcher_UnknownRune_NoOp` | An unmapped rune (`z`) calls none of the handler methods and returns `nil` |

### `internal/application/service/game_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestToConfirmViewState_NilInput` | `toConfirmViewState(nil)` returns `nil` |
| `TestToConfirmViewState_MapsFields` | `toConfirmViewState` maps a target's PID and name onto the view state |
| `TestBuildFrame_AliveTargetIncluded` | Alive target appears in the built frame with `Killing=false` |
| `TestBuildFrame_KillingTargetMarked` | Target mid-kill appears in the frame with `Killing=true` |
| `TestBuildFrame_DeadTargetExcluded` | Dead target is omitted from the frame |
| `TestBuildFrame_HUDReflectsStats` | Frame's HUD carries current kills and freed memory |
| `TestBuildFrame_StatusBarAliveCount` | `StatusBar.Alive` counts only `Alive`-state targets |
| `TestBuildFrame_NoConfirmPending` | `StatusBar.Confirming` is `nil` when no confirmation is pending |
| `TestGameService_FinderError` | A finder error during `Play` is surfaced to the caller |
| `TestGameService_ApplyKills_CompletesPendingKill` | `applyKills` drains one kill from the channel, transitions the target to `Killing`, and increments the kill count to 1 |
| `TestGameService_ApplyKills_EmptyChannelNoOps` | `applyKills` on an empty channel does not block and leaves kills at 0 |
| `TestGameService_Kill_ProtectedPID_ReturnsError` | `kill` refuses a protected PID (e.g. PID 1) without calling the process adapter's `Kill` |
| `TestGameService_Kill_LookupError_ReturnsError` | `kill` surfaces a `LookupName` error without calling `Kill` |
| `TestGameService_Kill_NameMismatch_SkipsKillWithoutError` | `kill` returns `killed=false, err=nil` when the live process name no longer matches the target — process was likely recycled |
| `TestGameService_Kill_NameMatch_InvokesKill` | `kill` calls the process adapter's `Kill` once the live name is verified to match |
| `TestGameService_NoProcesses` | An empty process list exits cleanly with `nil` error |

### `internal/application/service/validation_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestValidateSearchPatterns_EmptySlice` | Empty pattern slice is rejected with `process.ErrNoPatterns` |
| `TestValidateSearchPatterns_EmptyTerm` | Pattern slice containing an empty string is rejected |
| `TestValidateSearchPatterns_TooShort` | Patterns of 1 or 2 characters (below `MinPatternLength`) are rejected |
| `TestValidateSearchPatterns_ExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestValidateSearchPatterns_TooLong` | Pattern exceeding `MaxPatternLength` is rejected |
| `TestValidateSearchPatterns_Valid` | Multiple valid patterns are accepted |
| `TestValidateSearchPatterns_ExactMaxLength` | Pattern at exactly `MaxPatternLength` is accepted |

### `internal/application/service/game_integration_test.go` (9 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegration_GameService_HappyPath` | Pre-queued `q` key runs `Play` to completion; Save is called; session entry has `kills=0` and `duration>0` |
| `TestIntegration_GameService_LoadError_PrintsWarningAndSkipsSave` | A `Load` failure prints `"could not load scores"` to stderr; `Play` still returns `nil`; Save is skipped |
| `TestIntegration_GameService_SaveError_PrintsWarning` | A `Save` failure prints `"score not saved"` and the underlying error to stderr; `Play` still returns `nil` |
| `TestIntegration_GameService_QuitOnQ` | `q` key sent 50 ms into a live game session causes `Play` to return without error |
| `TestIntegration_GameService_QuitOnEscape` | Escape key sent 50 ms into a live game session causes `Play` to return without error |
| `TestIntegration_GameService_QuitOnCtrlC` | Ctrl+C sent 50 ms into a live game session causes `Play` to return without error |
| `TestIntegration_GameService_QuitOnCtrlZ` | Ctrl+Z sent 50 ms into a live game session causes `Play` to return without error |
| `TestIntegration_GameService_TimeLimitExpires` | Game with a 1-second time limit exits within 3 seconds |
| `TestIntegration_GameService_SignalGoroutineDoesNotAccumulate` | Signal goroutine started inside `runLoop` exits when `Play` returns; three sequential games do not accumulate goroutines |

---

## Entrypoint Layer

### `internal/entrypoint/cli/cli_test.go` (12 tests, 12 subtests)

| Test | Description |
|------|-------------|
| `TestRun_NoArgs_PrintsUsageAndReturnsNil` | `Run` with no arguments prints usage and returns `nil` |
| `TestRun_HelpFlag_PrintsUsageAndReturnsNil` | Table-driven: `--help` and `-h` each print usage and return `nil` |
| `TestRun_BasicPattern` | Single pattern sets defaults: `ConfirmMode=false`, `Speed=2.0`, `TimeLimit=30` |
| `TestRun_MultiplePatterns` | Three patterns are all captured in order |
| `TestRun_ConfirmFlag` | `--confirm` sets `ConfirmMode=true` |
| `TestRun_SpeedFlag` | Table-driven: valid (0.5–5.0), boundary, too low (0.0), too high (6.0), non-numeric |
| `TestRun_TimeFlag` | Table-driven: valid, zero (no limit), negative (error), non-numeric (error) |
| `TestRun_UnknownFlag` | Unrecognised flags return an error |
| `TestRun_NoPatterns` | Flags-only args (no process patterns) return `process.ErrNoPatterns` |
| `TestRun_PatternTooShort` | Patterns of 1 or 2 characters (below `MinPatternLength`) are rejected at CLI parse time |
| `TestRun_PatternExactMinLength` | Pattern at exactly `MinPatternLength` is accepted |
| `TestRun_PatternTooLong` | Pattern exceeding `MaxPatternLength` is rejected at CLI parse time |

---

## Infrastructure Layer

### `internal/infrastructure/osprocess/osprocess_test.go` (4 tests)

| Test | Description |
|------|-------------|
| `TestList_ReturnsResults` | A real `ps` call via the adapter returns at least one process |
| `TestOwnPid_MatchesOSGetpid` | `OwnPid` matches `os.Getpid()` |
| `TestLookupName_ReturnsOwnName` | `LookupName` on the running test binary's own PID returns a non-empty name |
| `TestKiller_NonexistentPID` | `Kill` on a PID that doesn't exist returns an error |

### `internal/infrastructure/osprocess/osprocess_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegration_List_ReturnsResults` | Real `ps` call returns at least one process |
| `TestIntegration_List_ValidFields` | Every listed process has a positive PID, non-empty name, and non-negative RSS |
| `TestIntegration_List_ShortProcessNames` | Process names are base names, not full paths |
| `TestIntegration_Find_ExcludesOwnAndInitPID` | `process.Find` excludes the caller's own PID and PID 1 when run against real adapter output |

### `internal/infrastructure/scorefilestore/score_file_store_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestLoad_FileNotExist_ReturnsEmptyBoard` | Missing score file returns an empty board with zero high score |
| `TestLoad_InvalidJSON_ReturnsError` | Corrupted JSON file returns a parse error |
| `TestSave_CreatesFile` | Saving an empty board creates the file on disk |
| `TestSave_FilePermissions` | Saved file has `0600` permissions (not world-readable) |
| `TestSave_Load_RoundTrip` | Single entry survives a save/load cycle with correct field values |
| `TestSave_Load_MultipleEntries` | Three entries persist correctly; high score is the max kills |
| `TestSave_OverwritesPreviousFile` | A second Save replaces the first; only the latest data is visible |

### `internal/infrastructure/tcellui/tcellui_test.go` (14 tests)

| Test | Description |
|------|-------------|
| `TestPollGoroutineExitsAfterCleanup` | Poll goroutine count returns to baseline after `Cleanup()` — regression for issue #6 |
| `TestDrawHUD_NarrowTerminalSuppressesCenter` | On a narrow (30-col) terminal the centre `Highscore` element is suppressed — regression for issue #14 |
| `TestDrawHUD_WideTerminalDrawsAllThree` | On a standard (80-col) terminal all three HUD elements (`FREED`, `Highscore`, `KILLS`) are visible |
| `TestRender_MultiByteLabel_ColumnLayout` | Multi-byte kill-animation rune (`✦`) occupies one column, not multiple — regression for issue #13 |
| `TestPoll_TranslatesEscape` | tcell `KeyEscape` → `outbound.KeyEvent{Key: KeyEscape}` |
| `TestPoll_TranslatesCtrlC` | tcell `KeyCtrlC` → `outbound.KeyEvent{Key: KeyCtrlC}` |
| `TestPoll_TranslatesCtrlZ` | tcell `KeyCtrlZ` → `outbound.KeyEvent{Key: KeyCtrlZ}` |
| `TestPoll_TranslatesRune` | Plain rune `'q'` → `outbound.KeyEvent{Key: KeyNone, Ch: 'q'}` |
| `TestPoll_MouseButton1_EmitsClickEvent` | Button1 mouse click → `outbound.ClickEvent{X, Y}` |
| `TestPoll_NonButton1_DropsEvent` | Button2 click is dropped; next event is the following key |
| `TestDrawStatusBar_Normal` | Status bar contains `Targets:`, `Speed:`, and `Click to kill` |
| `TestDrawStatusBar_Confirming` | Confirm bar shows PID, name, `(Y)es`, `(N)o` |
| `TestDrawStatusBar_WithTimeLimit` | Status bar shows `Time:` and remaining seconds when `TimeLimit > 0` |
| `TestDrawStatusBar_NoTimeLimit` | `Time:` segment is absent when `TimeLimit == 0` |

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

**Domain logic split out of `game` into standalone packages** — What was previously tested only through `game`/`osprocess` now lives in dedicated packages with their own unit tests: `core/process` (pattern validation, `Info.IsProtected`, `Find` filtering) and `core/movement` (`Bounds`, `Motion`, `Throttle`). This makes the filtering, protection, and physics rules independently testable and reusable outside the game loop.

**`game` itself is decomposed by responsibility** — Instead of one large test file, each concern gets its own: `state_test.go` (atomic FSM: `Pending`/`Running`/`Stopped`), `confirmation_test.go` (hold/accept/cancel a pending kill), `stats_test.go` (kill count, freed memory, high score), `timer_test.go` (elapsed/remaining time, unlimited mode), `loop_test.go` (kill + per-tick `Update`), `game_test.go` (lifecycle and `Snapshot`), and `target_test.go` (per-target physics, hit-detection, lifecycle).

**Kill safety moved to the application layer** — `osprocess_test.go` no longer tests PID/name guards directly (its own comment explains why: exercising `Kill` against PID 0/1/-1/self would send a real signal). That responsibility now lives in `application/service.kill`, which calls `process.Info.IsProtected()` and re-verifies the live process name via `LookupName` before ever invoking the adapter's `Kill` — covered by `TestGameService_Kill_ProtectedPID_ReturnsError`, `TestGameService_Kill_LookupError_ReturnsError`, `TestGameService_Kill_NameMismatch_SkipsKillWithoutError`, and `TestGameService_Kill_NameMatch_InvokesKill`. The integration suite (`TestIntegration_Find_ExcludesOwnAndInitPID`) replicates the exclusion policy against a live process table.

**New input seam: `Dispatcher` → `InputHandler`** — `application/event.Dispatcher` translates raw `outbound.KeyEvent`/`ClickEvent` values into calls on a `core/handler.InputHandler` interface (`OnQuit`, `OnYes`, `OnNo`, `OnSpeedUp`, `OnSpeedDown`, `OnClickAt`). The two halves are tested independently: `application/event/input_test.go` verifies event-to-method routing (including key aliases `=`/`+` and `_`/`-`) against a fake handler, while `core/handler/input_test.go` verifies the handler's actual effect on a real `game.Game`.

**Validation is checked at every seam it's used** — `core/process.Validate` is the canonical boundary check (`MinPatternLength`/`MaxPatternLength`); `application/service.validateSearchPatterns` and `entrypoint/cli`'s tests re-assert the same min/max boundaries at the service and CLI layers using the shared constants, so a change to the limits is caught everywhere a pattern can enter the system.

**Integration vs unit split** — Two files carry `//go:build integration` tags (`osprocess_integration_test.go`, `game_integration_test.go`). These tests touch the real OS process table or run a full game loop. All other tests are pure unit tests that run offline with fake/stub dependencies. Run with `go test -tags integration ./...`.

**Unicode correctness is enforced by regression tests** — `TestTarget_Update_MultiByteRightWall` and `TestTarget_Contains_MultiByteProcessName` guard against a byte-count bug where multi-byte UTF-8 names caused targets to bounce too early or accept out-of-bounds click hits. `TestRender_MultiByteLabel_ColumnLayout` guards the same class of bug in the tcell renderer. All three use multi-byte fixtures (`"café"`, `"✦"`).

**Frame pipeline isolation** — `game.Snapshot()` / `Target.Snapshot()` and `service.buildFrame()` are tested directly against structured data rather than terminal output, decoupling rendering correctness from game logic and TUI library from domain state.

**Score domain logic is fully covered** — `score_test.go` covers the `beats()` ranking predicate, all `Board.Add` invariants (sort order, tiebreak, cap), and all five `PrintHighScore` branches including the tie case. Output is captured via `internal/testutil/capture` without mocking stdout globally.

**Score persistence** — `score_file_store_test.go` covers the full save/load contract including corruption, overwrite semantics, file permissions (`0600`), and ordering invariants.

**Load failure skips save** — `TestIntegration_GameService_LoadError_PrintsWarningAndSkipsSave` confirms that if the score store is unreadable at session start, no save is attempted at the end. This prevents a corrupted file from being silently overwritten with a zero-kill entry.

**All quit signals are exercised** — Integration tests verify that `q`, Escape, Ctrl+C, and Ctrl+Z each independently terminate a live session, providing full coverage of the four supported exit paths.

**Signal goroutine lifecycle** — `TestIntegration_GameService_SignalGoroutineDoesNotAccumulate` runs three sequential game sessions and verifies that goroutine count does not grow, confirming the signal goroutine in `runLoop` exits cleanly when `Play` returns. `TestPollGoroutineExitsAfterCleanup` does the same for the tcell poll goroutine.

**Test infrastructure is organised as shared sub-packages** — `internal/testutil/fake` provides typed test doubles (`Process`, `Store`, `Renderer`, `InputSource`, `InputHandler`) used across multiple test files. `internal/testutil/capture` provides dependency-free stdout/stderr capture helpers. Both are structured as proper sub-packages to prevent duplication and circular imports.