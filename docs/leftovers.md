# pidshooter — Leftovers

Consolidated, deduplicated list of every finding still open across `docs/cli-findings.md`, `docs/code-review.md`, `docs/fable-review.md`, and `docs/security-issues.md`, as re-verified against current source on 2026-09-21. Each entry keeps its origin ID(s) for traceability back to the fuller writeup (evidence, exact line numbers, suggested fixes) in the source document.

**Not included:** `docs/osprocess-findings.md`, `docs/filescore-findings.md`, `docs/tcellui-findings.md` — these weren't part of today's re-verification pass, so their open items aren't reflected here. Check them separately before assuming this list is exhaustive.

---

## Security

### ~~1. No safeguard against running as root — partially mitigated~~ ✓ Addressed 2026-09-21
*Origin: `security-issues.md` SEC-07, `fable-review.md` S1*

No `os.Geteuid()` (or equivalent) check anywhere in the composition root or the process adapter. As root, every process on the machine (except `PID ≤ 1` and pidshooter's own PID) becomes a one-click `SIGKILL` target.

**Mitigated:** `core/process.IsKillableBy` + `Find`'s UID filter now scope a non-root player's target list to processes they own. **Still open:** the root case itself — no `os.Geteuid()` check exists.

**Resolved differently than the original suggested fix (refuse to start as root):** rather than blocking root entirely, `IsKillableBy`/`Find` gained an `includeRoot` parameter, exposed as a new `--include-root` CLI flag threaded through `config.Config`. By default, root-owned processes are excluded from results for non-root players (not just refused at kill time, as before — now not even shown as targets). `--include-root` opts a non-root player in explicitly; running pidshooter itself as root continues to bypass the restriction entirely (`ownUID == 0` in `IsKillableBy`), matching normal Unix permission semantics. This directly addresses the original concern (accidental targeting of root-owned processes) by making it an explicit, deliberate choice rather than an automatic consequence of how pidshooter was launched. Documented in `docs/features.md` and `usageText`. Covered by `TestFindIncludeRootIncludesRootOwnedProcesses`/`TestFindWithoutIncludeRootExcludesRootOwnedProcesses` (`core/process`), `TestFindProcessesIncludeRootIncludesRootOwnedProcesses` (`application/process`), and `TestRunIncludeRootFlag`/`TestRunIncludeRootDefaultsToFalse` (`entrypoint/cli`).

### ~~2. Residual TOCTOU window between name re-verification and SIGKILL~~ ✓ Addressed 2026-09-21
*Origin: `security-issues.md` SEC-08, `fable-review.md` S2*

Already closed on Linux, for free: `process.Pin` calls `os.FindProcess(pid)`, and on Linux 5.3+ Go's stdlib (since Go 1.20; this project targets 1.26.5) transparently opens a pidfd for that call, which the kernel keeps bound to the exact process regardless of PID reuse. No code change was needed there.

macOS has no pidfd equivalent, so `os.FindProcess` falls back to bare-PID signaling there, leaving a narrow window between `LookupName`'s re-verification and `Kill()`'s `SIGKILL`. Documented as an accepted residual risk directly on `osprocess.process.Pin` (`internal/infrastructure/osprocess/process.go`) rather than closed further — the window is syscall-latency-scale and a Darwin-specific mitigation wasn't judged worth the added complexity.

### ~~3. Mouse-drag fires a kill on every motion sample, not just on press~~ ✓ Non-issue, confirmed 2026-09-21
*Origin: `security-issues.md` SEC-09, `fable-review.md` S3*

`tui.go`'s `Init` calls `screen.EnableMouse(tcell.MouseButtonEvents)` — tcell's own doc for that flag is "click events only", as distinct from `MouseDragEvents`/`MouseMotionEvents`. Under this mode the terminal protocol (xterm mode 1000) reports only a press and a release event; no intermediate reports are sent while a button is held and the cursor moves. `translateMouseEvent` already fires a `ClickEvent` only when `Buttons() == Button1` exactly, so exactly one click is produced per press and a held drag produces nothing further before release. The originally-described scenario doesn't reproduce against the mouse mode actually enabled; no code change made.

### ~~4. Unbounded concurrent kill goroutines~~ ✓ Non-issue, confirmed 2026-09-21
*Origin: `security-issues.md` SEC-10, `fable-review.md` S4*

`application/game/service.go`'s `drainEventQueue` spawns one `killOrReap` goroutine per dispatched target, but `game.Target.FireShot`/`shotFired` (`core/game/target.go`) already prevents a target from being dispatched a second time while its kill is in flight — so concurrency is capped at one in-flight kill per live target. With #3 confirmed a non-issue, the only trigger for a target is a discrete click or confirm keypress, so total concurrency is inherently rate-limited by interactive input speed. The suggested worker-pool fix doesn't apply to this code path.

The planned "kill all findings, no selection" CLI mode (see `project_planned_cli_modes` memory) will *not* route through `game.Service` — it bypasses game/target/session logic entirely, calling `process.Service.Kill` directly for every match at once with no interactive pacing. That mode will need its own concurrency design when it's built; it isn't an extension of this code, so it doesn't retroactively make this a live concern today.

### ~~5. SIGKILL only — no graceful termination option~~ ✓ Settled, confirmed 2026-09-21
*Origin: `security-issues.md` SEC-11, `fable-review.md` S5*

`Kill` always sends `SIGKILL` — no flush, no cleanup handler for the target process. Confirmed intentional: pidshooter's premise is a shooting game against processes with immediate, unambiguous kill feedback, not a production-safe process manager. A `SIGTERM`-first escalation mode is out of scope. No code change.

### ~~6. `ps` output parsing trusts field structure~~ ✓ Settled, confirmed 2026-09-21
*Origin: `security-issues.md` SEC-13, `fable-review.md` S7*

`parseProcesses` (`internal/infrastructure/osprocess/process.go`) validates each numeric column (uid/pid/rss) via `strconv`, skipping any row that fails to parse, and `isProcessHeader` confirms the header line's first column really is "uid" before trusting positional order at all. The name field is simply everything after column 4, so it can't desync the earlier fixed columns. Any parse failure degrades to skipping the row (fail-closed), never to misattributing one field's value to another. What theoretical gap remains is fully covered by kill-time name re-verification (#2 above), so this is confirmed defense-in-depth rather than a standalone exploit path. No code change.

---

## Functional bugs

### ~~7. Kills completing at/after session end can be lost from the score~~ ✓ Fixed 2026-09-21
*Origin: `fable-review.md` F3*

`frameLoop` (`internal/application/game/service.go`) now tracks every spawned `killOrReap` goroutine with a `sync.WaitGroup`. Once its main loop exits (whether by the natural `session.IsRunning()` condition or the internal `break`), a new `awaitOutstandingKills` step waits for all outstanding goroutines to finish — concurrently draining `killSignals` as results arrive to avoid deadlocking against the buffered channel — then does one final non-blocking drain to catch anything already buffered. Only after that does `frameLoop` return, so `runLoop`'s deferred `close(done)` can no longer race a still-executing kill into being silently dropped. The early error-return path (`drainEventQueue` failing) intentionally skips this wait, since `Play` discards the whole `PlayResult` on that path anyway.

The wait is bounded by a `Service.killGracePeriod` (default 5s, set in `NewService`): if a `killOrReap` goroutine is still outstanding once the grace period elapses, `awaitOutstandingKills` gives up on it and returns rather than risking session shutdown hanging indefinitely on a wedged `processKiller`. Any outcome arriving after that is dropped, matching pre-fix behavior for that edge case.

Covered by `TestIntegrationServiceFrameLoopWaitsForKillInFlightWhenSessionStops` (fires a click immediately followed by a quit event against a deliberately slow fake killer, asserts `frameLoop` doesn't return until the kill lands and is counted — verified to fail against the pre-fix code before being kept) and `TestIntegrationAwaitOutstandingKillsReturnsOnceGracePeriodElapses` (a permanently wedged `WaitGroup`, asserts `awaitOutstandingKills` still returns once the configured grace period elapses), both in `internal/application/game/service_integration_test.go`.

### ~~8. Hit-detection truncates while the rendered position rounds — clicks miss ~50% of the time~~ ✓ Fixed 2026-09-21
*Origin: `code-review.md` #34, `fable-review.md` F6*

`isHitAt` (`internal/core/game/target.go`) now rounds `Position.X`/`Position.Y` with `math.Round` before computing the hit box, matching `toTargetViewState` (`application/game/converter.go`), which already rounded the rendered position. Previously `isHitAt` truncated with `int()`, so whenever a position's fractional part was `≥ 0.5` the tag was drawn one row/column away from where a click actually registered.

Covered by `TestTargetHitAtMatchesRoundedRenderPositionNotTruncated` (`internal/core/game/target_test.go`), using a position of `(10.6, 5.6)`: asserts a click at the rounded cell `(11, 6)` — where the tag is actually rendered — hits, and clicks at the truncated column/row (`10, 6` and `11, 5`) miss. Verified to fail against the pre-fix truncating code before being kept.

### ~~9. `LookupName` failures of any kind treated as "process already exited"~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` F7*

`process/service.go`'s `Kill` now distinguishes the two cases `LookupName` can fail with: a `outbound.NotFoundError` (the process is actually gone) still maps to `shouldReap = true`, but any other error (a transient `ps` failure, a timeout, a parse error) now returns `shouldReap = false` alongside a `CodeProcessDiscoveryFailed` error — so the caller sees a failed kill attempt instead of a silent, uncredited reap. `outbound.Process.LookupName`'s port doc comment was updated to state this as an explicit contract (not-found is signaled via `NotFoundError`, not just left to the one current implementation's behavior), since `Kill` now relies on it via `errors.As`.

Regression tests: `TestKillReturnsErrorWithoutReapingIfProcessLookupByNameFailsTransiently` (renamed from the old test that asserted the opposite, now asserts `shouldReap = false` for a plain `errors.New` lookup failure) and the new `TestKillReturnsShouldReapIfProcessLookupByNameReportsProcessNotFound` (asserts `shouldReap = true` for a `outbound.NotFoundError`), both in `internal/application/process/service_test.go`.

### ~~10. Zero-kill sessions are still recorded to the score board~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` F8*

Fixed at the domain level rather than in the application-layer caller: `core/score.Board.Add` now silently ignores an entry with `Kills <= 0` — a zero-kill session is not a score. This keeps the invariant in one place: `application/score.Service.RecordScore` and its `mergeWithLatest` helper both route through `Board.Add` already, so quitting instantly no longer appends anything to the in-memory board or the persisted one, without either call site needing its own guard. `Score.Service.RecordScore` still calls `s.store.Save` in this case (an unconditional-looking write when `err == nil`), but since the board it saves is unchanged, this is a no-op write, not a behavioral gap.

Regression tests: `TestBoardAddIgnoresZeroKillEntry`/`TestBoardAddIgnoresNegativeKillEntry` (`internal/core/score/board_test.go`), `TestRecordScoreDoesNotAddZeroKillEntryToBoardOrSave` (`internal/application/score/service_test.go`), and the pre-existing `TestIntegrationRunnerRunIsSuccessful` (`internal/application/runner_integration_test.go`) was updated — it exercised exactly this scenario (an immediate quit) and previously asserted the buggy behavior (a persisted 0-kill entry); it now asserts the board stays empty.

### ~~11. Timer display truncates instead of rounding up~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` F9*

`timer.SecondsLeft()` (`internal/core/game/timer.go`) now uses `math.Ceil` instead of truncating: `int(math.Ceil(t.Remaining().Seconds()))`. A session with 0.9s left now displays "1s" instead of "0s", and "0s" is only ever shown once `Remaining()` actually hits zero (`math.Ceil(0) == 0`), so the unlimited case (`Remaining()` always 0) is unaffected.

Regression tests in `internal/core/game/timer_test.go`: `TestTimerSecondsLeftRoundsUpPartialSecond` (0.9s remaining → `1`, verified to fail against the pre-fix truncating code), `TestTimerSecondsLeftExactWholeSecondIsUnaffected` (an exact 3s remaining stays `3`, not bumped to `4`), and `TestTimerSecondsLeftZeroWhenExpired` (an already-expired timer still reports `0`).

### ~~12. Score-list tie-break is nondeterministic (narrowed to a 5-way exact tie)~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` F10*

`board.go`'s `sortByRank` now uses `sort.SliceStable` instead of `sort.Slice`, so two entries tying on all five of `Entry.beats`'s ranked dimensions now deterministically keep their pre-sort relative order — in practice, the entry already on the board outranks one just appended by `Add`, rather than which one survives being unspecified.

Caveat on the regression test: `TestBoardAddFullTieKeepsInsertionOrder` (`internal/core/score/board_test.go`) asserts the now-guaranteed contract (insertion order preserved across a full 5-way tie), but this specific input pattern — every tied entry identical except an unranked `Date` marker, added one at a time via `Add` — was checked empirically and does **not** reproduce disorder against the pre-fix `sort.Slice` for slice lengths up to 30; Go's current sort implementation happens to preserve order for this exact shape of input. The bug being fixed was a documented API contract gap (`sort.Slice` explicitly does not guarantee tie order), not a case with a known concrete repro, and `Board.Add`'s own `maxScores` cap (10) keeps every real call site's slice length inside the range checked. `sort.SliceStable` closes the gap outright regardless, at no measurable cost given these slice sizes.

### ~~13. Score file contents trusted after unmarshal~~ ✓ Fixed 2026-09-22
*Origin: `code-review.md` #36*

Fixed at the domain boundary rather than in the infrastructure adapter (the suggested fix targeted `Store.Load`, but `filescore` is one of several planned score-store backends per `project_filescore_refactor_planned` memory, and `core/score.NewBoard` is the single choke point every backend's loaded entries already pass through — normalizing there covers all of them, present and future, for free): `Entry` gained an `isScore() bool` predicate (`Kills > 0 && FreedMem >= 0` — a session with no kills isn't a score, and freed memory can never legitimately be negative from a real session). `Board.Add`'s existing `Kills <= 0` guard was widened to use it, and `NewBoard` now builds its board by routing every persisted entry through `Add` instead of assigning `Scores` directly — so a hand-edited-but-valid JSON file with negative `Kills`/`FreedMem`, wrong sort order, or more than `maxScores` entries gets filtered, ranked, and capped right at load time, not left wrong until the next session's `Add` call.

Regression tests: `TestEntryIsScoreFalseForZeroKills`/`ForNegativeKills`/`ForNegativeFreedMem` (`entry_test.go`), `TestBoardAddIgnoresNegativeFreedMemEntry` (`board_test.go`, extending the existing zero/negative-kills coverage), and `TestNewBoardDropsEntriesThatAreNotGenuineScores`/`TestNewBoardSortsAndCapsOversizedUnsortedInput` (`board_test.go`) — the latter builds a 15-entry, ascending (i.e. wrongly ordered), oversized input and asserts the loaded board is ranked descending and capped at 10, verified to fail against the pre-fix `NewBoard` (which stored entries as-is) before confirming it passes against the fix.

---

## Design issues

### ~~14. `application/process/service.go`'s match-count message bypasses every port~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` D2 (half — the `application/score` half is already resolved)*

Fixed with a real port, mirroring the already-resolved `ScoreReporter` half of this same finding — considered and rejected routing this through `util.Logger` (which `apperror` already calls directly, unported, for `Warn`/`Error`) instead: that pattern is a deliberate, accepted exception to the port-boundary rule specifically because diagnostic output is cross-cutting and nobody swaps or asserts on it, whereas this message — like the score summary — is core, business-meaningful program output a test can reasonably want to assert on. New `outbound.ProcessReporter` port (`Report(count int, patterns []string)`), implemented by `console.ProcessReporter` (same shape as `console.ScoreReporter`: an injectable `io.Writer`, defaulting to `os.Stdout`), wired in via the composition root.

The call site landed back in its original location: `process.Service` holds the `outbound.ProcessReporter` directly (`NewService(proc, reporter)`) and `FindProcesses` calls `s.reporter.Report(len(matches), patterns)` itself, once discovery and validation succeed. An intermediate version of this fix moved the call up into `Runner` instead, on the reasoning that "Starting game..." isn't really `process.Service`'s business to assert — but that phrasing is entirely the console adapter's own choice of words behind `Report(count, patterns)`; `process.Service` itself only ever asserts a match count, so routing it back through `process.Service` (matching how `score.Service` and `game.Service` each own their own ports) was simpler and just as correct, once that was pointed out. `Runner`'s constructor reflects this: `NewRunner(processSvc, scoreSvc, gameSvc)`, three already-assembled collaborators, no loose port passed through `Runner` at all.

Regression tests: `TestProcessReporterReportPrintsCountAndPatterns` (`internal/infrastructure/console/process_reporter_test.go`), and `TestFindProcessesReportsMatchCountAndPatterns`/`TestFindProcessesDoesNotReportWhenDiscoveryFails`/`TestFindProcessesDoesNotReportWhenNoProcessIsFound` (`internal/application/process/service_test.go`).

### ~~15. `Target` embeds `movement.Motion`, promoting `Move()` and bypassing the `State`-gated `Update()`~~ ✓ Fixed 2026-09-22
*Origin: `fable-review.md` D5*

An intermediate version of this fix shadowed the promoted method instead (`Target` declaring its own exported `Move` with the same name, which Go always prefers over one promoted from an embedded field) — this worked, but the owner rejected it as too fragile: it silently depends on `Target.Move` and `Motion.Move` staying spelled identically forever, with no compiler error if that ever drifts (e.g. a rename that "forgets" the shadow relationship), and requires a reader to already know about Go's promotion-shadowing rule to see why the method exists at all.

Fixed instead by removing the embedding outright: `Target.Info` and `Target.Motion` are now ordinary named fields (`Info process.Info`, `Motion movement.Motion`), not anonymous/embedded ones. This is a one-token change per field (adding the explicit name in front of the type), but it's what actually removes Go's method/field promotion — with a named field, `target.Move(...)`, `target.PID`, `target.Position`, etc. simply don't exist as expressions at all; the compiler rejects them, rather than a convention keeping them safe. Every existing call site that relied on promotion needed the same one-token fix — inserting `.Info.` or `.Motion.` — since struct literals (`Target{Info: ..., Motion: ...}`) use the field name either way and were unaffected, only *reads* of promoted fields/methods needed updating: `internal/application/game/{converter,kill_tracker,service}.go` (production) and five test files across `core/game`, `application/game`, and `application/input`. The private `move(bounds, speed)` helper (called only from `Update`'s `Alive` case) keeps its own `State != Alive` no-op guard as defense in depth, consistent with this codebase's general preference for guarding even currently-unreachable misuse.

Regression tests in `internal/core/game/target_test.go`: `TestTargetMoveNoOpWhenKilling`/`WhenFleeing`/`WhenDead` (each constructs a target in that state, calls the private `move` directly, asserts `Motion.Position` unchanged) and `TestTargetMoveAdvancesPositionWhenAlive` (control case). Verified by temporarily removing `move`'s internal guard that these tests fail exactly as expected before confirming they pass against the real code.

### ~~16. `sync/atomic` in core exists only for app-layer signal concurrency~~ ✓ Fixed 2026-09-22
*Origin: `code-review.md` #42, `fable-review.md` D6*

The premise of the app-layer half was already stale by the time this pass reached it: `application/game/service.go`'s `frameLoop` doesn't have a separate signal-handling goroutine calling `Session.Stop()` concurrently — `registerTermSignalWatcher` just returns the `<-chan struct{}` from `signal.NotifyContext(...).Done()`, and `frameLoop`'s own `select` (inside its own single goroutine, the same one that calls `session.IsRunning()`/`Update()`) reads that channel and calls `session.Stop()` itself, synchronously. That's exactly the "funnel the signal into the loop's own select" fix this finding's own suggestion described — it must have landed in an earlier, unrelated refactor, and neither `code-review.md` nor `fable-review.md`'s re-verification caught that it made `atomicLifecycle` redundant. Confirmed by tracing every call site: `Session.Stop`/`IsRunning`/`Update` are only ever invoked from `frameLoop`'s single goroutine; the two goroutines this service does spawn (`killOrReap`, and `awaitOutstandingKills`'s internal `waitGroup.Wait()` watcher) never touch `Session` at all.

With no concurrent access left to guard against, `core/game/lifecycle.go`'s `atomicLifecycle` (`atomic.Int32` wrapper) was replaced with a plain `lifecycle` field on `Session`, dropping `sync/atomic` from `core/game` entirely. `lifecycle_test.go` (which only tested the now-deleted wrapper's `Store`/`Load`) was removed — `session_test.go` already independently covers the `pending`/`running`/`stopped` transitions through `Session`'s own public API.

