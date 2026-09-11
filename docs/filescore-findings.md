# filescore — Cross-Referenced Findings (Fable × Opus)

**Date:** 2026-09-09
**Method:** Two independent from-scratch reviews of `internal/infrastructure/filescore/filescore.go` and its callers, one on Claude Fable 5, one on Claude Opus, each with full repo context (`docs/hexagonal-arc.md`, the `outbound.ScoreStore` port, the sole caller `application/score.Service` and its converter, the `core/score` domain types, the `runner.go`/`main.go` wiring, the `fake.Store` test double, and the prior review `docs/fable-review.md` finding S6, which already touched this file). Findings below are cross-referenced, deduped, and independently re-verified by a third pass (Sonnet, orchestrating) — every load-bearing empirical claim in the Moderate section was re-executed against real code on this machine (Darwin, Go 1.26.5) rather than taken on either model's word; see each finding's "Verified" line.
**Scope:** `internal/infrastructure/filescore/filescore.go` (the `Store` adapter) and its full caller chain: `application/contract/outbound/score.go` + `error.go` (the port), `application/score/service.go` + `converter.go` (the sole application-layer consumer), `core/score/board.go` + `entry.go` (the domain), `application/runner.go` + `cmd/pidshooter/main.go` (wiring), `internal/testutil/fake/store.go` (test double).
**Supersedes:** `docs/fable-review.md` finding **S6** ("Score file written non-atomically") — this document supersedes S6's one-line description with a full trace of its actual downstream consequence (see Finding 2 below).
**Platform:** produced and verified on macOS (Darwin, Go 1.26.5). See §8 for Linux applicability — unlike `docs/osprocess-findings.md`'s subject file, this one has no `runtime.GOOS` branching, so almost every finding here is platform-neutral by construction.
**Note:** the package reviewed here (`internal/infrastructure/scorefilestore`, file `score_file_store.go`) was renamed to `internal/infrastructure/filescore`/`filescore.go` after this review was written; references throughout this document have been updated to the new name for consistency, but the `:NN-NN` line numbers cited alongside them reflect the file's shape at review time and may no longer line up exactly with current code following the Moderate-findings remediation that followed.
**Remediation:** All 5 Moderate findings (§3) were fixed 2026-09-10, via a new `internal/infrastructure/fsutil` package, a new `outbound.CorruptedDataError` sentinel, and a `NewStore`/`NewStoreAt` constructor split — the fix was independently cross-reviewed again (Fable × Opus) before being finalized. Findings 6-9 (§4) were also fixed the same day: 6 (cross-process last-writer-wins) via reload-merge-save in `Service.RecordScore`; 7 (stale directory permissions) via an unconditional `Chmod` in `fsutil.WriteFileAtomic`; 8 (inconsistent error handling) via modernizing `Load`'s not-found check to `errors.Is`; 9's under-verification gap via fuller round-trip/on-disk-schema tests. Finding 10 (naming collision) was deliberately accepted, won't-fix — both functions are unexported, so it's a readability nit with no functional or compile-time risk, and neither renaming nor sharing the conversion logic across packages was judged worth its cost. See the **Status** line under each Moderate finding and the **Status** column in §4/§5 for exact detail — §5's Low findings were left open, deliberately out of scope for this pass.

---

## 1. Architecture conformance — both models agree, clean

Both reviews independently concluded `filescore.Store` is a correct "dumb adapter":

- No reverse imports into `core/*` or other `application/*` packages — the only non-stdlib import is `application/contract/outbound`, used solely for the `ScoreStore`/`ScoreBoard`/`ScoreEntry`/`NotFoundError` types this file structurally satisfies or returns.
- Zero embedded business logic: no ranking, no top-N truncation, no "is this safe to overwrite" decision anywhere in this file. That responsibility correctly and entirely lives in `core/score.Board.Add`/`sortByRank` and `application/score.Service.RecordScore`'s `errors.As(err, &outbound.NotFoundError{})` gate.
- No leakage outward: `encoding/json` tags, `os.FileMode`, and `filepath` never appear above the adapter. The port (`outbound.ScoreEntry`/`ScoreBoard`) is tag-free and stdlib-`os`-free.
- The port's granularity (whole-board `Load`/`Save`, not per-entry) is the right level of abstraction for a domain that owns the whole ranked top-10 table — a finer `Append(entry)` port would be more atomicity-friendly but would push ranking/truncation policy down into the adapter, which is the worse trade.

**Verdict: no hexagonal-boundary violations, no changes needed on the layering axis.** The one real defect below (Finding 1) is an *internal cohesion* bug — the adapter's own instance state (`s.path`) and a package-level global (`defaultDir()`) disagree with each other — not a cross-layer leak. The issues in this document are about correctness, robustness, and testability of the adapter and the thinness of its test coverage, not about placement or coupling.

A secondary, out-of-scope observation both reviews made in passing while tracing callers: `core/score.Board.PrintHighScores` (`board.go:41-63`) and `application/score.Service.PrintResults` (`service.go:69-73`) both do direct console I/O from the domain/application layers — already tracked as `docs/fable-review.md` D1/D2, with a self-aware `TODO` at `board.go:38`. Not re-litigated here; noted only because both reviewers passed through this code while tracing the score-persistence path.

## 2. Test design quality — both models agree: tests pass, but the design forces workarounds

