# osprocess.go — Cross-Referenced Findings (Fable × Opus)

> **✅ macOS VERIFICATION COMPLETE — 2026-09-07**
>
> The original findings below were produced on a **Linux** machine:
> ```
> Linux fafnir 7.1.13-200.fc44.x86_64 #1 SMP PREEMPT_DYNAMIC Wed Sep 2 13:58:38 UTC 2026 x86_64 GNU/Linux
> ```
> This document has since been independently verified on real **macOS**
> hardware:
> ```
> Darwin Sleipnir.home.local 25.6.0 Darwin Kernel Version 25.6.0: Fri Jul 31 19:17:26 PDT 2026; root:xnu-12377.161.14~5/RELEASE_ARM64_T6041 arm64
> macOS 26.6.2 (25G83), Apple BSD ps, Go 1.26.5 darwin/arm64
> ```
> using the same two-model cross-review method as the original pass: one
> independent verification run on Claude Fable 5, one on Claude Opus, each
> executing the "macOS verification checklist" (Section 6) for real on this
> machine — building test processes, diffing `ps` invocations, reading Go's
> darwin-specific `os` package source — then reconciled by a third pass on
> Claude Sonnet 5 (orchestrating), which re-ran the one item where the two
> independent passes disagreed (see Section 6) to settle it empirically
> rather than by argument.
>
> **Bottom line for macOS:** Finding 1's *mechanism* (GNU `ps` truncating by
> `$COLUMNS` even when piped) does **not** reproduce on macOS — BSD `ps`
> here only truncates by width when attached to a real tty, and pidshooter's
> actual runtime always invokes `ps` via a pipe (`exec.Command(...).Output()`),
> so this specific bug cannot fire in production on this platform. However,
> the two `ps` invocations **can** still disagree on macOS, through a
> different, already-documented path: Finding 7's whitespace-canonicalization
> mismatch — and it is *more* exploitable here, because macOS's `ps -c comm=`
> reports the process's own `argv[0]` (unbounded length, process-controlled,
> can contain arbitrary embedded whitespace or be entirely spoofed) rather
> than the kernel-owned, length-capped `comm` field Linux exposes. See the
> updated Finding 1 and Finding 4 sections below for detail. Findings 2 and
> 10 are confirmed identical to Linux (shared Go source, not GOOS-specific).
> Finding 11 (pidfd leak) does not apply to Darwin at all.

**Date:** 2026-09-07
**Method:** Two independent from-scratch reviews of `internal/infrastructure/osprocess/osprocess.go`, one on Claude Fable 5, one on Claude Opus, each with full repo context (hexagonal architecture doc, the `outbound.ProcessManager` port, the sole caller `application/process.Service`, existing tests, and the prior full-codebase review in `docs/fable-review.md`). Findings below are cross-referenced, deduped, and — where practical — independently re-verified by a third pass (Sonnet, orchestrating) on the same Linux machine.
**Scope:** `internal/infrastructure/osprocess/osprocess.go` only.
**Supersedes:** nothing — this is additive to `docs/fable-review.md`'s F11/S2/S7, which already touched this file; this document supersedes those three items' detail (see "Relationship to prior review" below).

---

## 1. Architecture conformance — both models agree, no issues

Both reviews independently concluded `osprocess.go` fully honors the hexagonal boundary and the "dumb adapter" contract documented on `outbound.ProcessManager`:

- No reverse imports into `core/*` or other `application/*` packages — the only non-stdlib import is `application/contract/outbound`, used solely for the `ProcessManager`/`ProcessInfo` types this file structurally satisfies.
- Zero embedded business/safety logic: no `IsProtected` check, no name-match verification, no pattern matching anywhere in this file. That responsibility correctly and entirely lives in `application/process.Service.Kill` and `core/process.Find`/`ValidateProcesses`.
- Stateless after construction (`psPath` is the only field, set once in `NewProcess`) — safe under the game loop's concurrent `killOrReap` goroutines by design, not by accident.

**Verdict: no boundary violations, no changes needed on the architecture axis.** The issues below are about *fidelity* of the OS→domain translation, not about placement or coupling.

---

## 2. Critical — verified bug (Linux-confirmed; macOS-verified, different mechanism)

### Finding 1: `List()` and `LookupName()` use different `ps` invocations that truncate names differently, breaking the kill-time safety check — ✅ FIXED 2026-09-07

**Evidence:**
- `list()` — `osprocess.go:70-75` — `ps -eo pid,rss,comm` (three columns).
- `LookupName()` — `osprocess.go:44-48` — `ps -p <pid> -o comm=` (one column, no header).
- The doc comment on `LookupName` (`osprocess.go:42-43`) claims: *"using the same ps flags as List so truncation and format are consistent."* **This claim is false on Linux** — the two invocations have different column sets and therefore different width budgets under GNU `ps`'s terminal-width-aware truncation.

