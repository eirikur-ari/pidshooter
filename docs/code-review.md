# pidshooter — Code Review Report

*Updated 2026-07-09. Reflects current package layout: `core/` domain, `application/service` orchestration with `contract/inbound` and `contract/outbound` port packages, `entrypoint/cli` delivery adapter, `infrastructure/osprocess`/`tcellui`/`scorefilestore` driven adapters.*

---

## Bugs

### ~~21. Corrupt score file silently overwrites all prior scores — `application/service/game.go`~~ ✓ Resolved

`loadErr` is now propagated past the game loop. On a non-nil load error a `warning: could not load scores: <err>` line is printed to stderr and the save is skipped, preserving the file for manual recovery. On a nil load error (including the "file not found" path, which `scorefilestore.Load` already converts to nil) behaviour is unchanged. `TestGameService_LoadError_PrintsWarningAndSkipsSave` (integration) verifies the warning appears, `Play` returns nil, and `Save` is not called.

### ~~22. Trophy not shown when player ties the existing high score — `core/score/score.go`~~ ✓ Resolved

The old code (`main.go:105`) used `game.kills >= scoreBoard.HighScore()` — a tie earned the trophy. The new `PrintHighScore` uses strict greater-than:

```go
func (b *Board) PrintHighScore(kills int) {
    if kills > 0 && kills > b.highScore {   // was >= in old code
        fmt.Println("  🏆 New high score!")
    }
}
```

`b.highScore` is captured in `Add()` as the pre-add top score. If the player matches it exactly (e.g. both 5 kills), `5 > 5` is false and no trophy is shown. The operator should be `>=` to restore the previous behaviour.

`PrintHighScore` now uses `kills >= b.highScore`, restoring the original behaviour from `main.go`. `TestBoard_PrintHighScore_PrintsWhenTiesRecord` pins the exact tie case.

### ~~23. Score `Duration` includes `renderer.Cleanup()` time — `application/service/game.go`~~ ✓ Resolved

`runLoop` defers `s.renderer.Cleanup()` before returning. `Play` computes the recorded duration after `runLoop` returns:

```go
if err := s.runLoop(g); err != nil { ... }  // Cleanup() already ran inside here
// ...
duration := time.Since(g.StartTime()).Seconds()   // includes Cleanup() time
```

The old code computed elapsed time inside the loop before cleanup. The saved `Entry.Duration` now includes terminal teardown time (`tcell.Screen.Fini`). The fix is to snapshot the duration before `runLoop` returns, e.g. via a return value or by reading `g.StartTime()` inside `runLoop` just before the deferred cleanup fires.

`runLoop` now returns `(time.Time, error)`, capturing `time.Now()` as the last statement before the deferred `Cleanup()` fires. `Play` computes `duration := endTime.Sub(g.StartTime()).Seconds()` from that snapshot.

---

## Design Issues

### ~~3. `parseArgs` calls `os.Exit` — `entrypoint/cli/cli.go`~~ ✓ Resolved

`parseArgs` now returns `errUsage` (a package-level sentinel) instead of calling `os.Exit`. `CLI.Run()` detects `errors.Is(err, errUsage)`, prints usage, and returns `nil`. The zero-args and `--help`/`-h` paths are now covered by `TestParseArgs_NoArgs_ReturnsErrUsage`, `TestParseArgs_HelpFlag_ReturnsErrUsage`, `TestRun_NoArgs_PrintsUsageAndReturnsNil`, and `TestRun_HelpFlag_PrintsUsageAndReturnsNil`.

### ~~4. `parseArgs` returns five values — `entrypoint/cli/cli.go`~~ ✓ Resolved

`parseArgs` now returns `(inbound.GamePlayConfig, error)`. Defaults (`Speed: 2.0`, `TimeLimit: 30`) are set directly on the struct; each flag mutates the field in place. `CLI.Run()` passes the returned config straight to `service.Play()`, eliminating the five-value destructure and the manual field assignment.

### ~~5. `defaultPath()` creates a directory as a side effect~~ ✓ Resolved

`defaultPath()` is now a pure path function. Directory creation moved to `makeConfigDir()`, called only from `Save()`, with the error propagated to the caller.

### ~~6. Event poll goroutine can block permanently after game exits~~ ✓ Resolved

