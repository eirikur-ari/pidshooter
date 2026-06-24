# pidshooter — Code Review Report

*Updated 2026-06-24. Reflects fully refactored package layout: `runner`, `game`, `process`, `score`, `util`, `testutil`.*

---

## Bugs

### 1. Data race on `game.running` — `game.go:22`, `loop.go:59,71,84`, `event_handler.go:48–59`

`running` is a plain `bool`. The signal handler goroutine (spawned in `game.go:59`) calls `g.stop()` which writes `g.running = false` at `game.go:68`, while the main loop reads `for g.running` at `loop.go:59` and mutates it at `loop.go:71,84`. These are unsynchronised cross-goroutine accesses. `go test -race ./...` will catch this.

```go
// fix: replace plain bool with atomic
import "sync/atomic"

type Game struct {
    running atomic.Bool
    // ...
}
func (g *Game) stop() { g.running.Store(false) }
// loop: for g.running.Load() { ... }
// update: g.running.Store(false)
```

Test constructions that write `running: true` directly (e.g. `loop_test.go:63`, `event_handler_test.go:13`) must be updated to call `g.running.Store(true)`.

### 2. Broken `isHighScore` logic — `score.go:72–73`

```go
isHighScore := len(b.Scores) <= maxScores ||
    (len(b.Scores) > 0 && b.Scores[len(b.Scores)-1] != entry)
```

After sorting, `entry` can land at any index. Checking whether the *last* element equals `entry` is unreliable: `Entry` contains a `time.Time`, so two entries with the same kills/mem but different timestamps are never `==`, making the condition almost always `true`. The return value is also discarded at the call site (`runner.go:47` calls `scoreBoard.Add(...)` without capturing the result), making it dead code. Simplest fix: drop the bool return from `Add()`. The check in `runner.go:59` (`g.Kills() >= scoreBoard.HighScore()`) already does what the caller actually needs.

---

## Design Issues

### 3. `parseArgs` calls `os.Exit` — `main.go:58–60, 92–94`

Two `os.Exit(0)` calls remain:
- `main.go:58` — zero-argument invocation prints usage and exits
- `main.go:93` — `--help`/`-h` flag prints usage and exits

These make `parseArgs` untestable for those paths. Return a sentinel error (e.g. `errUsage`) and let `run()` detect and handle it, printing usage there instead of inside the parser.

### 4. `parseArgs` returns five values — `main.go:56`

```go
func parseArgs(args []string) ([]string, bool, float64, int, error)
```

`runner.Config` already exists with exactly these four data fields. `parseArgs` could return `(runner.Config, error)` to eliminate the five-value tuple and reduce construction boilerplate in `run()`.

### 5. `filePath()` creates a directory as a side effect — `score.go:38`

A function returning a path string should not create directories. `os.MkdirAll` runs on every `Load()` and `Save()` call. Move it into `Save()` only, where the directory is actually needed.

### 6. Event goroutine can block permanently after game exits — `loop.go:49–57`

```go
go func() {
    for {
        ev := g.screen.PollEvent()
        if ev == nil { return }
        eventCh <- ev   // blocks if channel is full
    }
}()
```

`eventCh` has capacity 10. If the channel is full and `PollEvent` returns another event, the goroutine blocks on the send. When the game loop exits and `cleanup()` calls `screen.Fini()`, `Fini()` unblocks `PollEvent` — but the goroutine is blocked on the *channel send*, not on `PollEvent`. The goroutine leaks for the lifetime of the process. Fix with a context or done channel:

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

