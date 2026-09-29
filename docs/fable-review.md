# pidshooter — Fable AI Code Review

**Reviewed by:** Claude Fable 5
**Date:** 2026-08-20
**Scope:** Full codebase — functional correctness, design, security, architecture
**Supersedes:** the 2026-07-10 review. This is a from-scratch review against the current package layout, written after the service-layer refactor that split the former monolithic `internal/application/service` package into `internal/application/{game,process,score}` plus `internal/application/{runner.go,event/dispatcher.go}`. No file paths or findings were carried forward from the prior review without re-verifying them against current line numbers and logic.

**Re-verified 2026-09-21** against current source (package/file renames since: `scorefilestore`→`filescore`, `osprocess.go` split into multiple files, `entrypoint/cli/cli.go`→`program.go`, `tcellui.go` split into several files including `tui.go`/`translator.go`, `killer`→`process.Service` still structurally satisfied, `applyKills`/`s.kills`→`applyKillSignals`/local `killSignals` chan). Every finding below was individually re-checked against current source, not assumed. F1/F2/F4/D3/D7 were superseded by `docs/cli-findings.md`'s own later, more thorough pass — see that document for the fuller writeup; only a brief pointer is left here.

---

## 1. Executive Summary

The refactor is a real improvement, not just a reshuffle: `core/game`, `core/movement`, `core/process`, and `core/score` have zero imports outside the standard library, `internal/util`, and each other; the former `core/game ↔ application/process` coupling is now avoided entirely through structural typing (`application/game/service.go` declares its own unexported `killer` interface, satisfied by `application/process.Service` without either package importing the other); and pattern-length validation is unified into one implementation (`core/process.Validate`) called from both the CLI and the application layer instead of being duplicated.

That said, several real bugs and boundary violations remain, and one is a fresh regression introduced by the refactor:

- **The `--speed` flag's help text and `docs/features.md` both advertise a range of `0.1–5.0`, but the enforced minimum is `0.5`** (`movement.MinSpeed`) — confirmed by actually running `--speed=0.2`, which is rejected.
- **`--speed=NaN` still bypasses range validation** — confirmed by direct comparison: NaN fails both `<` and `>` checks, so `cli.go`'s guard silently lets it through.
- **Kills that complete asynchronously right as a session ends can still be lost from the score** — the async kill-confirmation channel is only drained while the session is still reported as running, and there's no final drain or wait for in-flight kill goroutines.
- **Console I/O still leaks into `core/score` (domain) and into both `application/process` and `application/score` (use-case layer)**, undermining the hexagonal boundary the rest of the codebase is careful to respect.
- **No inbound-port validation** — `Speed`/`TimeLimit` are checked only by the one driving adapter (`entrypoint/cli`); the application service and its `Config` struct accept anything.
- **Security posture is essentially unchanged from before**: no safeguard against running as root, and the residual PID-recycling TOCTOU window (inherent to signal-by-PID) is unchanged. Both are pre-existing, not refactor regressions.

No critical or new-and-exploitable security findings; the SIGKILL-path safety layering (name re-verification, `PID ≤ 1` guard, minimum pattern length) is intact and was re-verified against the new package layout.

**2026-09-21 status of the bullets above:** the `--speed` help text (F1), `--speed=NaN` (F2), and no-inbound-port-validation (D3) items are resolved. Lost final-frame kills (F3) and the console-I/O leak (D1 resolved, D2 half-resolved — see below) are the two genuinely still-open items from this list.

**2026-09-29 status:** the root-safety bullet is now also resolved — see S1 below. F6–F10 and S3 (found during this pass to already be fixed by earlier, unrelated refactoring work, but left stale as "still open" in this document) are corrected in place below.

---

## 2. Functional Issues

