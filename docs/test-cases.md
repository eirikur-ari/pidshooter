# Test Cases

Total: 14 test files across 7 packages — 112 test functions (including 8 subtests).

---

## Infrastructure Layer

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

### `internal/infrastructure/osprocess/osprocess_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestValidate_EmptySlice` | Empty pattern slice is rejected |
| `TestValidate_EmptyTerm` | Pattern slice containing an empty string is rejected |
| `TestValidate_TooLong` | Pattern exceeding `MaxPatternLength` is rejected |
| `TestValidate_ExactMaxLength` | Pattern at exactly `MaxPatternLength` is accepted |
| `TestValidate_Valid` | Multiple valid patterns are accepted |
| `TestFind_EmptyPatterns` | `Find` with empty patterns returns an error |
| `TestFind_EmptyTerm` | `Find` with an empty-string term returns an error |
| `TestFilter_MatchesByName` | Exact name match returns only the matching process |
| `TestFilter_SubstringMatch` | Partial name match is supported |
| `TestFilter_CaseInsensitive` | Name matching is case-insensitive |
| `TestFilter_MultipleTerms` | OR-semantics: multiple terms each match independently |
| `TestFilter_NoMatch` | No matching name returns an empty slice |
| `TestFilter_ExcludesPID1` | PID 1 (init/launchd) is always excluded |
| `TestFilter_ExcludesOwnPID` | The running process's own PID is excluded from results |
| `TestProc_Fields` | `Pid`, `Name`, and `Rss` fields on `process.Info` hold correct values |
| `TestKiller_InvalidPID` | Killing PID -1 returns an error on all platforms |

### `internal/infrastructure/osprocess/osprocess_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegration_List_ReturnsResults` | Real `ps` call returns at least one process |
| `TestIntegration_List_ValidFields` | Every listed process has a positive PID, non-empty name, and non-negative RSS |
| `TestIntegration_List_ShortProcessNames` | Process names are base names, not full paths |
| `TestIntegration_Find_ExcludesOwnPID` | Own PID and PID 1 are absent from `Find` results on a live system |

### `internal/infrastructure/tcellui/tcellui_test.go` (15 tests)

| Test | Description |
|------|-------------|
| `TestPollGoroutineExitsAfterCleanup` | Poll goroutine count returns to baseline after `Cleanup()` — regression for issue #6 |
| `TestDrawHUD_NarrowTerminalSuppressesCenter` | On a narrow (30-col) terminal the centre `Highscore` element is suppressed — regression for issue #14 |
| `TestDrawHUD_WideTerminalDrawsAllThree` | On a standard (80-col) terminal all three HUD elements (`FREED`, `Highscore`, `KILLS`) are visible |
| `TestRender_MultiByteLabel_ColumnLayout` | Multi-byte kill-animation rune (`✦`) occupies one column, not multiple — regression for issue #13 |
| `TestPoll_TranslatesEscape` | tcell `KeyEscape` → `game.KeyEvent{Key: KeyEscape}` |
| `TestPoll_TranslatesCtrlC` | tcell `KeyCtrlC` → `game.KeyEvent{Key: KeyCtrlC}` |
| `TestPoll_TranslatesCtrlZ` | tcell `KeyCtrlZ` → `game.KeyEvent{Key: KeyCtrlZ}` |
| `TestPoll_TranslatesRune` | Plain rune `'q'` → `game.KeyEvent{Key: KeyNone, Ch: 'q'}` |
| `TestPoll_MouseButton1_EmitsClickEvent` | Button1 mouse click → `game.ClickEvent{X, Y}` |
| `TestPoll_NonButton1_DropsEvent` | Button2 click is dropped; next event is the following key |
| `TestPoll_ResizeEvent_EmitsResizeEvent` | Screen resize → `game.ResizeEvent` |
| `TestDrawStatusBar_Normal` | Status bar contains `Targets:`, `Speed:`, and `Click to kill` |
| `TestDrawStatusBar_Confirming` | Confirm bar shows PID, name, `(Y)es`, `(N)o` |
| `TestDrawStatusBar_WithTimeLimit` | Status bar shows `Time:` and remaining seconds when `TimeLimit > 0` |
| `TestDrawStatusBar_NoTimeLimit` | `Time:` segment is absent when `TimeLimit == 0` |

---

## Entrypoint Layer

### `internal/entrypoint/cli/cli_test.go` (8 tests, 10 subtests)