### 7. Signal goroutine leaks after game ends — `game.go:56–62`

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
defer signal.Stop(sigCh)
go func() {
    <-sigCh
    g.stop()
}()
```

After `g.run()` returns, the goroutine is blocked on `<-sigCh`. `signal.Stop(sigCh)` stops routing signals to the channel but does not close it, so the goroutine stays blocked until a signal arrives or the process exits. For a single game session this is harmless, but in integration tests that create multiple `Game` instances it accumulates. Close the channel after `signal.Stop`, or switch to a done-channel pattern.

### 8. Kill score incremented even if `SIGKILL` fails — `loop.go:191`

```go
func (g *Game) killTarget(e *Target) {
    if e.State != Alive { return }
    _ = e.Kill()           // error silently discarded
    e.StartKillAnim()
    g.Session.RecordKill(e.Rss())
}
```

`Target.Kill()` returns an error (`target.go:95`), but it is always discarded and the score is always incremented. On Linux, `os.FindProcess` succeeds for any PID, so a process that has already exited passes the guard and the score is inflated. At minimum log the error; ideally only call `RecordKill` after a successful signal.

### 9. High score never updates mid-game — `loop.go:125`

```go
hiStr := fmt.Sprintf(" Highscore: %d ", g.highScore)
```

`g.highScore` is set once in `Play()` via `g.SetHighScore(highScore)` and never changes during the session. If the player beats the record mid-game, the HUD still shows the pre-game value. Compare `g.kills` against `g.highScore` inside `killTarget` and update via `g.SetHighScore(g.kills)` when exceeded.

---

## Minor / Polish

### 10. `len()` counts bytes, not display columns — `target.go:37, 71, 83`

`len(label)` counts UTF-8 bytes. Kill animation frames contain multi-byte emoji (`"💥"` — 4 bytes, 2 display columns). `Update()` passes `float64(len(e.Label()))` as the label width for bounce bounds; `Contains()` uses `len(e.Label())` for hit detection. Both are wrong for frames containing multi-byte characters. Use `utf8.RuneCountInString()` for character count, or `uniseg.StringWidth()` for display column width.

### 11. HUD elements can overlap on narrow terminals — `loop.go:117–147`

`FREED`, `Highscore`, and `KILLS` are drawn independently on row 0 with no awareness of each other. On terminals narrower than ~50 columns they overwrite each other. A single formatted status line or an explicit minimum-width guard would be more robust.

### 12. `ps` resolved via `$PATH` — `finder.go:54`

`exec.Command("ps", ...)` relies on `$PATH`. On Linux, reading `/proc` directly would eliminate the subprocess and the PATH dependency. The existing TODO comment in the file acknowledges this.

### 13. Score table header does not match data — `score.go:97`

```go
//TODO: Replace Speed with Time
fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
```

`Entry` has both `Speed` and `Time` fields, but the table only shows Speed. The time-limit column was never added to the display.

### 14. Grammar error in `validate()` — `finder.go:100`

`"at least one search patterns is required"` should be `"at least one search pattern is required"`.

### 15. Stale TODO comment — `target.go:36`

```go
//TODO: Is label the correct word for the game entity that represents the process?
```

This is stale. `Target` is the entity; "label" refers to its rendered string. Remove the comment or replace it with a short inline explanation.

### 16. `runner` has no happy-path test — `runner_test.go`

`runner_test.go` covers `TestStart_FinderError` and `TestStart_NoProcesses` but not the case where processes are found and the game runs to completion. The internal `start()` function accepts a `process.Finder`, and `Game.UseScreen()` allows injecting a `tcell.SimulationScreen`, so a full-path integration test is achievable without mocking.

---

## Resolved Since Previous Review

| Previous issue | Resolution |
|---|---|
| Double `game.Cleanup()` call | `Play()` owns the full lifecycle via `defer g.cleanup()`; runner calls only `g.Play()` |
| Duplicate mock types in game tests | Consolidated into `internal/testutil.NewFakeProcess()`, used across all test packages |
| Monolithic `game.go` | Split into `game.go`, `loop.go`, `event_handler.go`, `target.go`, `motion.go`, `session.go` |
| `main.go` owning game/session orchestration | Extracted into `internal/runner`; `main.go` is now parse → start → error |
| `Entity` / `EntityState` naming | Renamed to `Target` / `TargetState` throughout |
| `process.New()` naming | Renamed to `process.NewFinder()` to follow Go multi-constructor conventions |
| `game_util.go` in wrong package | Moved to `internal/util/util.go` |
| No testability for game screen | `Game.UseScreen(tcell.Screen)` injection added; `game_integration_test.go` uses `tcell.NewSimulationScreen` |

---

## Summary

| # | File | Severity | Issue |
|---|------|----------|-------|
| 1 | game.go:22, loop.go:59 | **Bug** | Data race on `game.running` |
| 2 | score.go:72 | **Bug** | `isHighScore` logic wrong; return value unused |
| 3 | main.go:58,93 | Design | `parseArgs` calls `os.Exit` (zero args + --help paths untestable) |
| 4 | main.go:56 | Design | 5 return values; `runner.Config` already exists |
| 5 | score.go:38 | Design | `filePath()` creates directory as side effect |
| 6 | loop.go:49 | Design | Event goroutine can block on send after game exits |
| 7 | game.go:59 | Design | Signal goroutine leaks after game ends |
| 8 | loop.go:191 | Design | Kill score incremented even if SIGKILL fails |
| 9 | loop.go:125 | Design | High score display is stale during play |
| 10 | target.go:37,71,83 | Minor | `len()` counts bytes, not display columns |
| 11 | loop.go:117 | Minor | HUD elements overlap on narrow terminals |
| 12 | finder.go:54 | Minor | `ps` found via `$PATH` |
| 13 | score.go:97 | Minor | Score table shows Speed; TODO says replace with Time |
| 14 | finder.go:100 | Minor | Grammar error in `validate()` message |
| 15 | target.go:36 | Minor | Stale TODO comment |
| 16 | runner_test.go | Minor | No happy-path integration test for runner |

The highest priority fixes are **#1** (data race), **#6** (goroutine leak on channel send), and **#7** (signal goroutine leak).

---

## Test Coverage

Fourteen test files span all packages. Pure logic and state transitions are well covered. Terminal rendering (`render`, `drawStatusBar`) and the full game loop are exercised by integration tests using `tcell.NewSimulationScreen`.

### Integration tests — `game_integration_test.go`

Uses `//go:build integration` and `tcell.NewSimulationScreen`. Covers quit-on-Q, quit-on-Escape, time-limit expiry, and session state after quit. Run with `go test -tags integration ./...`.

### Race still live in unit tests

Tests in `game_test.go`, `loop_test.go`, and `event_handler_test.go` construct `Game` structs with `running: true` and read `g.running` directly. These are single-threaded and do not trigger the race detector, but they will require updating when issue #1 is fixed to `atomic.Bool`.

### Score cap boundary check is loose — `score_test.go`

`TestBoard_Add_CapsAtMax` asserts `lowestKills < 5`, but the loop adds entries `0..maxScores+4` and the lowest retained kill count is `5` exactly. The assertion should be `!= 5` (or derive the expected minimum from `maxScores` and loop bounds) to tighten the invariant.

### `TestHandleKeyPress_ConfirmYes` does not verify the signal — `event_handler_test.go:72`

The test verifies in-game state (`kills`, `freedMem`, `State`) but not whether `SIGKILL` was actually delivered. Issue #8 (score incremented regardless of signal success) remains untested. `Target.Kill()` is a standalone method (`target.go:95`), so it can be replaced with an injectable function type to cover the signal-error path.

### Undocumented confirm-cancel-with-q behaviour

`TestHandleKeyPress_QCancelsConfirm` asserts that `q` during a confirmation dialog cancels the confirm without quitting the game. This intentional UX is not documented in the `usage` string. A one-line addition under Controls would prevent future maintainers from treating it as a bug.
