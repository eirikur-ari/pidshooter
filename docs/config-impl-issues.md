# Config Implementation — Design & Functional Review

**Date:** 2026-09-27
**Method:** Full review of everything changed on branch `add-config-file` relative to `master` (the persisted `config.yaml` feature: `application/config`, `outbound/config.go`, `infrastructure/filestore`, the `application/runner` package, `entrypoint/cli`, `apperror.Code.In`), run on Claude Opus with full repo context and empirical verification (`go build`/`go vet`/`make test`, probe scripts against real config files, and a diff against `master` to confirm which behaviors are regressions vs. pre-existing). Findings 1, 2, and 4 below were independently re-verified against `master` and current code in this session (Sonnet, orchestrating) before being recorded.
**Scope:** `internal/application/config/` (`Request`, `Result`, `Service`), `internal/application/contract/outbound/config.go`, `internal/infrastructure/filestore/` (`config_file.go`, `score_file.go`, `file.go`, `converter.go`), `internal/application/contract/inbound/runner.go`, `internal/application/runner/`, `internal/entrypoint/cli/` (`program.go`, `flag_mapper.go`, `flag_splitter.go`), `internal/application/apperror/error.go`, `internal/composition/runner_creator.go`, `internal/core/game/session.go` (`Config`).
**Status:** Finding 1 resolved as a non-issue 2026-09-27 (see its updated note below — the first fix attempt was itself incorrect and was reverted). Finding 2 resolved as won't-fix 2026-09-27 (intentional behavior — see its updated note below). Finding 3 fixed 2026-09-27. Finding 4 fixed 2026-09-27 (doc comment only — the implementation was already correct). Finding 5 fixed 2026-09-28. Finding 6 partially resolved 2026-09-28 (see its updated note below — the empty-file half is intentional, not a defect; only the negative-version half was fixed). Finding 8 fixed 2026-09-28. Finding 9 fixed 2026-09-28. Finding 10 fixed 2026-09-28. Finding 11 fixed 2026-09-28. All other findings below are open (unfixed) as of this writing.

---

## Functional bugs

### ~~Finding 1: Argument validation now happens after the terminal screen is constructed — confirmed regression against `master`~~ ✓ Non-issue, resolved 2026-09-27

**Status: Resolved as a non-issue (2026-09-27).** A first fix attempt added an early `req.Config.Validate()` call in `program.go` before `p.creator.Create()`, mirroring `master`'s ordering — but this was wrong on two counts, caught in review: (1) it only validated `Game.Speed`/`Game.TimeLimit`, missing `Patterns` entirely (which validates elsewhere, inside `process.Service.FindProcesses`), and (2) it duplicated the *already-existing* `apperror.CodeInvalidConfig.In(err)` check further down in the same function, making that check dead code for the real CLI path. On reflection, `p.creator.Create()` only constructs/wires adapter instances (including the tcell screen) — nothing executes as a side effect of construction that validation would need to guard against, unlike `master`-era reasoning which treated screen construction as something worth avoiding on bad input. Since `process.Service.FindProcesses` and `config.Service.Load` already classify their respective validation failures identically (`apperror.CodeInvalidConfig`, `apperror.SeverityFatal`), both already flow correctly through `runner.Service.Run`'s returned error into the pre-existing `CodeInvalidConfig.In(err)` check in `program.go` — no code change was actually needed. The first attempt was fully reverted (`config.Request.Validate()` removed, `config.Service.Load` reverted to inline-wrap `validate()`'s error itself, `program.go`'s early check removed).

**Verified.** On `master`, `internal/entrypoint/cli/program.go`'s `NewProgram` doc comment states the invariant directly: "`creator.Create` is called only once argument parsing and configuration validation have both succeeded, so a service that's expensive or fallible to construct never affects `--help`, a parse error, or a rejected configuration." Concretely, `master`'s `run()` calls `cfg.Validate()` *before* `p.creator.Create()`.

On this branch, `program.go:65-75` no longer validates speed/time in `prepare`; validation moved into `config.Service.Load` (`internal/application/config/service.go:26`), which only runs inside `runner.Service.Run` — i.e. strictly *after* `p.creator.Create()` has already built the tcell screen.

**Reproduced:** `env -u TERM pidshooter sleep --speed=99`.
- `master` → `error: invalid configuration: speed must be between 0.5 and 5, got: 99` + usage text, exit code **2**.
- This branch → `error: failed to create screen: terminal entry not found`, exit code **1**.