| Test | Description |
|------|-------------|
| `TestParseArgs_BasicPattern` | Single pattern sets defaults: `confirm=false`, `speed=2.0`, `timeLimit=30` |
| `TestParseArgs_MultiplePatterns` | Three patterns are all captured in order |
| `TestParseArgs_ConfirmFlag` | `--confirm` sets `confirm=true` |
| `TestParseArgs_SpeedFlag` | Table-driven: valid (0.1–5.0), boundary, too low (0.0), too high (6.0), non-numeric |
| `TestParseArgs_TimeFlag` | Table-driven: valid, zero (no limit), negative (error), non-numeric (error) |
| `TestParseArgs_UnknownFlag` | Unrecognised flags return an error |
| `TestParseArgs_NoPatterns` | Flags-only args (no process patterns) return an error |
| `TestParseArgs_PatternTooLong` | Pattern exceeding `MaxPatternLength` is rejected at CLI parse time |

---

## Application Layer

### `internal/application/runner_test.go` (2 tests)

| Test | Description |
|------|-------------|
| `TestGameRunner_FinderError` | A finder error during `Play` is surfaced to the caller |
| `TestGameRunner_NoProcesses` | An empty process list exits cleanly with `nil` error |

### `internal/application/runner_integration_test.go` (6 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestGameRunner_HappyPath` | Pre-queued `q` key runs `Play` to completion; Save is called; session entry has `kills=0` and `duration>0` |
| `TestGameRunner_SaveError_PrintsWarning` | A `Save` failure prints `"score not saved"` to stderr; `Play` still returns `nil` |
| `TestGameRunner_QuitOnQ` | `q` key sent 50 ms into a live game session causes `Play` to return without error |
| `TestGameRunner_QuitOnEscape` | Escape key sent 50 ms into a live game session causes `Play` to return without error |
| `TestGameRunner_TimeLimitExpires` | Game with a 1-second time limit exits within 3 seconds |
| `TestGameRunner_SignalGoroutineDoesNotAccumulate` | Signal goroutine started inside `runLoop` exits when `Play` returns; three sequential games do not accumulate goroutines |

---

## Core Domain Layer

### `internal/core/game/game_test.go` (1 test)

| Test | Description |
|------|-------------|
| `TestNew` | `New` sets `confirmMode`, `speed`, `timeLimit`, `running=true` via `atomic.Bool`; kills and freedMem at zero |

### `internal/core/game/loop_test.go` (12 tests)

| Test | Description |
|------|-------------|
| `TestGame_timeRemaining_WithinLimit` | `timeRemaining` returns a positive value within the configured limit |
| `TestGame_timeRemaining_Expired` | `timeRemaining` returns zero when the clock has already passed the limit |
| `TestCompleteKill_TransitionsToKilling` | `CompleteKill` on an alive target sets state to `Killing`, increments kills and freedMem |
| `TestCompleteKill_NoOpWhenNotAlive` | `CompleteKill` on a dead target has no effect on kill count |
| `TestUpdate_StopsWhenTimeLimitExpired` | `Update` stops the game when elapsed time exceeds the limit |
| `TestUpdate_StopsWhenAllTargetsDead` | `Update` stops the game when all targets are in the `Dead` state |
| `TestFrame_AliveTargetIncluded` | Alive target appears in the frame at the correct coordinates with `Killing=false` |
| `TestFrame_DeadTargetExcluded` | Dead target is omitted from the frame |
| `TestFrame_KillingTargetMarked` | Killing target appears in the frame with `Killing=true` |
| `TestFrame_HUDReflectsSession` | HUD carries current kills, freed memory, and high score |
| `TestFrame_ConfirmStateInStatusBar` | When confirming a kill, the status bar includes the target's PID and name |
| `TestFrame_StatusBarAliveCount` | `StatusBar.Alive` counts only `Alive`-state targets |

### `internal/core/game/event_handler_test.go` (16 tests)

| Test | Description |
|------|-------------|
| `TestHandleKeyPress_Quit` | `q` sets `running=false` |
| `TestHandleKeyPress_QuitUppercase` | `Q` sets `running=false` |
| `TestHandleKeyPress_Escape` | Escape key sets `running=false` |
| `TestHandleKeyPress_CtrlC` | Ctrl+C sets `running=false` |
| `TestHandleKeyPress_CtrlZ` | Ctrl+Z sets `running=false` |
| `TestHandleKeyPress_SpeedUp` | `+` increases speed by 0.5 |
| `TestHandleKeyPress_SpeedDown` | `-` decreases speed by 0.5 |
| `TestHandleKeyPress_SpeedCapsAtMax` | Speed is capped at 5.0 and does not exceed it |
| `TestHandleKeyPress_SpeedCapsAtMin` | Speed is capped at 0.1 and does not go below it |
| `TestHandleKeyPress_ConfirmYes_ReturnsKillRequest` | `y` during confirmation returns a `*KillRequest` for the pending target and clears `confirming` |
| `TestHandleKeyPress_ConfirmNo` | `n` during confirmation returns `nil`; target stays `Alive`; `confirming` cleared |
| `TestHandleKeyPress_QCancelsConfirm` | `q` during confirmation cancels it without quitting the game |
| `TestHandleMouseClick_ReturnsKillRequest` | Clicking an alive target returns a `*KillRequest` for that target |
| `TestHandleMouseClick_SetsConfirmingInConfirmMode` | In confirm mode, clicking a target sets it as `confirming`; returns `nil` |
| `TestHandleMouseClick_NoOpWhenAlreadyConfirming` | A second click while already confirming returns `nil` and leaves `confirming` unchanged |
| `TestHandleMouseClick_NoOpOnMiss` | Clicking empty space returns `nil` |