`UI` now holds a `done chan struct{}` field. `Cleanup()` closes it before calling `screen.Fini()`. `poll()` resolves each event into a typed value first, then uses a single `select` to either send it or exit on `<-done`. This guarantees the goroutine exits even when the game loop has stopped consuming events and the channel buffer is full.

### ~~7. Signal goroutine leaks after game ends~~ ✓ Resolved

`signal.Stop(sigCh)` now runs before `close(done)` via deferred calls (LIFO). The goroutine uses a `select` on both `sigCh` and `done`, so it exits promptly when `Play()` returns rather than blocking indefinitely on `<-sigCh`.

### ~~8. Kill score recorded even if SIGKILL fails — `loop.go:103–110`~~ ✓ Resolved

`killTarget` now checks the error from `Kill()`. If it returns a non-nil error, the function returns early: no kill animation starts and neither `StartKillAnim` nor `RecordKill` is called. The target remains `Alive`. Tests added for both the failure path (click and confirm-mode) and the existing success-path tests now assert that `Kill` was actually invoked.

### ~~9. High score display is stale during play — `loop.go:89–92`~~ ✓ Resolved

`Session.RecordKill` now updates `highScore` in-place whenever `kills` exceeds it, so `g.highScore` tracks the live kill count and the HUD reflects the new record as soon as it is set. `TestSession_RecordKill_UpdatesHighScore` verifies the boundary: no update while kills ≤ previous high score, then increments correctly on each kill that beats it.

### ~~10. Score save errors silently discarded — `application/service/game.go`~~ ✓ Resolved

`s.store.Save(board)` error is now checked; on failure a `warning: score not saved: <err>` line is printed to stderr. `Play` still returns `nil` — a save failure is not fatal. `fake.Store` now has separate `LoadErr`/`SaveErr` fields so the two paths can be controlled independently. `TestGameService_SaveError_PrintsWarning` (integration) verifies the warning appears and `Play` returns `nil`.

### ~~20. `contract/` mixes inbound and outbound ports in flat files — `application/contract/`~~ ✓ Resolved

`contract/inbound.go` and `contract/outbound.go` replaced by two sub-packages with topic-focused files:

```
contract/
  inbound/
    gameplay.go     → GamePlay, GamePlayConfig
  outbound/
    ui.go           → Renderer, InputEvent, InputSource
    process.go      → ProcessFinder, ProcessKiller
    store.go        → ScoreStore
```

Three interfaces renamed for consistency: `Finder` → `ProcessFinder`, `EventSource` → `InputSource`, `Store` → `ScoreStore`. Port direction is now encoded in the import path — callers write `inbound.GamePlay`, `outbound.ProcessFinder`, `outbound.ScoreStore`.

### 24. `Kill()` blocks the 50 fps game loop with a synchronous `ps` subprocess — `application/service/game.go`

`drainEvents` is called every tick inside the frame loop. When a `KillRequest` is returned, it calls `s.process.Kill()` synchronously before the tick continues:

```go
for g.Running() {
    s.drainEvents(g)      // Kill() called here, may block for hundreds of ms
    w, h = s.renderer.Size()
    g.Update(w, h)
    s.renderer.Render(g.Frame())
    <-ticker.C
}
```

`osprocess.Kill()` spawns a `ps` subprocess (`currentName`) to verify the process name before sending SIGKILL. On a loaded system this can take hundreds of milliseconds, stalling the render loop, skipping kill-animation frames, and starving the `poll()` goroutine (buffer size 10) so subsequent user events are dropped. Consider performing the OS kill asynchronously or finding the name via a faster method.

### 25. Speed-up key `=` alias dropped — requires Shift on standard US keyboard — `core/game/event_handler.go`

The old `handleKeyPress` handled both `r == '+' || r == '='` and `r == '-' || r == '_'`. On a standard US keyboard `+` requires Shift while `=` does not; the old aliases let players adjust speed without modifier keys. The new `HandleKey` only handles bare `'+'` and `'-'`:

```go
case ch == '+':
    g.speed += 0.5
case ch == '-':
    g.speed -= 0.5
```

Speeding up now requires holding Shift on the main keyboard. The usage string (`+/- speed`) does not hint at the change. Fix: restore `ch == '+' || ch == '='` and `ch == '-' || ch == '_'`.

### 11. `PrintScores()` couples the domain to stdout — `core/score/score.go`