Verified with `make test-race` both before and after the change: clean either way, confirming there was no latent race being masked and none introduced.

Follow-up cleanup: `lifecycle.go` no longer had a reason to exist as its own file once the wrapper was gone — the `lifecycle` type and its `pending`/`running`/`stopped` consts were only ever used by `Session`. Moved both into `session.go` (deleting `lifecycle.go`), placed at the very top of the file with the type immediately followed by its const block — matching `core/game/target.go`'s existing `State` enum, which uses the same adjacent type-then-const shape rather than splitting them across the file's separate const/type sections. `Config` and `Session` follow.

### ~~17. `timer.now` clock side effect in core domain~~ ✓ Fixed 2026-09-22
*Origin: `code-review.md` #41*

Took the "inject the clock explicitly" option over the tick-driven redesign — the latter would have meant changing `Session.Update`'s public signature to take a time delta and moving wall-clock reads into `application/game/service.go`'s frame loop, a much larger and riskier change for a finding that's really about one constructor. `newTimer(limitSeconds int, now func() time.Time) timer` now takes the clock as a required parameter instead of lazily defaulting it to `time.Now` inside `Start()`/`Remaining()` the first time either was called — both methods now just call `t.now()` directly, since a `nil` clock is no longer a reachable state once construction requires one. `Session.NewSession` (the one real production call site) passes `time.Now` explicitly; nothing about its own public signature changed.

