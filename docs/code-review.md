# pidshooter — Code Review Report

*Updated 2026-06-17. Previous findings noted with resolution status.*

---

## Bugs

### 1. Data race on `game.running` — `game.go:25,58,104` *(unchanged)*
`running` is a plain `bool`. The signal handler goroutine calls `g.Stop()` (which writes `g.running = false` at `game.go:58`) while the main loop reads it in `for g.running` at `game.go:104`. These are unsynchronised accesses. `go test -race ./...` would catch this.

```go
// fix: replace plain bool with atomic
type Game struct {
    running atomic.Bool
    // ...
}
func (g *Game) Stop() { g.running.Store(false) }
// for loop: for g.running.Load() { ... }
```

### 2. Broken `isHighScore` logic — `score.go:72–73` *(unchanged)*
```go
isHighScore := len(b.Scores) <= maxScores ||
    (len(b.Scores) > 0 && b.Scores[len(b.Scores)-1] != entry)
```
After sorting, `entry` can land anywhere. Checking whether the *last* element equals `entry` is unreliable because `Entry` contains a `time.Time` field: two entries with identical kills/mem but different timestamps are never equal, so the condition is almost always `true`. The return value is also silently discarded at `main.go:94` (`scoreBoard.Add(entry)`), making it dead code. Since `main.go` has its own check (`g.Kills() >= scoreBoard.HighScore()`), the simplest fix is to drop the bool return from `Add()`.

### 3. `game.Cleanup()` called twice — `main.go:71,97` *(unchanged)*
`defer g.Cleanup()` is registered at line 71, then called explicitly at line 97. The explicit call restores the terminal before printing results. The deferred one fires afterwards but is a nil-safe no-op. It works, but the `defer` is misleading — it looks like the safety net but the explicit call already did the work. Remove the `defer` and rely solely on the explicit call; document in a comment that callers must call `Cleanup()` before printing post-game output.

---

## Design Issues

### 4. `parseArgs` calls `os.Exit` — `main.go:112,145` *(partially resolved)*
Two `os.Exit(0)` calls remain:
- `main.go:112` — zero-argument invocation prints usage and exits
- `main.go:145` — `--help`/`-h` flag prints usage and exits

The missing-pattern case (e.g. `parseArgs([]string{"--confirm"})`) was fixed to return an error, and `TestParseArgs_NoPatterns` covers it. But the zero-arg and `--help` paths are still untestable. Return a sentinel error (e.g. `errUsage`) that `run()` detects and exits cleanly without `os.Exit` in `parseArgs`.

### 5. `filePath()` creates a directory as a side effect — `score.go:36–38`
A function that returns a path string shouldn't create directories. `os.MkdirAll` runs on every `Load()` and `Save()` call. Move the `MkdirAll` into `Save()` only, where the directory is actually needed.

### 6. Five return values from `parseArgs` — `main.go:109` *(unchanged)*
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

### 7. Event goroutine can block permanently after game exits — `game.go:93–109`
The goroutine started in `Run()` sends to `eventCh` (capacity 10). Once the game loop exits (`g.running` turns false), the loop stops calling `drainEvents()`. If the channel fills and the goroutine tries to send one more event, it blocks on `eventCh <- ev`. At that point `Cleanup()` calls `screen.Fini()` — but the goroutine is blocked on the channel send, not on `PollEvent()`, so `Fini()` cannot unblock it. The goroutine leaks until the process exits. For a short-lived CLI this is harmless in practice, but a `context.Context` or done channel would make the lifetime explicit and correct:

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

### 8. `game.highScore` is stale during play — `main.go:66` *(unchanged)*
The high score is loaded once before the game starts. If kills exceed the stored high score mid-game, the HUD still shows the old number. Compare `g.kills` against `g.highScore` live inside `render()` and update `g.highScore` in `killEntity` if the current kill count exceeds it.

### 9. Kill score incremented even if signal fails — `game.go:323–330` *(unchanged)*
```go
proc, err := os.FindProcess(e.PID)
if err == nil {
    _ = proc.Signal(syscall.SIGKILL)
}
e.StartKillAnim()
g.kills++
g.freedMem += e.RSS
```
On Linux, `os.FindProcess` never fails (it only validates the PID numerically), so even a process that has already exited passes the guard. The signal error is discarded and the kill is counted regardless. Log signal errors at minimum; ideally only increment the score after a successful signal.

---

## Minor / Polish

### 10. `len()` counts bytes, not terminal columns — `entity.go:37,102,142` *(unchanged)*
`len(label)` counts UTF-8 bytes. Kill animation frames include multi-byte emoji (`"💥"`). During the `StateKilling` phase, `Update()` uses `len(e.Label())` for bounce bounds, and `Contains()` uses it for hit detection — both will be wrong for any frame that contains a multi-byte character. Use `utf8.RuneCountInString()` or, for the emoji case, `uniseg.StringWidth()` for correct column widths.