**Reproduced independently, twice, on the Linux machine above** (once by the Opus review, once by a follow-up verification pass):

```
$ cd /tmp && ln -sf $(which sleep) reallylongprocessnametotest12345
$ ./reallylongprocessnametotest12345 30 &
$ P=$!
$ for c in 20 24 28 40; do
    echo "COLUMNS=$c:"
    echo "  List-style:       $(COLUMNS=$c ps -eo pid,rss,comm | awk -v p=$P '$1==p{$1="";$2="";print}')"
    echo "  LookupName-style: $(COLUMNS=$c ps -p $P -o comm=)"
  done
```

```
COLUMNS=20   List: "really"              LookupName: "reallylongproce"
COLUMNS=24   List: "reallylong"          LookupName: "reallylongproce"
COLUMNS=28   List: "reallylongproc"      LookupName: "reallylongproce"
COLUMNS=40   List: "reallylongproce"     LookupName: "reallylongproce"   (only value where they agree)
```

GNU `ps` truncates to `$COLUMNS` **even when stdout is piped**, and `List`'s extra `pid`/`rss` columns eat into the width budget in a way the single-column `LookupName` invocation doesn't. They only happen to agree once the terminal/COLUMNS budget is wide enough that neither truncates below the kernel's own 15-character `comm` cap (see Finding 4).

**Full failure path, traced against real caller code:**
1. Discovery (`list()`) stores a truncated `Name`, e.g. `"reallylong"`.
2. Player clicks the target; `application/process/service.go:53` calls `process.ValidateName(target.Name, name)` where `name` comes from `LookupName`, e.g. `"reallylongproce"`.
3. Mismatch → `service.go` returns `(false, true, err)` — `shouldReap = true`.
4. `application/game/service.go` checks `shouldReap` **before** looking at the error, so the mismatch error is discarded.
5. The target is silently reaped (`target.Reap()`, `core/game/target.go`) — `State = Dead`, no kill credit, no failure shown.

**Net effect: the target vanishes on click, the score reflects nothing, no error is ever surfaced to the player, and the real OS process is never signaled at all.** This can turn into a silent no-op game whenever `$COLUMNS` is set narrow enough — which is common in tmux panes, some CI harnesses, and any shell config that exports `COLUMNS` explicitly.

**Status: CONFIRMED on Linux via the `$COLUMNS` mechanism.**

**macOS status (verified 2026-09-07): the `$COLUMNS` mechanism does NOT reproduce when piped.** Empirically, on this machine, `env COLUMNS=20 ps -ceo pid,rss,comm` and `env COLUMNS=20 ps -c -p <pid> -o comm=` both returned the full, untruncated 40-character test name — at every `$COLUMNS` value tried (unset/20/40/80), piped or redirected to a file. BSD `ps` on this machine ignores `$COLUMNS` entirely whenever stdout is not a tty. Since `osprocess.go` always invokes `ps` via `exec.Command(...).Output()` — always a pipe, never a tty — **this specific bug cannot occur in production on macOS.**

Under a *real* pty (verified via `script -q /dev/null env COLUMNS=N ps ...`), the two invocations do diverge, and worse than on Linux — `list()`'s effective width budget is `COLUMNS - 13` (eaten by the `pid`/`rss` columns) while `LookupName`'s budget is the full `COLUMNS`:
```
COLUMNS=20  list()="reallyl" (7 chars)             LookupName="reallylongprocessnam" (20 chars)
COLUMNS=40  list()="reallylongprocessnametotest" (28)  LookupName=<full 40-char name>
COLUMNS=80  list()=<full 40-char name>             LookupName=<full 40-char name>
```
This is real but operationally irrelevant to the shipped binary (no tty is ever attached to these subprocess calls); it matters only if this code is ever exercised interactively (e.g. a developer running the underlying `ps` commands by hand while debugging).

**The two paths can still disagree on macOS, but via Finding 7, not Finding 1's mechanism.** macOS's `ps -c ... comm=` derives its output from the process's `argv[0]` (see Finding 4 below), which can contain runs of internal whitespace. `list()` collapses these via `strings.Fields`+`strings.Join`; `LookupName` only does `strings.TrimSpace`, preserving internal runs. Reproduced directly: a process launched with `argv[0] = "weird  double   spaces"` (double/triple internal spaces) is reported identically by both raw `ps` invocations, but after `osprocess.go`'s own parsing, `list()` would normalize it to `"weird double spaces"` while `LookupName` returns `"weird  double   spaces"` verbatim — a mismatch that reproduces the exact same silent-reap failure path described above for Linux, just triggered differently.

