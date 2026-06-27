# Test Cases

Total: 14 test files across 5 packages — 101 test functions (including 18 subtests).

---

## Adapter Layer

### `internal/adapter/driven/jsonscores/jsonscores_test.go` (6 tests)

| Test | Description |
|------|-------------|
| `TestLoad_FileNotExist_ReturnsEmptyBoard` | Missing score file returns an empty board with zero high score |
| `TestLoad_InvalidJSON_ReturnsError` | Corrupted JSON file returns a parse error |
| `TestSave_CreatesFile` | Saving an empty board creates the file on disk |
| `TestSave_Load_RoundTrip` | Single entry survives a save/load cycle with correct field values |
| `TestSave_Load_MultipleEntries` | Three entries persist correctly; high score is the max kills |
| `TestSave_OverwritesPreviousFile` | A second Save replaces the first; only the latest data is visible |

### `internal/adapter/driven/osprocess/osprocess_test.go` (16 tests)

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
| `TestProc_Fields` | `Pid()`, `Name()`, and `Rss()` accessors return correct values |
| `TestKiller_InvalidPID` | Killing PID -1 returns an error on all platforms |

### `internal/adapter/driven/osprocess/osprocess_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestIntegration_List_ReturnsResults` | Real `ps` call returns at least one process |
| `TestIntegration_List_ValidFields` | Every listed process has a positive PID, non-empty name, and non-negative RSS |
| `TestIntegration_List_ShortProcessNames` | Process names are base names, not full paths |
| `TestIntegration_Find_ExcludesOwnPID` | Own PID and PID 1 are absent from `Find` results on a live system |

### `internal/adapter/driving/cli/cli_test.go` (8 tests, 10 subtests)

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

### `internal/app/runner_test.go` (2 tests)

| Test | Description |
|------|-------------|
| `TestGameService_FinderError` | A finder error during `Play` is surfaced to the caller |
| `TestGameService_NoProcesses` | An empty process list exits cleanly with `nil` error |

---

## Domain Layer

### `internal/domain/game/game_test.go` (1 test)

| Test | Description |
|------|-------------|
| `TestNew` | `New` sets `confirmMode`, `speed`, `timeLimit`, `running=true` via `atomic.Bool`, kills and freedMem at zero |

### `internal/domain/game/loop_test.go` (12 tests)

| Test | Description |
|------|-------------|
| `TestGame_timeRemaining_WithinLimit` | `timeRemaining` returns a positive value within the configured limit |
| `TestGame_timeRemaining_Expired` | `timeRemaining` returns zero when the clock has already passed the limit |
| `TestKillTarget_TransitionsToKilling` | Killing an alive target sets state to `Killing`, increments kills and freedMem |
| `TestKillTarget_NoOpWhenNotAlive` | Killing a dead target has no effect on kill count |
| `TestUpdate_StopsWhenTimeLimitExpired` | `update` stops the game when elapsed time exceeds the limit |
| `TestUpdate_StopsWhenAllTargetsDead` | `update` stops the game when all targets are in the `Dead` state |
| `TestRender_AliveTargetIncluded` | Alive target appears in the rendered frame at the correct coordinates |
| `TestRender_DeadTargetExcluded` | Dead target is omitted from the rendered frame |
| `TestRender_KillingTargetMarked` | Killing target appears in the frame with `Killing=true` |
| `TestRender_HUDReflectsSession` | HUD carries current kills, freed memory, and high score |
| `TestRender_ConfirmStateInStatusBar` | When confirming a kill, the status bar includes the target's PID and name |
| `TestRender_StatusBarAliveCount` | `StatusBar.Alive` counts only `Alive`-state targets |

### `internal/domain/game/motion_test.go` (7 tests)

| Test | Description |
|------|-------------|
| `TestNewMotion_WithinBounds` | 20 iterations: spawn position stays within `[1, maxX-labelLen-1]` × `[1, maxY-2]` |
| `TestNewMotion_SmallTerminal` | Small terminal (5×5) does not panic; position stays ≥ 1 |
| `TestMotion_Update_BounceLeft` | Moving left past x=0 reverses X velocity and keeps position non-negative |
| `TestMotion_Update_BounceRight` | Moving right past right bound reverses X velocity and clamps position |
| `TestMotion_Update_BounceTop` | Moving up past y=0 reverses Y velocity and keeps position non-negative |
| `TestMotion_Update_BounceBottom` | Moving down past bottom bound reverses Y velocity and clamps position |
| `TestMotion_Update_SpeedMultiplier` | Position delta is velocity × speed multiplier each tick |