This closes the finding's actual complaint — tests no longer reach in and overwrite `now`/`start` as unexported fields post-construction. A small `fakeClock` helper (`t time.Time`, `now()`, `advance(d)`) in `timer_test.go` gives tests a controllable clock supplied at construction time via `newTimer(limit, clock.now)`, and `Start()` is called normally to set `start` — no test anywhere sets `timer.start` or `timer.now` directly anymore. `core/game/session_test.go`'s `TestUpdateStopsWhenTimeLimitExpired` was rewritten the same way, constructing a `Session` with a pre-built `timer` via a raw struct literal (matching the existing `TestUpdateStopsWhenAllTargetsDead` pattern already in that file) instead of calling `NewSession` and then reaching into `s.timer.start`.

One regression surfaced by removing the lazy-default fallback: `TestUpdateStopsWhenAllTargetsDead` constructed a `Session` via raw literal without ever setting `timer`, relying on `Start()`'s old nil-safe defaulting to avoid a nil-func-call panic on the zero-value `timer.now`. Caught immediately by `make test` (a real panic, not a subtle miss) and fixed by giving that literal an explicit `timer: newTimer(0, time.Now)` too. Verified with `make test-race` that no concurrency assumption was disturbed.

### ~~18. Renderer view-state type names mix `State`/`ViewState` suffixes inconsistently~~ ✓ Fixed 2026-09-22
*Origin: `code-review.md` #44*