Both reviews reached the same verdict independently: the eleven tests in `filescore_test.go` verify round-trip/overwrite/not-found/invalid-JSON/permission behavior competently, but the *design* — not the test authors — is what's creating the gaps:

- **`NewStore()` is never called by any test.** Every test constructs `&Store{path: ...}` directly (`filescore_test.go:15-18`), reaching into the unexported `path` field to work around the constructor's hardcoded `defaultPath()`. This isn't a style nitpick — see Finding 5's verified coverage numbers.
- **Round-trip tests assert one field out of six.** `TestSaveLoadRoundTrip`/`TestSaveLoadMultipleEntries` (`:49-75`) only ever check `.Kills`; a transposition bug in `toBoard`/`toScoreBoard`'s five other hand-mapped fields would pass a fully green suite.
- **The on-disk JSON shape itself is never asserted.** No test reads the raw bytes `Save` writes and checks them against the six `json:` tags (`:21-26`) — those tags are this adapter's actual compatibility contract with a user's existing `highscores.json`, and nothing protects them from an accidental rename.
- **Ironic confirmation of the design smell:** the test suite has a real, unintended side effect — see Finding 1, verified below, where every `Save`-touching test was found to create `$HOME/.config/pidshooter` on the real filesystem, because the code the tests can't reach (`defaultDir()`) still runs regardless.

By contrast, `application/score/service_test.go` — one layer up — is genuinely good: error classification, severity, and cause-chain are each asserted explicitly (`:33-63`), and the "don't overwrite on a real load failure" policy is tested directly (`:98-114`). The application layer's tests reflect a testable design; the adapter's current shape doesn't let its own tests do the same.

**Status (2026-09-10): all four gaps above are closed.** `filescore_test.go` now constructs stores exclusively via `NewStore()`/`NewStoreAt()` (no more reaching into the unexported `path` field); `TestSaveLoadRoundTrip`/`TestSaveLoadMultipleEntries` assert full 6-field equality instead of just `.Kills`; `TestSaveWritesJSONMatchingOnDiskSchema` asserts the raw on-disk JSON against the `json:` tags; and the `$HOME`-pollution side effect no longer exists (Finding 1's fix) — verified by re-running the original repro command against current code.

---

## 3. Moderate findings

### Finding 1: `Save()`'s directory creation ignores `s.path` entirely, using a package-level global instead — the test suite has been writing to the developer's real `$HOME` as an unintended side effect

**Status: Fixed (2026-09-10).** Directory creation now derives from the actual path being written to — `internal/infrastructure/fsutil.WriteFileAtomic` calls `os.MkdirAll(filepath.Dir(path), 0700)` on its `path` argument every time, and `defaultDir()`/`makeConfigDir()` were deleted outright. Regression test: `TestSaveToNestedNonexistentDirectoryCreatesParentDirs`; the `$HOME`-pollution repro below no longer reproduces (re-verified against current code).

**Evidence:**
```go
// filescore.go:56-67
func (s *Store) Save(sb outbound.ScoreBoard) error {
	err := makeConfigDir()          // no argument — ignores s.path
	...
	return os.WriteFile(s.path, data, 0600)   // uses s.path
}

// filescore.go:113-119
func makeConfigDir() error {
	dir := defaultDir()             // package-level, not derived from any Store instance
	err := os.MkdirAll(dir, 0700)
	...
}
```
`Store` has an instance field, `path` (`:16`), that every test carefully points at a `t.TempDir()`. `Save` never uses it to decide which directory to create — `makeConfigDir()` always resolves `defaultDir()` (`~/.config/pidshooter`) independently of whatever `s.path` is.

**Reproduced independently on this machine (2026-09-09):**
```
$ HOME=/tmp/filescore-verify/fakehome go test ./internal/infrastructure/filescore/... -run TestSave -v
--- PASS: TestSaveCreatesFile / TestSaveFilePermissions / TestSaveLoadRoundTrip / TestSaveLoadMultipleEntries / TestSaveOverwritesPreviousFile
$ find /tmp/filescore-verify/fakehome
/tmp/filescore-verify/fakehome/.config
/tmp/filescore-verify/fakehome/.config/pidshooter        ← created by the test run, despite every test using t.TempDir()
```
Every `Save`-touching test in the suite passes, and every one of them creates `~/.config/pidshooter` in whatever `$HOME` the test process has, as an unrequested side effect. It's currently harmless — an empty directory, never a file, since `s.path` always points into the test's temp dir — but it is unsandboxed state mutation from `go test`/`make test`, and it also means:

**A non-default path is silently broken for `Save`.** Verified directly:
```
dir := t.TempDir() + "/nested/deeper"
s := &Store{path: dir + "/scores.json"}
s.Save(...)  →  err: open .../nested/deeper/scores.json: no such file or directory
```
`makeConfigDir()` created `~/.config/pidshooter` (which already existed, so this was invisible) instead of `nested/deeper`. Any future caller that constructs a `Store` with a custom path — exactly what Finding 5's injectable-constructor fix would add — inherits this bug immediately unless it's fixed first.

**Suggested fix:** Make directory creation derive from instance state: `os.MkdirAll(filepath.Dir(s.path), 0700)`, either inlined into `Save` or as a method (`s.makeConfigDir()`) rather than a free function. One-line root-cause fix; removes the `$HOME` side effect from the test suite and unblocks Finding 5.

---

### Finding 2: Non-atomic write (re-assessment of `docs/fable-review.md` S6) — traced to a self-perpetuating, silent save-lockout

Both reviews independently re-verified S6 still reproduces against current code and, independently, both traced the *same* downstream consequence through real caller code — this convergence is a stronger signal than either review alone.

**Status: Fixed (2026-09-10).** `Save` now goes through `fsutil.WriteFileAtomic`: write to a temp file in the same directory, `Chmod`, `Sync`, `Close`, then `os.Rename` into place (plus a best-effort directory fsync after rename). `path` is never opened or truncated directly, so a crash mid-write can no longer leave a truncated or zero-byte file behind. See `internal/infrastructure/fsutil/fsutil.go` and its tests, including a deterministic permission-induced-failure test proving the original file survives untouched when the write fails partway through.

**Evidence:**
```go
// filescore.go:67
return os.WriteFile(s.path, data, 0600)
```
`os.WriteFile` opens with `O_TRUNC` (truncating any existing content to zero length) before writing the new bytes, and never calls `fsync` — confirmed against this Go version's stdlib (`$GOROOT/src/os/file.go`'s `WriteFile`).

