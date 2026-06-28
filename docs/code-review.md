# pidshooter — Code Review Report

*Updated 2026-06-27. Reflects fully implemented hexagonal architecture: `app` application service, `cli`/`tcellui`/`osprocess`/`jsonscores` adapters, port interfaces in `ports/driven` and `ports/driving` sub-packages.*

---

## Bugs

All reported bugs have been resolved — see Resolved table below.

---

## Design Issues

### 3. `parseArgs` calls `os.Exit` — `cli/cli.go:63, 97`

Two `os.Exit(0)` calls remain inside `parseArgs`:

```go
if len(args) == 0 {
    fmt.Println(usage)
    os.Exit(0)  // line 63 — untestable
}
// ...
case arg == "--help" || arg == "-h":
    fmt.Println(usage)
    os.Exit(0)  // line 97 — untestable
```

Both paths are unreachable in tests, meaning there is no coverage for the zero-args or `--help` code paths. Return a sentinel error (e.g. `errUsage`) and let `CLI.Run()` detect and print usage before returning `nil`, keeping `parseArgs` a pure function.

### 4. `parseArgs` returns five values — `cli/cli.go:60`

```go
func parseArgs(args []string) ([]string, bool, float64, int, error)
```

`driving.Config` already has exactly these four data fields. `parseArgs` could return `(driving.Config, error)`, eliminating the five-value return and the manual field assignment inside `CLI.Run()`.

### ~~5. `defaultPath()` creates a directory as a side effect~~ ✓ Resolved

`defaultPath()` is now a pure path function. Directory creation moved to `makeConfigDir()`, called only from `Save()`, with the error propagated to the caller.

### 6. Event poll goroutine can block permanently after game exits — `tcellui/tcellui.go:154,157,160`

```go
func (a *UI) poll() {
    for {
        ev := a.screen.PollEvent()
        if ev == nil { return }
        switch ev := ev.(type) {
        case *tcell.EventMouse:
            a.ch <- driven.ClickEvent{...}   // blocks if channel full
        case *tcell.EventKey:
            a.ch <- driven.KeyEvent{...}     // blocks if channel full
        case *tcell.EventResize:
            a.ch <- driven.ResizeEvent{}     // blocks if channel full
        }
    }
}
```

`a.ch` has capacity 10. If the game loop exits (and `drainEvents` stops consuming), `poll` blocks on the next channel send. When `Cleanup()` then calls `screen.Fini()`, `PollEvent` would return `nil` to unblock `poll` — but `poll` is stuck at the *channel send*, not at `PollEvent`. The goroutine leaks for the lifetime of the process. Fix with a done channel:

```go
func (a *UI) poll(done <-chan struct{}) {
    for {
        ev := a.screen.PollEvent()
        if ev == nil { return }
        // ...
        select {
        case a.ch <- event:
        case <-done: return
        }
    }
}
```