### `internal/core/game/session_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestSession_RecordKill` | Each `RecordKill` increments `Kills` by 1 and accumulates freed bytes |
| `TestSession_SetHighScore` | `SetHighScore` stores the supplied value |
| `TestSession_StartTime_ZeroOnNew` | `StartTime` is zero on a freshly created session |

### `internal/core/game/target_test.go` (10 tests)

| Test | Description |
|------|-------------|
| `TestNewTarget_WithinBounds` | Newly created target spawns within terminal bounds with correct PID/name/RSS |
| `TestNewTarget_SmallTerminal` | Small terminal (5×5) returns a non-nil target |
| `TestTarget_Label_Alive` | Alive target label is `"[PID name]"` |
| `TestTarget_Label_Dead` | Dead target returns an empty label |
| `TestTarget_Label_Killing` | Killing target returns a non-empty animation label |
| `TestTarget_Update_KillingState` | Target advances to `Dead` after the final kill animation frame |
| `TestTarget_Update_DeadNoOp` | Dead target does not move when updated |
| `TestTarget_Contains` | Hit-detection: true for the label rectangle, false outside it |
| `TestTarget_Contains_NotAlive` | Non-alive target returns `false` for all positions |
| `TestTarget_StartKillAnim` | `StartKillAnim` sets state to `Killing` and resets `KillAnimFrame` to 0 |

### `internal/core/score/score_test.go` (11 tests)

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
| `TestBoard_PrintHighScore_PrintsForFirstEntry` | Trophy message printed for the very first entry (previous high score was 0) |

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

**Integration vs unit split** — Two files carry `//go:build integration` tags (`osprocess_integration_test.go`, `runner_integration_test.go`). These tests touch the real OS process table or run a full game loop. All other tests are pure unit tests that run offline with fake/stub dependencies. Run with `go test -tags integration ./...`.

**Security properties are tested explicitly** — `TestFilter_ExcludesPID1` and `TestFilter_ExcludesOwnPID` guard the two hard-coded exclusion rules that prevent the game from killing system init or itself. The integration tests replicate these checks on a live system.

**Boundary conditions on input validation** — CLI argument parsing and process pattern validation are tested at every edge: empty, too-long, exactly-max-length, out-of-range numeric values, and unknown flags. This mirrors the `validate` function in `osprocess` which has its own dedicated boundary tests.

**State machine coverage** — Target lifecycle (`Alive → Killing → Dead`) is exercised across `target_test.go`, `event_handler_test.go`, `loop_test.go`, and `runner_integration_test.go`. Every transition and every state's observable behaviour (label, hit-detection, motion) is covered.

**KillRequest pattern tested end-to-end** — `event_handler_test.go` verifies that `HandleEvent` returns a `*KillRequest` (not a kill effect). `runner_integration_test.go` verifies the application layer correctly executes the two-phase kill: `killer.Kill()` then `g.CompleteKill()`.

**Frame pipeline isolation** — `loop_test.go` tests `g.Frame()` directly, asserting on structured `Frame` data rather than terminal output. This decouples rendering correctness from any TUI library.

**Score domain logic is fully covered** — `score_test.go` covers the `beats()` ranking predicate, all `Board.Add` invariants (sort order, tiebreak, cap), and all four `PrintHighScore` branches. Output is captured via `internal/testutil/capture` without mocking stdout globally.

**Score persistence** — `score_file_store_test.go` covers the full save/load contract including corruption, overwrite semantics, file permissions (`0600`), and ordering invariants.

**Signal goroutine lifecycle** — `TestGameRunner_SignalGoroutineDoesNotAccumulate` runs three sequential game sessions and verifies that goroutine count does not grow, confirming the signal goroutine in `runLoop` exits cleanly when `Play` returns.

**Test infrastructure is organised as shared sub-packages** — `internal/testutil/fake` provides typed test doubles (`Killer`, `Finder`, `Store`) used across six test files. `internal/testutil/capture` provides a dependency-free stdout/stderr capture helper. Both are structured as proper sub-packages to prevent duplication and circular imports.