**Reproduced on this machine (2026-09-09):** a zero-byte file (the state a crash mid-`O_TRUNC` leaves behind) produces `Load() err = "unexpected end of JSON input"`, with `os.IsNotExist(err) == false` — i.e. **not** classified as `outbound.NotFoundError`.

**Full failure path, traced against real caller code:**
1. `Save` is called exactly once per session, from `application/score.Service.RecordScore` (`service.go:59`), itself called once from `application/runner.go:68`.
2. A crash/kill in the narrow window between truncation and the new bytes landing leaves a corrupt (possibly zero-byte) file.
3. Next launch: `Load` returns the raw `json` error, unwrapped — not `NotFoundError`.
4. `application/score.Service.LoadScoreBoard` (`service.go:31-39`) classifies this `CodeScoreLoadFailed`/`SeverityWarning`; `apperror.Handler` (`SeverityWarning` is absorbed) lets the game proceed with an empty in-memory board.
5. At end of session, `RecordScore`'s deliberate safety gate — `errors.As(err, &outbound.NotFoundError{})`, `service.go:58` — is **false** for this generic corruption error, so **`Save` is skipped**, by design, to avoid overwriting a file that "failed to load for a real reason."
6. **Every subsequent run repeats steps 3-5.** The corrupt file is never rewritten, backed up, or reset — score-saving is silently and permanently disabled until a human manually deletes `highscores.json`. The only symptom is one `warning:` line on stderr per session; the game itself keeps working normally.