Converged on `ViewState` (not the majority `State`): `core/game` already has its own domain `State` type (`Alive`/`Killing`/`Fleeing`/`Dead`), and bare `*State` on these outbound view models risked reading as a reference to that unrelated domain concept, whereas `ViewState` is unambiguous about being a presentation-layer value. Renamed `outbound.FrameState`→`FrameViewState`, `HUDState`→`HUDViewState`, `StatusState`→`StatusViewState` (`TargetViewState`/`ConfirmViewState` were already correctly named). Followed through to the sibling private converter functions in `application/game/converter.go` that build these values (`toFrameState`→`toFrameViewState`, `toHUDState`→`toHUDViewState`, `toStatusState`→`toStatusViewState`), which sit right next to already-consistently-named `toTargetViewState`/`toConfirmViewState` — and to the test names exercising them, so nothing in the codebase still says bare `*State` for a rendering value. Purely mechanical rename; no behavior change. `go build`/`go vet`/`make test`/`gofmt` all clean across every touched file.

### 19. `StatusViewState.TimeLimit`/`TimeLeft` encode one optional value as two ints
*Origin: `code-review.md` #45*

Two fields plus a "meaningful only when `TimeLimit > 0`" precondition, sitting right next to `Confirming *ConfirmViewState`'s nilable-pointer idiom in the same struct. A single `TimeLeft *int` (nil = untimed) would be more idiomatic and drop the precondition.