Any non-TTY caller (CI, a pipe, `TERM` unset) gets a screen-construction error instead of the real argument error in the case where both problems coincide, and a bad flag value allocates a terminal (an expensive, fallible resource) before being rejected. Pattern-length validation has the same characteristic — it lives in `process.Service.FindProcesses`, also reachable only after `Create()`.

**Accepted, no fix needed:** in the normal case (a real interactive TTY, which is this program's only legitimate use case besides `--help`), `Create()` succeeds regardless of whether the request is valid, and the actual validation error surfaces correctly afterward via the existing `CodeInvalidConfig.In(err)` check. The only observable difference from `master` is in the coincidental case of *simultaneously* invalid input and no usable terminal — judged not worth the complexity of guarding against, since `Create()` performs no side effect beyond object construction.

---

### ~~Finding 2: "Config file values ignored" warning fires even when a CLI flag overrides the invalid file value~~ ✓ Won't fix — intentional, 2026-09-27

**Status: Won't fix — this is intentional (2026-09-27).** A fix was attempted (making `validateStore` skip reporting a field the request also overrides) and then reverted at the owner's direction: the warning is deliberately unconditional. A bad value in the persisted config file is worth surfacing regardless of whether a CLI flag happens to override it for the *current* run — the file stays broken, and a future run without that flag would silently fall back to the domain default instead of the user's intended value. Suppressing the warning whenever an override happened to be present would hide that. Reverting this "fix" restored the original always-warn behavior.

**Original report, for reference:** `internal/application/config/result.go:61-64`:

```go
func (cfg Result) apply(stored outbound.ConfigStoreResult, req Request) (Result, error) {
	cfg, err := cfg.fromStore(stored)
	return cfg.fromRequest(req), err
}
```

`apply` returns whatever error `fromStore` produced, unconditionally — even when `fromRequest` immediately afterward overwrites the exact field that error was about.

**Reproduced:** config file contains `game.speed: 99.0` (invalid — out of `[0.5, 5.0]`); run `pidshooter foo --speed=3.0`. Final `Result.Game.Speed` is correctly `3.0` (the CLI override wins), but the run still prints `warning: config file values ignored: speed: ...` about a value that was never actually used for anything.

**Suggested fix:** Only surface a stored-field-rejected warning for fields the request didn't also explicitly override, e.g. compute the warning after `fromRequest` has run, or skip validating/reporting a stored field whose corresponding `Request` field is non-nil.

---

### ~~Finding 3: No `KnownFields(true)` on the YAML decoder — misspelled/misplaced keys are silently dropped~~ ✓ Fixed 2026-09-27

**Status: Fixed.** `config_file.go`'s `Load` now decodes via `yaml.NewDecoder(bytes.NewReader(data)).KnownFields(true)` instead of `yaml.Unmarshal`, so an unrecognized key becomes a decode error, reported the same way as any other malformed file (`outbound.CorruptedDataError`, surfacing as a `apperror.SeverityWarning`). An empty file is deliberately exempted (`len(data) > 0` guard before decoding) and still treated as "no config, use defaults" — `yaml.v3`'s `Decoder.Decode` returns `io.EOF` on empty input, which would otherwise have turned an intentionally-empty file into a spurious warning too. Separately, `config.Service.Load`'s warning message was corrected from "config defaults not loaded" (which is wrong — a default `Result` is always available regardless) to **"config store not loaded"**, naming the actual tier that failed. Regression tests added: `TestConfigLoadUnknownKeyReturnsCorruptedDataError`, `TestConfigLoadEmptyFileReturnsNoError`.

**Original report, for reference:** `internal/infrastructure/filestore/config_file.go:82` calls `yaml.Unmarshal` directly, with no `yaml.Decoder.KnownFields(true)`. A file containing a typo'd key, e.g.:

```yaml
game:
  timelimit: 55
  include_root: true
```

loads with no error and no warning — both keys are silently ignored (the correct key is `time_limit`, and `include_root` belongs under `process`, not `game`). This matters more than the same gap would in most codebases because `ConfigStore.Save` currently has **zero production callers** (grep confirms only tests call it) — hand-editing is the *only* path that ever produces this file, so there is no typo feedback anywhere in the loop.

---

### ~~Finding 4: `apperror.Code.In`'s doc comment ("is, or wraps") overstates what `errors.As` actually checks~~ ✓ Fixed 2026-09-27 (doc only)

**Status: Fixed — doc comment only, implementation was already correct.** Considered making `In` walk the whole chain (checking every `*Error`, not just the first), but that would have been a real bug, not a fix: nesting one `*Error` as another's `Cause` (e.g. `runner.Service.logKillFailures` wrapping a `CodeProcessDiscoveryFailed`/etc. error inside a new `CodeKillFailed` one) represents "problem B occurred, caused by problem A" — two distinct classified events, not one event wrapped with extra context. Walking the whole chain would let an unrelated `Code` buried in some inner cause falsely match, misclassifying the outer (and actually-relevant) error. Stopping at the first `*Error` — `errors.As`'s actual behavior — is correct. Only the doc comment was wrong; it now states the real, narrower contract.

**Original report, for reference:** `internal/application/apperror/error.go:19-23`:

```go
// In reports whether error is, or wraps, an *Error with this Code.
func (c Code) In(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == c
}
```

`errors.As` stops at the *first* value in the chain assignable to `*Error` — it does not keep unwrapping past that match to find a *different* `*Error` deeper in the chain with a matching `Code`. So `CodeInvalidConfig.In(NewError(CodeKillFailed, SeverityWarning, "", someInvalidConfigErr))` returns `false` even when a `CodeInvalidConfig` error genuinely exists deeper in the chain — the doc's "or wraps" claim is only true when the *outermost* `apperror.Error` in the chain is the one with the matching code.

**Not reachable today:** tracing the current call graph, nothing that reaches `program.go`'s sole `CodeInvalidConfig.In(err)` check nests one `*apperror.Error` inside another. But `runner.Service.logKillFailures`/`logDuds` already wrap errors that themselves originate as `apperror.Error`s in some paths — this is one future change away from silently mismatching.

---

### ~~Finding 5: `ConfigStoreResult.Mode` is parsed and round-tripped but never validated or consumed~~ ✓ Fixed 2026-09-28

**Status: Fixed.** `config.Result` now has a `Mode` field (a new, local `config.Mode` type — deliberately not the `outbound.Mode` DTO, keeping the same decoupling `GameResult`/`ProcessResult` already have from `GameConfig`/`ProcessConfig`), defaulting to `ModeGame` in `newResult()`. `validateStore` (`internal/application/config/result.go`) now also checks `stored.Mode` against the known `outbound.ModeGame`/`ModeYolo`/`ModeList` values; an unrecognized value is rejected (kept at the current default) and reported the same way an invalid `speed`/`time_limit` already is — via the existing warning path, not a new one. `fromStore` overlays a valid, non-empty stored `Mode` onto `cfg.Mode`. Regression tests added: `TestResultApplyOverlaysValidStoredMode`, `TestResultApplyRejectsUnrecognizedStoredModeAndKeepsDefault`; existing `Service.Load` tests updated to expect `Mode: ModeGame` as part of the default `Result`.

**Original report, for reference:** `internal/application/contract/outbound/config.go:32` and `filestore/converter.go:42` carry `Mode` through the YAML round-trip, but `config.Result` has no `Mode` field and nothing reads it anywhere. A file containing `mode: wobble` loads with no error, producing `Mode("wobble")`, which is then silently discarded. Dead data with no validation — likely a forward-looking field for planned CLI modes (see project memory on planned game/yolo/list modes) that hasn't been wired up yet, but as-is it's an unvalidated no-op.

**Suggested fix:** Either validate `Mode` against its known set of values at load time (even if unused downstream, so a typo is surfaced) or leave a comment noting it's intentionally unused pending the CLI-modes feature.

---

### ~~Finding 6: Empty config file and negative `version` both pass silently~~ Partially fixed 2026-09-28 — negative version rejected; empty file confirmed intentional

**Status: Partially fixed.** The two halves of this finding turned out to have different answers:

- **Empty file treated as "no config file"**: confirmed **intentional, not a defect** — this codebase deliberately treats "no config file" as a fully supported, normal state (see `ConfigStore.Load`'s doc comment, and Finding 3's resolution, which already made this an explicit `len(data) > 0` guard rather than an accident). A zero-byte file is indistinguishable from a missing one by design; the theoretical "truncated file from a crash mid-write" scenario the original report raised doesn't actually apply here, since `fsutil.WriteFileAtomic` (temp file + rename) already makes a partial/truncated file on disk impossible via `Save`'s own write path — the only way to get a zero-byte file is deliberate manual truncation, which is equivalent to deleting it. No fix needed; no code changed for this half.
- **Negative `version` passing silently**: **fixed**, and generalized. `config_file.go` and `score_file.go` had byte-for-byte identical version-bounds checks (only checking `Version > current`, both missing a lower bound — `score_file.go` had the exact same gap this finding raised for `config_file.go`). Extracted a shared `schemaVersion` type (`internal/infrastructure/filestore/schema_version.go`) with a `validate(kind string, got int) error` method, used by both `Load` implementations, that rejects `got < 0` in addition to `got > current`. The threshold is `< 0`, not `<= 0` — `version: 0` and a file that omits the `version` key entirely (which also decodes to `0`) both remain valid, since a hand-edited file omitting `version` is a legitimate, already-tested case (`TestConfigLoadAcceptsFileWithoutVersionField`), consistent with "no config file is okay." Regression tests added: `TestConfigLoadRejectsNegativeSchemaVersion`, `TestConfigLoadAcceptsExplicitZeroSchemaVersion`, `TestLoadRejectsNegativeSchemaVersion` (score file), plus direct `schemaVersion.validate` unit tests in `schema_version_test.go`.

**Original report, for reference:**

- A zero-byte `config.yaml` loads as all-nil/all-zero with no error or warning — indistinguishable from "no file at all," so a truncated file (e.g. from a crash mid-write, pre-`fsutil` atomic-write days, or manual corruption) is invisible.
- `version: -5` passes the `cd.Version > currentConfigSchemaVersion` staleness gate (`config_file.go:86`) since it's compared only against the upper bound, not a lower one.

**Suggested fix:** Treat an empty file the same as a missing one only if that's actually the intended semantics (currently true by accident, not by a written check); reject `Version <= 0` explicitly alongside the existing `Version > current` check.

---

## Design / layering issues

### Finding 7: `contract/inbound/runner.go` imports `application/config`

A port package (`contract/inbound`) now depends on a concrete application-layer service package (`application/config`) for `RunRequest.Config config.Request`. This is asymmetric with `contract/outbound/config.go`, which deliberately declares its own `GameConfig`/`ProcessConfig` DTOs rather than importing `application/config`, for exactly the coupling reason this finding raises. This was a conscious tradeoff made during this branch's work (accepted: `inbound`/`config` both live under `internal/application`, and a second inbound adapter would need the identical shape anyway) — recorded here so it's a visible, deliberate decision rather than something to "discover" and re-litigate later.

---

### ~~Finding 8: `flagMapper` aliases the `flag.FlagSet`'s own destination pointers, and isn't reentrant~~ ✓ Fixed 2026-09-28

**Status: Fixed.** Both sub-issues addressed together by restructuring `toRunRequest`: the in-progress `inbound.RunRequest` is now a function-local variable, built inside a closure passed to `flagSet.Visit`, instead of a `req` field mutated across the struct and a separate `visit` method. This eliminates the shared-mutable-state problem structurally (no reset-at-top-of-call needed, no comment required) — confirmed by temporarily reverting the fix and re-running the new concurrency test under `-race`, which reproduced a real data race on the old `m.req` field on the first try, then passed clean once the fix was restored. Each cloned value is now written via `util.ClonePtr`, matching the `converter.go` precedent, so `RunRequest.Config`'s pointers no longer alias the `FlagSet`'s own storage. Regression tests added: `TestFlagMapperToRunRequestClonesFlagSetPointers`, `TestFlagMapperToRunRequestConcurrentCallsDoNotRace` (run under `make test-race`).

**Original report, for reference:** `internal/entrypoint/cli/flag_mapper.go:46-52` assigns `m.confirm`/`m.speed`/etc. (pointers owned by the `FlagSet` itself) straight into the built `RunRequest`, rather than cloning them (contrast `infrastructure/filestore/converter.go`, which clones carefully via `util.ClonePtr` when crossing a similar boundary). Additionally, `flagMapper` holds the in-progress `req inbound.RunRequest` as struct state (`flag_mapper.go:17`), mutated across `toRunRequest`/`visit` — calling `toRunRequest` twice on the same `*flagMapper` would leave stale/mixed state from the first call. Not currently a live bug (each `flagMapper` is used exactly once per CLI invocation), but worth a comment noting the single-use assumption, or a `req` reset at the top of `toRunRequest`.

---

### ~~Finding 9: `validateStore`/`fromStore` are coupled through string literals with no shared constant~~ ✓ Fixed 2026-09-28

**Status: Fixed.** Replaced all four raw string-literal occurrences (`"mode"`, `"speed"`, `"time_limit"`, each written twice — once in `validateStore`, once in `fromStore`) with a `fieldMode`/`fieldSpeed`/`fieldTimeLimit` constant block, referenced from both functions. A rename now has to touch a single shared identifier, so a rename that only updates one side is a compile error instead of a silent `slices.Contains` mismatch.

**Original report, for reference:** `internal/application/config/result.go:76,79` check `slices.Contains(rejected, "speed")`/`"time_limit"` against literals produced at `result.go:114,120`. A future rename on either side (e.g. renaming the reported field name) silently starts *applying* an invalid persisted value instead of rejecting it, since the `slices.Contains` check would just stop matching. No compiler or test would catch a rename that touched only one side.

**Suggested fix:** Define named constants (e.g. `fieldSpeed = "speed"`) shared between `validateStore` and `fromStore`, or restructure `validateStore` to return a `map[string]bool`/set that `fromStore` and the two functions share the same source of truth for.

---

### ~~Finding 10: `usageText` hardcodes stale defaults and doesn't mention the config file at all~~ ✓ Fixed 2026-09-28

**Status: Fixed.** `--speed`/`--time`'s descriptions in `usageText` (`internal/entrypoint/cli/program.go`) no longer claim a fixed numeric default — they now say "defaults to the config file value, or 2.0/30 if unset". A new "Config file:" section documents the file's location (`~/.config/pidshooter/config.yaml`, noting the `$XDG_CONFIG_HOME` override) and that any flag passed on the command line overrides it for that run, with `--include-root=false` as the example.

**One claim in the original report turned out to be wrong, caught before fixing:** the report said "`--include-root` is a bool flag with no way to force it back to `false`... other than editing the file directly." Verified against Go's actual `flag` package behavior (`flag.FlagSet.Visit`/`Bool`) before writing the fix: `--include-root=false` **is** valid syntax and **is** visited by `FlagSet.Visit` (confirmed with a standalone repro), so it already correctly overrides a persisted `true`. The fix documents this working behavior rather than inventing a new flag to solve a problem that didn't exist — see [[feedback_verify_call_graph_before_fixing]]. Separately, also worth noting: `outbound.ConfigStore.Save` still has zero production callers (per Finding 3's note), so a persisted `include_root: true` can currently only come from hand-editing the file, never from running `--include-root` once — the original report's framing ("silently arms... for every future run") slightly overstated how a value would get persisted in the first place.

**Original report, for reference:** `internal/entrypoint/cli/program.go:29-30` hardcodes `(default 2.0)`/`(default 30)` in the `--speed`/`--time` flag descriptions. These are no longer the actual effective defaults for a user with a config file that sets different values — `--help` can now show a number that isn't what running the command bare will actually do. The help text also never mentions that a config file exists, where it lives, or that a persisted `include_root: true` silently arms root-process targeting for every future run until explicitly overridden with `--include-root=false`-equivalent behavior (there currently isn't one — `--include-root` is a bool flag with no way to force it back to `false` once persisted as `true`, other than editing the file directly).

**Suggested fix:** At minimum, soften the flag descriptions to not claim a specific default (e.g. "defaults to the config file value, or 2.0 if unset"), and document the config file's existence and location in `usageText` or `--help`'s output.

---

## Test coverage gaps

### ~~Finding 11: No test proves a persisted config value actually reaches gameplay~~ ✓ Fixed 2026-09-28

**Status: Fixed.** Added a `newServiceWithConfigStore` helper to `service_integration_test.go` (the existing `newService`/`newServiceWithRenderer` now delegate to it, still pinned to `outbound.NotFoundError{}`) and three new integration tests, each with no request-level override for the field under test — proving the value flows from a fake `ConfigStore` result all the way through `config.Service.Load` and `runner.Service.Run`:
- `TestIntegrationServicePersistedIncludeRootIncludesRootOwnedProcess` — a persisted `include_root: true` reaches `process.Service.FindProcesses`.
- `TestIntegrationServicePersistedTimeLimitQuitsGameplay` — a persisted `time_limit` reaches `game.PlayRequest` and actually ends the session (mirrors the existing request-level `TestIntegrationServiceRunWillQuitWhenTimeLimitExpires`, sourced from the store instead). `speed` flows through the identical one-line mapping in `toPlayRequest` as `time_limit`, so this one field stands in for both — a `speed`-specific behavioral assertion would need to observe frame-by-frame movement timing, which trades a solid, deterministic test for a flaky one for no real extra coverage; see [[feedback_pidshooter_avoid_flaky_tests]].
- `TestIntegrationServiceFallsBackAndWarnsOnUnreadableConfigStore` — a `CorruptedDataError` from the store still lets the game play successfully (`Run` returns nil) and is reported through the logger as a warning, not swallowed silently.

**Verified each test is load-bearing, not trivially passing**, by temporarily breaking the corresponding production code and confirming the right test failed: hardcoding `FindProcesses(req.Patterns, false)` in `runner/service.go` failed the include-root test with "no processes found"; hardcoding `TimeLimit: 0` in `converter.go`'s `toPlayRequest` hung the time-limit test (no time limit + no quit event = runs forever), confirmed via a bounded timeout rather than left to hang. Both files were restored immediately after (`git diff` confirmed clean). See [[feedback_pidshooter_avoid_flaky_tests]].

**Original report, for reference:** Every `runner` package test (`runner/service_test.go:68`, `runner/service_integration_test.go:226`) pins the config store to `outbound.NotFoundError{}`. The 3-tier merge (domain default < file < request) is verified in isolation inside `application/config`'s own tests, but nothing exercises the full path: a real persisted `include_root: true` actually reaching `process.Service.FindProcesses`, or a persisted `speed`/`time_limit` actually reaching `game.PlayRequest`. Also untested: a corrupted/unreadable config file with the game still playing correctly on the resulting warning-and-fallback path.

### Finding 12: Untested edge cases matching Findings 3, 5, 6 above

Empty file, unknown/typo'd YAML keys, invalid `mode`, negative `version` — none have a test. `testutil/fake.ConfigStore`'s `Saved`/`SaveErr` fields are dead code from the test double's own perspective — no test calls `Save` through the `outbound.ConfigStore` port at all (only `Load` is exercised).

---

## Checked and confirmed clean

- **Pointer/zero-value handling through the YAML layer.** `time_limit: 0` and `confirm_mode: false` both round-trip correctly as non-nil (yaml.v3's `omitempty` on a pointer field checks `IsNil`, not the pointed-to zero value) — verified on disk.
- **`fsutil.WriteFileAtomic`** — temp file + chmod + fsync + rename + best-effort directory fsync, with cleanup on failure. Sound.
- **Precedence** (domain default < file < request) is correct for all four fields, on every path *except* the spurious-warning case in Finding 2 — the actual *values* used are always correct, only the diagnostic message is sometimes wrong.
- **Severity classification** (Fatal vs. Warning) is otherwise consistent: an invalid request is Fatal; a file load failure or a rejected individual file field is a Warning, with the game still proceeding on a usable `Result`.
- **Concurrent `Save`** can't tear a file — unique temp file + atomic rename. (The score board's separate read-modify-write lost-update window is pre-existing, not introduced by this branch.)

---

## Stale documentation (not a code defect, flagged for follow-up)

`docs/hexagonal-arc.md`, `docs/diagram.md`, and `docs/test-cases.md` still describe types and structures removed on this branch: `config.Config`, `NewRunner(...)`, a free-function `application/runner.go`, the `filescore` package name, and `CodeScoreLoadFailed`. These were last synced 2026-09-22, before the config subsystem existed. Needs a resync pass once the config feature itself settles.

---

## Appendix — file:line references

- `internal/entrypoint/cli/program.go:65-75` (validation timing), `:29-30` (hardcoded defaults)
- `internal/entrypoint/cli/flag_mapper.go:17`, `:46-52`
- `internal/application/config/service.go:25-39`
- `internal/application/config/result.go:61-64`, `:76,79`, `:109-129`, `:36-39`
- `internal/application/contract/inbound/runner.go:3`
- `internal/application/contract/outbound/config.go:3-11,32,44-45`
- `internal/application/apperror/error.go:19-23`; `error_handler.go:27-43`
- `internal/infrastructure/filestore/config_file.go:13,48,82,86,96`
- `internal/infrastructure/filestore/converter.go:42,57`
- `internal/application/runner/service.go:45-49,51`; `service_test.go:68`; `service_integration_test.go:226`
- `internal/testutil/fake/config_store.go:9,19-22`