`Board.PrintScores()` calls `fmt.Println` and `fmt.Printf` directly, making the core domain package responsible for terminal output. This violates the layering used everywhere else in the codebase (where output goes through the `Renderer` port or the application layer). The display logic belongs in `application/game.go`'s `Play()` after the game completes — the domain should only provide the data.

### ~~12. `NewKiller()` returns a concrete type~~ ✓ Resolved

`NewKiller()` now returns `outbound.ProcessKiller`, consistent with `NewFinder()` returning `outbound.ProcessFinder`.

---

## Minor / Polish

### ~~13. Multi-byte characters: byte counts used for bounds, hit-detection, and rendering~~ ✓ Resolved

`target.go` now uses `utf8.RuneCountInString()` in `NewTarget` (spawn bounds), `Update` (right-wall boundary), and `Contains` (hit-detection). The `tcellui` render loop now uses an independent `col` counter instead of the byte offset from `range`, so multi-byte kill-animation frames (`"✦ KILLED ✦"`, `"· · ·"`) render at the correct screen columns.

### ~~14. HUD elements can overlap on narrow terminals~~ ✓ Resolved

`drawHUD` now pre-computes the start positions of all three elements and only draws the centre `Highscore` label when it fits without overlapping the left `FREED` or right `KILLS` elements. On narrow terminals where the guard fires, `FREED` and `KILLS` remain intact.

### ~~15. `ps` resolved via `$PATH` — `osprocess/osprocess.go:52`~~ ✓ Resolved

`NewFinder()` now calls `exec.LookPath("ps")` at construction time and stores the absolute path in `Finder.psPath`; `List()` uses `f.psPath` instead of the bare string `"ps"`. `NewFinder` returns `(outbound.ProcessFinder, error)` so callers fail fast if `ps` is absent. `main.go` and both test files updated accordingly.

### 26. Seven inline implementation comments removed from `entity.go` not carried into `target.go` — CLAUDE.md violation

CLAUDE.md rule: *"Never remove comments, Javadoc, loggers, or annotations unless explicitly asked to."*

The deleted `entity.go` contained these inline body comments explaining non-obvious choices:

```go
// Ensure entity fits within bounds
spawnMaxY := maxY - 2 // Leave room for status bar
// Random position
// Random velocity scaled by speed multiplier (default slower)
// Kill animation frames
// Bounce off horizontal walls
// Bounce off vertical walls (leave bottom row for status)
```

None appear in the replacement `internal/core/game/target.go`. The godoc lines were re-written under new names (acceptable), but the inline implementation notes explaining *why* were dropped.

### 16. Score table header does not match data — `core/score/score.go`

```go
//TODO: Replace Speed with Time
fmt.Println("  ║  # ║ Kills ║   Freed    ║ Speed ║    Date    ║")
```

`Entry` has both `Speed` and `Time` fields, but only `Speed` is rendered. The time-limit column was never added to the table display.

### ~~17. Grammar error in `validate()` — `osprocess/osprocess.go:96`~~ ✓ Resolved

`"patterns is"` → `"pattern is"`.

### ~~18. `GameService` has no happy-path test — `application/service/game_test.go`~~ ✓ Resolved

`TestGameService_HappyPath` (integration) pre-queues a quit event, runs `Play` to completion, and asserts: `Play` returns `nil`; `store.Saved` is non-nil (Save was called); the board contains one entry with `kills=0` and `duration>0` (confirming `StartTime` was recorded). `fake.Store` gained a `Saved *score.Board` field to capture the argument passed to `Save`.

### ~~19. `tcellui` package has no tests — `infrastructure/tcellui/`~~ ✓ Resolved

Added tests using `tcell.NewSimulationScreen()`: `translateKey` is covered via four `TestPoll_Translates*` cases (Escape, CtrlC, CtrlZ, plain rune → KeyNone); `poll` routing is covered by `TestPoll_MouseButton1_EmitsClickEvent`, `TestPoll_NonButton1_DropsEvent`, and `TestPoll_ResizeEvent_EmitsResizeEvent`; `drawStatusBar` is covered by four cases (normal, confirming, with time limit, without time limit). Shared `newUI`, `nextEvent`, and `rowContent` helpers keep boilerplate out of each test.

---

## Resolved Since Previous Review

