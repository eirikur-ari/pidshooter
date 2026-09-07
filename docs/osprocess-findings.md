# osprocess.go — Cross-Referenced Findings (Fable × Opus)

> **⚠️ PLATFORM NOTICE — READ FIRST IF YOU ARE OPENING THIS ON macOS**
>
> Every finding, reproduction, and empirical measurement in this document was
> produced on a **Linux** machine:
>
> ```
> Linux fafnir 7.1.13-200.fc44.x86_64 #1 SMP PREEMPT_DYNAMIC Wed Sep 2 13:58:38 UTC 2026 x86_64 GNU/Linux
> ```
> `ps` on that machine is **procps-ng** (GNU/Linux). None of the truncation
> behavior described below — especially Finding 1 — was tested against
> **BSD `ps`** (macOS's `ps`), which has different flag semantics and a
> different (and, as far as the Linux-side authors know, undocumented-here)
> relationship between piped output, terminal width, and `$COLUMNS`.
>
> `osprocess.go` already has a separate `darwin` branch (`runtime.GOOS ==
> "darwin"` in both `LookupName` and `list()`, using `-c` instead of plain
> `-o comm=`/`-eo`), which exists precisely because BSD `ps` doesn't behave
> like GNU `ps`. **Do not assume any fix validated here also fixes the
> `darwin` branch — it must be independently verified on real macOS
> hardware.** See the "macOS verification checklist" section at the bottom —
> run it first, then update this doc with the results (and remove or revise
> this notice once macOS is verified).

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

## 2. Critical — verified bug (Linux-confirmed; macOS status unknown)

### Finding 1: `List()` and `LookupName()` use different `ps` invocations that truncate names differently, breaking the kill-time safety check

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

**Status: CONFIRMED on Linux. UNVERIFIED on macOS — see checklist below before attempting any fix on that platform.**

**Recommended fix — deliberately platform-agnostic:** Don't chase `-ww`/width flags (that's GNU-`ps`-specific syntax and is not confirmed to have equivalent semantics under BSD `ps`). Instead, make `LookupName` issue **the same command shape** as `list()` — same flags, same column order, same darwin/non-darwin branching — just filtered to one `pid`, e.g. reuse `list()`'s underlying invocation and grep/filter by PID in Go, rather than maintaining a second, differently-shaped `ps` call. Whatever truncation policy the platform's `ps` applies then applies identically to both code paths *by construction*, without this codebase needing to know or hard-code either OS's width rules. This is the fix to implement and then verify on macOS (see checklist).

---

## 3. Moderate findings (both models, deduped)

