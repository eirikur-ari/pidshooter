# Security Issues Report

**Project:** pidshooter  
**Date:** 2026-08-20 (updated; originally 2026-07-06)  
**Branch:** refactoring-after-ai-creation  

**Re-verified 2026-09-21** against current source (package renames since: `scorefilestore`→`filescore`, `osprocess.go` split into multiple files, `tcellui.go` split into several files including `translator.go`). These findings map directly onto `docs/fable-review.md`'s S1–S7 (re-verified the same day) — SEC-07↔S1, SEC-08↔S2, SEC-09↔S3, SEC-10↔S4, SEC-11↔S5, SEC-12↔S6, SEC-13↔S7 — so the same confirmations are cited here rather than re-derived from scratch.

**Cleanup pass 2026-09-29:** SEC-07 (root guard) and SEC-09 (mouse-drag) fixed — see their entries below. SEC-08, SEC-10, SEC-11, and SEC-13 remain open (SEC-10 not yet addressed; SEC-08/SEC-11/SEC-13 are deliberate accepted-risk / by-design / low-priority calls).

---

## Summary

pidshooter is a terminal game that kills OS processes. Its core attack surface is the ability to send `SIGKILL` to arbitrary PIDs, which makes process-identity validation the most critical security concern. Six issues from the original audit were found and fixed (SEC-01–SEC-06). A follow-up pass (informed by `docs/fable-review.md`, 2026-08-20) against the current package layout — after the service-layer refactor that split `internal/application/service` into `internal/application/{game,process,score}` — found seven additional issues. Two (SEC-07, SEC-09) have since been fixed. The rest (SEC-08, SEC-10, SEC-11, SEC-13) remain open, either as deliberate accepted-risk/by-design calls or not yet addressed. None are critical; the previously-fixed process-identity checks (name re-verification, `PID ≤ 1` guard, minimum pattern length) remain intact.

---

## Issues

### ~~[CRITICAL] SEC-01 — TOCTOU: PID recycling between discovery and kill~~ ✓ FIXED

**File:** `internal/infrastructure/osprocess/osprocess.go:36-45`, `internal/application/runner.go` (`drainEvents`)

Processes are discovered once at game start via `Find()`, and their PIDs are stored in `[]*Target`. When the user clicks a target, the stored PID is sent directly to `Kill()`. A game session can last 30+ seconds (or longer with `--time=0`). During that window, a discovered process may exit and its PID may be recycled by the OS to a completely different — potentially critical — process.

**Example attack path:**
1. User starts `pidshooter node`
2. A `node` process with PID 1234 is discovered
3. That `node` process exits naturally
4. PID 1234 is reassigned to a new critical system daemon
5. User clicks the stale target — the new daemon receives `SIGKILL`

