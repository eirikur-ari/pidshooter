# Security Issues Report

**Project:** pidshooter  
**Date:** 2026-07-01  
**Branch:** refactoring-after-ai-creation  

---

## Summary

pidshooter is a terminal game that kills OS processes. Its core attack surface is the ability to send `SIGKILL` to arbitrary PIDs, which makes process-identity validation the most critical security concern. Five issues were found across three severity levels.

---

## Issues

### [CRITICAL] SEC-01 — TOCTOU: PID recycling between discovery and kill

**File:** `internal/adapter/driven/osprocess/osprocess.go:36-45`, `internal/domain/game/loop.go:103-110`

Processes are discovered once at game start via `Find()`, and their PIDs are stored in `[]*Target`. When the user clicks a target, the stored PID is sent directly to `Kill()`. A game session can last 30+ seconds (or longer with `--time=0`). During that window, a discovered process may exit and its PID may be recycled by the OS to a completely different — potentially critical — process.

**Example attack path:**
1. User starts `pidshooter node`
2. A `node` process with PID 1234 is discovered
3. That `node` process exits naturally
4. PID 1234 is reassigned to a new critical system daemon
5. User clicks the stale target — the new daemon receives `SIGKILL`

**Fix:** At kill time, re-validate that the process still exists and its name still matches the original discovery name before sending the signal. In `osprocess.go` `Kill()`, read `/proc/<pid>/comm` (Linux) or `ps -p <pid> -o comm=` (macOS) and compare to the stored name before issuing `SIGKILL`.

---

### ~~[CRITICAL] SEC-02 — PID 0 is not excluded from Kill~~ ✓ FIXED

**File:** `internal/adapter/driven/osprocess/osprocess.go:113-115`, `internal/adapter/driven/osprocess/osprocess.go:135-141`

The filter in `filter()` excludes only the current process PID and PID 1. On Unix, sending a signal to PID 0 sends it to **every process in the current process group**, not to a single process. If a process with PID 0 were ever returned by `ps` parsing (e.g., due to malformed output or a future platform difference), calling `os.FindProcess(0).Signal(SIGKILL)` would kill the entire process group.

**Fix applied:** Changed `filter()` exclusion condition from `p.Pid() == 1` to `p.Pid() <= 1`, and added an explicit guard at the top of `Kill()` that returns an error for any `pid <= 1` (covers PID 0, PID 1, and all negative values) before any syscall is made.

---

### [HIGH] SEC-03 — Insufficient system process exclusion

**File:** `internal/adapter/driven/osprocess/osprocess.go:113-115`

Only PID 1 (init/launchd) and the current process are excluded. Many other low-numbered PIDs are critical kernel threads or system daemons (e.g., PID 2 = `kthreadd` on Linux, launchd helpers on macOS). A broad pattern like `a` would match process names containing the letter "a" — which includes most system processes — and they would all be surfaced as valid kill targets.

There is also no minimum pattern length. A single-character pattern is accepted and can match hundreds of system processes.

**Fix:**
- Enforce a minimum pattern length of 3 characters in `validate()`.
- Optionally, warn users when a pattern matches a disproportionate number of processes (e.g., > 20) before starting.

---

### [MEDIUM] SEC-04 — Kill errors silently discarded

**File:** `internal/domain/game/loop.go:107`

```go
_ = g.killer.Kill(e.Pid())
e.StartKillAnim()
g.Session.RecordKill(e.Rss())
```

The kill error is explicitly ignored. If `Kill` fails (e.g., permission denied, process already gone), the game still plays the kill animation and increments the kill count. This misrepresents game state and, more importantly, hides failures silently — including permission errors that could indicate unexpected privilege conditions.

**Fix:** Propagate or at minimum log the kill error. Only start the animation and record the kill if the signal was sent successfully.

```go
if err := g.killer.Kill(e.Pid()); err == nil {
    e.StartKillAnim()
    g.Session.RecordKill(e.Rss())
}
```

---

### [MEDIUM] SEC-05 — Score file and directory have world-readable permissions

**File:** `internal/adapter/driven/jsonscores/jsonscores.go:53`, `internal/adapter/driven/jsonscores/jsonscores.go:71`

```go
os.WriteFile(s.path, data, 0644)   // world-readable
os.MkdirAll(dir, 0755)             // world-executable and world-readable
```

While scores are not sensitive credentials, the file records game sessions with timestamps and memory statistics, which could reveal information about what processes were running on the machine. On shared systems (e.g., CI/CD runners, development servers with multiple users), this leaks session activity.

**Fix:** Use `0600` for the file and `0700` for the directory:

```go
os.WriteFile(s.path, data, 0600)
os.MkdirAll(dir, 0700)
```

---

### [LOW] SEC-06 — Outdated transitive dependencies with known vulnerabilities

**File:** `go.mod`

The following transitive dependencies (pulled in by `golang.org/x/tools`) are pinned to old versions that contain known CVEs:

| Package | Version | Notes |
|---|---|---|
| `golang.org/x/crypto` | `v0.0.0-20210921155107` | Multiple CVEs in SSH, TLS packages since 2021 |
| `golang.org/x/net` | `v0.6.0` | HTTP/2 DoS CVEs (e.g., CVE-2023-44487) |

Neither package is imported directly by pidshooter's runtime code, reducing actual risk. However, they surface in `go list -m all` and would appear in any automated dependency scan.

**Fix:** Run `go get -u golang.org/x/crypto golang.org/x/net` to update to current versions, then run `go mod tidy`.

---

## Summary Table

| ID | Severity | Title | File |
|---|---|---|---|
| SEC-01 | Critical | TOCTOU PID recycling | `osprocess.go`, `loop.go` |
| SEC-02 | Critical | ~~PID 0 not excluded from Kill~~ ✓ FIXED | `osprocess.go` |
| SEC-03 | High | Insufficient system process exclusion | `osprocess.go` |
| SEC-04 | Medium | Kill errors silently discarded | `loop.go:107` |
| SEC-05 | Medium | Score file world-readable permissions | `jsonscores.go` |
| SEC-06 | Low | Outdated transitive dependencies | `go.mod` |

## Fix Priority

1. **SEC-02** — trivial one-liner guard; eliminates the process-group kill risk entirely.
2. **SEC-01** — requires re-validating process name at kill time; most impactful fix.
3. **SEC-03** — add minimum pattern length in `validate()`; low effort, high benefit.
4. **SEC-04** — stop recording kills when `Kill()` returns an error.
5. **SEC-05** — change two permission constants.
6. **SEC-06** — `go get -u` + `go mod tidy`.