| Previous issue | Resolution |
|---|---|
| Stale TODO "Is label the correct word?" (`target.go`) | Removed; `Target` / `Label()` naming is now self-evident |
| Monolithic `game.go` | Split into `game.go`, `loop.go`, `event_handler.go`, `target.go`, `motion.go`, `session.go` |
| `main.go` owning game/session orchestration | Extracted to `application/service.GameService`; `main.go` is now wiring only |
| Duplicate test double types across test packages | Reorganised into `internal/testutil/fake` package (`Killer`, `Finder`, `Store`, `NewProcess`) with idiomatic Go naming (no `Fake` prefix, no package name in file names) |
| #1 — Latent data race on `game.running` | `running bool` → `atomic.Bool`; `stop()`, `update()`, and `event_handler.go` use `Store/Load`; test files use `newRunningGame()` helper for two-step init |
| #2 — `Board.Add()` bool return value dead code | Removed return value; `Board` captures `highScore` before append; new `PrintHighScore(kills int)` method encapsulates the new-high-score decision and output |
| Double `game.Cleanup()` call | `Play()` owns the full lifecycle via `defer g.renderer.Cleanup()` |
| `Entity` / `EntityState` naming | Renamed to `Target` / `TargetState` throughout |
| `process.New()` naming | Renamed to `process.NewFinder()` |
| `game_util.go` in wrong package | Moved to `internal/util/util.go` |
| Testability required `tcell.SimulationScreen` | Game ports (`Renderer`, `EventSource`) are now plain interfaces; integration tests use hand-rolled stubs defined inline, no tcell dependency in tests |
| No separation between domain and infrastructure | Full hexagonal layout: port interfaces in `application/contract/`; core domain in `core/`; adapters in `infrastructure/` and `entrypoint/` |
| #8 — Kill score recorded even if SIGKILL fails | `drainEvents` in `game.go` guards `CompleteKill` behind a nil error check; target stays `Alive` on failure; no score credit |
| #9 — High score display stale mid-game | `Session.RecordKill` now updates `highScore` in-place when `kills` exceeds it; covered by `TestSession_RecordKill_UpdatesHighScore` |
| #10 — Score save error silently discarded | `game.go` now prints `warning: score not saved: <err>` to stderr on save failure; `fake.Store` split into `LoadErr`/`SaveErr`; covered by `TestGameService_SaveError_PrintsWarning` (integration) |
| #15 — `ps` resolved via `$PATH` | `NewFinder()` calls `exec.LookPath("ps")` at construction, stores absolute path in `Finder.psPath`; signature changed to `(outbound.ProcessFinder, error)`; `main.go` and test files updated |
| #18 — No happy-path test for `GameService` | `TestGameService_HappyPath` (integration) verifies end-to-end success, Save called, session state recorded; `fake.Store` gained `Saved` field |
| #19 — No tests for `tcellui` | Added goroutine-exit, HUD, multi-byte rendering, `translateKey` (×4), poll routing (×3), and `drawStatusBar` (×4) tests using `tcell.NewSimulationScreen()` |
| #3 — `parseArgs` calls `os.Exit` | `errUsage` sentinel returned instead; `Run()` prints usage and returns `nil`; four new tests cover both zero-args and `--help`/`-h` paths |
| #4 — `parseArgs` returns five values | Returns `(inbound.GamePlayConfig, error)`; defaults set on struct literal; `Run()` passes config directly to `service.Play()` |
| #20 — `contract/` flat file layout | Split into `contract/inbound/` and `contract/outbound/` sub-packages; outbound split into `ui.go`, `process.go`, `store.go`; `Finder` → `ProcessFinder`, `EventSource` → `InputSource`, `Store` → `ScoreStore` |

---

## Summary

