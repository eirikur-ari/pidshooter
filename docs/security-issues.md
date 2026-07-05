# Security Issues Report

**Project:** pidshooter  
**Date:** 2026-07-01  
**Branch:** refactoring-after-ai-creation  

---

## Summary

pidshooter is a terminal game that kills OS processes. Its core attack surface is the ability to send `SIGKILL` to arbitrary PIDs, which makes process-identity validation the most critical security concern. Five issues were found across three severity levels.

---

## Issues

### ~~[CRITICAL] SEC-01 — TOCTOU: PID recycling between discovery and kill~~ ✓ FIXED

**File:** `internal/adapter/driven/osprocess/osprocess.go:36-45`, `internal/domain/game/loop.go:103-110`

Processes are discovered once at game start via `Find()`, and their PIDs are stored in `[]*Target`. When the user clicks a target, the stored PID is sent directly to `Kill()`. A game session can last 30+ seconds (or longer with `--time=0`). During that window, a discovered process may exit and its PID may be recycled by the OS to a completely different — potentially critical — process.

**Example attack path:**
1. User starts `pidshooter node`
2. A `node` process with PID 1234 is discovered
3. That `node` process exits naturally
4. PID 1234 is reassigned to a new critical system daemon
5. User clicks the stale target — the new daemon receives `SIGKILL`

**Fix applied:** The `ProcessKiller` interface was extended to `Kill(pid int, name string) error`. `loop.go` now passes `e.Name()` alongside the PID. In `osprocess.go`, `Kill()` runs `ps -p <pid> -o comm=` (with `-c` on macOS to match discovery flags) to get the process's current name, then compares it to the expected name before issuing `SIGKILL`. A mismatch or a dead process returns an error, which causes `killTarget` to skip the animation and kill-count increment — so a recycled PID is never killed.

---

### ~~[CRITICAL] SEC-02 — PID 0 is not excluded from Kill~~ ✓ FIXED

**File:** `internal/adapter/driven/osprocess/osprocess.go:113-115`, `internal/adapter/driven/osprocess/osprocess.go:135-141`

The filter in `filter()` excludes only the current process PID and PID 1. On Unix, sending a signal to PID 0 sends it to **every process in the current process group**, not to a single process. If a process with PID 0 were ever returned by `ps` parsing (e.g., due to malformed output or a future platform difference), calling `os.FindProcess(0).Signal(SIGKILL)` would kill the entire process group.

**Fix applied:** Changed `filter()` exclusion condition from `p.Pid() == 1` to `p.Pid() <= 1`, and added an explicit guard at the top of `Kill()` that returns an error for any `pid <= 1` (covers PID 0, PID 1, and all negative values) before any syscall is made.

---

### ~~[HIGH] SEC-03 — Insufficient system process exclusion~~ ✓ FIXED (minimum pattern length)

**File:** `internal/adapter/driven/osprocess/osprocess.go:113-115`

Only PID 1 (init/launchd) and the current process are excluded. Many other low-numbered PIDs are critical kernel threads or system daemons (e.g., PID 2 = `kthreadd` on Linux, launchd helpers on macOS). A broad pattern like `a` would match process names containing the letter "a" — which includes most system processes — and they would all be surfaced as valid kill targets.

There is also no minimum pattern length. A single-character pattern is accepted and can match hundreds of system processes.

**Fix applied:** Added `MinPatternLength = 3` constant to the driven port. `validate()` in `osprocess.go` now rejects any pattern shorter than 3 characters. `parseArgs()` in the CLI layer mirrors the same check for fast, user-facing feedback. The optional broad-match warning was not implemented.

---

### ~~[MEDIUM] SEC-04 — Kill errors silently discarded~~ ✓ FIXED (resolved during SEC-01)

**File:** `internal/domain/game/loop.go:107`

The kill error was explicitly ignored (`_ = g.killer.Kill(...)`). The animation and kill count were recorded unconditionally, misrepresenting game state and hiding failures silently.

**Fix applied:** Resolved as part of the SEC-01 fix. `killTarget` now checks the error from `Kill` and returns early on failure — `StartKillAnim()` and `RecordKill()` are only called when the signal was sent successfully.

---

### ~~[MEDIUM] SEC-05 — Score file and directory have world-readable permissions~~ ✓ FIXED

**File:** `internal/adapter/driven/jsonscores/jsonscores.go:53`, `internal/adapter/driven/jsonscores/jsonscores.go:71`

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

## Summary Table

| ID | Severity | Title | File |
|---|---|---|---|
| SEC-01 | Critical | ~~TOCTOU PID recycling~~ ✓ FIXED | `osprocess.go`, `loop.go` |
| SEC-02 | Critical | ~~PID 0 not excluded from Kill~~ ✓ FIXED | `osprocess.go` |
| SEC-03 | High | ~~Insufficient system process exclusion~~ ✓ FIXED | `osprocess.go` |
| SEC-04 | Medium | ~~Kill errors silently discarded~~ ✓ FIXED | `loop.go:107` |
| SEC-05 | Medium | ~~Score file world-readable permissions~~ ✓ FIXED | `jsonscores.go` |
| SEC-06 | Low | ~~Outdated transitive dependencies~~ ✓ FIXED | `go.mod` |

## Fix Priority

1. **SEC-02** — trivial one-liner guard; eliminates the process-group kill risk entirely.
2. **SEC-01** — requires re-validating process name at kill time; most impactful fix.
3. **SEC-03** — add minimum pattern length in `validate()`; low effort, high benefit.
4. **SEC-04** — stop recording kills when `Kill()` returns an error.
5. **SEC-05** — change two permission constants.
6. **SEC-06** — `go get -u` + `go mod tidy`.