### `internal/domain/game/event_handler_test.go` (16 tests)

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
| `TestHandleKeyPress_ConfirmYes` | `y` during confirmation kills the target, increments kills and freedMem |
| `TestHandleKeyPress_ConfirmNo` | `n` during confirmation leaves the target alive and clears the confirm state |
| `TestHandleKeyPress_QCancelsConfirm` | `q` during confirmation cancels it without quitting the game |
| `TestHandleMouseClick_KillsTargetOnClick` | Clicking a target transitions it to `Killing` and records a kill |
| `TestHandleMouseClick_SetsConfirmingInConfirmMode` | In confirm mode, clicking a target sets it as the pending confirmation |
| `TestHandleMouseClick_NoOpWhenAlreadyConfirming` | A second click while confirming is ignored |
| `TestHandleMouseClick_NoOpOnMiss` | Clicking empty space has no effect |

### `internal/domain/game/session_test.go` (3 tests)

| Test | Description |
|------|-------------|
| `TestSession_RecordKill` | Each `RecordKill` increments `Kills` by 1 and accumulates freed bytes |
| `TestSession_SetHighScore` | `SetHighScore` stores the supplied value |
| `TestSession_StartTime_ZeroOnNew` | `StartTime` is zero on a freshly created session |

### `internal/domain/game/target_test.go` (10 tests)

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

### `internal/domain/game/game_integration_test.go` (4 tests) `// go:build integration`

| Test | Description |
|------|-------------|
| `TestPlay_QuitOnQ` | `q` key sent 50ms into a live game session causes `Play` to return without error |
| `TestPlay_QuitOnEscape` | Escape key sent 50ms into a live game session causes `Play` to return without error |
| `TestPlay_TimeLimitExpires` | Game with a 1-second time limit exits within 3 seconds |
| `TestPlay_SessionStateAfterQuit` | After quitting immediately: kills remain 0, `StartTime` is recorded |

### `internal/domain/score/score_test.go` (11 tests)

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

**Integration vs unit split** — Two files carry `//go:build integration` tags (`osprocess_integration_test.go`, `game_integration_test.go`). These tests touch the real OS process table and run a full game loop. All other tests are pure unit tests that run offline with fake/stub dependencies.

**Security properties are tested explicitly** — `TestFilter_ExcludesPID1` and `TestFilter_ExcludesOwnPID` guard the two hard-coded exclusion rules that prevent the game from killing system init or itself. The integration tests replicate these checks on a live system.

**Boundary conditions on input validation** — CLI argument parsing and process pattern validation are tested at every edge: empty, too-long, exactly-max-length, out-of-range numeric values, and unknown flags. This mirrors the `validate` function in `osprocess` which has its own dedicated boundary tests.

**State machine coverage** — Target lifecycle (`Alive → Killing → Dead`) is exercised across `target_test.go`, `event_handler_test.go`, `loop_test.go`, and `game_integration_test.go`. Every transition and every state's observable behaviour (label, hit-detection, motion) is covered.

**Render pipeline isolation** — `loop_test.go` captures rendered frames via a `stubRenderer` and asserts on the structured `Frame` output rather than terminal output, decoupling rendering correctness tests from any TUI library.

**Score domain logic is fully covered** — `score_test.go` covers the `beats()` ranking predicate directly, all `Board.Add` invariants (sort order, tiebreak, cap), and all four `PrintHighScore` branches (new record, equal, below, zero kills). Output is captured via `internal/testutil/capture` without mocking stdout globally.

**Score persistence** — `jsonscores_test.go` covers the full save/load contract including corruption, overwrite semantics, and ordering invariants that live in `score_test.go`.

**Motion physics** — `motion_test.go` tests all four wall-bounce directions and the speed multiplier in isolation, ensuring the collision logic is independent of the game loop.

**Test infrastructure is organised as shared sub-packages** — `internal/testutil/fake` provides typed test doubles (`Killer`, `Finder`, `Store`, `NewProcess`) used across six test files. `internal/testutil/capture` provides a dependency-free stdout capture helper. Both are structured as proper sub-packages rather than inline helpers to prevent duplication and circular imports.