**Reachability is higher than "requires a hardware fault or OS crash."** `core/process.Find` (`internal/core/process/info.go:43-57`) excludes only the caller's own PID, PID ≤ 1, and processes the caller doesn't own by UID — it does **not** exclude *other* `pidshooter` processes. Two `pidshooter` instances run by the same user (`pidshooter pidshooter` — pattern length 10, passes the length-3 minimum; matches by substring against the other instance's own `comm`) let one instance `SIGKILL` the other, including mid-`Save`. The hit window is small (one ~1.5KB `write(2)`), so this doesn't make the bug *likely*, but it removes "no plausible trigger without a hardware fault" as a reason to discount it — pidshooter ships a first-party way to trigger its own worst case.

**Severity reconciliation:** Fable rated this Minor; Opus rated it Moderate. Both independently reached the identical "self-perpetuating lockout" mechanism through separate code tracing — that convergence, combined with the concrete self-inflicted-SIGKILL reachability path (verifiable from code already read for this review) and the triviality of the fix relative to "permanent silent data loss requiring manual file surgery," settles this at **Moderate**.

**Suggested fix:** Write-temp-then-rename in the same directory as `s.path` (so `os.Rename` stays on one filesystem, making it atomic on POSIX), with an `f.Sync()` before the rename for durability, `Chmod` (or rely on `CreateTemp`'s `0600` default) for permissions, and a `defer os.Remove(tmp)` on every error path. This also incidentally fixes Finding 7 below (permissions on pre-existing files), since a renamed-over file is always freshly created.

---

### Finding 3: `Load` cannot distinguish "corrupt/unparseable" from "unreadable," so the one safety policy that's correct for one case is wrongly applied to the other

**Status: Fixed (2026-09-10).** Took the first suggested option below: added `outbound.CorruptedDataError` (a bare marker type, same shape as `NotFoundError` — no `Cause` field, since nothing downstream needed to unwrap into it) alongside `NotFoundError`. `Load` returns it on `json.Unmarshal` failure; `application/score.Service.RecordScore`'s overwrite gate now treats it as safe-to-overwrite, same as `NotFoundError`. Named `CorruptedDataError` rather than the `CorruptError` first drafted during implementation — kept storage-agnostic (no "File" in the name, matching `NotFoundError`'s port-neutral style) per two independent post-implementation reviews (Fable × Opus) that both flagged the original name, but deliberately not generalized further than "persisted data failed to parse" until a second `ScoreStore` implementation actually needs a broader concept.

**Evidence:** `filescore.go:39-54` — `Load` has exactly two error shapes: `outbound.NotFoundError{}` (file absent) and everything else, raw and undifferentiated. A `json.SyntaxError`, an `EACCES`, an `ENOTDIR` (config dir exists as a *file*, not a directory), and an `io` error are indistinguishable to `application/score.Service`, which therefore has only one lever (`service.go:58`): `NotFoundError` → safe to overwrite; anything else → never overwrite.

**"Never overwrite" is correct for `EACCES`/`EIO`** (the data may be intact; the caller just couldn't read it right now) **and actively wrong for a confirmed-corrupt JSON body** (there's nothing left to protect by refusing to write). Because the port can't tell these apart, the conservative policy — correct for one case — is applied to both, which is exactly what produces Finding 2's permanent lockout for the JSON-corruption case specifically.

**Verified on this machine (2026-09-09):** a directory blocked by a pre-existing regular file at the expected path (`ENOTDIR`) also fails `os.IsNotExist` — `open .../blocker/scores.json: not a directory`, `os.IsNotExist=false` — so it falls into the same permanent-lockout bucket as a corrupt file, not a fresh-install one. Also verified: a file containing literal `null` loads *successfully* as an empty board (`{Scores:[]}`, `err=nil`) — the one malformed shape currently tolerated silently rather than surfaced at all.

**Suggested fix (two viable options, in order of architectural cleanliness):**
- Add a distinct `outbound.CorruptError` sentinel alongside the existing `NotFoundError` (`outbound/error.go`), returned by `Load` on JSON unmarshal failure. `application/score.Service` then makes the "corrupt is safe to overwrite" call explicitly — keeping the policy decision in the application layer, consistent with how the rest of this codebase places business logic (Finding 1 of `docs/osprocess-findings.md`'s methodology: the "dumb adapter" contract). Costs one new port type and one new branch in `service.go:58`.
- Cheaper, adapter-local alternative: on unmarshal failure, best-effort-rename the bad file to `highscores.json.corrupt` and return `NotFoundError{}` (wrapped, with the parse error as cause) — the existing `errors.As` gate then treats it as a fresh install with no application-layer change, and the corrupt data is preserved for inspection rather than silently discarded.

---

### Finding 4: `defaultDir()`'s `os.UserHomeDir()` failure silently redirects to the current working directory

**Status: Fixed (2026-09-10).** `defaultDir()` is gone; `internal/infrastructure/fsutil.ConfigDir(appName)` now returns an error instead of silently falling back to `"."` when `os.UserHomeDir()` fails, and the adapter's `NewStore()` propagates it as `NewStore() (*Store, error)` — exactly the pattern this finding's suggested fix names, matching `osprocess.NewProcess()`. `cmd/pidshooter/main.go` handles the new error the same way it already handled `osprocess.NewProcess()`'s.

**Evidence:**
```go
// filescore.go:100-107
func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "pidshooter")
}
```
`os.UserHomeDir()` fails when `$HOME` (Unix) or `%USERPROFILE%` (Windows) is unset — realistic under systemd units without `Environment=HOME`, some CI/cron harnesses, minimal containers, or `env -i`. On failure, the score path silently becomes `./.config/pidshooter/highscores.json`, relative to wherever `pidshooter` happens to be launched from:
1. Running from different directories produces *different, unrelated* score files, each silently believing it's canonical — high scores appear to randomly vanish depending on launch directory, with no diagnostic.
2. It litters arbitrary working directories (potentially a git checkout) with a hidden `.config/` folder.
3. There is no way to log this — the adapter has no `outbound.Logger`, by design — which is itself the argument for returning an error instead of guessing.

Not currently reachable by any test — see Finding 5.

**Suggested fix:** Don't silently redirect to cwd on failure. Either propagate the error — `NewStore() (*Store, error)`, matching the pattern `osprocess.NewProcess()` already establishes (`osprocess/process.go:30`) and that `cmd/pidshooter/main.go:29-33` already knows how to handle for a sibling adapter — or, at minimum, fall back to `os.TempDir()` rather than `"."` so a `$HOME`-less run can't pollute the user's working tree, and log the fallback once a logger is available.

---

### Finding 5: `NewStore()` hardcodes `defaultPath()` with no injectable path — verified 0% coverage on the constructor itself

**Status: Fixed (2026-09-10).** Split into `NewStore() (*Store, error)` (resolves the default path, can fail) and `NewStoreAt(path string) *Store` (explicit path, infallible) — exactly the suggested fix below. All tests now go through one of these two constructors instead of `&Store{path: ...}`. Verified via `go tool cover`: `NewStore`/`defaultPath` went from 0.0% to 100% coverage.

**Evidence:**
```go
// filescore.go:35-37
func NewStore() *Store { return &Store{path: defaultPath()} }
```
Every test bypasses this via `&Store{path: filepath.Join(t.TempDir(), "scores.json")}` (`filescore_test.go:15-18`).

**Reproduced on this machine (2026-09-09), `go tool cover`:**
```
NewStore       0.0%
defaultPath    0.0%
defaultDir    75.0%   ← covered only as a side effect of Finding 1's bug, not by any test targeting it
makeConfigDir 80.0%   ← same
```
The adapter's only exported constructor, and the path-resolution logic behind it, are completely untested — which is exactly why Finding 4's `os.UserHomeDir()` failure branch has never been exercised, and why Finding 1 went unnoticed for as long as it has.

This is the same class of question `docs/osprocess-findings.md` Finding 13 raised for `osprocess.NewProcess()` — there, the owner declined to change the constructor's return type, deferring to a planned wiring revisit. The situations differ in one relevant way: `osprocess` already has a working injectable seam (`psPath`) reached by a constructor that can fail and is therefore exercised; here, the seam (`path`) exists on the struct but no exported constructor reaches it at all, so the constructor itself is 0%-covered, not just returning a wider type than ideal.

**Suggested fix:** Add an explicit-path constructor alongside the zero-config one:
```go
func NewStore() (*Store, error)        // resolves defaultPath(); can fail, per Finding 4
func NewStoreAt(path string) *Store    // explicit path — used by tests, and any future --scores-file flag
```
`NewStore` becomes a thin wrapper over `NewStoreAt(defaultPath())`. This must land together with Finding 1 — a non-default path doesn't work for `Save` until that bug is fixed.

---

## 4. Minor findings

| # | Finding | Verified? | Status |
|---|---|---|---|
| 6 | **Cross-process last-writer-wins.** Confirmed no *in-process* race first — traced every call site: `application/runner.go` calls `Load` once (`:56`) and `Save` once (`:68`), strictly sequentially on one goroutine, separated by the blocking `Play` call; `application/game.Service` never references the score store at all (its kill-verification goroutines touch only `application/process`). **No mutex is needed for the current wiring.** But `Load`-at-session-start / `Save`-at-session-end means the read-modify-write window spans an entire game session (unbounded with `--time=0`) — two overlapping `pidshooter` runs will have the second `Save` silently discard the first player's entire session, every time, not as a narrow race but as a guaranteed outcome of the whole-board design. | Traced through all real call sites; not executed as a live two-process repro. | Fixed (2026-09-10) — reload-merge-save, no locking. `Service.RecordScore` no longer saves the board loaded at session start; it re-`Load()`s the persisted board immediately before `Save`, applies just this session's new entry to that fresh copy (`Service.mergeWithLatest`), and saves that instead. A concurrent save from another process is now merged with, not discarded by, this one. Still not fully race-free — two `Save` calls landing at the exact same instant could still interleave — deemed an acceptable residual for a single-user CLI game rather than adding file locking (which would otherwise need to be held for an unbounded, whole-session duration). Regression test: `TestRecordScoreMergesWithConcurrentlyPersistedEntries`. |
| 7 | **`0600`/`0700` permissions are applied only on file/directory creation**, never tightened on a pre-existing looser-mode file or directory (`os.WriteFile`'s `perm` and `os.MkdirAll`'s `mode` are both ignored by POSIX `open(2)`/`mkdir(2)` when the target already exists). `docs/fable-review.md` S6's claim that permissions are "correctly restrictive" holds only for the fresh-install path; `TestSaveFilePermissions` (`:41-47`) only exercises that path. | **Reproduced on this machine:** `Save` over a pre-existing `0644` file left it `0644`; `MkdirAll(0700)` over a pre-existing `0777` dir left it unchanged. | Fixed (2026-09-10) — the file-permission half was already fixed as a side effect of Finding 2 (a renamed-over file is always freshly `Chmod`'d). The directory-permission half is now fixed too: `fsutil.WriteFileAtomic` calls `os.Chmod(dir, 0700)` unconditionally right after `os.MkdirAll`, so a pre-existing looser-mode directory gets tightened on every write, not just left alone. Regression test: `TestWriteFileAtomicTightensExistingDirectoryPermissions`. (This also required redesigning `TestWriteFileAtomicLeavesExistingFileUntouchedOnFailure` — it used to force a failure by restricting `dir`'s own permissions, which this fix now unconditionally resets, so the failure is forced via the parent directory's execute bit instead.) |
| 8 | **Inconsistent error handling across `Load`/`Save`'s six return paths.** Three of six are bare passthroughs with no added context (a logged "unexpected end of JSON input" names no file); `makeConfigDir` wraps with `%v` (`:117`) rather than `%w`, silently breaking `errors.Is`/`errors.As` chain-walking through that path; and `Load`'s not-found check uses the pre-1.13 `os.IsNotExist`, which — per its own doc — does not unwrap errors, making it a latent trap the moment any of the other findings here add a `fmt.Errorf("...: %w", ...)` wrap upstream of it. Contrast `osprocess.go`, where every `ps`-invocation error path is funneled through one context-adding helper (`docs/osprocess-findings.md` Finding 9). Currently latent, not live — the sole caller (`apperror.NewError`) re-wraps with its own context regardless, so no information is lost end-to-end today. | Confirmed by direct code reading; not independently re-executed (no live-bug claim being made). | Fixed (2026-09-10) — the `makeConfigDir` `%v`-wrap defect no longer exists (function deleted; `fsutil.go` wraps consistently with `%w` throughout). The pre-1.13 `os.IsNotExist` non-unwrapping trap is also closed: `Load` now uses `errors.Is(err, fs.ErrNotExist)`. The remaining bare-passthrough branch for non-`NotFound`/non-corrupt errors (e.g. `EACCES`) is intentionally left as-is — `os.ReadFile`'s own `*fs.PathError` already includes the file path in its `Error()` string, so nothing is actually missing there; and `NotFoundError`/`CorruptedDataError` being equally terse (no path, no cause) is now a deliberate, symmetric convention rather than an oversight (see Finding 3's status). |
| 9 | **Triple-duplicated struct definitions** (`core/score.Entry` / `outbound.ScoreEntry` / `filescore.entry`) with four hand-written mapping functions. Both reviews, independently, judged this an *earned* translation boundary, not accidental coupling: core↔port avoids a reverse import from `infrastructure` into `core` (the property `docs/hexagonal-arc.md` names as central); port↔wire-format keeps JSON tags out of the port so a hypothetical second `ScoreStore` implementation (SQLite, remote) doesn't inherit file-format concerns it has no use for, and lets the on-disk schema evolve independently of the port. **Verdict: keep the duplication, but it is currently under-verified** — see Finding 4 in §2 (round-trips check 1 of 6 fields; no test asserts the actual on-disk JSON bytes against the `json:` tags that are this adapter's real compatibility contract). | Structural judgment call, not an empirical claim. | Fixed — the under-verification gap is closed: round-trip tests now assert full 6-field equality, and `TestSaveWritesJSONMatchingOnDiskSchema` asserts the raw on-disk JSON against the `json:` tags. The "keep the duplication" verdict itself stands unchanged (it was always an endorsement, not a defect). |
| 10 | `toBoard`/`toScoreBoard` are defined **twice**, in two different packages (`application/score/converter.go` and `filescore/filescore.go`), mapping between different pairs of types and in some cases opposite directions of meaning between the two files. Purely a naming collision — reading `service.go:59`'s `s.store.Save(toScoreBoard(board))` next to `filescore.go:62`'s `toBoard(sb)` reads like inverse operations chained together, and they aren't. No functional defect. | Confirmed by direct code reading. | Accepted, won't fix (2026-09-10) — both functions are unexported, so there's no actual ambiguity or compiler risk, only a human-readability nit when tracing code across both files; this finding's own text already concedes "no functional defect." A rename was tried (`toWireBoard`/`fromWireBoard`) and reverted — the owner preferred the original names once weighed against the (very small) benefit. Sharing the conversion logic instead of renaming was also considered and rejected: the two pairs convert between genuinely different type pairs for the reasons Finding 9 already establishes (`outbound.ScoreBoard`↔`*core/score.Board` vs `outbound.ScoreBoard`↔the adapter-private JSON wire type), and no neutral package can see both without either `filescore` importing `application/score` (no adapter in this codebase depends on the application layer today) or the domain/adapter layers reaching across each other — real architectural coupling that isn't worth avoiding ~15 lines of duplicated field-copying. |