**Fix applied:** The `ProcessKiller` interface requires `Kill(pid int, name string) error`. The application layer (`runner.go`'s `drainEvents`) passes `req.Target.Name` alongside the PID. In `osprocess.go`, `Kill()` runs `ps -p <pid> -o comm=` (with `-c` on macOS to match discovery flags) to get the process's current name, then compares it to the expected name before issuing `SIGKILL`. A mismatch or a dead process returns an error, which causes `drainEvents` to skip `CompleteKill` — so a recycled PID is never killed and no kill animation or score credit is recorded.

---

### ~~[CRITICAL] SEC-02 — PID 0 is not excluded from Kill~~ ✓ FIXED

**File:** `internal/infrastructure/osprocess/osprocess.go` (`filter`, `Kill`)

The filter in `filter()` excludes only the current process PID and PID 1. On Unix, sending a signal to PID 0 sends it to **every process in the current process group**, not to a single process. If a process with PID 0 were ever returned by `ps` parsing (e.g., due to malformed output or a future platform difference), calling `os.FindProcess(0).Signal(SIGKILL)` would kill the entire process group.

**Fix applied:** Changed `filter()` exclusion condition from `p.Pid() == 1` to `p.Pid() <= 1`, and added an explicit guard at the top of `Kill()` that returns an error for any `pid <= 1` (covers PID 0, PID 1, and all negative values) before any syscall is made.

---

### ~~[HIGH] SEC-03 — Insufficient system process exclusion~~ ✓ FIXED (minimum pattern length)

**File:** `internal/infrastructure/osprocess/osprocess.go` (`filter`)

Only PID 1 (init/launchd) and the current process are excluded. Many other low-numbered PIDs are critical kernel threads or system daemons (e.g., PID 2 = `kthreadd` on Linux, launchd helpers on macOS). A broad pattern like `a` would match process names containing the letter "a" — which includes most system processes — and they would all be surfaced as valid kill targets.

There is also no minimum pattern length. A single-character pattern is accepted and can match hundreds of system processes.

**Fix applied:** Added `MinPatternLength = 3` constant to the driven port. `validate()` in `osprocess.go` now rejects any pattern shorter than 3 characters. `parseArgs()` in the CLI layer mirrors the same check for fast, user-facing feedback. The optional broad-match warning was not implemented.

---

### ~~[MEDIUM] SEC-04 — Kill errors silently discarded~~ ✓ FIXED (resolved during SEC-01)

**File:** `internal/application/runner.go` (`drainEvents`)

The kill error was explicitly ignored. The animation and kill count were recorded unconditionally, misrepresenting game state and hiding failures silently.

**Fix applied:** Resolved as part of the SEC-01 fix. `drainEvents` now checks the error from `killer.Kill`; only on `nil` error does it call `g.CompleteKill(target)`, which starts the kill animation and records the kill. On error the target remains `Alive` with no score credit.

---

### ~~[MEDIUM] SEC-05 — Score file and directory have world-readable permissions~~ ✓ FIXED

**File:** `internal/infrastructure/scorefilestore/score_file_store.go`

File was written with `0644` (world-readable) and directory created with `0755` (world-readable/executable). On shared systems this leaks session activity recorded in the score file.

**Fix applied:** Changed `WriteFile` permission to `0600` and `MkdirAll` permission to `0700`.

---

### ~~[LOW] SEC-06 — Outdated transitive dependencies with known vulnerabilities~~ ✓ FIXED

**File:** `go.mod`

The following transitive dependencies (pulled in by `golang.org/x/tools`) were pinned to old versions containing known CVEs:

| Package | Old version | New version |
|---|---|---|
| `golang.org/x/crypto` | `v0.0.0-20210921155107` | `v0.53.0` |
| `golang.org/x/net` | `v0.6.0` | `v0.56.0` |

**Fix applied:** Ran `go get -u golang.org/x/crypto golang.org/x/net` and `go mod tidy`. Also pulled in updated `golang.org/x/sys`, `golang.org/x/term`, and `golang.org/x/text`. The minimum Go version in `go.mod` was bumped from `1.22.0` to `1.25.0` as required by the new `crypto` version.

---

### ~~[MEDIUM] SEC-07 — No safeguard against running as root~~ ✓ FIXED

**File:** `cmd/pidshooter/main.go`, `internal/infrastructure/osprocess/osprocess.go`

There is no `os.Geteuid()` (or equivalent) check anywhere in the composition root or the process adapter. Run as root, every process on the machine — except `PID ≤ 1` and pidshooter's own PID — becomes a one-click `SIGKILL` target in a game whose core mechanic is fast, imprecise clicking (compounded by SEC-11, drag-to-kill).

**Suggested fix:** Refuse to start when `os.Geteuid() == 0`, or require an explicit `--i-am-root` override flag.

**Partially mitigated, confirmed 2026-09-21** — `core/process.IsKillableBy` plus `Find`'s UID filter now scope a non-root player's target list to processes they own, closing the non-root case. The root case this finding is actually about is unchanged: still no `os.Geteuid()` check anywhere.

**Fix applied (2026-09-29):** `core/process.ValidateNotRoot(ownUID int, allowRoot bool) error` rejects `ownUID == 0` unless `allowRoot` is set. `application/process.Service.FindProcesses` calls it first, using the injected `outbound.Process.OwnUID()`, before pattern validation or discovery — wrapped as `CodeInvalidConfig`/`SeverityFatal`, refusing before anything else in the run happens. The override is a CLI-only `--i-am-root` flag, mapped to `config.ProcessRequest.AllowRoot`/`config.ProcessResult.AllowRoot`. It deliberately flows through the same `config.Request`→`config.Result` pipeline `--include-root` uses (for consistency — one path for all process-discovery parameters), but unlike `IncludeRoot` it is never wired to the persisted config file: `config.Result.fromStore` never touches `Process.AllowRoot`, so it always reflects only the current invocation's flags, never a forgotten config file setting. Covered by `TestValidateNotRoot*` (`core/process/info_test.go`), `TestFindProcessesReturnsErrorWhenRunningAsRootWithoutAllowRoot`/`TestFindProcessesRefusesRootBeforeValidatingPatterns`/`TestFindProcessesAllowsRootWithAllowRoot` (`application/process/service_test.go`), `TestServiceRunReturnsErrorWhenRunningAsRootWithoutOverride` (`application/runner/service_test.go`), `TestIntegrationServiceRunAsRootWithOverrideReachesProcessDiscovery` (`application/runner/service_integration_test.go`), `TestResultApplySetsAllowRootFromRequestOnly` (`application/config/result_test.go`), and `TestFlagMapperToRunRequestSetsAllowRootWhenPassed` (`entrypoint/cli/flag_mapper_test.go`).

---

### [MEDIUM] SEC-08 — Residual TOCTOU window between name re-verification and SIGKILL — OPEN (accepted residual risk)

**File:** `internal/application/process/service.go:48-63`, `internal/infrastructure/osprocess/osprocess.go:44-68`

SEC-01's fix (name re-verification before `Kill`) narrowed the race dramatically but did not eliminate it: `LookupName` shells out to `ps` (a subprocess round-trip, milliseconds of latency) and only afterward calls `Kill(pid)` (`os.FindProcess` + `Signal`). If the target PID exits and is recycled by the OS inside that window, the new, unrelated process at that PID is the one that receives `SIGKILL`.

This is inherent to signal-by-PID and not fully closable on macOS. On Linux, `pidfd_open` + `pidfd_send_signal` (opening a stable handle to the process at discovery time, then signaling the handle instead of the PID) would close the window completely for that platform.

**Suggested fix:** Document the residual window in the `outbound.ProcessManager.Kill` contract as a known, accepted risk; consider a Linux-only `pidfd`-based implementation to close it on that platform.

**Still open, confirmed 2026-09-21** — inherent to signal-by-PID; unchanged, as expected. No `pidfd`-based implementation added.

---

### ~~[LOW] SEC-09 — Mouse-drag fires a kill on every motion sample while the button is held~~ ✓ FIXED

**File:** `internal/infrastructure/tcellui/tcellui.go:158-163`

```go
case *tcell.EventMouse:
    if ev.Buttons() != tcell.Button1 {
        continue
    }
    x, y := ev.Position()
    inputEvent = outbound.ClickEvent{X: x, Y: y}
```

This forwards every `EventMouse` where `Button1` is held — including motion events generated while dragging, not just the initial press. Sweeping the cursor across the screen with the button held down produces a `ClickEvent`, and therefore a kill attempt, for every target the cursor passes over in one continuous gesture — a "drag-to-kill" that bypasses the one-target-per-click intent of the game and, combined with SEC-07, meaningfully raises the blast radius of a single misclick.

**Suggested fix:** Track the previous button state and only emit a `ClickEvent` on the press transition (`prev == 0 && ev.Buttons() == Button1`).

**Still open, confirmed 2026-09-21** — `translator.go` (this logic moved out of `tcellui.go` when that file was split) still checks only `mouseEvent.Buttons() != tcell.Button1`, with no press-vs-motion distinction. Unchanged.

**Fix applied (2026-09-29):** `translator` (`infrastructure/tcellui/translator.go`) now tracks `prevButtons tcell.ButtonMask` across calls. `translateMouseEvent` only emits a `ClickEvent` on the press transition (`buttons == tcell.Button1 && prevButtons != tcell.Button1`); a motion sample reporting `Button1` still held from an earlier press is dropped, same as any other non-transition event. Covered by `TestTranslateMouseEventDragOnlyFiresOnPressNotMotion` (`translator_test.go`), which drives a press, a motion sample, a release, and a second press through the same `translator` and asserts only the two presses fire.

---

### [LOW] SEC-10 — Unbounded concurrent kill goroutines — OPEN

**File:** `internal/application/game/service.go:145-172` (`drainEvents`)

Every accepted click spawns a new goroutine that runs `s.killer.Kill(target)` — a `ps` subprocess lookup plus a signal syscall — with no cap on concurrency and no dedup of in-flight requests for the same target. Combined with SEC-09 (drag-to-kill), a single mouse sweep can spawn dozens of concurrent `ps` invocations, a resource-exhaustion vector even if unlikely to be exploited maliciously in this single-player context.

**Suggested fix:** Bound concurrency with a small worker pool, or dedup by tracking an in-flight "killing" flag per target so repeat clicks/drags on the same target are no-ops until the first attempt resolves.

**Still open, confirmed 2026-09-21** — `application/game/service.go` still spawns `go s.killOrReap(target, killSignals, done)` per accepted target with no cap; the `killSignals` channel's buffer (10) bounds queued results, not concurrent goroutines. Unchanged.

---

### [LOW] SEC-11 — SIGKILL only — no graceful termination option — OPEN

**File:** `internal/infrastructure/osprocess/osprocess.go:64` (`Kill`)

`Kill` always sends `SIGKILL` unconditionally — no flush, no cleanup handler, guaranteed data loss for whatever the target process was doing. A less destructive default (`SIGTERM`, escalating to `SIGKILL` after a grace period) would make the tool safer to point at real workloads.

**Suggested fix:** Add a `--sigterm` mode, or default to `SIGTERM` with a timeout-based escalation to `SIGKILL`.

**Still open, confirmed 2026-09-21** — by design; unchanged.

---

### ~~[LOW] SEC-12 — Score file written non-atomically~~ ✓ FIXED

**File:** `internal/infrastructure/scorefilestore/score_file_store.go:67`

`os.WriteFile(s.path, data, 0600)` truncates and rewrites `highscores.json` in place. A crash or power loss mid-write corrupts the file permanently; the existing corrupt-file handling then deliberately skips saving on every subsequent run to preserve the file for recovery (a documented, tested behavior), so a single interrupted write silently disables score persistence going forward. File permissions (`0600` on the file, `0700` on the directory, `score_file_store.go:115`) are already correctly restrictive — this is an atomicity issue, not a permissions one.

**Suggested fix:** Write to a temp file in the same directory and `os.Rename` over the target to make the update atomic.

**Fixed, confirmed 2026-09-21** — `filescore` (renamed from `scorefilestore`) now saves via `fsutil.WriteFileAtomic`, which writes to a temp file and renames it into place, exactly the suggested fix. A crash mid-write can no longer corrupt `highscores.json`.

---

### [LOW] SEC-13 — `ps` output parsing trusts field structure — OPEN

**File:** `internal/infrastructure/osprocess/osprocess.go:70-113` (`list`)

`list()` splits raw `ps` output on lines and whitespace with only a `len(fields) < 3` sanity check before treating `fields[0]` as a PID and the remainder as the process name. A process that names itself with embedded newlines or unusual whitespace could, on a platform where `ps` doesn't escape control characters, produce a malformed or spoofed row. Exploitability is substantially neutralized by the kill-time name re-verification from the SEC-01 fix (a spoofed discovery-time name still has to match at kill time), so this is defense-in-depth rather than a standalone exploit path.

**Suggested fix:** Reject rows with implausible field counts or non-numeric PIDs more defensively; no change needed to the kill-time verification, which already provides the real backstop.

**Still open, minor tightening only, confirmed 2026-09-21** — the sanity check is now `len(fields) < 5` (was `< 3`), plus a header-row detection heuristic, but parsing still trusts positional field structure once that check passes. Exploitability still substantially neutralized by the kill-time name re-verification, same as originally noted.

---

## Summary Table

| ID | Severity | Title | File |
|---|---|---|---|
| SEC-01 | Critical | ~~TOCTOU PID recycling~~ ✓ FIXED | `osprocess.go`, `runner.go` (`drainEvents`) |
| SEC-02 | Critical | ~~PID 0 not excluded from Kill~~ ✓ FIXED | `osprocess.go` |
| SEC-03 | High | ~~Insufficient system process exclusion~~ ✓ FIXED | `osprocess.go` |
| SEC-04 | Medium | ~~Kill errors silently discarded~~ ✓ FIXED | `runner.go` (`drainEvents`) |
| SEC-05 | Medium | ~~Score file world-readable permissions~~ ✓ FIXED | `score_file_store.go` |
| SEC-06 | Low | ~~Outdated transitive dependencies~~ ✓ FIXED | `go.mod` |
| SEC-07 | Medium | ~~No safeguard against running as root~~ ✓ FIXED | `core/process/info.go`, `application/process/service.go`, `application/runner/service.go`, `entrypoint/cli` |
| SEC-08 | Medium | Residual TOCTOU window (name check vs. SIGKILL) — OPEN | `application/process/service.go`, `osprocess.go` |
| SEC-09 | Low | ~~Mouse-drag fires kill events continuously while button held~~ ✓ FIXED | `infrastructure/tcellui/translator.go` |
| SEC-10 | Low | Unbounded concurrent kill-goroutine spawning — OPEN | `application/game/service.go` |
| SEC-11 | Low | SIGKILL only — no graceful shutdown option — OPEN | `osprocess.go` |
| SEC-12 | Low | ~~Score file written non-atomically~~ ✓ FIXED | `filescore/file_score.go` |
| SEC-13 | Low | `ps` output parsing trusts field structure — OPEN (minor tightening) | `osprocess.go` |

## Fix Priority

1. **SEC-02** — trivial one-liner guard; eliminates the process-group kill risk entirely. *(fixed)*
2. **SEC-01** — requires re-validating process name at kill time; most impactful fix. *(fixed)*
3. **SEC-03** — add minimum pattern length in `validate()`; low effort, high benefit. *(fixed)*
4. **SEC-04** — stop recording kills when `Kill()` returns an error. *(fixed)*
5. **SEC-05** — change two permission constants. *(fixed)*
6. **SEC-06** — `go get -u` + `go mod tidy`. *(fixed)*
7. **SEC-07** — root guard; one `os.Geteuid()` check at startup, high benefit for low effort. *(fixed 2026-09-29 — `core/process.ValidateNotRoot` + `--i-am-root` override)*
8. **SEC-09** — press-vs-drag fix for mouse input; low effort, closes the drag-to-kill amplifier. *(fixed 2026-09-29 — press-transition tracking in `translator`)*
9. **SEC-12** — temp-file + rename for score persistence; low effort. *(fixed — `fsutil.WriteFileAtomic`)*
10. **SEC-10** — worker pool or in-flight dedup for kill goroutines.
11. **SEC-11** — optional `SIGTERM`-first mode.
12. **SEC-08** — document as accepted risk now; revisit with a Linux `pidfd` implementation later.
13. **SEC-13** — defense-in-depth parser hardening; lowest priority given the kill-time backstop. *(field-count minimum tightened 3→5 in a later pass; still open in substance)*