### 20. `Renderer` Init/Cleanup ordering contract exists only because the service owns both calls
*Origin: `code-review.md` #46*

`application/game/service.go`'s `runLoop` calls `Init()`/defers `Cleanup()` itself, deep in the application layer — that's the only reason the ordering contract needs documenting on the port at all. If the composition root owned that lifecycle instead (`Init` before constructing the service, `Cleanup` after it returns), the ordering question would disappear rather than needing to be documented.

---

## Docs / deferred feature work

### 21. `docs/hexagonal-arc.md` is substantially stale
*Origin: `cli-findings.md` Finding 14*

Still describes `cobra`/`buildCommand()` (removed), `scorefilestore`/`stderrlog` (renamed), `Config` living in the `inbound` package (moved), a `Lister`/`Killer` port split (superseded), a nonexistent `testutil/capture` package, and a `main.go` code sample that no longer matches. Deliberately deferred — a full architecture-doc rewrite is a separate, substantial task, not a quick fix.

### 22. `--list`/`--dry-run` preview flag — deferred, part of a larger planned CLI feature set
*Origin: `cli-findings.md` Finding 16*

Pattern matching is undocumented case-insensitive substring matching, and there's no way to preview what a pattern will match before entering the game screen where a click is fatal. Deliberately not built standalone: the owner has a planned 3-mode CLI feature set (list+pick-and-kill, kill-all-without-selection, dry-run preview — see the `project_planned_cli_modes` memory) that this belongs to. Don't implement piecemeal without confirming the overall design first.

---

## Explicitly not carried forward (settled, not leftover work)

These were reviewed today and judged as intentional decisions rather than open TODOs — listed here so they aren't accidentally re-raised, not because anything needs doing:

- **`--speed`/`--time` accepting Go numeric-literal syntax** (octal/hex/binary/underscores) — inherited stdlib `flag` behavior, harmless, documented as expecting plain decimal rather than code-changed (`cli-findings.md` Finding 12).
- **`util.NewLogger()` constructed independently at multiple call sites** — harmless given `Logger` is stateless; design note only, no action recommended (`cli-findings.md` Finding 17).
- **`splitArgs`/`flagNameAndValue` as free functions** — judged a non-issue by two independent reviewers, matches this codebase's established norm for small stateless string-processing helpers (`cli-findings.md` Finding 18).