## 5. Low findings

| # | Finding | Status |
|---|---|---|
| 11 | **No override for the score file's location** (no env var, no CLI flag) and no use of `os.UserConfigDir()`/`$XDG_CONFIG_HOME` — the hand-rolled `filepath.Join(home, ".config", "pidshooter")` gets `$XDG_CONFIG_HOME` wrong unconditionally (ignores it even when set) and gets Windows wrong (`docs/osprocess-findings.md` Finding 10 already establishes Windows is a real, if untested, compile target for this codebase — see also its Finding 15 on `runtime.GOOS` branching). **Reviewer disagreement, reconciled:** Fable additionally flagged the macOS row (`~/Library/Application Support` vs `~/.config`) as wrong; Opus pushed back, arguing `~/.config` is increasingly idiomatic for CLI/terminal tools on macOS specifically and arguably the *better* fit for a game that's never going to be a GUI app — this document sides with Opus's framing: the macOS row is not treated as a defect, only the ignored-XDG-override and wrong-on-Windows rows are. | Partially fixed (2026-09-11) — `fsutil.ConfigDir` now honors `$XDG_CONFIG_HOME` when it's set to an absolute path (per the XDG Base Directory spec, a relative value is ignored and the `~/.config` fallback is used instead), falling back to `$HOME/.config/<appName>` otherwise. The Windows row is deliberately **not** addressed — Windows is not a supported platform for this project, so that half of the finding is out of scope rather than deferred. |
| 12 | **No schema/version field.** `encoding/json`'s default behavior (ignore unknown keys, zero-fill missing ones, no `DisallowUnknownFields`) means a future field rename or removal would silently and permanently erase that column from every historical entry the next time a legitimate `Save` runs — and no current test would catch it, since none assert the on-disk shape (§2, and Finding 9 above). Purely structural; no schema change has shipped yet to trigger this. | Fixed (2026-09-11) — `board` now carries a `version` field (`filescore.go`'s `currentSchemaVersion`, currently `1`), stamped on every `Save`. A missing/zero `version` (pre-existing files, or a legacy install) is treated as version 1 and loads normally — no migration needed yet, since the schema hasn't changed since versioning was added. `Load` rejects a `version` greater than `currentSchemaVersion` (a file written by a newer pidshooter build, encountered on a downgrade) with a plain unwrapped error — not `NotFoundError`, and deliberately not `CorruptedDataError` either, since the data is valid, just unrecognized; this routes it through the existing "don't overwrite on a real load failure" gate in `application/score.Service.RecordScore` (`service.go:67`) with no service-layer changes needed, so an old binary can never silently round-trip away fields a newer one wrote. Regression tests: `TestLoadAcceptsFileWithoutVersionField`, `TestLoadRejectsNewerSchemaVersion`, and `TestSaveWritesJSONMatchingOnDiskSchema` now also asserts the top-level `version` key. The migration step itself (for when `currentSchemaVersion` actually bumps) is not yet written — there's nothing to migrate from yet — but the field and the version-gate now exist so a future rename has somewhere to hook in. |
| 13 | `NewStore`'s doc comment claims it "returns an `outbound.ScoreStore`"; it returns `*Store` (the better choice — Go's "accept interfaces, return structs," and what makes Finding 5's fix easy) — stale comment, not a defect in the code. No adapter in this codebase asserts port conformance at compile time (`grep -rn "var _ outbound"` returns zero hits repo-wide) — conformance is currently checked only transitively, at `main.go:44`, so a port-signature drift would surface as a `main.go` compile error rather than one localized to the adapter that actually broke. Separately: `Load`/`Save` — the adapter's entire implementation of the port — carry no doc comments at all, unlike every other exported method in the infrastructure layer (`stderrlog.go`, `osprocess/process.go`); this is exactly where file-specific operational notes (which directory gets created, non-atomicity, the not-found/corrupt distinction) belong, per the project's stated rule that *port*-side comments should stay contract-only. | Partially fixed — the stale doc comment is corrected (`NewStore`/`NewStoreAt` now accurately say they construct a `*Store`), and `Load`/`Save` now carry the doc comments this finding says belong there (directory creation, atomicity, the not-found/corrupt distinction). The repo-wide `var _ outbound.X` conformance-assertion gap is unchanged — applies to every adapter, not just this one, so it's out of scope here. |
| 14 | **No timeout/cancellation on the two file syscalls**, unlike `osprocess`'s subprocess calls (`docs/osprocess-findings.md` Finding 5, fixed via `exec.CommandContext`). Flagged deliberately as a **considered, not-recommended** fix: a wedged network home directory (NFS/SMB) could hang `Load` at startup or `Save` at the score screen, but Go's blocking file syscalls — unlike a subprocess — cannot be cancelled by a `context`; the only way to "time out" would be to abandon the call on a goroutine, leaking that goroutine and its file descriptor for as long as the mount stays wedged. The cure is worse than the rare disease here; recorded so a future pass doesn't "fix" this by reflex from the `osprocess` precedent. | No change — deliberately not implemented, per this finding's own "considered, not-recommended" conclusion; nothing in this pass revisited that reasoning. |
| 15 | `os.ReadFile` is unbounded. Harmless today — the board is capped at 10 entries (`core/score/board.go:10`) so a legitimate file is ~1.5KB, and the file lives at `0600` under the user's own home (self-DoS only) — but a generous size cap (e.g. 1MB) before parsing would double as a cheap, cooperative corruption signal, composable with Finding 3's recovery path. | Fixed (2026-09-11) — `Load` now rejects a file over `maxScoreFileSize` (1 MiB) before calling `json.Unmarshal`, via the same `outbound.CorruptedDataError` path Finding 3 established (safe to overwrite, same as any other corrupt file). To keep the size-limit case distinguishable from an actual JSON parse failure in a log line — both would otherwise be the identical bare `"corrupted data"` string — `outbound.CorruptedDataError` gained an optional `Message` field, prefixed onto `Error()`'s text when set; matching via `errors.As` is unaffected since it's type-based, not value-based, so the safety gate in `application/score.Service` needed no changes. Regression tests: `TestLoadRejectsFileOverMaxSize`, plus new coverage for `CorruptedDataError.Message` itself in `outbound/error_test.go`. |
| 16 | `internal/testutil/fake/store.go`'s `Store` test double has two latent (currently harmless) fidelity gaps versus the real adapter: it can represent a state the real `Load` cannot construct (a populated `Board` returned alongside a non-nil `LoadErr` — the real adapter always returns an empty board on any error), and `Save` captures the board by reference (`f.Saved = &b`) rather than snapshotting, unlike the real adapter which serializes to bytes at call time. Traced and confirmed harmless today: nothing constructs the impossible state, and `application/score/converter.go:24`'s `toScoreBoard` always allocates a fresh slice before `Save` is called, so there's nothing live to alias. Worth a one-line fix (`return outbound.ScoreBoard{}, f.LoadErr` when `LoadErr != nil`; deep-copy in `Save`) only if/when a future test would otherwise rely on the gap. | Not addressed — explicitly out of scope for this pass; `internal/testutil/fake/store.go` untouched. |
| 17 | `Entry.Time` (the configured session time limit) is persisted through all four layers (`core/score.Entry` → `outbound.ScoreEntry` → `filescore.entry` → JSON `"time_limit"` and back) but is never read by anything: `Entry.beats` ranks on kills/speed/duration/freed-memory only, and the score table's column headed "Time" (`board.go:52`) actually displays `Duration`, not `Time` — a reader could reasonably assume the two are the same value. Not a bug (persisting it now is reasonable in case a future ranking or display change wants it), but worth a comment or a column rename to prevent the "Time" ↔ `Time` field confusion from compounding. | Not addressed — `Entry.Time` is still persisted but unread, and no clarifying comment/rename was added. |

