# entrypoint/cli + cmd/pidshooter — Cross-Referenced Findings (Fable × Opus)

**Date:** 2026-09-20
**Method:** Two independent from-scratch reviews of `internal/entrypoint/cli/cli.go` (+ `cli_test.go`) and `cmd/pidshooter/main.go`, one on Claude Fable 5, one on Claude Opus, each with full repo context (`docs/hexagonal-arc.md`, the `inbound.Runner` port, `application/config` and its validating `NewConfig` constructor, `application/apperror`, `util.Logger`, `application/runner.go`, `testutil/fake/runner.go`, and the prior full-codebase review in `docs/fable-review.md`). Both reviews built the real binary and drove it through dozens of real argv/env combinations rather than reasoning about behavior in the abstract. Findings below are cross-referenced, deduped, and — for the highest-value/most load-bearing claims — independently re-verified by a third pass (Sonnet, orchestrating) on the same machine.
**Scope:** `internal/entrypoint/cli/cli.go`, `internal/entrypoint/cli/cli_test.go`, `cmd/pidshooter/main.go`.
**Supersedes:** `docs/fable-review.md`'s **F1**, **F2**, **F4**, **D3**, **D7** (all of which touched an old, pre-rewrite version of these files — `cli.go` has since had cobra removed, then a validating `config.NewConfig` introduced) and partially **S1**. See §5 for full reconciliation.
**Platform:** produced and verified on macOS (Darwin, go1.26.5, arm64). Nothing in these findings is platform-specific — no `runtime.GOOS` branching exists in either file.

---

## 1. Architecture conformance

Both reviews agree: **the inbound boundary itself is clean.** `cli.go` imports only stdlib (`errors`, `flag`, `fmt`, `io`, `os`, `strings`) plus `application/apperror`, `application/config`, `application/contract/inbound`, and `util` — never `application/game`/`process`/`score` or anything under `core/*`. Its only reach into the domain is `config.NewConfig`, the validating constructor the application layer owns; the CLI does not re-implement any domain rule itself. This is a genuine improvement over the state `docs/fable-review.md` reviewed, where the CLI owned its own `validate()`.

`main.go` is business-logic-free and matches the documented principle ("main.go is not logic — it is wiring"), with two real deviations both reviews converged on independently:

- ~~`main.go` is the only file outside `infrastructure/tcellui` that imports `tcell` directly~~ — **resolved as a side effect of Finding 1's fix (2026-09-20):** all adapter construction, including the `tcell.NewScreen()` call and its error wrapping, moved into `internal/composition.RunnerFactory`. `main.go` itself now imports neither `tcell` nor any `infrastructure/*` package at all — it is exactly `cli.NewCLI(composition.NewRunnerFactory()).Run(os.Args[1:])` and nothing else.
- ~~The three early-adapter-construction failures are logged via three manual `logger.Error(...)` calls, entirely bypassing `apperror`~~ — **resolved as the same side effect:** those three blocks no longer exist in `main.go`; a `composition.RunnerFactory.Create()` failure is now a plain returned error that flows through `cli.report()` like any other CLI-returned error. See Finding 9, now moot.

**A third point, found only by Opus and worth treating as its own architecture-conformance item:** `cli.go` hardcodes `os.Stdout`/`os.Stderr` at five call sites and constructs a fresh `util.NewLogger()` inline inside `report`, rather than holding an injectable writer/logger field — unlike the sibling adapter `console.ScoreReporter` (`infrastructure/console/score_reporter.go`), which takes a `writer io.Writer` field specifically so its own tests can substitute a `bytes.Buffer` and assert on real output. See Finding 7 below — this is not a stray inconsistency, it's the direct cause of `cli_test.go` having zero assertions on anything the adapter actually prints.

**Verdict: no boundary violations.** Every finding below is about functional correctness, UX/error-handling behavior, test coverage, or code quality — not placement or coupling.

---

## 2. Moderate findings

### Finding 1: Composition root builds every adapter before parsing any argument — `pidshooter --help` fails outright (exit 1, empty stdout) whenever `$TERM` is unset — ✅ FIXED 2026-09-20

**Both reviews found this** (Fable as M4, Opus as Finding 1); this is `docs/fable-review.md`'s **D7**, confirmed still present and materially worse than D7 originally described — D7 said `--help` "pays for, and can fail on" the adapters; measured, it doesn't merely pay, it **cannot succeed at all** in a routine environment.

**Evidence:** `main.go:29-45` constructs `osprocess.NewProcess()`, `filescore.NewFileScore()`, and `tcell.NewScreen()` — in that order — all before `main.go:50`'s `cli.NewCLI(runner).Run(os.Args[1:])`, which is the first point flags are ever inspected.

