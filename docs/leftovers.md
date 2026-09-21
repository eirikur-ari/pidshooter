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

### 2. Residual TOCTOU window between name re-verification and SIGKILL
*Origin: `security-issues.md` SEC-08, `fable-review.md` S2*

`LookupName` (a `ps` subprocess round-trip) and `Kill(pid)` are two separate calls; a PID recycle in between signals the wrong process. Inherent to signal-by-PID on macOS; a Linux `pidfd_open`/`pidfd_send_signal` implementation could close it on that platform only. Accepted residual risk, not currently planned.

### 3. Mouse-drag fires a kill on every motion sample, not just on press
*Origin: `security-issues.md` SEC-09, `fable-review.md` S3*

`translator.go`'s `translateMouseEvent` accepts any event reporting `Button1` down, with no press-vs-motion distinction. Sweeping the cursor with the button held triggers a kill attempt on every target the cursor passes over. Suggested fix: track previous button state, emit only on the press transition.

### 4. Unbounded concurrent kill goroutines
*Origin: `security-issues.md` SEC-10, `fable-review.md` S4*

`application/game/service.go` spawns one goroutine (`ps` subprocess + signal syscall) per accepted kill target, no cap. Combined with #3 above, a single drag can spawn many concurrent `ps` invocations. Suggested fix: worker pool, or an in-flight-per-target dedup flag.

### 5. SIGKILL only — no graceful termination option
*Origin: `security-issues.md` SEC-11, `fable-review.md` S5*

`Kill` always sends `SIGKILL` — no flush, no cleanup handler for the target process. By design; a `SIGTERM`-first mode with escalation would be safer for real workloads.

### 6. `ps` output parsing trusts field structure
*Origin: `security-issues.md` SEC-13, `fable-review.md` S7*

Sanity check tightened from `len(fields) < 3` to `< 5` plus a header-row heuristic since these findings were written, but parsing still trusts positional field structure once that check passes. Exploitability substantially neutralized by the kill-time name re-verification (#2 above), so this is defense-in-depth, not a standalone exploit path.

---

## Functional bugs

### 7. Kills completing at/after session end can be lost from the score
*Origin: `fable-review.md` F3*

`runLoop` calls `frameLoop` then immediately returns and closes `done` via `defer`, with no final drain of `killSignals` and no `sync.WaitGroup` wait for in-flight `killOrReap` goroutines. A kill whose result arrives after the loop's last drain — including one landing in the buffered channel just as the session stops — is never applied: the process **is** killed, but the score never reflects it. Fix: drain once more after the loop exits, or wait on an explicit `WaitGroup` before returning.

### 8. Hit-detection truncates while the rendered position rounds — clicks miss ~50% of the time
*Origin: `code-review.md` #34, `fable-review.md` F6*

`isHitAt` (`core/game/target.go`) truncates with `int()`; `toTargetViewState` (`application/game/converter.go`) rounds with `math.Round`. Whenever a position's fractional part is `≥ 0.5`, the tag is drawn one row/column away from where a click actually registers. Fix: use `math.Round` in both places.

### 9. `LookupName` failures of any kind treated as "process already exited"
*Origin: `fable-review.md` F7*

`process/service.go`'s `Kill` maps *any* `LookupName` error (including a transient `ps` failure) directly to `shouldReap = true`, silently reaping a target that may still be alive — no kill, no score credit, no player feedback.

### 10. Zero-kill sessions are still recorded to the score board
*Origin: `fable-review.md` F8*

`RecordScore` calls `board.Add(entry)` unconditionally regardless of kill count, and the caller (`application/runner.go`) invokes it unconditionally too. Quitting instantly still creates and persists a 0-kill entry, polluting the top-10 table.

### 11. Timer display truncates instead of rounding up
*Origin: `fable-review.md` F9*

`timer.SecondsLeft()` floors (`int(t.Remaining().Seconds())`), so e.g. 0.9s remaining still displays "0s" for up to a full second before the game actually ends.

### 12. Score-list tie-break is nondeterministic (narrowed to a 5-way exact tie)
*Origin: `fable-review.md` F10*

`board.go`'s `sortByRank` uses `sort.Slice`, not `SliceStable`. `Entry.beats` now ranks Kills → Duds → Speed → Duration → FreedMem (a fuller chain than when this finding was written), so two entries now have to match on all five dimensions — including an exact float `Duration` — to hit the ambiguity. Much narrower in practice than originally described, but the underlying non-stable sort is unchanged.

### 13. Score file contents trusted after unmarshal
*Origin: `code-review.md` #36*

`filescore.Load()` handles a missing/oversized/malformed/too-new-schema file, each with a distinct error, but does not re-validate field values, sort order, or entry count after a successful unmarshal. `core/score.NewBoard` also stores whatever it's given as-is — sorting/truncation only happen inside `Add()`, i.e. only after a new score is recorded this session. A hand-edited-but-valid-JSON file with negative values, wrong sort order, or too many entries stays wrong until the next `Add()`, and negative entries are never dropped even then.

---

## Design issues

### 14. `application/process/service.go`'s match-count message bypasses every port
*Origin: `fable-review.md` D2 (half — the `application/score` half is already resolved)*

`fmt.Printf("Found %d process(es) matching %v. Starting game...\n", ...)` is a direct, unrouted print — no `outbound` port involved. Also blocks a planned `--list`/`--dry-run` feature (#22 below), since a list-only invocation would need this message to not say "Starting game...".

### 15. `Target` embeds `movement.Motion`, promoting `Move()` and bypassing the `State`-gated `Update()`
*Origin: `fable-review.md` D5*

Any code holding a `*Target` can call the promoted `target.Move(...)` directly, moving a `Dead` target's position while leaving `State` unchanged — the guard inside `Target.Update()` that's supposed to gate all movement is circumventable.

### 16. `sync/atomic` in core exists only for app-layer signal concurrency
*Origin: `code-review.md` #42, `fable-review.md` D6*

`atomicLifecycle` (`core/game/lifecycle.go`) wraps `atomic.Int32` solely so `application/game/service.go`'s signal-handling goroutine can call `Session.Stop()` concurrently with the loop's own reads. If the application layer instead funneled the signal into the loop's own `select`, core could drop `sync/atomic` entirely.

### 17. `timer.now` clock side effect in core domain
*Origin: `code-review.md` #41*

`timer` reads `time.Now` directly (defaulted lazily in `Start()`/`Remaining()`); tests reach in and overwrite `now`/`start` as unexported same-package fields. Fix: inject the clock explicitly, or make the timer tick-driven instead of wall-clock-driven.

### 18. Renderer view-state type names mix `State`/`ViewState` suffixes inconsistently
*Origin: `code-review.md` #44*

`FrameState`/`HUDState`/`StatusState` use `State`; `TargetViewState`/`ConfirmViewState` use `ViewState`. Purely cosmetic.

### 19. `StatusState.TimeLimit`/`TimeLeft` encode one optional value as two ints
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