---

## 6. Relationship to `docs/fable-review.md`

- **S6** ("Score file written non-atomically") is confirmed still present and is superseded in detail by Finding 2 above, which traces its actual consequence (a self-perpetuating save-lockout, not merely "corruption") through real caller code and adds a concrete reachability path (self-inflicted `SIGKILL` via a second `pidshooter` instance) that S6's original one-line description didn't have. **Fixed 2026-09-10, together with Finding 2 above.**
- **D1/D2** (console I/O in `core/score` and `application/score`) are unchanged and out of scope for this document; noted in §1 only because both reviews passed through the same files while tracing this adapter's callers.

## 7. Where the two independent reviews disagreed, and how it was resolved

1. **Severity of the non-atomic-write finding** (Finding 2): Fable → Minor, Opus → Moderate. Both reached the identical "self-perpetuating lockout" mechanism independently by tracing the same code path, which is a stronger signal than either review's severity label alone. Resolved at **Moderate**, weighing the convergent mechanism, the verified self-inflicted-`SIGKILL` reachability path, and the triviality of the fix against the severity of "permanent silent data loss requiring manual file deletion."
2. **Severity of the constructor/testability finding** (Finding 5): Fable → Minor, Opus → Moderate. Resolved at **Moderate** based on the verified 0% coverage on `NewStore`/`defaultPath` (an objective measurement, not a judgment call) and because it's the shared root blocker for cleanly fixing both Finding 1 and Finding 4.
3. **Whether hardcoding `~/.config` is "wrong" on macOS** (Finding 11): Fable said yes (should be `~/Library/Application Support`, matching `os.UserConfigDir()`'s per-platform behavior); Opus disagreed, arguing `~/.config` is the more idiomatic choice for a terminal-only CLI game on macOS specifically. This document adopts Opus's framing — the macOS row is not counted as a defect — while keeping both reviews' agreement that ignoring `$XDG_CONFIG_HOME` and mishandling Windows are real, unrelated gaps.

## 8. Linux applicability

Both reviews, and this document's own re-verification, were produced on macOS (Darwin). Asked directly whether the same findings hold on Linux: **almost all of them do, identically, and this can be established from stdlib source rather than needing a live Linux machine** — the deciding factor being that `filescore.go`, unlike `osprocess.go`, contains **zero `runtime.GOOS` branches** and calls nothing outside portable `os`/`encoding/json`/`filepath`.

**Verified by reading Go 1.26.5's stdlib** (`$GOROOT/src/os/`) rather than assumed: none of `WriteFile`, `MkdirAll`, or `UserHomeDir` — the three functions every Moderate finding above depends on — has a `_darwin.go`/`_linux.go` override file. All three live in shared files (`file.go`, `path.go`) with, at most, a `switch runtime.GOOS` that treats every non-`windows`/`plan9`/`android`/`ios` target (Linux and Darwin both fall through to the same default) identically:

- `os.UserHomeDir()` (`file.go:605-624`) — reads `$HOME` on both Linux and Darwin (only `windows`/`plan9` differ). **Finding 4's fallback-to-cwd bug applies identically to Linux.**
- `os.WriteFile()` (`file.go:932-942`) — `O_TRUNC`-then-write, no `fsync`, on every platform. **Finding 2's non-atomic-write consequence applies identically to Linux**, and the self-inflicted-`SIGKILL` reachability argument transfers at least as cleanly: `docs/osprocess-findings.md` confirms Linux's `ps -eo ...comm` reports `pidshooter`'s full, untruncated name (it's only 10 characters, under Linux's 15-character `comm` cap), so a second same-user `pidshooter` instance is discoverable and killable the same way.
- `os.MkdirAll()` (`path.go:19-27`) — returns immediately, without `Chmod`, whenever `Stat` shows the target already exists as a directory; this is plain Go control flow layered on `Stat`, not a platform-conditional syscall wrapper. **Finding 7's not-tightened-on-existing-directory claim applies identically to Linux** (the not-tightened-on-existing-*file* half of Finding 7 rests on POSIX `open(2)`'s `mode`-ignored-without-`O_CREAT` semantics, which both Linux and Darwin implement per POSIX).