### 7. Signal goroutine leaks after game ends — `game/game.go:64–70`

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
defer signal.Stop(sigCh)
go func() {
    <-sigCh   // blocks here forever if no signal arrives
    g.stop()
}()
```

`signal.Stop(sigCh)` stops routing future signals to the channel but does not close it, so the goroutine stays blocked on `<-sigCh` after the game loop returns. For a single-session process this is harmless, but integration tests that create multiple `Game` instances accumulate blocked goroutines. Close the channel after `signal.Stop`, or switch to a context/done-channel pattern to guarantee cleanup.

### 8. Kill score recorded even if SIGKILL fails — `loop.go:103–110`

```go
func (g *Game) killTarget(e *Target) {
    if e.State != Alive { return }
    _ = g.killer.Kill(e.Pid())   // error silently discarded
    e.StartKillAnim()
    g.Session.RecordKill(e.Rss())
}
```

`Kill()` returns an error, but it is always discarded and the kill and freed-memory counters are always incremented. On platforms where sending SIGKILL to an already-exited PID succeeds silently, scores are inflated. At minimum log the error; ideally call `RecordKill` only when `Kill` returns `nil`.

### 9. High score display is stale during play — `loop.go:89–92`

```go
HUD: gamedriven.HUDState{
    // ...
    HighScore: g.highScore,  // set once before the loop, never updated
},
```

`g.highScore` is initialised in `Play()` via `g.SetHighScore(highScore)` and never changed during the session. If the player beats the record mid-game, the HUD continues to display the pre-game high score. Update `g.highScore` inside `killTarget` when `g.kills` exceeds it.

### 10. Score save errors silently discarded — `app/runner.go:75`

```go
_ = s.store.Save(board)
```

If the score file cannot be written (disk full, permission denied, stale NFS mount), the user sees the summary and score table printed from the in-memory board, with no indication that the result was not persisted. The error should at minimum be logged to stderr.

### 11. `PrintScores()` couples the domain to stdout — `score/score.go:64–83`

`Board.PrintScores()` calls `fmt.Println` and `fmt.Printf` directly, making the domain package responsible for terminal output. This violates the layering used everywhere else in the codebase (where output goes through the `Renderer` port or the application layer). The display logic belongs in `app.GameService.Play()` after the game completes — the domain should only provide the data.

### ~~12. `NewKiller()` returns a concrete type~~ ✓ Resolved

`NewKiller()` now returns `gamedriven.ProcessKiller`, consistent with `NewFinder()` returning `driven.Finder`.

---

## Minor / Polish

### 13. Multi-byte characters: byte counts used for bounds, hit-detection, and rendering — `target.go:34,68,79`, `tcellui/tcellui.go:64`

This is a cross-layer issue. In the domain:

- `target.go:34` — `NewTarget` computes spawn bounds with `len(fmt.Sprintf("[%d %s]", ...))`, which counts UTF-8 bytes.
- `target.go:68` — `Update` passes `float64(len(e.Label()))` as the right-wall boundary for bounce calculations.
- `target.go:79` — `Contains` uses `len(e.Label())` for click hit-detection.

In the adapter:

- `tcellui.go:64` — `for i, ch := range tv.Label` gives `i` as a **byte offset**, then uses `tv.X+i` as the screen column. For kill-animation frames containing multi-byte characters (`"✦ KILLED ✦"`, `"· · ·"`), the rendered characters land at wrong column positions, leaving visual gaps.

Fix: use `utf8.RuneCountInString()` for character count in the domain. Fix the rendering loop to track the column index independently:

```go
col := 0
for _, ch := range tv.Label {
    a.screen.SetContent(tv.X+col, tv.Y, ch, nil, style)
    col++
}
```

### 14. HUD elements can overlap on narrow terminals — `tcellui/tcellui.go:87, 99`

`FREED`, `Highscore`, and `KILLS` are drawn independently on row 0 — left-aligned, centred, and right-aligned respectively — with no awareness of each other. On terminals narrower than ~50 columns they overwrite each other. A single formatted status line or a minimum-width guard would be more robust.

### 15. `ps` resolved via `$PATH` — `osprocess/osprocess.go:52`

```go
cmd := exec.Command("ps", flags, "pid,rss,comm")
```

`ps` is found by searching `$PATH`. On Linux, reading `/proc` directly would eliminate the subprocess and the PATH dependency entirely. The existing TODO comment acknowledges this.

### 16. Score table header does not match data — `score/score.go:71`

```go
//TODO: Replace Speed with Time
fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
```

`Entry` has both `Speed` and `Time` fields, but only `Speed` is rendered. The time-limit column was never added to the table display.

### 17. Grammar error in `validate()` — `osprocess/osprocess.go:96`

```go
return fmt.Errorf("at least one search patterns is required")
```

Should be `"at least one search pattern is required"` (singular).

### 18. `GameService` has no happy-path test — `app/runner_test.go`

`runner_test.go` covers `FinderError` and `NoProcesses` but not the normal path where processes are found and the game runs to completion. With `FakeFinder`, `FakeKiller`, `FakeStore`, and the stub renderer/event source already in `testutil`, a complete happy-path test is straightforward. The session state (`Kills`, `FreedMem`, `StartTime`) and score persistence call are currently uncovered.

### 19. `tcellui` package has no tests — `adapter/driven/tcellui/`

`tcellui` implements both `Renderer` and `EventSource`. `drawHUD`, `drawStatusBar`, `translateKey`, and the event loop in `poll()` have no test coverage. `tcell.NewSimulationScreen()` provides a headless screen suitable for unit tests without a real terminal.

---

## Resolved Since Previous Review

| Previous issue | Resolution |
|---|---|
| Stale TODO "Is label the correct word?" (`target.go`) | Removed; `Target` / `Label()` naming is now self-evident |
| Monolithic `game.go` | Split into `game.go`, `loop.go`, `event_handler.go`, `target.go`, `motion.go`, `session.go` |
| `main.go` owning game/session orchestration | Extracted to `app.GameService`; `main.go` is now wiring only |
| Duplicate test double types across test packages | Reorganised into `internal/testutil/fake` package (`Killer`, `Finder`, `Store`, `NewProcess`) with idiomatic Go naming (no `Fake` prefix, no package name in file names) |
| #1 — Latent data race on `game.running` | `running bool` → `atomic.Bool`; `stop()`, `update()`, and `event_handler.go` use `Store/Load`; test files use `newRunningGame()` helper for two-step init |
| #2 — `Board.Add()` bool return value dead code | Removed return value; `Board` captures `highScore` before append; new `PrintHighScore(kills int)` method encapsulates the new-high-score decision and output |
| Double `game.Cleanup()` call | `Play()` owns the full lifecycle via `defer g.renderer.Cleanup()` |
| `Entity` / `EntityState` naming | Renamed to `Target` / `TargetState` throughout |
| `process.New()` naming | Renamed to `process.NewFinder()` |
| `game_util.go` in wrong package | Moved to `internal/util/util.go` |
| Testability required `tcell.SimulationScreen` | Game ports (`Renderer`, `EventSource`) are now plain interfaces; integration tests use hand-rolled stubs defined inline, no tcell dependency in tests |
| No separation between domain and infrastructure | Full hexagonal layout: domain ports in `ports/driven` and `ports/driving` sub-packages; adapters in `adapter/driven/` and `adapter/driving/` |

---

## Summary

| # | File | Severity | Issue |
|---|------|----------|-------|
| 1 | `game.go:24`, `loop.go:26`, `game.go:76` | ✓ Resolved | Latent data race on `game.running` (signal goroutine vs game loop) |
| 2 | `score/score.go:36`, `app/runner.go:67` | ✓ Resolved | `Board.Add()` bool return value is dead code; never consumed at call site |
| 3 | `cli/cli.go:63,97` | Design | `parseArgs` calls `os.Exit` — zero-args and `--help` paths untestable |
| 4 | `cli/cli.go:60` | Design | Five return values; `driving.Config` already exists |
| 5 | `jsonscores/jsonscores.go:29` | ✓ Resolved | `defaultPath()` creates directory as side effect with silently discarded error |
| 6 | `tcellui/tcellui.go:154,157,160` | Design | Poll goroutine blocks on channel send after game exits — goroutine leak |
| 7 | `game/game.go:64–70` | Design | Signal goroutine leaks after game ends |
| 8 | `loop.go:103–110` | Design | Kill score recorded even if SIGKILL fails |
| 9 | `loop.go:89–92` | Design | High score display stale mid-game |
| 10 | `app/runner.go:75` | Design | Save error silently discarded |
| 11 | `score/score.go:64–83` | Design | `PrintScores()` on domain type — stdout I/O belongs in app layer |
| 12 | `osprocess/osprocess.go:130` | ✓ Resolved | `NewKiller()` returns `*Killer` not the port interface |
| 13 | `target.go:34,68,79`; `tcellui.go:64` | Minor | Byte count/offset used for bounds, hit-detection, and rendering — breaks for multi-byte chars |
| 14 | `tcellui/tcellui.go:87,99` | Minor | HUD elements overlap on narrow terminals |
| 15 | `osprocess/osprocess.go:52` | Minor | `ps` found via `$PATH` |
| 16 | `score/score.go:71` | Minor | Score table shows Speed column; TODO says replace with Time |
| 17 | `osprocess/osprocess.go:96` | Minor | Grammar: "patterns is" → "pattern is" |
| 18 | `app/runner_test.go` | Minor | No happy-path test for `GameService` |
| 19 | `adapter/driven/tcellui/` | Minor | No tests for `tcellui` package |

The highest-priority open fixes are **#6** (goroutine leak in tcellui poll) and **#7** (signal goroutine leak).

---

## Test Coverage

Fourteen test files cover all packages except `tcellui` and the port declaration packages. Pure logic and state transitions are well covered.

### Integration tests

Two files carry `//go:build integration` tags:

- `game_integration_test.go` — runs a full game loop with hand-rolled `testRenderer`/`testEventSource` stubs (no tcell dependency). Covers quit-on-Q, quit-on-Escape, time-limit expiry, and session state after quit. Run with `go test -tags integration ./...`.
- `osprocess_integration_test.go` — calls the real `ps` command. Covers non-empty results, valid fields, short names, and own-PID exclusion on a live system.

### Race detector and signal goroutine

Issue #1 (data race on `game.running`) has been resolved — `running` is now an `atomic.Bool` and all reads/writes go through `Store`/`Load`. The signal goroutine (issue #7) still leaks after game exit: it blocks on `<-sigCh` indefinitely because `signal.Stop` does not close the channel. This is harmless for a single-session process but accumulates blocked goroutines in integration tests that create multiple `Game` instances.

### Score cap assertion is loose — `score_test.go`

`TestBoard_Add_CapsAtMax` adds entries `0..maxScores+4` and asserts:

```go
if lowestKills < 5 {
    t.Errorf(...)
}
```

The loop produces entries with kills `0–14`; the top `maxScores` retained entries are kills `5–14`, making the lowest retained exactly `5`. The assertion `lowestKills < 5` passes vacuously for any value ≥ 5 and would not catch an off-by-one error that retained kill-count 4 instead. The assertion should be `!= 5` (or derive the expected minimum from `maxScores` and the loop bounds) to tighten the invariant.

### `TestHandleKeyPress_ConfirmYes` does not verify the signal — `event_handler_test.go`

The test verifies in-game state (`kills`, `freedMem`, target `State`) but not whether `Kill()` was actually invoked. Issue #8 (score recorded regardless of kill success) is therefore not exercised. Since `FakeKiller` records called PIDs, asserting `len(fakeKiller.KilledPIDs) == 1` would close this gap without any new infrastructure.

### Undocumented confirm-cancel-with-q behaviour

`TestHandleKeyPress_QCancelsConfirm` tests that pressing `q` during a confirmation dialog cancels the confirm without quitting. This intentional UX decision is not reflected in the `usage` string in `cli.go`. A one-line addition under Controls (`q  Cancel confirmation / Quit`) would prevent future maintainers from treating it as a bug.