**Verified independently (Sonnet, this pass):**
```
$ env -u TERM /tmp/pidshooter-verify --help </dev/null
exit=1
stdout bytes: 0
error: failed to create screen: terminal entry not found: term not set
```
`$TERM` is routinely unset in CI runners, `docker build` `RUN` steps, systemd units, cron, and `ssh host 'cmd'`. A packaging/CI smoke test of the form `pidshooter --help >/dev/null` will fail in every one of those. Both reviews additionally confirmed the same failure mode with `$PATH` stripped (`osprocess.NewProcess()` fails first) and, less severely, `$HOME`/`$XDG_CONFIG_HOME` unset (`filescore.NewFileScore()` fails first, but only when *combined* with the other two being satisfied — checked and confirmed this is the one adapter that doesn't break `--help` on its own).

**Recommended fix:** parse first, construct second. Give the CLI adapter lazily-constructed dependencies (e.g. `cli.NewCLI` takes a `func() (inbound.Runner, error)` factory that `play` invokes only after `help`/no-args/parse/validate have all passed), so the invariant becomes: *no adapter is constructed on any code path that doesn't reach `service.Run`.*

**Implemented (2026-09-20):** `inbound.RunnerFactory` (`Create() (Runner, error)`) was added as a new inbound port alongside `Runner` — named `Create` rather than `NewRunner` specifically to avoid colliding, in both name and meaning, with the pre-existing `application.NewRunner` constructor it calls internally. A new package, `internal/composition` — the composition root's wiring logic, allowed to import both `application` and `infrastructure/*` so that `application` itself never has to — holds `composition.RunnerFactory`, the concrete implementation of `inbound.RunnerFactory` that constructs `osprocess.NewProcess()`, `filescore.NewFileScore()`, `tcell.NewScreen()`, and `application.NewRunner(...)`. `CLI` now holds a `RunnerFactory` (field `factory`) instead of a `Runner`, and `run()` calls `c.factory.Create()` only at its very last step — after `help`/no-args/parse-errors/`config.NewConfig` have all already had the chance to return first — so `main.go` shrinks to `cli.NewCLI(composition.NewRunnerFactory()).Run(os.Args[1:])` with zero adapter construction of its own. Verified against the real binary: `--help`, bare invocation, and `--help` with `$PATH` stripped all now exit 0 with full usage output regardless of `$TERM`/`$PATH`/`$HOME`, while an actual run still correctly fails when the terminal really can't be created. Covered by `TestRunNoArgsDoesNotConstructRunner`, `TestRunHelpDoesNotConstructRunner`, `TestRunInvalidFlagsDoesNotConstructRunner`, `TestRunInvalidConfigDoesNotConstructRunner`, and `TestRunReturnsErrorWhenRunnerFactoryFails` (all in `cli_test.go`, using a call-counting `fake.RunnerFactory`), directly pinning both the laziness invariant and the still-open Finding 7 gap (`cmd/pidshooter` itself still has no test file, since `composition.RunnerFactory`'s real adapter construction isn't separately covered here).

This also incidentally simplifies Finding 9: `main.go` no longer has any error-handling logic of its own to bypass `apperror` with — the three former manual `logger.Error(...)` calls are gone along with the eager construction they guarded, and any factory failure now flows through the same single `cli.report()` path as every other CLI-returned error.

---

### Finding 2: `--speed=NaN` still bypasses range validation — ✅ FIXED 2026-09-21

**Both reviews found this identically**; this is `docs/fable-review.md`'s **F2**, still present, unchanged in mechanism despite validation having moved twice (out of `cli.go`, into `config.NewConfig`, and now also independently re-checked by `application.Runner.Run`).

**Evidence:** `internal/core/movement/throttle.go:46-51` — `if speed < MinSpeed || speed > MaxSpeed` — both comparisons are `false` for `NaN`, so it's silently accepted at every layer that calls this function.

**Verified independently (Sonnet, this pass):**
```
$ /tmp/pidshooter-verify zzznoproc --speed=NaN
error: no processes found
```
(reaches process discovery — i.e. passed `config.NewConfig` cleanly, rather than failing with "invalid configuration"). `±Inf` *is* correctly rejected (`Inf > MaxSpeed`/`-Inf < MinSpeed` are both true) — only `NaN` slips through both comparisons.

**Downstream consequence, traced by Opus and not previously documented:** a `NaN` speed makes `movement.Position` computations return `NaN` on the very first tick — every target's on-screen coordinate collapses to `int(NaN) == 0` simultaneously (confirmed via a throwaway probe) — and the failure is **unrecoverable in-game**, because `Throttle.Increase`/`Decrease` compute `NaN ± 0.5 = NaN` and their own clamp comparisons are likewise always `false`, so the `+`/`-` speed keys can't rescue the session. `LowestSpeed` also stays `NaN` and would fail to JSON-marshal when the session tries to save its score.

**Recommended fix:** in `movement.ValidateSpeed`, explicitly reject non-finite values: `if math.IsNaN(speed) || speed < MinSpeed || speed > MaxSpeed`. Fixing it in `core/movement` closes it for the adapter and the port simultaneously — the whole point of having centralized validation there.

**Implemented (2026-09-21):** `movement.ValidateSpeed` now rejects `math.IsNaN(speed)` in addition to the existing range check, exactly as recommended (`±Inf` was already correctly rejected by the pre-existing range comparisons). Covered by `TestValidateSpeedNaN`, `TestValidateSpeedPositiveInf`, and `TestValidateSpeedNegativeInf` in `throttle_test.go`.

---

### Finding 3: `--time` has no upper bound; sufficiently large values silently become "no time limit" — ✅ FIXED 2026-09-21

**Both reviews found this identically**; this is `docs/fable-review.md`'s **F4**, still present, same mechanism, now living in `core/game.ValidateTimeLimit` instead of `cli.go`.

**Evidence:** `internal/core/game/timer.go:69-79` only rejects `limit < 0`. `secondsToDuration` computes `time.Duration(seconds) * time.Second` — `time.Duration` is `int64` nanoseconds, so this overflows past ~9.2×10⁹ seconds. `Timer.Expired()`/`Remaining()` both treat `limit <= 0` as unlimited.

**Exact threshold (Opus measured this precisely):**
```
--time=9223372036           -> limit=2562047h47m16s              (works as expected)
--time=9223372037           -> limit=-2562047h47m16.709551616s   (overflows negative -> treated as unlimited)
--time=9223372036854775807  -> limit=-1s                          (max int, silently unlimited)
```
The only value `flag` itself rejects is one that overflows `int64` at *parse* time (`9223372036854775808` → `"value out of range"`); the entire band below that is silently accepted and inverts the user's intent (they asked for the longest possible limit and got none at all).

**Recommended fix:** add an explicit upper bound to `ValidateTimeLimit` — either a domain-meaningful cap (e.g. 24h) or the exact arithmetic cap (`math.MaxInt64 / int64(time.Second)` = 9223372036), and state it in the help text.

**Implemented (2026-09-21):** rather than an arithmetic-overflow cap, a domain-meaningful one was chosen: `game.MaxTimeLimitSeconds = 300` (5 minutes) — the game is about quickly clearing processes, not hour-long sessions, so a short, deliberately restrictive cap fits the domain better than the widest value that avoids overflow. `ValidateTimeLimit` now rejects `limit > MaxTimeLimitSeconds`; `0` still means unlimited (unaffected — it's a distinct sentinel, not the bottom of the bounded range, so the error message reads "0 (no limit) or between 1 and 300" rather than "between 0 and 300"). Stated in `cli.go`'s `usageText` and in `docs/features.md`. Covered by `TestValidateTimeLimitMaxBoundary` and `TestValidateTimeLimitExceedsMax` in `timer_test.go`.

**Related design question raised and resolved this session:** whether `Throttle`/`timer`/`confirmation` should each validate on their own construction (mirroring `config.NewConfig`), or even whether `ValidateSpeed`/`ValidateTimeLimit` belong in `config.go` instead of their domain packages. Decided against both: the range constants (`MinSpeed`/`MaxSpeed`, `MaxTimeLimitSeconds`) are load-bearing domain facts already used by the types' own runtime behavior (`Throttle.Increase`/`Decrease`'s clamping), so co-locating the validator with them keeps a single source of truth; moving the check into `config.go` would still need to reach into the domain package for those constants, just relocating the comparison away from the numbers it compares against — the same two-places-must-agree risk that caused Finding 4's staleness in the first place. `config.NewConfig` remains the sole orchestration point that calls each domain validator once, at the boundary where external input becomes a trusted `Config`; no additional self-validating constructors were added.

---

### Finding 4: `--speed` help text still says `0.1-5.0` (enforced minimum is `0.5`) — and the mechanism guarantees this drifts again — ✅ FIXED 2026-09-21

**Both reviews found this**; this is `docs/fable-review.md`'s **F1**, half-fixed: `docs/features.md` was corrected to `0.5`–`5.0`, but the CLI's own `--help` output was never updated.

**Evidence:** `cli.go:33` (`usageText`) and `cli.go:66` (the `Float64Var` description) both still read `"Speed multiplier (range: 0.1-5.0)"` against `movement.MinSpeed = 0.5`.

**Verified independently (Sonnet, this pass):**
```
$ /tmp/pidshooter-verify --speed=0.2 zzznoproc
error: invalid configuration: speed must be between 0.5 and 5, got: 0.2
```
The binary's own `--help` and its own validation error disagree with each other.

**The structural reason this will drift again (both models independently found this same mechanism — strong corroboration):** every flag's description string exists in *two* places — `usageText` and the `*Var` registration call — and the second copy is **dead code**. `cli.go:64` sets `fs.SetOutput(io.Discard)`, and neither `fs.PrintDefaults()` nor `fs.Usage` is ever called:
```
$ grep -n "PrintDefaults\|\.Usage(" internal/entrypoint/cli/cli.go
(no matches — confirmed dead)
```
So the `"Speed multiplier (range: 0.1-5.0)"` at `cli.go:66` can never be seen by a user; it exists only as a second place for a maintainer to forget to update, which is exactly what happened.

**Recommended fix:** fix `0.1` → `0.5` in `usageText`. Then remove the duplication — either pass `""` for the now-provably-unused `*Var` description strings, or drop the hand-written `Flags:` block in favor of rendering `fs.PrintDefaults()` into a buffer, so there's exactly one source of truth per flag.

**Implemented (2026-09-21):** `usageText` now reads `0.5-5.0` and also states the `--time` cap added in Finding 3 (`0 = no limit, max 300`). The dead `*Var` description strings were deleted (passed as `""`) rather than switching to `fs.PrintDefaults()` — that alternative was evaluated and rejected: `flag`'s renderer always uses a single dash (`-speed`, not `--speed`, inconsistent with the rest of `usageText`) and would render `-h`/`-help` as two separate, duplicated entries since they're two distinct `BoolVar` registrations. The hand-written `Flags:` block stays as the one place these are documented, with no second, unreachable copy of the same text to drift out of sync again.

---

### Finding 5: Error/usage print order is inconsistent across failure classes, and the mechanism is an ownership inversion, not a formatting choice — ✅ FIXED 2026-09-21

**Both reviews found this** (Fable as M6, Opus as Finding 5, with a fuller verification table).

**Evidence and verified behavior (Sonnet spot-checked, matches both reports exactly):**
```
$ /tmp/pidshooter-verify zzznoproc --speed=0.2   # config-class error
error: invalid configuration: speed must be between 0.5 and 5, got: 0.2
Process ID Shooter
... (19 more lines of usage) ...

$ /tmp/pidshooter-verify proc --unknown          # parse-class error
Process ID Shooter
... (19 more lines of usage) ...
error: flag provided but not defined: -unknown
```
Config-class errors print error-then-usage; parse-class errors print usage-then-error — for two failures that are, from the user's perspective, the same kind of mistake. On a normal terminal, the config-class message can scroll off-screen above 19 lines of usage text.

**Root cause (the useful part of this finding):** `config.NewConfig` calls `apperror.Handle` *internally* on failure, which logs to stderr right there, before `cli.go` ever gets control back to decide when to print `usageText`. By contrast, a `splitArgs`/`fs.Parse` failure is a plain error, logged later by `report()`. **The application layer decides when the error is printed; the adapter decides when the usage is printed — so the adapter cannot control its own presentation order**, and `config.NewConfig`'s doc comment ("returning an error if any parameter fails validation") doesn't mention that it also has this stderr side effect.

**No double-logging or silent swallowing was found** — both reviews traced all failure classes and confirmed each error appears exactly once.

**Recommended fix (in order of preference):** (1) have `config.NewConfig` return the `*apperror.Error` *without* calling `Handle`, and let `cli.report` be the single place that decides both whether and when to log — this also removes `report`'s "assumed to already be logged" premise (see Finding 8) since there'd be nothing to assume; or (2) at minimum, make both branches print in the same order.

**Implemented (2026-09-21):** by the time of this fix, `report()` no longer existed (it was folded into `CLI.Run` during an unrelated refactor: `return apperror.LogError(c.run(args))`), and `LogError` carried the exact same "assume an `*apperror.Error` is already logged" premise this finding's root cause describes. A first attempt at fix (1) — just removing `config.NewConfig`'s internal `Handle` call — was tried and immediately caught by re-running the real binary: it silently dropped the config-error message entirely, because `LogError` skipped it too, and nothing else logged it. The actual fix went further than either recommended option: `apperror.Error` gained an unexported `logged bool` field, and `Handle` was rewritten to be idempotent — it logs an error only the first time it sees it (checking/setting the flag), and is otherwise safe to call again later on the same error as it propagates through further layers. `LogError` was deleted outright (and its tests), `CLI.Run` now calls `apperror.Handle` directly, and `config.NewConfig` no longer calls `Handle` internally — its error is first logged only when it reaches `CLI.Run`'s boundary, at the exact same point split/parse errors already were. All three CLI-boundary failure classes (split, parse, config) now consistently print usage-then-error. Deeper runtime-fatal paths (`application/runner.go`'s `FindProcesses`/`gameSvc.Play` failures, already logged via their own `Handle` call at the point of origin) are unaffected — the boundary's later `Handle` call sees the flag already set and doesn't re-log. One behavior change ships with this: `Handle` given a plain (non-`*apperror.Error`) error now *propagates* it (needed for `CLI`'s plain split/parse errors), where it previously logged-and-absorbed it under `SeverityUnknown` — verified as safe since no existing call site ever passed a plain error to `Handle` before this fix. Covered by `TestHandlePropagatesUnknownSeverity` and `TestHandleLogsFatalErrorOnlyOnce` in `apperror_test.go`.

---

### Finding 6: `--help` does not short-circuit other argument errors — ✅ FIXED 2026-09-21

**Found by Fable only (M5); independently verified by Sonnet this pass — confirmed.**

`cli.go`'s `help` check only runs after `splitArgs` and `fs.Parse` have both already succeeded, so a parse error anywhere in `args` wins regardless of where `--help`/`-h` appears.

```
$ /tmp/pidshooter-verify --help --unknown
exit=1
error: flag provided but not defined: -unknown
```
The usage text does still get shown (as part of the error path), but the program exits 1 with an error instead of the clean "you asked for help, here it is" exit-0 a user would reasonably expect regardless of what else is on the line — most CLIs, including the cobra-based predecessor this repo used before the stdlib-`flag` rewrite, treat `--help` as an unconditional early exit.

**Recommended fix:** scan `args` for `-h`/`--help` and short-circuit before `splitArgs`/`fs.Parse`'s strict validation ever runs, so `--help` always wins regardless of position or other errors on the line.

**Implemented (2026-09-21):** added `help(args []string) bool` in `cli.go`, called first thing in `run()` (right after the no-args check, before the `flag.FlagSet` is even built), scanning every token in `args` for an exact `-h` or `--help` match regardless of position. `--help=true`/`--help=false` are deliberately *not* accepted as help requests — only the two bare forms are — so an inline value falls through to the ordinary "flag provided but not defined: -help" rejection like any other bad flag. Two design points settled along the way: (1) this couldn't live next to `flagNameAndValue` in `flag_splitter.go` — `flagSplitter` is deliberately agnostic about which specific flags exist (it just calls `flagSet.Lookup`), and hardcoding `"h"`/`"help"` there would break that; it belongs in `cli.go`, which already knows about every specific flag. (2) folding the scan into `split()`'s existing loop (to avoid a second pass over `args`) was considered and rejected: `split()` fails fast on the first bad flag, but `--help` must win even when an error occurs *earlier* in `args` than the `--help` token — supporting that would require `split()` to keep scanning to the end and stash its first error rather than returning immediately, a real change to a small, well-tested, fail-fast function, just to avoid a second pass over what's normally 2-5 tokens. Since the new `help()` func now catches every accepted help request before the `flag.FlagSet` is built, the old registered `help`/`h` `BoolVar`s and the post-`Parse` `if help` check on their variable became fully unreachable and were deleted. Covered by `TestRunHelpWinsRegardlessOfPositionOrOtherErrors` (4 cases, help in every position, mixed with unknown-flag and missing-value errors) and `TestRunHelpWithInlineValueIsNotAccepted` in `cli_test.go`.

---

### Finding 7: `cli_test.go` has 100% statement coverage and zero assertions on anything the adapter actually prints or which stream it uses

**Found in depth by Opus (Finding 11); Fable's l1 independently flagged the narrower "no test for the M5/Finding-6 scenario" instance of the same gap.**

```
$ go tool cover -func=/tmp/cli.cov
…cli.go:44:  NewCLI            100.0%
…cli.go:49:  Run               100.0%
…cli.go:53:  play              100.0%
…cli.go:96:  report            100.0%
…cli.go:107: splitArgs         100.0%
…cli.go:139: flagNameAndValue  100.0%
total:                          100.0%
```
Every line executes under test, but `cli_test.go` contains zero references to `os.Stdout`, `os.Stderr`, or captured output — `TestRunNoArgsPrintsUsageAndReturnsNil` and `TestRunHelpFlagPrintsUsageAndReturnsNil` are *named* for printing usage but assert only `assert.NoError`. This is the direct consequence of the architecture point in §1: without an injectable writer, there's nothing for a test to substitute and inspect. It's exactly why Finding 4 (stale help text), Finding 5 (print ordering), and Finding 6 (`--help` not short-circuiting) all survived multiple rewrites undetected — the lines that produce the wrong output are "covered," but nothing checks what they produced.

`cmd/pidshooter/` has no test file at all, so Finding 1 (the most severe finding in this document) has no regression guard either.

**Recommended fix:** give `CLI` `out`/`errOut io.Writer` fields (mirroring `console.ScoreReporter`, which already does exactly this for the identical "adapter writes text" reason), defaulted to `os.Stdout`/`os.Stderr` in `NewCLI`, then assert destination and content directly — including a speed-range assertion built from `movement.MinSpeed`/`MaxSpeed` so Finding 4 can't recur silently. For `main.go`, a single test that execs the built binary with `--help` under `env -u TERM` would pin Finding 1.

**Worth flagging directly: this reverses a decision made earlier in this same session.** The `out`/`errOut` fields were deliberately removed from `CLI` a few turns before this review, reasoning by analogy to `util.Logger` (which is genuinely stateless and always targets `os.Stderr`, with no per-instance configuration to inject). The disanalogy Opus's finding exposes: for `Logger`, losing that injection point was an explicit, discussed tradeoff ("stop asserting on logged content") the user chose knowingly. For `CLI.out`/`errOut`, the same removal happened as a "for consistency" follow-on without separately surfacing that it would zero out the CLI's own output-assertion capability — which `console.ScoreReporter`'s test suite shows is a real, exercised need for this class of adapter, not a hypothetical one. This is worth deciding deliberately rather than by extension of the `Logger` precedent.

---

### Finding 8: `report`'s doc comment is now inaccurate, and states implementation rationale about a different package — MOOT as of 2026-09-21

**Found by Opus (Finding 13).**

```go
// report logs err through the adapter's logger, unless err is (or wraps) an
// *apperror.Error, which is assumed to already be logged.
func (c *CLI) report(err error) error {
```
"The adapter's logger" describes a field the type no longer has — `CLI` was left with a single `service` field once the injected `*util.Logger` was removed a few turns before this review; `report` now constructs `util.NewLogger()` fresh inline. And "which is assumed to already be logged" isn't `report`'s own contract — it's the rationale for the guard, describing `apperror.Handle`'s behavior in a different package. This is exactly the "doc comments state contract only, never implementation rationale" issue that's been corrected elsewhere in this codebase multiple times this session — this is a fresh instance introduced by the `Logger`-field removal, not caught at the time.

The assumption is also unenforced: any future `inbound.Runner` implementation that returns an `*apperror.Error` without having passed it through `apperror.Handle` first gets silently swallowed here. Not reachable today (the sole implementer, `application.Runner`, routes every error through `Handle`), but invisible from the port's own doc comment, which says only "returning an error when something happens."

**Recommended fix:** restate as behavior only — e.g. *"report writes err to standard error unless it is (or wraps) an `*apperror.Error`, and returns err unchanged."* Adopting Finding 5's fix #1 (move all logging decisions into `report`) would remove the assumption entirely rather than just rewording it.

**Overtaken by events:** `report()` no longer existed even before this finding was acted on — an earlier, unrelated refactor this same session (runner-factory work) folded its job into `CLI.Run` calling `apperror.LogError` directly, which carried the exact same "assumed already logged" premise this finding flagged, just relocated. Finding 5's fix (2026-09-21) replaced that premise entirely: `apperror.Handle` is now idempotent (an unexported `logged` flag on `*apperror.Error`) and is the single function used both at error-origin call sites and at `CLI.Run`'s boundary — there is no separate "is this assumed logged" doc-comment claim left anywhere to be stale. No action needed.

---

### Finding 9: `main.go`'s three early-failure blocks bypass `apperror` entirely, and their explanatory comment only covers the first of the three — MOOT as of 2026-09-20

**Found by Opus (Finding 12).** Overtaken by Finding 1's fix: the three blocks this finding describes (`osprocess.NewProcess`/`filescore.NewFileScore`/`tcell.NewScreen`, each followed by a manual `logger.Error(...)`) no longer exist in `main.go` — they moved into `internal/composition.RunnerFactory.Create()`, which returns plain errors with no logging of its own, letting `cli.report()` be the sole logging site for them like any other CLI-returned error. Recorded below for history; no action needed.

```go
// A failure here means the program can't run at all, so it's logged
// and the program exits immediately rather than continuing on.
proc, err := osprocess.NewProcess()
if err != nil { logger.Error(err.Error()); return err }
store, err := filescore.NewFileScore()
if err != nil { logger.Error(err.Error()); return err }
screen, err := tcell.NewScreen()
...
```
The comment sits above the first block but is the rationale for all three — a reader looking only at the `filescore`/`tcell` blocks finds an unexplained repeated pattern. More substantively: these are the *only* three error sites in the entire program that don't route through `apperror.Handle` (every path in `application/runner.go` does). The output happens to look identical today (`Handle` calls `logger.Error` for `SeverityFatal` too), so there's no visible bug — but the `Code`/`Severity` classification `apperror` exists to provide is skipped for precisely the three most fatal failures in the program.

**Recommended fix:** move the comment above the group it governs (or give each its own), and route these three through `apperror.NewError(apperror.CodeUnknown, apperror.SeverityFatal, ...)` + `apperror.Handle`, so `main.go` no longer needs its own logger and gets Finding 10's exit-code classification for free if that's adopted.

---

### Finding 10: Exit codes don't distinguish usage errors from runtime failures, and bare invocation vs. `--confirm` alone disagree on whether "no pattern" is success or failure

**Found by Opus (Finding 10); independently verified by Sonnet this pass.**

```
$ /tmp/pidshooter-verify </dev/null >/dev/null 2>&1; echo $?
0
$ /tmp/pidshooter-verify --confirm </dev/null >/dev/null 2>&1; echo $?
1
```
Both are "you gave me no pattern," yet one is success (shows usage, exit 0) and the other is failure (rejected by `config.NewConfig`, exit 1) — purely because one has a flag attached. Separately, *every* failure class (a mistyped flag, an out-of-range value, and "no processes matched") all share exit 1, so a caller can't distinguish "you typed it wrong" from "nothing matched, which is a legitimate outcome" from "the game itself crashed."

**Recommended fix:** decide whether bare `pidshooter` is help (exit 0) or a usage error, and make `--confirm`-alone agree. `apperror.Code` already carries the classification needed for distinct exit statuses, if that's judged worth doing.

---

## 3. Minor findings

| # | Finding | Source | Status |
|---|---|---|---|
| 11 | A bool flag immediately followed by a bare value token (`--confirm true`) is not consumed as the flag's value (correctly, per Go convention) — but the leftover token then silently becomes an extra search pattern (`patterns=[true, node]`) rather than being rejected. Conversely, a *non*-bool flag with no `=` whose next token happens to itself look like a flag (`--speed --confirm`) gets that next token force-consumed as its value, producing a confusing `invalid value "--confirm" for flag -speed` instead of `flag needs an argument: -speed`. | Opus (Finding 8) + Fable (m1) — complementary, opposite-direction cases of the same root cause (`splitArgs`'s value-consumption heuristic has no lookahead for "does the next token look like a registered flag") | Open |
| 12 | `--time`/`--speed` silently accept Go numeric-literal syntax — octal (`010` → `8`), hex (`0x10` → `16`), binary (`0b101` → `5`), underscore separators (`1_000` → `1000`) — via `strconv.ParseInt`/`ParseFloat`'s base-0 parsing, completely undocumented in `usageText`. | Opus (Finding 9), unique | Open |
| 13 | `--` (conventional end-of-options terminator) and a bare `-` both produce the misleading error `flag provided but not defined: -` (naming a flag the user never typed), and there is no way to express a search pattern that begins with `-`. Additionally: the `a == "-"` special case in `splitArgs` that lets a bare `-` through as a "pattern" is dead in practice — `core/process`'s minimum-pattern-length-3 rule rejects it unconditionally anyway, so both paths converge on rejection regardless. | Fable (m2) + Opus (Findings 6 & 7, more complete — found the dead-code angle) — verified by Sonnet this pass | Open |
| 14 | `docs/hexagonal-arc.md` is substantially stale for these two files: still describes `cobra`/`buildCommand()` (removed), `scorefilestore`/`stderrlog` (renamed to `filescore`/`console`+`util.Logger`), `Config` living inside the `inbound` package itself (moved to `application/config` with a validating constructor), a `Lister`/`Killer` port split (superseded, see `docs/osprocess-findings.md`), a `testutil/capture` package (doesn't exist), and a full `main.go` code sample that no longer matches (wrong package names, wrong param count, wrong error-handling shape). | Both, merged — Opus's list is the more exhaustive of the two | Open — same deferred doc-sync effort noted in `docs/osprocess-findings.md` Finding 14 |
| 15 | `fs.NArg()` is never checked after `fs.Parse` succeeds. Correct today only because of an *unstated* invariant of `splitArgs` (every token is pre-classified as pattern-or-flag before `fs.Parse` ever runs, so it can never stop early on a leftover positional) — if that invariant is ever broken by a future edit, the affected arguments would be silently dropped rather than reported. Note: an explicit `fs.NArg() > 0` check *was* present earlier in this file's history and was deliberately removed once `splitArgs` was rewritten to make it provably unreachable — this finding is really "that invariant should be documented (or defensively re-guarded), not just relied upon silently." | Opus (Finding 14) | Open — worth a one-line comment on `splitArgs` at minimum, matching this repo's stated taste for defensive guards against currently-impossible states |
| 16 | Security/UX: pattern matching is case-insensitive **substring** matching (`core/process`), completely undocumented in `usageText` — `pidshooter ssh` matches `sshd`, `pidshooter agent` matches `ssh-agent`/`gpg-agent`. Combined with Finding 11 (unconsumed tokens silently becoming extra patterns) and Finding 13 (no `--` escape), there's no way to preview what a pattern will match before entering the game screen where a click is fatal. No injection surface exists (patterns never reach a shell). | Opus (Finding 17), ties to 11 | Open — candidate fix: a `--dry-run`/`--list` flag |
| 17 | `util.NewLogger()` is constructed independently at three separate call sites (`main.go`, `apperror.Handle`, `cli.report`) rather than shared. Harmless given `Logger` is stateless, but signals it's really a package-level capability being passed around as if it were an object. | Opus (Finding 12c) | Low priority / design note only |
| 18 | `splitArgs`/`flagNameAndValue` are free functions, which is in tension with this repo's stated "behavior belongs on a type" preference — but both are small, genuinely stateless string-processing helpers, and free helpers of this shape are the established norm in this codebase's other adapter packages (`osprocess`, `filescore`). Opus itself judged this a non-issue given the "don't over-anchor on precedent" counter-preference already on record. | Opus (Finding 16) | Not actioned — judgment call, no action recommended by either reviewer |

---

## 4. Low findings

- **No test exercises the `--help` + other-error interaction (Finding 6), and `cmd/pidshooter/` has no test file at all** (both reviews; Opus additionally proved this via 100% coverage + zero output assertions, Finding 7 above).
- **No root/privilege safeguard at the composition root** — carryover of `docs/fable-review.md`'s **S1**. Both reviews confirmed `main.go`/`cli.go` still perform no `os.Geteuid()` check, but noted this is now substantially mitigated one layer down: `core/process.IsKillableBy` + `Find`'s UID filter (`docs/osprocess-findings.md` Finding 3, implemented 2026-09-08) already scope a non-root player's target list to processes they own. S1's *root* case (as root, everything is still a target) remains open, but that's a `process`/`osprocess` concern, not specific to these two files.

---

## 5. Relationship to `docs/fable-review.md` (2026-08-20)

| Prior finding | Status now |
|---|---|
| **F1** — `--speed` help text / `docs/features.md` claim `0.1–5.0`; enforced minimum `0.5` | ✅ **Fully fixed 2026-09-21.** `docs/features.md` was already corrected; `cli.go`'s own `--help` text now reads `0.5-5.0` too, and the dead-code mechanism that let the two drift (unreachable `*Var` description strings) is removed. See **Finding 4**. |
| **F2** — `--speed=NaN` bypasses validation | ✅ **Fixed 2026-09-21.** `movement.ValidateSpeed` now rejects `math.IsNaN`. See **Finding 2**. |
| **F4** — `--time` has no upper bound, overflows to "unlimited" | ✅ **Fixed 2026-09-21**, via a domain-meaningful cap (`MaxTimeLimitSeconds = 300`) rather than the arithmetic-overflow boundary. See **Finding 3**. |
| **D3** — No inbound-port validation (only the CLI adapter validated) | ✅ **Fixed at the architecture level.** `application.Runner.Run` independently re-validates via `config.NewConfig`, verified by calling it directly with bad configs, bypassing the CLI entirely. **Caveat:** the port is now uniformly validated, but by the same buggy validators (F2/F4 above) — the architectural gap is closed, two of the specific holes it was meant to catch are not. |
| **D7** — Composition root builds adapters that can fail/block before argument parsing | ✅ **Fixed 2026-09-20**, after being confirmed measurably worse than originally described — `--help` didn't just risk paying the cost, it reliably failed in any environment without a `$TERM` (routine for CI/containers/cron). See **Finding 1**. |
| **S1** — No safeguard against running as root | ⚠️ **Partially mitigated, out of primary scope.** Non-root case closed one layer down (`IsKillableBy`); root case unchanged; not specific to `cli.go`/`main.go`. |
| (implied by task brief) cobra / `buildCommand()` staleness in `docs/hexagonal-arc.md` | 🔴 **Confirmed, and found to extend much further** than those two mentions — see **Finding 14**. |

**What genuinely improved:** the CLI-owned `validate()` that F1/F2/F4/D3 all originally cited is gone entirely; validation is centralized in `config.NewConfig`+`core` validators and independently re-checked by the application layer; cobra is gone in favor of a much smaller stdlib-`flag`-based implementation. The residual F1/F2/F4 defects survived every rewrite of `cli.go` because they live in `core`, not in the adapter that kept getting rewritten — the most useful thing this cross-reference established is that **no amount of further `cli.go` work could have fixed them, and reviewing `cli.go` in isolation would never have found them either.**

---

## 6. Verification notes

Independently re-run by Sonnet (orchestrating) against a fresh build (`go build ./cmd/pidshooter`, go1.26.5 darwin/arm64) after both independent reviews completed: Finding 1 (TERM-unset `--help` failure), Finding 2 (NaN bypass), Finding 4 (dead flag-description strings via direct grep), Finding 5 (print-order spot check for both failure classes), Finding 6 (`--help --unknown`), Finding 10 (bare vs. `--confirm`-alone exit codes), and Finding 13 (`--`/bare-`-` behavior). All matched both reports exactly; no discrepancies required a tie-break re-test this time (contrast `docs/osprocess-findings.md` §6, where the two independent passes disagreed on one point and needed one).