**Net risk to macOS players: same failure mode (silent reap, no error surfaced, target vanishes), reached via Finding 7's whitespace path rather than Finding 1's `$COLUMNS` path.**

**Recommended fix — deliberately platform-agnostic:** Don't chase `-ww`/width flags (that's GNU-`ps`-specific syntax and is not confirmed to have equivalent semantics under BSD `ps`). Instead, make `LookupName` issue **the same command shape** as `list()` — same flags, same column order, same darwin/non-darwin branching — just filtered to one `pid`, e.g. reuse `list()`'s underlying invocation and grep/filter by PID in Go, rather than maintaining a second, differently-shaped `ps` call. Whatever truncation policy the platform's `ps` applies then applies identically to both code paths *by construction*, without this codebase needing to know or hard-code either OS's width rules. **Verified on macOS (2026-09-07):** using `list()`'s exact darwin flags filtered to one PID (`ps -c -p <pid> -o pid,rss,comm`) produces byte-identical output to `ps -ceo pid,rss,comm` across the full COLUMNS/tty matrix tested. **Important caveat found during macOS verification: unifying the flags alone is not sufficient** — `LookupName` must also adopt `list()`'s *parsing* (`strings.Fields`+`strings.Join`), not just its `ps` invocation, or Finding 7's whitespace mismatch survives the fix untouched on macOS.

**Implemented (2026-09-07):** `list()`'s row-parsing logic was extracted into a shared `parseProcesses([]byte) ([]outbound.ProcessInfo, error)`. `LookupName` now issues `ps [-c] -p <pid> -o pid,rss,comm` — the same columns and darwin branch as `list()` — and parses the result through the same `parseProcesses` function, picking out the row matching `pid`. Both code paths are now structurally incapable of diverging in truncation or whitespace handling, closing this finding and Finding 7 together. Covered by `TestParseProcesses` in `osprocess_test.go`, including the exact whitespace-collapse case reproduced above. Finding 4's `argv[0]`-spoofability risk is untouched by this fix — it's a separate, still-open concern.

---

## 3. Moderate findings (both models, deduped)