By the same reasoning, **Findings 1, 3, 5, 6, 8, 9, 10, 12, 13, 15, 16, and 17 are pure Go-level logic or POSIX-uniform error semantics** (`ENOTDIR`, permission bits, `errors.As`/`errors.Is` chain-walking, `encoding/json`'s unknown-field behavior) with no macOS-specific mechanism in any of them — nothing in this review's evidence for any of these findings cited a Darwin-only API, unlike `osprocess-findings.md`'s Finding 1 (`$COLUMNS`-aware `ps` truncation) or Finding 4 (`argv[0]`-derived `comm`), which were genuinely BSD-`ps`-specific and turned out to *not* reproduce on Darwin the way they did on Linux. There is no equivalent OS-specific mechanism anywhere in `filescore` for a platform difference to hide behind.

**The one finding that is inherently platform-relative is Finding 11** (config-directory convention) — and the Linux case is narrower, not wider, than the macOS one this document already debated: `~/.config` is already the *correct*, spec-compliant default on Linux (it's literally the XDG fallback), so Linux's only real miss is the same one already noted — the code never checks `$XDG_CONFIG_HOME` before falling back to `~/.config`, so a Linux user who *has* set that variable (common enough to be worth respecting) is silently ignored. There's no macOS-style "which directory convention is more correct" debate to have on Linux; it's a strict subset of the already-documented gap.

**What this section is not:** a live Linux execution pass. No Linux environment was available in this session (`docker info` failed — daemon not running) to reproduce Finding 1's `$HOME`-pollution repro or Finding 7's permission probes the way `docs/osprocess-findings.md`'s macOS section re-ran its checklist on real hardware. The claims above rest on reading the actual stdlib source for the functions in question (which is platform-*independent* reasoning — the same file compiles into both the Linux and Darwin binaries) rather than on inference from documentation or general POSIX familiarity. If a Linux machine becomes available, the cheapest high-value re-run would be Finding 2's `$HOME`-scoped repro (`HOME=<scratch> go test ./internal/infrastructure/filescore/... -run TestSave`) and Finding 7's permission probes, purely to convert "verified via shared stdlib source" into "verified via shared stdlib source and independently reproduced," matching this repo's existing bar for cross-platform claims.