**~~F1. [moderate] `--speed` help text and docs claim a `0.1–5.0` range; the enforced minimum is `0.5`.~~ ✓ Resolved — see `docs/cli-findings.md` Finding 4.**
`internal/entrypoint/cli/cli.go:57` registers the flag with `"Speed multiplier (range: 0.1-5.0)"`, and `docs/features.md:34` repeats the same `0.1–5.0` range. But `internal/core/movement/throttle.go:4-6` defines `MinSpeed = 0.5`, and `validate()` (`cli.go:88-90`) enforces exactly that constant. Confirmed by running the built binary:
```
$ pidshooter --speed=0.2 sleep
Error: speed must be between 0.5 and 5, got: 0.2
```
The error message correctly reports `0.5`, but the `--help` text a user reads first still says `0.1`. This is a regression from the refactor: `MinSpeed` used to be `0.1` (per the prior review's citation of the old bounds); it was tightened to `0.5` without updating the flag description or `docs/features.md`.

**~~F2. [moderate] `--speed=NaN` still bypasses range validation.~~ ✓ Resolved — see `docs/cli-findings.md` Finding 2.**
`cli.go:88-90`:
```go
if speed < movement.MinSpeed || speed > movement.MaxSpeed {
    return fmt.Errorf(...)
}
```
All comparisons with `NaN` are `false`, so neither branch fires and `NaN` passes. Confirmed directly: `NaN < 0.5 || NaN > 5.0` evaluates to `false`. `NaN` then flows unclamped into `movement.NewThrottle` (no clamping at construction — only `Increase`/`Decrease` clamp) and from there into every `Position`/`Velocity` computation in `core/movement/motion.go:27-31`, making targets un-renderable/unclickable. There is no second line of defense (see D3) — any caller that reaches `Play` without going through `cli.go`'s `validate()` gets no protection at all.

**F3. [moderate] Kills that complete asynchronously as (or after) a session ends can be lost from the score.**
`internal/application/game/service.go`'s `runLoop` (lines 80-116) drains completed kills via `applyKills` (lines 118-133) only once per iteration, while `gs.IsRunning()` is still `true`. `drainEvents` (lines 145-172) dispatches each accepted kill target to a goroutine that runs `s.killer.Kill(target)` (a `ps` lookup plus a signal syscall — not instantaneous) and only reports the result if it can still send on `s.kills` before `done` is closed. `Session.Update` (`internal/core/game/session.go:56-67`) can call `Stop()` independently of any in-flight kill — e.g. on timer expiry — so a kill request accepted in the same tick the timer expires races the loop's exit: the goroutine's result arrives after the final `applyKills` call and after `close(done)` (deferred at `service.go:92`), so it takes the `case <-done:` branch and the result — the process **was** SIGKILLed, but the score's kill/freed-memory count never reflects it and the target's on-screen state never updates. This is not covered by `internal/application/runner_integration_test.go`, which has no test exercising a kill in flight when the session stops.

**Still open, confirmed 2026-09-21** — renamed (`applyKills`/`s.kills` → `applyKillSignals`/local `killSignals` chan in `application/game/service.go`) but the mechanism is unchanged: `runLoop` calls `frameLoop` then immediately returns and `close(done)` fires via `defer`, with no final drain of `killSignals` and no `sync.WaitGroup` wait for in-flight `killOrReap` goroutines. A kill whose result arrives after the last in-loop `applyKillSignals` call — including one landing in the buffered channel just as the session stops — is never applied. The suggested fix (drain once more after the loop, or wait on the in-flight goroutines) has not been implemented.

**~~F4. [moderate] `--time` has no upper bound; large values silently become "no time limit."~~ ✓ Resolved — see `docs/cli-findings.md` Finding 3.**
`cli.go:92-94` only rejects negative values. `internal/core/game/timer.go:14-19` computes `limit: time.Duration(limitSeconds) * time.Second` — `time.Duration` is `int64` nanoseconds, so this multiplication overflows for `limitSeconds` beyond ~292 years. Confirmed directly: `time.Duration(9223372036854775807) * time.Second` evaluates to `-1s`. `Timer.Expired()` treats `limit <= 0` as "no limit" (`timer.go:31-33`), so e.g. `--time=9223372036854775807` — a value `IntVar` happily accepts — silently disables the time limit instead of erroring, the opposite of what was requested.

**~~F5. [moderate] A failed OS-level kill is silently swallowed — no reap, no error, no player feedback.~~ ✓ Resolved**
`internal/application/process/service.go:48-63`'s `Kill` returns `(killed=false, shouldReap=false, err=non-nil)` when the final `s.process.Kill(pid)` call fails (line 61-62) — e.g. `EPERM` because pidshooter isn't running as root/owner of the target process, a realistic case given there's no root check (S1). Back in `internal/application/game/service.go`'s `drainEvents` goroutine (lines 154-166), the `switch` only has cases for `err == nil && killed` and `shouldReap` — a real, non-reap error matches neither, so it is dropped entirely. Unlike the load/save-score paths (`application/score/service.go:28,50`), nothing is printed to stderr; the target just silently stays `Alive` forever with no indication to the player that their click did nothing.

**Resolved, confirmed 2026-09-21** — `applyKillSignals` (`application/game/service.go`) now has an explicit `case sig.err != nil: tracker.recordFailure(sig.target, sig.err)` branch — a failed kill is recorded, not dropped. `application.Runner.logKillFailures` reports every recorded failure via `apperror.Handle` (a `"could not kill %s (PID %d)"` warning to stderr) once the session ends. Feedback is post-session rather than in-the-moment, but the original complaint — nothing printed, ever, no indication the click did nothing — no longer holds.

**~~F6. [moderate] Hit-detection truncates while the rendered position rounds — clicks miss the visible target.~~ ✓ Resolved**
`internal/core/game/target.go:98-104`'s `isHitAt` uses `int(t.Position.Y)`/`int(t.Position.X)` (truncation toward zero), while `internal/application/game/converter.go:18-25`'s `toTargetViewState` uses `math.Round` for the same coordinates. Targets accumulate non-integer positions every tick via velocity (`core/movement/motion.go:27-31`), so whenever the fractional part of `Position.Y` (or `X`) is `≥ 0.5`, the tag is drawn one row (or column) away from where a click actually registers. Since fractional parts are effectively uniformly distributed over time, this affects roughly half of all click attempts on a moving target.

**Still open, confirmed 2026-09-21** — same finding as `docs/code-review.md` #34: `isHitAt` (`core/game/target.go:124` at current line numbers) still truncates, `toTargetViewState` (`application/game/converter.go:29-30`) still rounds. Unchanged.

**Resolved, confirmed 2026-09-29** — `Target.isHitAt` (`core/game/target.go`) now computes `row := int(math.Round(t.Motion.Position.Y))` and `col := int(math.Round(t.Motion.Position.X))`, matching `toTargetViewState`'s `math.Round` exactly. Both sides of the mismatch this finding described now agree. Not fixed as part of this cleanup pass — found already resolved by earlier, unrelated refactoring work — but the doc was left stale claiming "still open," which this entry corrects.

**~~F7. [minor] `LookupName` failures for any reason are treated as "process already exited."~~ ✓ Resolved**
`internal/application/process/service.go:53-56` maps *any* error from `s.process.LookupName(pid)` — which just wraps `ps`'s `exec.Command(...).Output()` error (`internal/infrastructure/osprocess/osprocess.go:49-52`) — to `shouldReap = true`. A transient `ps` failure (resource exhaustion, unexpected exit code, spawn failure) is indistinguishable from "PID no longer exists," so the target gets silently reaped (marked `Dead`, no kill, no score credit) even if the real process is still running.

**Still open, confirmed 2026-09-21** — `process/service.go`'s `Kill` still maps any `LookupName` error directly to `shouldReap = true` (`return true, apperror.NewError(...)` immediately after the `LookupName` call, no error-type discrimination). Unchanged.

**Resolved, confirmed 2026-09-29** — `process.Service.Kill` (`application/process/service.go`) now checks `errors.As(err, &outbound.NotFoundError{})` on the `LookupName` error specifically: only that sentinel maps to `shouldReap = true`; every other error returns `shouldReap = false` with a `"could not verify PID %d"` message, leaving the target alive rather than silently reaping it. Covered by `TestKillReturnsErrorWithoutReapingIfProcessLookupByNameFailsTransiently` and `TestKillReturnsShouldReapIfProcessLookupByNameReportsProcessNotFound` (`application/process/service_test.go`). Found already resolved by earlier, unrelated refactoring work, not this cleanup pass — the doc was left stale claiming "still open," which this entry corrects.

**~~F8. [minor] Zero-kill sessions are still recorded to the score board.~~ ✓ Resolved**
`internal/application/score/service.go:37-53`'s `RecordScore` unconditionally calls `board.Add(...)` regardless of `tracker.Kills`. Quitting instantly (`q`) still creates a 0-kill/0-freed entry, and it gets persisted whenever the initial load succeeded (`persist=true` is `success` from `LoadScoreBoard`, independent of whether any kills happened), polluting the top-10 table until it fills with real scores.

**Still open, confirmed 2026-09-21** — `score/service.go`'s `RecordScore` still calls `board.Add(entry)` unconditionally with no `Kills > 0` guard, and `application/runner.go`'s `Run` still calls `RecordScore` unconditionally for every session regardless of kill count. Unchanged.

**Resolved, confirmed 2026-09-29** — `core/score.Entry.isScore()` now requires `Kills > 0` (and `FreedMem >= 0`), and `Board.Add` returns early via `if !entry.isScore() { return }` before appending — a zero-kill entry is silently dropped rather than stored. Found already resolved by earlier, unrelated refactoring work, not this cleanup pass — the doc was left stale claiming "still open," which this entry corrects.

**~~F9. [minor] Timer display truncates instead of rounding up.~~ ✓ Resolved**
`internal/core/game/timer.go:54-56`'s `SecondsLeft()` returns `int(t.Remaining().Seconds())` — a floor, so e.g. 0.9s remaining still displays "0s" for up to a full second before the game actually ends.

**Still open, confirmed 2026-09-21** — `timer.go`'s `SecondsLeft()` is unchanged: `int(t.Remaining().Seconds())`.

**Resolved, confirmed 2026-09-29** — `timer.SecondsLeft()` (`core/game/timer.go`) now returns `int(math.Ceil(t.Remaining().Seconds()))`, rounding up instead of truncating; the doc comment states this explicitly ("rounded up so a session isn't shown as having 0 seconds left before it has actually expired"). Found already resolved by earlier, unrelated refactoring work, not this cleanup pass — the doc was left stale claiming "still open," which this entry corrects.

**~~F10. [minor] Score-list tie-break is nondeterministic.~~ ✓ Resolved**
`internal/core/score/board.go:78-82`'s `sortByRank` uses `sort.Slice` (not `SliceStable`), and `Entry.beats` (`internal/core/score/entry.go:19-30`) ties out fully identical entries (same kills, speed, duration, and freed memory) with no further tie-break. Which of two otherwise-identical entries survives the `maxScores` cutoff (`board.go:37-39`) is unspecified.

**Still open, but narrowed, confirmed 2026-09-21** — `board.go` still calls `sort.Slice` (not `SliceStable`). `Entry.beats` gained a new tie-break dimension since this finding was written: it now ranks by Kills → Duds → Speed → Duration → FreedMem (Duds didn't exist as a scoring field at review time). Two entries now have to match on all five dimensions, including an exact float `Duration` match, to hit the unstable-sort ambiguity — a much narrower real-world case than originally described, but the underlying mechanism (`sort.Slice`, not `SliceStable`) is unchanged.

**Resolved, confirmed 2026-09-29** — `board.sortByRank` (`core/score/board.go`) now calls `sort.SliceStable`, closing the unstable-sort ambiguity outright rather than merely narrowing it. Found already resolved by earlier, unrelated refactoring work, not this cleanup pass — the doc was left stale claiming "still open," which this entry corrects.

**~~F11. [minor] Process-name canonicalization can differ between discovery and kill-time verification.~~ ✓ Resolved**
`internal/infrastructure/osprocess/osprocess.go:105` (`list()`) collapses internal whitespace runs via `strings.Fields(...)` + `Join(..., " ")`, while `LookupName` (`osprocess.go:44-54`, `TrimSpace` at line 53) only trims the outer whitespace of the raw `ps` output without collapsing interior runs. A `comm` value with irregular internal spacing would never satisfy `validateProcessName` (`internal/application/process/validation.go:13-18`), silently refusing every kill attempt on that target with no player-visible feedback.

**Resolved, confirmed 2026-09-21** — `osprocess/process.go` now has a single shared `parseProcesses(output []byte)` helper, explicitly documented as "shared by" both call sites, used by both `Discover()` and `LookupName()`. Both paths now canonicalize identically — the mismatch this finding described no longer exists.

**~~F12. [minor, test-only] `capture.Output`/`capture.Stderr` don't restore the streams via `defer`.~~ Irrelevant — package removed**
`internal/testutil/capture/capture.go:11-24` and `:27-40` reassign `os.Stdout`/`os.Stderr` back to the original after calling `fn()`, without `defer`. A panicking `fn` leaves the package-level stream variable pointed at a closed pipe for every subsequent test in the process.

**Confirmed 2026-09-21** — `internal/testutil/capture` no longer exists at all. Tests that need to assert on printed output now construct an injectable `io.Writer` (e.g. `bytes.Buffer`) directly on the type under test (see `console.ScoreReporter`, `cli.Program`), rather than swapping the global `os.Stdout`/`os.Stderr`. The finding is moot — there's no shared capture helper left to have this bug.

---

## 3. Design Issues

**~~D1. [moderate] Domain package performs console I/O.~~ ✓ Resolved**
`internal/core/score/board.go:46-68`'s `PrintHighScores` calls `fmt.Println`/`fmt.Printf` directly from `core/score`. The refactor removed this class of violation from `core/game` (frame/view-model assembly now lives in `application/game/converter.go`) but it persists in `core/score` — the file even carries its own acknowledging TODO at line 42 ("*perhaps this should not be part of a domain layer... Maybe move to a view package?*").

**Resolved, confirmed 2026-09-21** — `core/score/board.go` has no `fmt` import and no printing method at all now; it's pure ranking/data logic. All score-table/trophy output goes through `infrastructure/console.ScoreReporter` (an `outbound.ScoreReporter` port implementation) via `application/score/service.go`. Same fact independently confirmed via `docs/code-review.md` #11.

**D2. [moderate] Application-layer services write directly to stdout/stderr instead of going through a port.**
`internal/application/process/service.go:34,38` (`fmt.Printf` for match-count messages) and `internal/application/score/service.go:28,50,58-59` (`fmt.Fprintf(os.Stderr, ...)` warnings plus the `fmt.Printf` game-over summary in `PrintResults`) all bypass the `outbound.Renderer` port that the rest of the application is careful to route through. None of this is testable without `testutil/capture`, which is exactly what `internal/application/runner_integration_test.go` has to reach for.

**Half-resolved, confirmed 2026-09-21** — `application/score/service.go` no longer prints anything directly at all: warnings route through `apperror.Handle`, and the game-over summary/table route through `console.ScoreReporter`, an actual outbound port. `application/process/service.go:34` is unchanged, though: `fmt.Printf("Found %d process(es) matching %v. Starting game...\n", ...)` still bypasses every port, unrouted. (This exact line was independently flagged again during this session's design discussion for a planned `--list`/`--dry-run` flag — see the `project_planned_cli_modes` memory — since a list-only invocation would need this message to not say "Starting game...".)

**~~D3. [moderate] The inbound port performs no validation of its own; boundary defense lives in exactly one driving adapter.~~ ✓ Resolved — see `docs/cli-findings.md` §5 (D3).**
`internal/application/contract/inbound/runner.go`'s `Config` and `internal/application/game/service.go`'s `Play` (lines 62-74) accept any `Speed`/`TimeLimit` without checking them. F2's NaN gap and F4's overflow gap only get caught today because `entrypoint/cli/cli.go`'s `validate()` happens to run first and is the *only* caller of `Runner.Run`. A hexagonal inbound port is supposed to be the architectural boundary; nothing here stops a second driving adapter (a future daemon, a test harness calling the runner directly, a scripting entrypoint) from skipping every one of these guards silently.

**~~D4. [minor] `Service.kills` is mutable struct state assigned inside `runLoop`, not passed as a parameter.~~ ✓ Resolved**
`internal/application/game/service.go:31` (field) and `:104` (assignment). A `*game.Service` is driven by exactly one `Runner.Run()` call in the current wiring (`internal/application/runner.go`), so this isn't exercised today, but it's a latent data race if the same `*Service` were ever reused for two concurrent `Play` calls, and it forces any would-be unit test of `applyKills`/`drainEvents` in isolation to reach into unexported state.

**Resolved, confirmed 2026-09-21** — `Service` no longer has a `kills` field at all. Kill tracking now lives in a `*killTracker` value created fresh per `Play` call and threaded explicitly as a parameter through `runLoop`/`frameLoop`/`applyKillSignals`; the `killSignals` channel is likewise a local variable inside `frameLoop`, not struct state. Two concurrent `Play` calls on the same `*Service` would no longer share any mutable kill-tracking state.

**D5. [minor] `Target` embeds `movement.Motion`, promoting `Move()` and bypassing the `State`-gated `Update()` that's supposed to be the only mutator.**
`internal/core/game/target.go:28-33` embeds both `process.Info` and `movement.Motion`. Confirmed directly: calling the promoted `target.Move(bounds, speed, tagWidth)` on a `Dead` target moves its position while leaving `State` unchanged — the guard inside `Target.Update()` (`target.go:56-66`) that's supposed to gate all movement is entirely circumventable by any code holding a `*Target`, since `Move` is exported via promotion regardless.

**Still open, confirmed 2026-09-21** — `Target` still embeds `movement.Motion` (`process.Info` + `movement.Motion` + `State`/`AnimationTick`/`shotFired`), `Motion.Move` is still exported, `Target.Update` still exists as the intended gated mutator. Unchanged.

**D6. [minor] `sync/atomic` in the "pure" core exists solely to support the application layer's concurrency design.**
`internal/core/game/lifecycle.go` wraps session state in `atomic.Int32` because `application/game/service.go`'s signal-handling goroutine (lines 93-99) calls `gs.Stop()` from a separate goroutine while the main loop goroutine concurrently reads `IsRunning()`/calls `Update()`. This is a real, currently-necessary requirement given how signals are wired today (funneling `os/signal` into the main loop's own `select` instead of a separate `Stop()`-calling goroutine would let `core/game` drop `sync/atomic` entirely) — noted as a design trade-off, not a bug.

**Still open, confirmed 2026-09-21** — same finding as `docs/code-review.md` #42: `atomicLifecycle` (`core/game/lifecycle.go`) is unchanged, still wrapping `atomic.Int32` solely for the signal-handling goroutine's benefit.

**~~D7. [minor] Composition root builds adapters that can fail or block before argument parsing.~~ ✓ Resolved — see `docs/cli-findings.md` Finding 1.**
`cmd/pidshooter/main.go:24` (`osprocess.NewProcess()`, requires `ps` resolvable via `exec.LookPath`) and `:30` (`tcell.NewScreen()`) both run before `cli.NewCLI(runner).Run(os.Args[1:])` ever parses flags (line 36-37). `pidshooter --help` still pays for, and can fail on, both adapters in the current layout.

**~~D8. [low] `UI.Cleanup()` has no re-entry guard.~~ ✓ Resolved**
`internal/infrastructure/tcellui/tcellui.go:44-47` closes `a.done` unconditionally; a second call panics on the already-closed channel. Not triggered by the single call site in `game/service.go` today, but the `outbound.Renderer` port doesn't document one-shot semantics either.

**Resolved, confirmed 2026-09-21** — `TUI.Cleanup()` (`infrastructure/tcellui/tui.go`) is now `t.cleanupOnce.Do(t.cleanup)`, a `sync.Once` guard. Safe to call any number of times.

---

## 4. Security Concerns

**~~S1. [moderate] No safeguard against running as root.~~ ✓ Resolved**
Confirmed: no `os.Geteuid()` or equivalent check exists anywhere in `cmd/pidshooter/main.go` or `internal/infrastructure/osprocess`. As root, every process on the machine (besides `PID ≤ 1` and pidshooter itself) becomes a one-click SIGKILL target, compounded by S3 (drag-to-kill) and F5 (a failed kill against a non-owned process now fails completely silently, which — combined with running as non-root, the safer default — at least fails closed rather than crashing).

**Partially mitigated, confirmed 2026-09-21** — matches `docs/cli-findings.md`'s own assessment: `core/process.IsKillableBy` plus `Find`'s UID filter now scope a non-root player's target list to processes they own, closing the non-root case. The root case this finding is actually about — running pidshooter as root still makes every process on the machine a target — is unchanged; no `os.Geteuid()` check exists anywhere.

**Resolved 2026-09-29** — see `docs/security-issues.md` SEC-07: `core/process.ValidateNotRoot` refuses to run as root unless the caller passes `--i-am-root`, enforced as the first check inside `application/process.Service.FindProcesses`.

**S2. [moderate] Residual TOCTOU window between name re-verification and SIGKILL.**
`internal/application/process/service.go:53-61` looks up the current name via `ps` (`LookupName`), compares it, then calls `s.process.Kill(pid)` (`internal/infrastructure/osprocess/osprocess.go:59-68`, a plain `os.FindProcess` + `Signal`). The window between the two — a `ps` subprocess round-trip — remains open for a PID recycle. Inherent to signal-by-PID on both Linux and macOS; unchanged by the refactor.

**Still open, confirmed 2026-09-21** — inherent to signal-by-PID; unchanged, as expected (this was never something a refactor would fix).

**~~S3. [low] Mouse-drag fires a kill on every motion sample while the button is held, not just on press.~~ ✓ Resolved**
`internal/infrastructure/tcellui/tcellui.go:158-163` forwards every `*tcell.EventMouse` where `Buttons() == Button1`, including motion events while the button stays down. In the default (non-confirm) mode, sweeping the cursor across the screen with the button held triggers a `ClickEvent`, and therefore a kill attempt, on every target the cursor passes over in one continuous gesture.

**Still open, confirmed 2026-09-21** — `translator.go`'s `translateMouseEvent` still checks only `mouseEvent.Buttons() != tcell.Button1` (i.e. accepts any event reporting Button1 down), with no distinction between a press and a motion sample taken while the button is held. Unchanged.

**Resolved 2026-09-29** — see `docs/security-issues.md` SEC-09: `translator` now tracks the previous button state and only emits a `ClickEvent` on the press transition.

**S4. [low] Unbounded concurrent kill goroutines.**
`internal/application/game/service.go:153` spawns one goroutine (a `ps` subprocess plus a signal syscall) per accepted kill target, with no cap. Combined with S3, a single drag gesture can spawn many concurrent `ps` invocations.

**Still open, confirmed 2026-09-21** — `go s.killOrReap(target, killSignals, done)` is still spawned once per accepted target with no cap; the `killSignals` channel's buffer (10) bounds only how many *results* can queue, not how many goroutines can run concurrently. Unchanged.

**S5. [low] SIGKILL only — no graceful termination option.**
`internal/infrastructure/osprocess/osprocess.go:64` always signals `SIGKILL`: no flush, no cleanup handler, guaranteed data loss for the target process. Unchanged from before.

**Still open, confirmed 2026-09-21** — by design; unchanged.

**~~S6. [low] Score file written non-atomically.~~ ✓ Resolved**
`internal/infrastructure/scorefilestore/score_file_store.go:67` — `os.WriteFile` truncates and rewrites `highscores.json` in place; a crash mid-write corrupts it permanently. File permissions (`0600`/`0700`, `score_file_store.go:67`, `:115`) are correctly restrictive.

**Resolved, confirmed 2026-09-21** — `filescore` (renamed from `scorefilestore`) now saves via `fsutil.WriteFileAtomic`, which writes to a temp file and renames it into place — a crash mid-write can no longer leave a truncated/corrupted `highscores.json`.

**S7. [low] `ps` output parsing trusts field structure.**
`internal/infrastructure/osprocess/osprocess.go:81-113` splits raw `ps` output on lines/whitespace with only a `len(fields) < 3` sanity check. Exploitability remains substantially neutralized by the kill-time name re-verification (S2's mechanism), unchanged from before.

**Still open, minor tightening only, confirmed 2026-09-21** — the sanity check is now `len(fields) < 5` (was `< 3`), plus a header-row detection heuristic, but parsing still trusts positional field structure once that check passes. Exploitability still substantially neutralized by S2's kill-time name re-verification, same as originally noted.

---

## 5. Architecture Observations & Proposals

**What the refactor got right.** `core/game`, `core/movement`, `core/process`, and `core/score` genuinely have zero outward imports — verified by reading every non-test file's import block, not just the package diagram. The former `core/game` ↔ `os`/`syscall`/`tcell` couplings documented in `docs/hexagonal-arc.md` are gone. The kill flow's two-phase design (core returns intent, application performs the OS side effect, result flows back through a channel) now runs the OS call off the render-loop goroutine entirely, and the `killer` interface in `application/game/service.go` is satisfied structurally by `application/process.Service` with no import between the two `application/*` sub-packages — a clean way to avoid what would otherwise be a cross-package dependency. Pattern-length validation duplication (flagged pre-refactor) is resolved: both `entrypoint/cli` and `application/process` now call the single `core/process.Validate`.

**What's left.**
- D1–D3 (console I/O in `core/score` and both application services, plus no inbound-port validation) are the largest remaining boundary violations. A `Presenter`/`Notifier` outbound port covering pre-game messages, warnings, and the end-of-game summary/table would resolve D1 and D2 in one move and make the application services fully deterministic under test without `testutil/capture`.
- F3 (lost final-frame kills) would be closed by either draining `s.kills` one more time after the loop exits but before `done` is closed, or — more completely — having `runLoop` wait on an explicit `sync.WaitGroup` for in-flight kill goroutines before returning.
- D3's inbound-validation gap and F2/F4's concrete symptoms share one fix: validate `Config` once, at the `application/game.Service.Play` boundary (or in `game.NewSession`), so every current and future driving adapter gets the same guarantees the CLI happens to provide today.
- F1 (help text vs. enforced range) is a one-line documentation fix once someone decides whether `0.1` or `0.5` is the intended floor — right now the error message and the `--help` text disagree with each other.

**2026-09-21 status:** D3 and F1/F2/F4 (third and fourth bullets) are resolved — see `docs/cli-findings.md`. F3 (second bullet) is still open, unchanged. The suggested `Presenter`/`Notifier` outbound port (first bullet) was not built as a single unified port, but `application/score/service.go`'s half of the problem was independently resolved via `console.ScoreReporter` + `apperror.Handle`; `application/process/service.go`'s direct `fmt.Printf` (the other half) is still open.

---

## 6. Positive Observations

1. **The hexagon is real, not cosmetic.** Every `core/*` package's imports were checked directly; none reach outward into `application` or `infrastructure`. `application/game`'s `killer` interface being satisfied structurally by `application/process.Service` — with neither package importing the other — is a clean example of Go interfaces avoiding an import cycle that a naive refactor could easily have introduced.

2. **The async kill path is well-built where it counts.** `drainEvents` (`application/game/service.go:145-172`) moves the `ps`-subprocess-plus-signal call off the 20fps render loop into a goroutine, and every channel send in it (and in `tcellui.poll`, `infrastructure/tcellui/tcellui.go:172-176`) is guarded by a `done`-channel `select`, avoiding goroutine leaks. The one gap (F3) is a genuine edge case, not a wholesale design flaw.

3. **Validation is now unified, not duplicated.** `core/process.Validate` is the single implementation called from both `entrypoint/cli/cli.go:84` and `application/process/service.go:23` (via `validateSearchPatterns`) — the drift risk flagged in the pre-refactor review is gone.

4. **Kill-safety layering is intact after the move.** `PID ≤ 1` refusal now lives at both discovery time (`core/process/process.go:44`) and kill time (`application/process/service.go:50`), name re-verification precedes every `SIGKILL`, and `ps` is resolved to an absolute path once at construction (`infrastructure/osprocess/osprocess.go:22-28`) rather than trusted from `$PATH` at call time.

5. **Ports are narrow and directional.** `application/contract/{inbound,outbound}` splits cleanly into one small file per concern (`input.go`, `process.go`, `score.go`, `ui.go`), each defining only the methods its consumers actually use.

6. **The codebase documents its own known rough edges.** `core/score/board.go:42` and `core/game/target.go:26` both carry TODOs pointing at exactly the design tension this review independently flagged (D1, and the domain's dependence on `process.Info`), which suggests these are known, tracked simplifications rather than oversights.

---

## Summary Table

| ID  | Severity | Category    | Short Description                                                  |
|-----|----------|-------------|----------------------------------------------------------------------|
| F1  | ✓ Resolved | Functional  | `--speed` help text/docs say 0.1–5.0; enforced minimum is 0.5        |
| F2  | ✓ Resolved | Functional  | `--speed=NaN` bypasses range validation                              |
| F3  | moderate | Functional  | Kills completing at/after session end can be lost from the score     |
| F4  | ✓ Resolved | Functional  | `--time` upper bound unchecked; overflows silently to "unlimited"    |
| F5  | ✓ Resolved | Functional  | Failed OS-level kill is silently swallowed, no player feedback       |
| F6  | ✓ Resolved | Functional  | Hit-detection truncates while render position rounds — clicks miss   |
| F7  | ✓ Resolved | Functional  | `LookupName` errors of any kind treated as "process already exited"  |
| F8  | ✓ Resolved | Functional  | Zero-kill sessions still recorded to the score board                 |
| F9  | ✓ Resolved | Functional  | Timer display truncates instead of rounding up                       |
| F10 | ✓ Resolved | Functional  | Unstable sort causes nondeterministic score-list tie-break           |
| F11 | ✓ Resolved | Functional  | Process-name canonicalization mismatch between discovery/kill-time   |
| F12 | Irrelevant | Functional  | `capture` test helpers don't restore streams via `defer` — package removed |
| D1  | ✓ Resolved | Design      | Domain (`core/score`) performs console I/O                           |
| D2  | Half-resolved | Design   | Application services write directly to stdout/stderr (score fixed; process service's match-count print still direct) |
| D3  | ✓ Resolved | Design      | Inbound port performs no validation; only the CLI adapter defends    |
| D4  | ✓ Resolved | Design      | `kills` channel field on `Service` is a latent race if `Service` reused |
| D5  | minor    | Design      | Embedded `Motion` promotes `Move()`, bypassing `Target.Update()` guard |
| D6  | minor    | Design      | Atomic lifecycle in core exists only for app-layer signal concurrency |
| D7  | ✓ Resolved | Design      | Composition root builds `ps`/tcell adapters before arg parsing       |
| D8  | ✓ Resolved | Design      | `UI.Cleanup` panics if called twice                                   |
| S1  | ✓ Resolved | Security | No guard against running as root (non-root case scoped via UID filter; root case now refused by default, `--i-am-root` to override) |
| S2  | moderate | Security    | Residual TOCTOU window in kill flow                                  |
| S3  | ✓ Resolved | Security    | Mouse-drag fires kill events continuously while button held          |
| S4  | low      | Security    | Unbounded concurrent kill-goroutine spawning                         |
| S5  | low      | Security    | SIGKILL only — no graceful shutdown option                           |
| S6  | ✓ Resolved | Security    | Score file written non-atomically                                    |
| S7  | low      | Security    | `ps` output parsing trusts field structure                           |