### Finding 2: `os.ErrProcessDone` is discarded, producing permanently un-killable "undead" targets — ✅ FIXED 2026-09-07
`osprocess.go:64-67` returns the raw `proc.Signal` error uninspected. Go's unix `findProcess`/`Signal` path maps `ESRCH` (process already gone) to the distinguishable sentinel `os.ErrProcessDone` — the adapter throws this signal away into an opaque `error`. Downstream, `application/process/service.go:57-58` returns `(false, false, err)` on any `Kill` error — note `shouldReap` is hardcoded `false` on this branch, unlike the `LookupName`-mismatch branch — so a target whose process exited in the discovery→kill window becomes stuck `Alive` forever with no way to reap it. Confirmed against Go's `os/exec_unix.go` and `pidfd_linux.go` source (both platforms map `ESRCH`→`ErrProcessDone`; behavior expected to hold on Darwin too since it's in the portable `os` package, not a GOOS-specific file).

**macOS status (verified 2026-09-07): CONFIRMED.** Killing an already-reaped child on Darwin and checking `errors.Is(err, os.ErrProcessDone)` returns `true`, reproduced twice independently (once via a direct kill+reap+re-signal, once racing a background `kill` against a concurrent lookup/kill). Root cause confirmed in source: `$GOROOT/src/os/exec_unix.go`'s `convertESRCH` (maps `syscall.ESRCH → os.ErrProcessDone`) lives in a file gated `//go:build unix`, shared verbatim between Linux and Darwin — this is not GOOS-specific inference, it's the same compiled code on both platforms. One added wrinkle found during verification: the sentinel *replaces* `ESRCH` rather than wrapping it, so checking `errors.Is(err, syscall.ESRCH)` directly returns `false` — code must check `os.ErrProcessDone`. Finding 3(a)'s EPERM "undead" trap (root-owned process, non-root player) was also reproduced identically on Darwin: `errors.Is(err, os.ErrProcessDone)` is `false` and `errors.Is(err, syscall.EPERM)` is `true`, so it hits the same un-reapable branch as on Linux.

**Fix:** in `Kill`, check `errors.Is(err, os.ErrProcessDone)` and let the caller (or this adapter itself) signal "already gone" distinctly from "kill failed" — resolves cleanly alongside Finding 8 (the redundant `bool` return). **macOS: identical fix applies unchanged** — see confirmation above.

**Implemented (2026-09-07):** rather than leaking `os.ErrProcessDone` (a stdlib sentinel) up into the application layer — which would mean `application/process.Service` importing `os`, breaking the hexagonal boundary documented in Section 1 — the adapter now translates it into a port-level sentinel. `Kill(2)`'s `ESRCH` ("No process ... can be found corresponding to that specified by pid") is what Go's `os.ErrProcessDone` actually maps from on this call path; that's a "not found at signal time" fact, not itself a claim about the process's history, so the sentinel is named `outbound.NotFoundError` — reusing (not duplicating) the existing type already used by `ScoreStore.Load`, now lifted into a shared `internal/application/contract/outbound/error.go` so both ports depend on the same generic "resource does not exist" error. `osprocess.go`'s `Kill` checks `errors.Is(err, os.ErrProcessDone)` and returns `outbound.NotFoundError{}` in its place; `application/process.Service.Kill` checks `errors.As(err, &outbound.NotFoundError{})` and returns `shouldReap = true`, exactly like the existing `LookupName`-mismatch branch. `os` never appears in the application layer. Covered by `TestKillReturnsShouldReapWhenProcessAlreadyExited` (`internal/application/process/service_test.go`, fake-based) and a strengthened `TestKillerNonexistentPID` (`osprocess_test.go`, real OS, asserts the sentinel type). **Not fixed by this change:** Finding 3(a)'s `EPERM` "undead" trap — `EPERM` isn't converted to any sentinel by Go, so a root-owned process killed by a non-root player still falls into the generic-error, `shouldReap=false` branch and stays stuck alive. That remains open.

### Finding 3: The port drops process ownership entirely — the domain can't express "don't touch processes I don't own" — ✅ FIXED 2026-09-08
`list()` uses `-eo` (all users). `outbound.ProcessInfo` carries only `PID`/`Name`/`Rss` — no UID. Grep confirms zero references to `Getuid`/`Geteuid`/ownership anywhere in the tree; the only domain guard is `Info.IsProtected()` (`PID <= 1`). Two consequences: (a) as non-root, root-owned processes become unkillable targets that hit the `EPERM`→generic-error path (same "undead" trap as Finding 2, since `EPERM` is *not* converted to a sentinel by Go); (b) as root, `pidshooter sshd` will happily kill every `sshd` on the box — `PID <= 1` is the only guard between the player and system daemons. This is a boundary-fidelity gap, not a leak: cheap for the adapter to surface (`ps -eo uid,pid,rss,comm`), and the domain is structurally unable to build ownership-based policy without it.

**Implemented (2026-09-08):** `outbound.ProcessInfo` and `core/process.Info` both gained a `UID` field; `osprocess.go`'s `list()` and `LookupName()` now request `ps -eo uid,pid,rss,comm` (darwin: `-ceo`/`-c -o`), and `parseProcesses` parses the extra leading column. A new `outbound.ProcessManager.OwnUID()` (backed by `os.Geteuid()`) surfaces the caller's effective UID alongside the existing `OwnPID()`. `core/process.Info` gained `IsKillableBy(ownUID int) bool` — root (`UID 0`) may target any process; everyone else only processes they themselves own — and `Find` now takes `ownUID` and excludes any process that fails this check, so unkillable processes never surface as targets in the first place rather than becoming selectable-but-broken. This resolves consequence (b) directly (root's own kill scope is unchanged — sees and can kill everything, matching real `kill(2)` semantics — but a non-root player's target list is now scoped to processes they can actually terminate) and resolves (a) for the normal case (a root-owned process is simply never offered as a target to a non-root player). The residual TOCTOU race this doesn't close (Finding 6: a PID recycled between discovery and `Kill` to a different-owner process with a coincidentally-matching name) was deliberately left without a dedicated `EPERM` sentinel — `application/process.Service.Kill` doesn't need to distinguish it from any other kill failure (no `shouldReap` branch depends on it, unlike `NotFoundError` in Finding 2), so it falls through the existing generic error path unchanged, where the underlying cause (`"operation not permitted"`) is already visible in the wrapped message. Covered by `TestInfoIsKillableBy` and `TestFind*OwnedByCaller`/`TestFindAsRootIncludes*` (`core/process/process_test.go`), `TestFindProcessesExcludesProcessesNotOwnedByCaller` (`application/process/service_test.go`), and updated `TestParseProcesses`/`TestOwnUIDMatchesOSGeteuid` (`osprocess_test.go`).

### Finding 4: Kernel `comm` truncation to 15 characters is a separate, more fundamental limit than Finding 1 — macOS spoofing gap ✅ FIXED 2026-09-08
Independently of the `COLUMNS` mismatch, Linux truncates `/proc/pid/comm`-derived `comm` values to `TASK_COMM_LEN - 1` = 15 visible characters — this applies **identically** to both `List` and `LookupName` once `COLUMNS` is wide enough not to truncate further (see the `COLUMNS=40` row above, where both already read `"reallylongproce"`, 15 chars). This weakens the kill-time name-recheck (the TOCTOU defense documented in `docs/fable-review.md` S2) for *any* process with a name ≥15 characters, independent of whether Finding 1 is ever fixed.

**macOS status (verified 2026-09-07): REFUTED for the `ps`-based code path this codebase actually uses — but with a worse finding underneath.** Darwin does have a kernel-level analog to `TASK_COMM_LEN`: `MAXCOMLEN = 16`, defined in the live SDK header (`.../MacOSX.sdk/usr/include/sys/param.h:95`), backing the kernel-owned `p_comm`/`pbi_comm` fields (`sys/proc.h:138`, `sys/proc_info.h:72-73`). But `osprocess.go`'s `ps -c ... comm=` does **not** read that field. Confirmed by direct comparison on the same PID: `ps -p <pid> -o ucomm=` (kernel-truthful, reads `p_comm` via `proc_pidinfo`) correctly returned the truncated/padded 16-byte name, while `ps -c -p <pid> -o comm=` (what `osprocess.go` uses) simultaneously returned the full, uncapped 66-character name for the same PID. BSD `ps`'s `-c`/`comm=` output is reconstructed from the process's own `argv[0]` (via `KERN_PROCARGS2`), not from the kernel's capped `p_comm` — so there is no length cap on this code path at all, up to the filesystem's own `NAME_MAX` (255 chars tested successfully; 500 chars failed at exec/`ln` time with "File name too long," a filesystem limit, not a `ps` limit).

**New finding, macOS-only, more severe than the truncation issue it replaces: the kill-time name recheck is spoofable on Darwin in a way it structurally cannot be on Linux.** Because `argv[0]` is fully process-controlled and unbounded, a process can present any name it wants to `LookupName`/`list()` regardless of its actual executable. Reproduced directly: a `sleep` process launched with `argv[0] = "TOTALLY-DIFFERENT-NAME"` is reported by `ps -c -o comm=` (this codebase's invocation) as `TOTALLY-DIFFERENT-NAME`, while the kernel-truthful `ps -o ucomm=` correctly reports `sleep` for the identical PID. This means on macOS, unlike Linux, the verify-then-kill safety property (`docs/fable-review.md` S2, sharpened by Finding 6) can be defeated by any process that controls its own `argv[0]` at exec time — no TOCTOU race or PID reuse required, just a mismatched `argv[0]`. Switching the darwin branch to `ucomm` instead of `-c comm=` would close this spoofing gap, at the cost of reintroducing a Linux-style ~16-character visible cap on macOS names (trading spoofability for truncation — the same tradeoff Linux already lives with).

**Implemented (2026-09-08):** `osprocess.go`'s darwin column now reads `ucomm` instead of `-c comm=`, closing the `argv[0]`-spoofing gap by construction — `ucomm` is the kernel-owned name (`p_comm`), unaffected by whatever a process claims as its own `argv[0]`. `list()` and `LookupName` share a single column-selection helper so the two invocations cannot drift again the way Finding 1 originally did. The `-c` flag itself turned out to be unnecessary once the column changed — confirmed empirically on real macOS hardware that `ucomm` is unaffected by `-c`'s presence — so it was dropped rather than carried forward unused. This intentionally reintroduces the ~16-character truncation Linux already lives with (`MAXCOMLEN = 16`), trading spoofability for truncation, exactly as this finding recommended; Linux's invocation is untouched. Verified against a real spoofed process (`argv[0] = "TOTALLY-DIFFERENT-SPOOFED-NAME"` launched via `sleep`) both manually via raw `ps` and via a new regression test, `TestIntegrationLookupNameResistsArgv0Spoofing` (`osprocess_integration_test.go`), which asserts `LookupName` reports the true binary name regardless of platform.

### Finding 5: No timeout/context on either `ps` subprocess call — ✅ FIXED 2026-09-08
`osprocess.go:49` and `:75` both use bare `exec.Command(...).Output()`. A wedged `ps` (stalled `/proc` reader, hung container runtime) hangs `List` at startup with no feedback, and hangs `LookupName` inside the game loop's `go s.killOrReap(...)` goroutine, which then never reaches its `done`-channel select. **Fix:** `exec.CommandContext` with a short deadline on both calls. This fix is platform-neutral — no macOS-specific concern.

**Implemented (2026-09-08):** both invocations now go through a shared `Process.run(args ...string) ([]byte, error)` helper that wraps the call in `context.WithTimeout(context.Background(), p.timeout)` and uses `exec.CommandContext`, so a wedged `ps` is killed and returns an error instead of hanging its caller indefinitely. `Process` gained a `timeout` field (`NewProcess` sets it to a `defaultPSTimeout` of 5s) rather than a hardcoded constant, specifically so a test can inject a short timeout and verify the enforcement itself deterministically rather than waiting out a real multi-second hang. Covered by `TestListTimesOutWhenPsHangs` and `TestLookupNameTimesOutWhenPsHangs` (`osprocess_test.go`), both of which point `psPath` at a throwaway script that hangs for 10s, inject a 50ms timeout, and assert the call returns an error within a small bounded window rather than the full hang duration.

### Finding 6: `Kill(pid int)` makes the documented verify-then-kill contract structurally unenforceable — ✅ FIXED 2026-09-08 (Linux; Darwin unaffected, see below)
The port's doc comment and `application/process.Service.Kill` both document a verify-then-kill sequence, but `LookupName(pid)` and `Kill(pid)` are two independent calls keyed only on an `int` — nothing pins the OS-level identity across them, so a PID recycled between the two calls is indistinguishable from the original target. Notably, `os.FindProcess` on Linux already opens a pidfd internally (which *does* pin identity) but the adapter discards it at the end of a single `Kill` call rather than letting it span the verify step. Flagged as design-level (the race window is narrow and needs PID wraparound), not a live bug — a handle-returning port shape (`Open(pid) (Handle, error)` → `Handle.Name()`/`Handle.Kill()`) would make the safety property enforceable rather than merely documented. **macOS status (verified 2026-09-07): the identity-pinning gap is structurally *wider* on Darwin than on Linux, not just differently shaped.** Reading `$GOROOT/src/os/exec_unix.go` and `pidfd_other.go` (build tag `unix && !linux`, which includes darwin): `findProcess` on Darwin calls `pidfdFind`, which unconditionally returns `syscall.ENOSYS`, so it always falls through to `newPIDProcess(pid)` — a bare-PID handle with no underlying OS object at all (confirmed empirically: `FindProcess(-1)`, `FindProcess(0)`, and an out-of-range PID all return `err == nil` on this machine). Linux at least opens a real pidfd that pins process identity — the adapter just discards it too early (the original point of this finding). Darwin has nothing to pin in the first place: there is no pidfd equivalent in the stdlib, so a handle-returning port redesign here cannot be a drop-in recompile — it would need a different mechanism entirely on macOS (e.g. re-validating via `proc_pidpath`/`libproc` immediately before signaling, which narrows but cannot close the race the way a real pidfd does), or an explicit acceptance that this particular race is unfixable-by-handle on Darwin.

**Implemented (2026-09-08):** the port gained `outbound.ProcessHandle` (`Kill() (bool, error)`) and `ProcessManager.Open(pid int) (ProcessHandle, error)`, replacing the old free-standing `Kill(pid int)`. `osprocess.go`'s `Open` wraps `os.FindProcess(pid)` in a `processHandle`, moving the actual `SIGKILL` send onto `processHandle.Kill()`. `application/process.Service.Kill` now calls `Open(pid)` *before* `LookupName(pid)` and kills through the returned handle rather than re-resolving `pid` at kill time — so a PID recycled between verification and kill can no longer be silently redirected to a different process; on Linux, `Signal` on the handle's already-open pidfd either hits the original process or cleanly reports `NotFoundError`, never a different one. This is a real, structural fix on Linux and is a no-op on Darwin (nothing to pin there — `newPIDProcess` remains a bare-PID reference either way), consistent with the finding's own analysis; it does not regress Darwin's existing behavior either. Covered by `TestOpenReturnsHandleForExistingPID`/`TestKillerNonexistentPID` (`osprocess_test.go`), `TestIntegrationOpenThenKillReportsNotFoundAfterProcessExits` (`osprocess_integration_test.go`, opens a handle to a real process, waits for it to actually exit, then confirms `Kill` on the stale handle reports `NotFoundError` rather than silently succeeding), and `TestKillReturnsErrorAndShouldReapIfOpenFails` (`application/process/service_test.go`).

---

## 4. Minor / low findings (both models, deduped)

| # | Severity | Finding | Linux-verified? | macOS status |
|---|---|---|---|---|
| 7 | Minor | `comm` whitespace collapsed in `List` (`strings.Fields`+`Join`) but not in `LookupName` (`TrimSpace` only) — extends prior review's F11. Consequence corrected from F11: causes silent **reap**, not silent refusal, per the `shouldReap`-before-`err` check in `application/game/service.go`. | Yes, by direct code reading | **CONFIRMED live on macOS (verified 2026-09-07)**, more exploitable than on Linux since macOS `comm` is unbounded, process-controlled `argv[0]` (see Finding 4). **✅ FIXED 2026-09-07 alongside Finding 1** — `LookupName` now shares `list()`'s `parseProcesses` parsing, so whitespace is collapsed identically on both paths. |
| 8 | Minor | `Kill`'s `bool` return is redundant: `killed ≡ (err == nil)` always, since `os.FindProcess` never fails on Unix and the only remaining path is `Signal`'s error. Makes a defensive branch in `application/game/service.go` (`case !killed: ...`) unreachable with this adapter. | Yes | **Confirmed on macOS (verified 2026-09-07)** — `FindProcess` never returns non-nil error on Darwin (see row 10); reasoning holds unchanged. |
| 9 | Minor | `ps` stderr (up to 32KB, captured in `exec.ExitError.Stderr`) is discarded by plain `%w` wrapping in both `list()` and `LookupName` | Yes | Platform-neutral fix — unaffected by macOS verification. |
| 10 | Minor | `os.FindProcess`'s error branch in `Kill` (`osprocess.go:60-62`) is dead code on Linux — `findProcess` never returns non-nil error on unix per Go source | Yes, confirmed against Go source | **CONFIRMED on Darwin too (verified 2026-09-07)**, for the identical reason, in shared code: `pidfd_other.go` (`//go:build unix && !linux`, includes darwin) makes `pidfdFind` unconditionally return `ENOSYS`; `findProcess` swallows that and always falls back to `newPIDProcess(pid), nil`. Empirically checked with `FindProcess(-1)`, `FindProcess(0)`, and an out-of-range PID — all returned `nil` error. |
| 11 | Minor | pidfd leaked until GC — `proc.Release()` is never called on the `*os.Process` from `FindProcess`; Go's own docs call `Release` required when `Wait` is never invoked | Yes | **REFUTED on Darwin (verified 2026-09-07) — non-issue on this platform, not merely "differently shaped."** Darwin's `findProcess` always returns a bare-PID `*os.Process` (`handle == nil`, no `runtime.AddCleanup` registered); `Release()`/`doRelease()`'s handle-closing branch (`if p.handle != nil`) never executes. There is no fd or kernel handle attached to a PID-only `*os.Process` on Darwin, so skipping `Release()` has zero resource consequence there. |
| 12 | Minor | Header row (`lines[1:]`) skipped positionally with no content validation; parse-failure policy is inconsistent across the three fields (skip row / skip row / default-to-zero) | Yes | Platform-neutral, both branches share this code — unaffected. |
| 13 | Low | `NewProcess` returns the `outbound.ProcessManager` interface instead of `*Process`, forcing `osprocess_test.go` to hand-construct `&Process{psPath: path}` to get a concrete value | Yes | N/A |
| 14 | Low | Doc-drift: `docs/hexagonal-arc.md` documents `OwnPid()`; code has `OwnPID()` | Yes | N/A |
| 15 | Low | Runtime `GOOS` branching instead of build-tagged platform files (`osprocess_linux.go`/`osprocess_darwin.go`) — compiles silently on unsupported platforms (e.g. Windows), reaches "ps not found" by accident rather than design | Yes | **Strengthened by macOS verification** — the platforms now have documented, materially different semantics beyond flag syntax (argv[0]-derived vs. kernel-derived `comm`, no length cap vs. a 15-char cap, no pidfd vs. a real one), which is exactly the kind of divergence build-tagged files would make explicit and testable in isolation, instead of buried in `runtime.GOOS` branches inside shared functions. |
| 16 | Low | SIGKILL to a zombie reports success (`kill(2)` on a zombie returns 0) even though nothing dies | Yes (POSIX-general) | Not independently retested on macOS during this pass; POSIX semantics make divergence unlikely, but this remains an assumption, not a verified result, unlike rows 7–11 above. |

---

## 5. Relationship to `docs/fable-review.md`

- **F11** ("whitespace-canonicalization mismatch") is confirmed still present — see Finding 7 above, which also corrects F11's stated consequence (silent *reap*, not silent *refusal*).
- **S2** ("residual TOCTOU window") is sharpened by Finding 6 above (pidfd is opened and discarded within `Kill`, but never spans the `LookupName`→`Kill` gap).
- **S7** ("ps output parsing trusts field structure") is extended by Findings 1, 4, 9, and 12 — the mitigation S7 cites ("substantially neutralized by kill-time name re-verification") is exactly what Finding 1 shows breaking down when the two `ps` calls disagree.

---

## 6. macOS verification — completed 2026-09-07

**Method:** Two independent from-scratch verification passes, each executing all 7 checklist items live on the same real macOS machine (`Darwin Sleipnir 25.6.0`, macOS 26.6.2, arm64, Apple BSD `ps`, Go 1.26.5) — one on Claude Fable 5, one on Claude Opus, each blind to the other's output — followed by a reconciliation pass on Claude Sonnet 5 (orchestrating). All results above (Sections 2–4) have been updated in place with the reconciled findings; this section records the raw per-item outcomes and, notably, the one point where the two independent passes disagreed and how it was resolved.

**Item-by-item outcome:**

1. **Finding 1 asymmetry** — REVISED. Piped (production-realistic): never diverges, any `$COLUMNS` value, confirmed by both passes and by the Sonnet tie-break re-test below. Under a real pty: does diverge, confirmed by the Sonnet re-test (see disagreement below) — operationally irrelevant since production never attaches a tty.
2. **BSD `ps` truncation trigger** — `$COLUMNS` is ignored entirely when stdout is not a tty (all three passes agree); `-c` selects `argv[0]`-basename vs. full command line and imposes no length limit of its own (all three passes agree).
3. **Kernel-level name cap** — `MAXCOMLEN = 16` exists (`sys/param.h`) but is not what `ps -c comm=` reports; that comes from unbounded `argv[0]` via `KERN_PROCARGS2`. Both passes independently found and cited the same header/struct locations and reached the same conclusion — see updated Finding 4.
4. **Recommended fix convergence** — CONFIRMED: `ps -c -p <pid> -o pid,rss,comm` is byte-identical to `ps -ceo pid,rss,comm` across the full COLUMNS/tty matrix (Opus pass), with the caveat that `LookupName`'s *parsing*, not just its flags, must also be unified to actually close Finding 7 (both passes flagged this independently).
5. **`os.ErrProcessDone` on Darwin** — CONFIRMED by both passes independently, plus direct code citation of the shared (`//go:build unix`) `exec_unix.go` source; not re-tested by Sonnet since both independent passes agreed and the mechanism is shared, unambiguous code.
6. **`os.FindProcess` dead-code branch** — CONFIRMED by both passes independently, from the same `pidfd_other.go`/`exec_unix.go` source citations; not independently re-tested by Sonnet for the same reason as item 5.
7. **pidfd/`Release()` semantics** — CONFIRMED as a non-issue on Darwin by both passes independently (no handle, no cleanup registered, nothing to leak); not independently re-tested by Sonnet for the same reason as item 5.

**Where the two passes disagreed, and how it was resolved:** Fable's pass reported "`$COLUMNS` had zero effect on either invocation, piped or tty — no width-based truncation at all." Opus's pass reported active, asymmetric truncation under a real pty (e.g. `COLUMNS=20` → `list()`-style `"abcdefg"` (7 chars) vs. `LookupName`-style `"abcdefghijklmnopqrst"` (20 chars)). Since this was a directly, cheaply re-testable empirical claim, Sonnet re-ran it independently rather than picking a side by argument:
```
$ script -q /dev/null env COLUMNS=20 ps -ceo pid,rss,comm | grep <pid>   → "reallyl"                (7 chars)
$ script -q /dev/null env COLUMNS=20 ps -c -p <pid> -o comm=            → "reallylongprocessnam"   (20 chars)
```
This reproduced Opus's result, not Fable's — the pty *does* truncate asymmetrically by `$COLUMNS`, matching Opus's finding. Fable's "zero effect" claim did not hold up on re-test (most likely explanation: Fable's pty test harness did not actually propagate `$COLUMNS` into the pty's reported window size, e.g. depending on how `script`/`stty` interact with an explicit `env COLUMNS=N` prefix versus the pty's actual `ioctl(TIOCGWINSZ)` size — GNU/BSD `ps` on a real terminal generally follows the ioctl window size, and an explicit `$COLUMNS` override only takes effect if the harness makes it visible to `ps` as such). This item is the reason the doc distinguishes "piped: never diverges" (confirmed by all three passes, and the only case that matters for the shipped binary) from "pty: does diverge, worse than Linux" (confirmed only after the tie-break, and irrelevant to production).

**Outstanding, not yet acted on:** Finding 4's spoofability discovery (macOS `comm` = unbounded `argv[0]`, no kernel truncation, fully process-controlled) is new relative to the original Linux-only review and was not something either model was originally asked to look for — it fell out of investigating the "kernel cap" checklist item. It has not yet been triaged into the severity table above as its own numbered finding; it's currently folded into the Finding 4 writeup. Consider promoting it to its own numbered finding if/when this doc is revised again, since its risk profile (safety-check bypass, no race required) is arguably more severe than the truncation finding it was found alongside.