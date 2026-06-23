# pidshooter — Code Review Report

*Updated 2026-06-23. Reflects refactored package layout (`game_play.go`, `game_entity.go`, `game_motion.go`, `game_session.go`, `game_util.go`).*

---

## Bugs

### 1. Data race on `game.running` — `game_play.go:24,41,87` *(unchanged)*
`running` is a plain `bool`. The signal handler goroutine (spawned in `main.go:78`) calls `g.Stop()` which writes `g.running = false` at `game_play.go:41`, while the main loop reads it in `for g.running` at `game_play.go:87`. These are unsynchronised accesses across goroutines. `go test -race ./...` (now available via `make test-race`) would catch this.

```go
// fix: replace plain bool with atomic
type Game struct {
    running atomic.Bool
    // ...
}
func (g *Game) Stop() { g.running.Store(false) }
// for loop: for g.running.Load() { ... }
```

Note: `game_play_test.go` directly initialises and reads `g.running` (e.g. `&Game{running: true}` at lines 50, 57, 65, …). Fixing to `atomic.Bool` requires updating all test constructions to `g.running.Store(true)` and assertions to `g.running.Load()`.

### 2. Broken `isHighScore` logic — `score.go:72–73` *(unchanged)*
```go
isHighScore := len(b.Scores) <= maxScores ||
    (len(b.Scores) > 0 && b.Scores[len(b.Scores)-1] != entry)
```
After sorting, `entry` can land anywhere in the slice. Checking whether the *last* element equals `entry` is unreliable because `Entry` contains a `time.Time` field — two entries with identical kills/mem but different timestamps are never equal, so the condition is almost always `true`. The return value is also silently discarded at `main.go:95` (`scoreBoard.Add(entry)`), making it dead code. The simplest fix is to drop the bool return from `Add()` and rely on the caller's `scoreBoard.HighScore()` check (already done at `main.go:101`).

### 3. `game.Cleanup()` called twice — `main.go:72,98` *(unchanged)*
`defer g.Cleanup()` is registered at line 72, then called explicitly at line 98. The explicit call restores the terminal before printing results; the deferred one fires afterwards. `Cleanup()` is now idempotent (it nil-checks `g.screen` before calling `Fini()`), so this is safe — but the `defer` is still misleading because it looks like the safety net when the explicit call already did the work. Remove the `defer` and rely solely on the explicit call; add a comment that `Cleanup()` must be called before printing post-game output.

---

## Design Issues

### 4. `parseArgs` calls `os.Exit` — `main.go:112,147` *(partially resolved)*
Two `os.Exit(0)` calls remain:
- `main.go:112` — zero-argument invocation prints usage and exits
- `main.go:147` — `--help`/`-h` flag prints usage and exits

The missing-pattern case was fixed to return an error (`main.go:158–160`), and `TestParseArgs_NoPatterns` covers it. But the zero-arg and `--help` paths are still untestable. Return a sentinel error (e.g. `errUsage`) that `run()` detects and exits cleanly without `os.Exit` inside `parseArgs`.

### 5. `filePath()` creates a directory as a side effect — `score.go:38`
A function that returns a path string shouldn't create directories. `os.MkdirAll` runs on every `Load()` and `Save()` call. Move the `MkdirAll` into `Save()` only, where the directory is actually needed.

### 6. Five return values from `parseArgs` — `main.go:110` *(unchanged)*
```go
func parseArgs(args []string) ([]string, bool, float64, int, error)
```
This should be a struct:
```go
type Config struct {
    Patterns    []string
    ConfirmMode bool
    Speed       float64
    TimeLimit   int
}
```