### Finding 2: `os.ErrProcessDone` is discarded, producing permanently un-killable "undead" targets
`osprocess.go:64-67` returns the raw `proc.Signal` error uninspected. Go's unix `findProcess`/`Signal` path maps `ESRCH` (process already gone) to the distinguishable sentinel `os.ErrProcessDone` — the adapter throws this signal away into an opaque `error`. Downstream, `application/process/service.go:57-58` returns `(false, false, err)` on any `Kill` error — note `shouldReap` is hardcoded `false` on this branch, unlike the `LookupName`-mismatch branch — so a target whose process exited in the discovery→kill window becomes stuck `Alive` forever with no way to reap it. Confirmed against Go's `os/exec_unix.go` and `pidfd_linux.go` source (both platforms map `ESRCH`→`ErrProcessDone`; behavior expected to hold on Darwin too since it's in the portable `os` package, not a GOOS-specific file — **worth confirming directly on macOS anyway**, see checklist).

**Fix:** in `Kill`, check `errors.Is(err, os.ErrProcessDone)` and let the caller (or this adapter itself) signal "already gone" distinctly from "kill failed" — resolves cleanly alongside Finding 8 (the redundant `bool` return).

### Finding 3: The port drops process ownership entirely — the domain can't express "don't touch processes I don't own"
`list()` uses `-eo` (all users). `outbound.ProcessInfo` carries only `PID`/`Name`/`Rss` — no UID. Grep confirms zero references to `Getuid`/`Geteuid`/ownership anywhere in the tree; the only domain guard is `Info.IsProtected()` (`PID <= 1`). Two consequences: (a) as non-root, root-owned processes become unkillable targets that hit the `EPERM`→generic-error path (same "undead" trap as Finding 2, since `EPERM` is *not* converted to a sentinel by Go); (b) as root, `pidshooter sshd` will happily kill every `sshd` on the box — `PID <= 1` is the only guard between the player and system daemons. This is a boundary-fidelity gap, not a leak: cheap for the adapter to surface (`ps -eo uid,pid,rss,comm`), and the domain is structurally unable to build ownership-based policy without it.

### Finding 4: Kernel `comm` truncation to 15 characters is a separate, more fundamental limit than Finding 1
Independently of the `COLUMNS` mismatch, Linux truncates `/proc/pid/comm`-derived `comm` values to `TASK_COMM_LEN - 1` = 15 visible characters — this applies **identically** to both `List` and `LookupName` once `COLUMNS` is wide enough not to truncate further (see the `COLUMNS=40` row above, where both already read `"reallylongproce"`, 15 chars). This weakens the kill-time name-recheck (the TOCTOU defense documented in `docs/fable-review.md` S2) for *any* process with a name ≥15 characters, independent of whether Finding 1 is ever fixed. **Not yet checked whether macOS's `comm`/`-c` has an equivalent or different length cap** — see checklist.

### Finding 5: No timeout/context on either `ps` subprocess call
`osprocess.go:49` and `:75` both use bare `exec.Command(...).Output()`. A wedged `ps` (stalled `/proc` reader, hung container runtime) hangs `List` at startup with no feedback, and hangs `LookupName` inside the game loop's `go s.killOrReap(...)` goroutine, which then never reaches its `done`-channel select. **Fix:** `exec.CommandContext` with a short deadline on both calls. This fix is platform-neutral — no macOS-specific concern.

### Finding 6: `Kill(pid int)` makes the documented verify-then-kill contract structurally unenforceable
The port's doc comment and `application/process.Service.Kill` both document a verify-then-kill sequence, but `LookupName(pid)` and `Kill(pid)` are two independent calls keyed only on an `int` — nothing pins the OS-level identity across them, so a PID recycled between the two calls is indistinguishable from the original target. Notably, `os.FindProcess` on Linux already opens a pidfd internally (which *does* pin identity) but the adapter discards it at the end of a single `Kill` call rather than letting it span the verify step. Flagged as design-level (the race window is narrow and needs PID wraparound), not a live bug — a handle-returning port shape (`Open(pid) (Handle, error)` → `Handle.Name()`/`Handle.Kill()`) would make the safety property enforceable rather than merely documented. **macOS uses a different process-handle mechanism (no pidfd equivalent as of this writing)** — any redesign here needs separate macOS-side research, not just a recompile.

---

## 4. Minor / low findings (both models, deduped)

| # | Severity | Finding | Linux-verified? | macOS status |
|---|---|---|---|---|
| 7 | Minor | `comm` whitespace collapsed in `List` (`strings.Fields`+`Join`) but not in `LookupName` (`TrimSpace` only) — extends prior review's F11. Consequence corrected from F11: causes silent **reap**, not silent refusal, per the `shouldReap`-before-`err` check in `application/game/service.go`. | Yes, by direct code reading | Same code path runs on darwin branch too — untested there |
| 8 | Minor | `Kill`'s `bool` return is redundant: `killed ≡ (err == nil)` always, since `os.FindProcess` never fails on Unix and the only remaining path is `Signal`'s error. Makes a defensive branch in `application/game/service.go` (`case !killed: ...`) unreachable with this adapter. | Yes | Same reasoning is POSIX-general, not Linux-specific — low risk of macOS divergence, but unverified |
| 9 | Minor | `ps` stderr (up to 32KB, captured in `exec.ExitError.Stderr`) is discarded by plain `%w` wrapping in both `list()` and `LookupName` | Yes | Platform-neutral fix |
| 10 | Minor | `os.FindProcess`'s error branch in `Kill` (`osprocess.go:60-62`) is dead code on Linux — `findProcess` never returns non-nil error on unix per Go source | Yes, confirmed against Go source | **Not necessarily true on Darwin** — verify against Go's `os` darwin-specific paths before assuming this branch is dead there too |
| 11 | Minor | pidfd leaked until GC — `proc.Release()` is never called on the `*os.Process` from `FindProcess`; Go's own docs call `Release` required when `Wait` is never invoked | Yes | **Darwin doesn't use pidfd** — the leak, if any, takes a different shape there (traditional process-handle bookkeeping); verify separately |
| 12 | Minor | Header row (`lines[1:]`) skipped positionally with no content validation; parse-failure policy is inconsistent across the three fields (skip row / skip row / default-to-zero) | Yes | Platform-neutral, both branches share this code |
| 13 | Low | `NewProcess` returns the `outbound.ProcessManager` interface instead of `*Process`, forcing `osprocess_test.go` to hand-construct `&Process{psPath: path}` to get a concrete value | Yes | N/A |
| 14 | Low | Doc-drift: `docs/hexagonal-arc.md` documents `OwnPid()`; code has `OwnPID()` | Yes | N/A |
| 15 | Low | Runtime `GOOS` branching instead of build-tagged platform files (`osprocess_linux.go`/`osprocess_darwin.go`) — compiles silently on unsupported platforms (e.g. Windows), reaches "ps not found" by accident rather than design | Yes | Directly relevant to this doc's theme — **this is the mechanism that would make future macOS-specific fixes maintainable**, worth doing alongside the Finding 1 fix |
| 16 | Low | SIGKILL to a zombie reports success (`kill(2)` on a zombie returns 0) even though nothing dies | Yes (POSIX-general) | Same POSIX semantics apply on Darwin, low risk of divergence |

---

## 5. Relationship to `docs/fable-review.md`

- **F11** ("whitespace-canonicalization mismatch") is confirmed still present — see Finding 7 above, which also corrects F11's stated consequence (silent *reap*, not silent *refusal*).
- **S2** ("residual TOCTOU window") is sharpened by Finding 6 above (pidfd is opened and discarded within `Kill`, but never spans the `LookupName`→`Kill` gap).
- **S7** ("ps output parsing trusts field structure") is extended by Findings 1, 4, 9, and 12 — the mitigation S7 cites ("substantially neutralized by kill-time name re-verification") is exactly what Finding 1 shows breaking down when the two `ps` calls disagree.

---

## 6. macOS verification checklist

Run this on real macOS hardware before implementing or trusting any fix for Finding 1, and update this document with the results (replace the "UNVERIFIED" markers above).

1. **Reproduce Finding 1's asymmetry, or rule it out.** Build the same test process (a long-named symlink to `sleep`, or any process with a name ≥ 16 characters) and run the two `ps` invocations `osprocess.go` actually uses on darwin:
   ```
   ps -ceo pid,rss,comm            # list() on darwin
   ps -c -p <pid> -o comm=         # LookupName() on darwin
   ```
   Compare their output for the same long-named process, with and without `$COLUMNS` set to several values (try unset, 20, 40, 80, and piped vs. attached to a real terminal). Confirm whether they ever disagree the way the Linux invocations do.
2. **Determine BSD `ps`'s truncation trigger.** Does it respond to `$COLUMNS` at all when output is piped (not a tty)? Does `-c` impose its own fixed-width truncation independent of terminal width? (On BSD-derived `ps`, `-c` changes *which* name is shown — short name vs. full command line — and may have separate truncation rules from GNU `ps`'s terminal-width logic.)
3. **Check the kernel-level name cap.** Does macOS have an equivalent to Linux's 15-character `TASK_COMM_LEN` cap (Finding 4), and if so what's the actual limit for `-c`-style short names?
4. **Verify the recommended fix actually converges the two paths.** After unifying `LookupName` to reuse `list()`'s exact darwin flag set (per the fix direction in Finding 1), re-run the comparison from step 1 and confirm they now always agree, across the same COLUMNS/tty matrix.
5. **Spot-check Finding 2 (`os.ErrProcessDone`)** — kill a process out from under `LookupName`→`Kill` (e.g. race a background `kill` command against the adapter call) and confirm Go still surfaces `os.ErrProcessDone` on Darwin the way it does on Linux, since Finding 2's fix depends on that sentinel being reliable cross-platform.
6. **Spot-check Finding 10** (whether `os.FindProcess`'s error branch is truly dead on Darwin) by reading Go's `os` package source for the darwin-specific process-finding path, or by attempting to trigger a non-nil error from `FindProcess` directly (e.g. against PID `-1` or `0`, off the main kill path so it's safe).
7. **Spot-check Finding 11** (pidfd leak) — Darwin has no pidfd; confirm what `*os.Process.Release()` actually does on Darwin and whether the "leak until GC" framing even applies there, or whether it's a non-issue on this platform.

Once these are run, fold the results back into Sections 2–4 above (replace "UNVERIFIED"/"untested" markers with concrete macOS findings), and only then implement the Finding 1 fix for the darwin branch specifically.