| # | File | Severity | Issue |
|---|------|----------|-------|
| 1 | `core/game/game.go` | ✓ Resolved | Latent data race on `game.running` (signal goroutine vs game loop) |
| 2 | `core/score/score.go`, `application/service/game.go` | ✓ Resolved | `Board.Add()` bool return value is dead code; never consumed at call site |
| 3 | `entrypoint/cli/cli.go` | ✓ Resolved | `parseArgs` calls `os.Exit` — zero-args and `--help` paths untestable |
| 4 | `entrypoint/cli/cli.go` | ✓ Resolved | Five return values; `inbound.GamePlayConfig` already exists |
| 5 | `scorefilestore/score_file_store.go` | ✓ Resolved | `defaultPath()` creates directory as side effect with silently discarded error |
| 6 | `infrastructure/tcellui/tcellui.go` | ✓ Resolved | Poll goroutine blocks on channel send after game exits — goroutine leak |
| 7 | `core/game/game.go` | ✓ Resolved | Signal goroutine leaks after game ends |
| 8 | `application/service/game.go` (`drainEvents`) | ✓ Resolved | Kill score recorded even if SIGKILL fails |
| 9 | `core/game/loop.go` | ✓ Resolved | High score display stale mid-game |
| 10 | `application/service/game.go` | ✓ Resolved | Save error silently discarded |
| 11 | `core/score/score.go` | Design | `PrintScores()` on domain type — stdout I/O belongs in app layer |
| 12 | `infrastructure/osprocess/osprocess.go` | ✓ Resolved | `NewKiller()` returns `*Killer` not the port interface |
| 13 | `core/game/target.go`; `infrastructure/tcellui/tcellui.go` | ✓ Resolved | Byte count/offset used for bounds, hit-detection, and rendering — breaks for multi-byte chars |
| 14 | `infrastructure/tcellui/tcellui.go` | ✓ Resolved | HUD elements overlap on narrow terminals |
| 15 | `infrastructure/osprocess/osprocess.go` | ✓ Resolved | `ps` found via `$PATH` |
| 16 | `core/score/score.go` | Minor | Score table shows Speed column; TODO says replace with Time |
| 17 | `infrastructure/osprocess/osprocess.go` | ✓ Resolved | Grammar: "patterns is" → "pattern is" |
| 18 | `application/service/game_integration_test.go` | ✓ Resolved | No happy-path test for `GameService` |
| 19 | `infrastructure/tcellui/tcellui_test.go` | ✓ Resolved | No tests for `tcellui` package |
| 20 | `application/contract/` | ✓ Resolved | Flat file layout mixes inbound/outbound; outbound groups four unrelated concerns; `Finder`, `EventSource`, `Store` names lack specificity |
| 21 | `application/service/game.go:57-59` | ✓ Resolved | Corrupt score file silently overwrites all prior scores |
| 22 | `core/score/score.go:51` | ✓ Resolved | Trophy not shown when player ties existing high score (`>` should be `>=`) |
| 23 | `application/service/game.go:71` | ✓ Resolved | Recorded `Duration` includes `renderer.Cleanup()` time, not pure game time |
| 24 | `application/service/game.go:145` | Design | `Kill()` blocks the 50 fps game loop with a synchronous `ps` subprocess |
| 25 | `core/game/event_handler.go:43` | Bug/Regression | Speed-up key `=` alias dropped — now requires Shift on standard US keyboard |
| 26 | `core/game/target.go` | CLAUDE.md | Seven inline comments from `entity.go` not carried into `target.go` |
| 27 | `core/game/loop.go:28-45` | Minor | `Frame()` iterates `g.targets` twice; alive count can be accumulated in the first pass |
| 28 | `core/game/target.go:74` | Minor | Kill-animation `frames` slice allocated on every `Label()` call; should be package-level var |
| 29 | `infrastructure/tcellui/tcellui.go:169` | Minor | `ResizeEvent` emitted but never consumed — dead abstraction |
| 30 | `entrypoint/cli/cli.go:76-108` | Minor | `--help` hint inconsistently appended to some cli error messages but not others |

---

## Test Coverage

Fourteen test files cover all packages. Pure logic and state transitions are well covered. All packages including `tcellui` have direct unit tests.

### Build and test commands

| Command | What it runs |
|---|---|
| `make build` | Compiles `bin/pidshooter` |
| `make test` | All tests (unit + integration) with verbose output |
| `make test-unit` | Unit tests only (no `integration` build tag) |
| `make test-integration` | All tests including integration files |
| `make test-short` | All tests, no verbose output |
| `make test-race` | All tests under the race detector |
| `make coverage` | Unit test coverage report in `bin/coverage.out` |
| `make vet` | `go vet ./...` |
| `make fmt` | `go fmt ./...` |
| `make run ARGS="firefox"` | Build and run with the given pattern |
| `make clean` | Remove `bin/` |

### Integration tests

Two files carry `//go:build integration` tags:

- `application/service/game_integration_test.go` — runs a full game loop with hand-rolled stub renderer and event source (no tcell dependency). Covers quit-on-Q, quit-on-Escape, time-limit expiry, happy-path save, corrupt-load warning, save-error warning, and signal goroutine lifecycle. All tests are named `TestIntegration_GameService_*`. Run with `make test-integration`.
- `infrastructure/osprocess/osprocess_integration_test.go` — calls the real `ps` command. Covers non-empty results, valid fields, short names, and own-PID exclusion on a live system.

### Race detector

Issue #1 (data race on `game.running`) has been resolved — `running` is now an `atomic.Bool` and all reads/writes go through `Store`/`Load`. Issue #7 (signal goroutine leak) has also been resolved — `runLoop()` uses a done channel so the signal goroutine exits via `select` when the game ends rather than blocking indefinitely on `<-sigCh`. Verified by `TestGameService_SignalGoroutineDoesNotAccumulate`.

### Score cap assertion is loose — `score_test.go`

`TestBoard_Add_CapsAtMax` adds entries `0..maxScores+4` and asserts:

```go
if lowestKills < 5 {
    t.Errorf(...)
}
```

The loop produces entries with kills `0–14`; the top `maxScores` retained entries are kills `5–14`, making the lowest retained exactly `5`. The assertion `lowestKills < 5` passes vacuously for any value ≥ 5 and would not catch an off-by-one error that retained kill-count 4 instead. The assertion should be `!= 5` (or derive the expected minimum from `maxScores` and the loop bounds) to tighten the invariant.

### ~~`TestHandleKeyPress_ConfirmYes` does not verify the signal — `core/game/event_handler_test.go`~~ ✓ Resolved

`TestHandleKeyPress_ConfirmYes_ReturnsKillRequest` verifies the returned `*KillRequest` carries the target. The kill-error path is now an application-layer concern tested by `TestGameService_*` integration tests. `TestHandleMouseClick_ReturnsKillRequest` verifies click-to-kill returns a `*KillRequest`.

### 27. `Frame()` iterates `g.targets` twice per tick — `core/game/loop.go`

`Frame()` has two independent for-range loops over the same slice: one builds the `TargetView` list (skipping `Dead`) and a separate one counts `Alive` targets. The alive count can be accumulated in the first pass, halving the iterations per render tick:

```go
// current: two passes
for _, e := range g.targets { /* build TargetView */ }
for _, e := range g.targets { if e.State == Alive { alive++ } }  // redundant

// fix: accumulate alive in first pass
for _, e := range g.targets {
    if e.State == Dead { continue }
    if e.State == Alive { alive++ }
    targets = append(targets, TargetView{...})
}
```

### 28. Kill-animation `frames` slice allocated on every `Label()` call — `core/game/target.go`

```go
func (e *Target) Label() string {
    case Killing:
        frames := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}  // new alloc every call
```

`Label()` is called at least twice per tick per killing target (in `Frame()` and in `Update()`). The slice literal is constant; it should be a package-level `var` to allocate once.

### 29. `ResizeEvent` emitted but never consumed — `infrastructure/tcellui/tcellui.go`

`poll()` emits `event.ResizeEvent{}` into the event channel on every terminal resize, but `drainEvents` has no case for it — the value is read from the channel and silently dropped. Actual resize handling works correctly by re-querying `s.renderer.Size()` at the top of each game tick. `ResizeEvent` is dead infrastructure: it occupies channel capacity on every resize without serving any purpose. Either remove it or wire it to a resize handler.

### 30. `--help` hint inconsistently appended in cli.go error messages — `entrypoint/cli/cli.go`

`"\nRun 'pidshooter --help' for usage"` is appended to some error messages (invalid flag, bad parse) but omitted from others (speed out-of-range, time out-of-range, missing pattern). The hint should be applied uniformly — either added once in `Run()` after any parse error, or appended consistently at every error site.

### Undocumented confirm-cancel-with-q behaviour

`TestHandleKeyPress_QCancelsConfirm` tests that pressing `q` during a confirmation dialog cancels the confirm without quitting. This intentional UX decision is not reflected in the `usage` string in `entrypoint/cli/cli.go`. A one-line addition under Controls (`q  Cancel confirmation / Quit`) would prevent future maintainers from treating it as a bug.