### 11. HUD elements can overlap on narrow terminals — `game.go:244–274` *(unchanged)*
FREED, Highscore, and KILLS are drawn independently on row 0 with no awareness of each other. On a terminal narrower than ~50 columns they overwrite each other. A single formatted status line (or explicit minimum-width guard) would be more robust.

### 12. `ps` resolved via `$PATH` — `process.go:69` *(unchanged)*
`exec.Command("ps", ...)` relies on `$PATH`. On Linux, reading `/proc` directly would eliminate the subprocess and the PATH dependency entirely.

---

## Resolved Since Previous Review

| Previous # | Issue | Resolution |
|---|---|---|
| Makefile `lint` | Phantom `.PHONY` target with no recipe | Removed from `.PHONY` entirely |
| Makefile `coverage` | Would fail on clean checkout | Added `@mkdir -p $(BUILD_DIR)` |
| #4 (partial) | `parseArgs` exits on missing pattern | Now returns error; `TestParseArgs_NoPatterns` covers it |

---

## Summary

| # | File | Severity | Issue |
|---|------|----------|-------|
| 1 | game.go:25 | **Bug** | Data race on `game.running` |
| 2 | score.go:72 | **Bug** | `isHighScore` logic wrong and return value unused |
| 3 | main.go:71,97 | **Bug** | Double `Cleanup()` — misleading defer |
| 4 | main.go:112,145 | Design | `parseArgs` calls `os.Exit` (zero args + --help) |
| 5 | score.go:36 | Design | `filePath()` creates directory as side effect |
| 6 | main.go:109 | Design | 5 return values; should be a config struct |
| 7 | game.go:93 | Design | Event goroutine can block on channel send after game exits |
| 8 | main.go:66 | Design | Stale high score in HUD |
| 9 | game.go:323 | Design | Kill score incremented even if SIGKILL fails |
| 10 | entity.go:37 | Minor | `len()` bytes vs rune width (breaks with emoji kill frames) |
| 11 | game.go:244 | Minor | HUD elements overlap on narrow terminals |
| 12 | process.go:69 | Minor | `ps` found via `$PATH` |

The most important to fix are **#1** (data race), **#7** (goroutine lifecycle), and **#4** (testability).

---

## Test Coverage

Five test files cover all packages. Pure logic is well tested; anything requiring a live terminal (`Run`, `Init`, `Cleanup`, `Draw`) is correctly excluded.

### Race still live in `game_test.go`
Tests construct `Game` structs with `running: true` and read `g.running` directly (e.g. `game_test.go:54`). These are single-threaded and don't trigger the race themselves, but they are structurally coupled to `running` being a plain `bool`. Fixing issue #1 to `atomic.Bool` requires updating all test constructions to use `g.running.Store(true)` and assertions to use `g.running.Load()`.

### `TestFind_ValidPattern` passes vacuously when binary is renamed — `process_test.go:31`
The test searches for `"process.test"` (the name the test binary gets from `go test`). If the binary is built with `-o` or a different name, the search finds zero results and the test passes vacuously — the only assertion is `p.PID != 1`. Add `if len(results) == 0 { t.Fatal("expected at least one match") }` to make the self-match check loud.

### Score cap boundary check is loose — `score_test.go:52`
`TestBoard_Add_CapsAtMax` checks `lowestKills < 5` but the loop adds entries `0..maxScores+4`. The lowest retained score is `5` exactly (the 10th highest out of 15 entries: kills 5..14). The assertion `< 5` is more permissive than the invariant; use `!= 5` or derive the expected minimum from `maxScores` and the loop bounds.

### `TestHandleKeyPress_ConfirmYes` doesn't exercise the signal — `game_test.go:123`
The test verifies in-game state (`kills`, `freedMem`, `State`) but not that SIGKILL was actually delivered. Issue #9 (score incremented whether or not the signal succeeds) remains untested. Extracting the kill dispatch into an injectable function (`type killer func(pid int) error`) would allow a table-driven test covering the signal-error path.

### Undocumented confirm-cancel-with-q behaviour — `game_test.go:157`
`TestHandleKeyPress_QCancelsConfirm` asserts that `q` during a confirmation dialog cancels the confirm without quitting. This is intentional UX but is not documented in the usage string or in a comment near `HandleKeyPress`. Add a line to the usage string under Controls so future maintainers don't mistake it for a bug.

---

## Makefile

### No `test-race` target *(unchanged)*
Given the confirmed data race on `game.running`, the Makefile should include:
```makefile
## test-race: Run tests with the race detector
test-race:
	go test -race -count=1 ./...
```
This would catch the race in CI without requiring contributors to remember the flag.

### `run` target ARGS quoting *(unchanged)*
```makefile
ARGS ?= sleep
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)
```
`ARGS` values containing `=` (e.g. `make run ARGS="sleep --speed=2.5"`) may be mis-parsed by some shells when expanded unquoted. Quote `$(ARGS)` in the recipe, or document the exact invocation form in the `## run:` comment.