### 7. Event goroutine can block permanently after game exits — `game_play.go:77–85` *(unchanged)*
The goroutine started in `Run()` sends to `eventCh` (capacity 10) with an unbuffered channel write: `eventCh <- ev`. Once the game loop exits (`g.running` turns false), `drainEvents` stops being called. If the channel is already full and `PollEvent` returns another event, the goroutine blocks indefinitely on the channel send. At that point `Cleanup()` calls `screen.Fini()`, but `Fini()` unblocks `PollEvent` — which the goroutine is no longer blocked on. The goroutine leaks until the process exits. A `context.Context` or done channel fixes this:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go func() {
    for {
        ev := g.screen.PollEvent()
        if ev == nil { return }
        select {
        case eventCh <- ev:
        case <-ctx.Done(): return
        }
    }
}()
```

### 8. `game.highScore` is stale during play — `main.go:67` *(unchanged)*
The high score is loaded once before the game starts (`g.SetHighScore(scoreBoard.HighScore())`). It is never updated live if the player breaks it mid-game. The HUD renders `g.highScore` (via `game_play.go:232`) but the value never changes. Compare `g.kills` against `g.highScore` inside `killEntity` and call `g.Session.SetHighScore(g.kills)` if it is exceeded.

### 9. Kill score incremented even if signal fails — `game_play.go:298` *(unchanged)*
```go
func (g *Game) killEntity(e *Entity) {
    if e.State != Alive { return }
    _ = e.Kill()            // error silently discarded
    e.StartKillAnim()
    g.Session.RecordKill(e.Rss())
}
```
`Entity.Kill()` now properly returns an error (`game_entity.go:94–100`), which is a structural improvement over the previous review. However, `killEntity` still discards it with `_` and always increments the score. On Linux, `os.FindProcess` never fails numerically, so a process that has already exited will pass the guard. Log signal errors at minimum; ideally only call `RecordKill` after a successful signal.

---

## New Issues

### 13. Score table header does not match data — `score.go:97–109`
The `PrintScores` method contains an unfixed TODO:
```go
//TODO: Replace Speed with Time
fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
```
The `Entry` struct has both `Speed` and `Time` fields, but the table only shows Speed. The time-limit column was never added to the display, making the recorded `Time` field invisible to the user.

### 14. Duplicate mock types in game package tests
`mockInfo` in `game_entity_test.go:8–16` and `fakeProcess` in `game_play_test.go:11–19` are structurally identical — both implement `process.Info` with `Pid()/Name()/Rss()`. Since both files share `package game`, one definition is redundant. Consolidate into a single `testProcess` type in a `game_test_helpers_test.go` file.

### 15. Grammar error in `validate()` error message — `process_finder.go:101`
`"at least one search patterns is required"` is grammatically wrong. Should be `"at least one search pattern is required"`.

---

## Minor / Polish

### 10. `len()` counts bytes, not terminal columns — `game_entity.go:36,70,83` *(unchanged)*
`len(label)` counts UTF-8 bytes. Kill animation frames include multi-byte emoji (`"💥"` — 4 bytes, 2 columns). During the `Killing` state, `Update()` passes `float64(len(e.Label()))` as the label width for bounce bounds, and `Contains()` uses `len(e.Label())` for hit detection — both are wrong for any frame containing a multi-byte character. Use `utf8.RuneCountInString()` or `uniseg.StringWidth()` for correct column widths.

### 11. HUD elements can overlap on narrow terminals — `game_play.go:224–254` *(unchanged)*
FREED, Highscore, and KILLS are drawn independently on row 0 with no awareness of each other. On a terminal narrower than ~50 columns they overwrite each other. A single formatted status line or explicit minimum-width guard would be more robust.

### 12. `ps` resolved via `$PATH` — `process_finder.go:54` *(unchanged)*
`exec.Command("ps", ...)` relies on `$PATH`. On Linux, reading `/proc` directly would eliminate the subprocess and the PATH dependency entirely. The existing TODO comment in the file acknowledges this gap.

---

## Resolved Since Previous Review

| Previous | Issue | Resolution |
|---|---|---|
| Makefile `lint` | Phantom `.PHONY` target with no recipe | Removed from `.PHONY` entirely |
| Makefile `coverage` | Would fail on clean checkout | Added `@mkdir -p $(BUILD_DIR)` |
| Makefile `test-race` | No race detector target | `test-race` target now present |
| Makefile `run` ARGS | Unquoted `$(ARGS)` could mis-parse values with `=` | `$(ARGS)` is now quoted: `"$(ARGS)"` |
| #4 (partial) | `parseArgs` exits on missing pattern | Now returns error; `TestParseArgs_NoPatterns` covers it |
| Process test vacuous pass | `TestFind_ValidPattern` could pass silently with a renamed binary | Removed; replaced by `TestIntegration_Find_ExcludesOwnPID` which asserts exclusion of own PID without relying on binary name |
| Monolithic `game.go` | Single file contained all game logic | Split into `game_play.go`, `game_entity.go`, `game_motion.go`, `game_session.go`, `game_util.go` — much easier to navigate and test |
| #9 (partial) | Kill dispatch was inlined with no error path | `Entity.Kill()` is now a standalone method returning an error; the caller still discards it (see #9 above) |

---

## Summary

| # | File | Severity | Issue |
|---|------|----------|-------|
| 1 | game_play.go:24 | **Bug** | Data race on `game.running` |
| 2 | score.go:72 | **Bug** | `isHighScore` logic wrong; return value unused |
| 3 | main.go:72,98 | **Bug** | Double `Cleanup()` — misleading defer |
| 4 | main.go:112,147 | Design | `parseArgs` calls `os.Exit` (zero args + --help) |
| 5 | score.go:38 | Design | `filePath()` creates directory as side effect |
| 6 | main.go:110 | Design | 5 return values; should be a config struct |
| 7 | game_play.go:77 | Design | Event goroutine can block on channel send after game exits |
| 8 | main.go:67 | Design | Stale high score in HUD |
| 9 | game_play.go:298 | Design | Kill score incremented even if SIGKILL fails |
| 10 | game_entity.go:36 | Minor | `len()` bytes vs rune width (breaks with emoji kill frames) |
| 11 | game_play.go:224 | Minor | HUD elements overlap on narrow terminals |
| 12 | process_finder.go:54 | Minor | `ps` found via `$PATH` |
| 13 | score.go:97 | Minor | Score table header shows "Speed" but TODO says replace with "Time" |
| 14 | game_entity_test.go:8 | Minor | Duplicate `mockInfo`/`fakeProcess` in same test package |
| 15 | process_finder.go:101 | Minor | Grammar error in `validate()` error message |

The most important to fix are **#1** (data race), **#7** (goroutine lifecycle), and **#4** (testability).

---

## Test Coverage

Six test files cover all packages. Pure logic is well tested; anything requiring a live terminal (`Run`, `Init`, `Cleanup`, `Draw`) is correctly excluded.

### Race still live in `game_play_test.go`
Tests construct `Game` structs with `running: true` and read `g.running` directly (e.g. `game_play_test.go:50`). These are single-threaded and don't trigger the race themselves, but they are structurally coupled to `running` being a plain `bool`. Fixing issue #1 to `atomic.Bool` requires updating all test constructions to `g.running.Store(true)` and assertions to `g.running.Load()`.

### Score cap boundary check is loose — `score_test.go:52` *(unchanged)*
`TestBoard_Add_CapsAtMax` checks `lowestKills < 5` but the loop adds entries `0..maxScores+4`. The lowest retained score is `5` exactly (the 10th highest out of 15 entries: kills 5..14). The assertion `< 5` is more permissive than the invariant; use `!= 5` or derive the expected minimum from `maxScores` and the loop bounds.

### `TestHandleKeyPress_ConfirmYes` doesn't exercise the signal — `game_play_test.go:109` *(unchanged)*
The test verifies in-game state (`kills`, `freedMem`, `State`) but not that SIGKILL was actually delivered. Issue #9 (score incremented whether or not the signal succeeds) remains untested. Now that `Entity.Kill()` is a standalone method, it can be replaced with an injectable function (`type killer func(pid int) error`) to cover the signal-error path with a table-driven test.

### Undocumented confirm-cancel-with-q behaviour — `game_play_test.go:143` *(unchanged)*
`TestHandleKeyPress_QCancelsConfirm` asserts that `q` during a confirmation dialog cancels the confirm without quitting. This is intentional UX but is not documented in the `usage` string or in a comment near `HandleKeyPress`. Add a line to the usage string under Controls so future maintainers don't mistake it for a bug.
