# Branch review: introduce-play-command

Scope: `git diff master...HEAD` (3 commits: ae4d3ce, a75f94a, 4f2a323). `make build`, `make lint`, `make vet`, `make test` and `make test-acceptance` all pass. I also ran the built binary by hand with an isolated `HOME`/`XDG_CONFIG_HOME`.

## Summary

The branch swaps the hand-rolled flag splitter/parser/mapper for cobra. It adds `play`/`list`/`kill` subcommands, typed `PlayRunner`/`ListRunner`/`KillRunner` contracts with request DTOs and a lazy `play` runner in composition. Patterns and options now travel per call (`Load(options)`, `FindRequest.Patterns`) and are no longer injected at construction. The old `Mode` config field is removed. The overall shape is clean and well tested. The remaining problems are at the edges: the `list`/`kill` stubs that the help text advertises as working, and the help text that lost its config-file and default-value documentation. The bare-`play` bug and improvements 1, 4, 6, 9 and 10 are fixed (see the notes under each), and improvement 11 is partly fixed.

No security findings. The root guard is unchanged for `play`, and `list`/`kill` are stubs that touch no processes.

---

## Security

None found in the lines this branch changed.

---

## Bugs

### 1. `pidshooter play` with no pattern needs a terminal and writes the config file before it reports the missing argument

**Fixed.** `play`, `list` and `kill` now reject a missing pattern up front with an `ArgumentError` (`rootCommand.requirePatterns`), before anything is built or written. Covered by the `Program.Run` tests. `--speed`/`--time` range validation still happens inside `Load`, as on master.

`internal/entrypoint/cli/play_command.go:43-52` (likewise `list_command.go:27`, `kill_command.go:27`)

The subcommands set no `Args` validator, so `play` with zero patterns goes straight to the lazy runner. That runner builds every dependency (`ps` lookup, score/config file paths, `tcell.NewScreen`) and then calls `config.Service.Load`, which creates `config.yaml` if it is missing. Only after that does `process.ValidatePatterns` reject the empty list. Observed:
- `TERM= pidshooter play` prints `error: failed to create screen: terminal entry not found` and exits 1, with no usage text.
- `TERM=xterm pidshooter play` prints `invalid configuration: at least one search pattern is required`, then usage, exits 2, and has already written `~/.config/pidshooter/config.yaml`.

On master, the equivalent bare invocation (`pidshooter`) printed usage with exit 0 and had no side effects. A missing positional is a command-line error and should be caught before anything is built.

The same ordering applies to `--speed 99` and `--time=-1`. These flags are validated only inside `Load`, after the build, so without `TERM` they also come back as the screen error (exit 1) instead of an argument error (exit 2). This ordering existed on master (`creator.Create` ran before `runner.Run`). Bare `play` is the new, most common way to hit it.

Suggested fix: add `Args: cobra.MinimumNArgs(1)` to all three subcommands, or have `play` show help with no args like the root does. The `Args` error must be wrapped as an `ArgumentError`, because cobra does not send `Args` errors through `SetFlagErrorFunc` (see Improvement 1). Add a `Program.Run` test case for `{"play"}`, `{"list"}` and `{"kill"}`.

---

## Improvements

### 1. `Program.Run` returns cobra-originated errors unlogged and without usage

**Fixed.** The `Args` error is now an `ArgumentError`, and `mapFlags` no longer exists, so no cobra-originated error exits silently today.

`internal/entrypoint/cli/program.go:56-69`

The root sets `SilenceErrors: true`. `Run` only logs errors that are `ArgumentError`s (produced by `SetFlagErrorFunc` and the root `RunE`) and relies on runners to log their own. Any other error cobra or the command tree produces gets no log line and no usage text, and exits 1 silently:
- `mapFlags` `GetBool` failures (`root_command.go:79-96`)
- `Args` validator errors
- required-flag or flag-group errors

None of these can happen today, but adding the `MinimumNArgs` from Bug 1 without wrapping would produce exactly this silent exit.

Suggested fix: make the root's own errors `ArgumentError`s at the source. Wrap `Args` with a validator that returns `ArgumentError`, and wrap the `mapFlags` error. That keeps the "a command's error is returned as reported, not logged again" contract intact.

### 2. `list` and `kill` are advertised in help but always fail with "unimplemented"
`internal/application/runner/list.go:21-24`, `internal/application/runner/kill.go:21-24`, `internal/entrypoint/cli/root_command.go:58-62`

`pidshooter --help` lists `kill` and `list` as available commands. Both accept `--include-root`/`--i-am-root` and silently ignore them, and both exit 1 with `error: unimplemented`. If this branch is merged as-is, a user-visible CLI ships with two dead commands.

Suggested fix: set `Hidden: true` on both cobra commands, or leave them unregistered, until the runners exist. When they are implemented, they must go through `process.Service.FindProcesses` (or an equivalent guard). The root refusal (`process.ValidateRoot`) currently lives only on that path, and `kill` is the command where the root guard matters most.

### 3. Help text lost the config-file and default-value documentation
`internal/entrypoint/cli/play_command.go:9-18, 57-61`, `root_command.go:70-73`

The removed `usageText` explained several things the new help no longer mentions:
- where the config file lives (`~/.config/pidshooter/config.yaml` / `$XDG_CONFIG_HOME`)
- that flags override it per run (for example, `--include-root=false` beats a persisted `true`)
- the effective defaults (speed 2.0, time 30)
- that `--include-root` is implied when running as root
- why `--i-am-root` exists

The new `--speed`/`--time` registrations use a `0` default, which is not the real default, so cobra prints no default at all.

Suggested fix: move the config-file paragraph into the root `Long` (or the `play` `Long`), and append "defaults to the config file value, or N" to the `--speed`/`--time` usage strings. Restore the root-related caveats on the two global flags.

### 4. `ProcessOptions` is built the same way in three places

**Fixed.** `rootCommand.toProcessOptions(flagSet)` builds it for all three commands.

`play_command.go:68`, `list_command.go:31`, `kill_command.go:31`

`inbound.ProcessOptions{IncludeRoot: c.root.includeRoot, AllowRoot: c.root.allowRoot}` appears three times. There are three real call sites and they convert data owned by `rootFlags`.

Suggested fix: add a method on `rootFlags`, for example `func (f *rootFlags) toProcessOptions() inbound.ProcessOptions`.

### 5. Flag definitions use two styles
`play_command.go:57-61, 71-79` vs `root_command.go` (`defineFlags`, `toProcessOptions`)

Root flags now use named constants with `BoolVar` plus `Changed`, which is the same style as the play flags. Play flags use `BoolVar`/`Float64Var`/`IntVar` and repeat the string literals `"confirm"`, `"speed"`, `"time"` in both `defineFlags` and `toRequest`. A typo in one place would silently drop that flag from the request: `Changed` returns false for an unknown name.

Suggested fix: pick one style. The smallest change is to add `confirmFlag`/`speedFlag`/`timeFlag` constants and use them in both methods.

### 6. `Program.Run` doc comment contradicts the `CodeInvalidConfig` branch

**Fixed.** The doc comment now describes the invalid-configuration case.

`internal/entrypoint/cli/program.go:44-48`

The comment says "The error of a command is returned as the command reported it". However, a command error carrying `apperror.CodeInvalidConfig` is wrapped in `ArgumentError` and printed with usage (lines 64-66), which is behavior a caller depends on for exit code 2.

Suggested fix: add a sentence saying that a command error classified as invalid configuration is returned as an `ArgumentError`, with usage printed to standard error.

### 7. `TestCreator_Create_BuildsNothingUntilACommandRuns` does not verify what its name claims
`internal/composition/creator_test.go:20-37`

The test runs `list` and `play` and checks that `play` fails with something other than "unimplemented". An eager `Create` would not return an error either (`Create` has no error return), so this assertion cannot tell lazy from eager. Laziness is already properly covered in `lazy_runner_test.go:13`.

Suggested fix: rename the test to what it actually checks (for example, that `play` surfaces its build error once it runs), or drop it in favor of the `lazyRunner` test.

### 8. `lazyRunner` is a generic with a single instantiation
`internal/composition/lazy_runner.go:6`

The two type parameters exist for one use (`PlayRequest`/`PlayRunner`). If `list`/`kill` will need the same deferred wiring once implemented, the generic is justified ("restructure to intended shape"). If they will stay cheap (no TTY), a concrete `lazyPlay` type would be simpler. *Uncertain: depends on the planned wiring for list/kill.*

### 9. Root flags travel through shared mutable state and a parent-only `PersistentPreRunE`

**Fixed.** Removed `PersistentPreRunE`, `mapFlags` and the nil-pointer fields. The flags are bound with `BoolVar` on the root, and each command calls `rootCommand.toProcessOptions(cmd.Flags())`.

`root_command.go:48` (`PersistentPreRunE: c.mapFlags`), `:78-99`; consumed at `play_command.go:68`, `list_command.go:33`, `kill_command.go:33`

Each subcommand reads `c.root.includeRoot`/`allowRoot`, relying on the parent's `PersistentPreRunE` having filled them. In cobra, if a subcommand ever defines its own `PersistentPreRun(E)`, the parent's hook is skipped unless `cobra.EnableTraverseRunHooks` is set. The root flags would then be silently unset, and with `--include-root=false` a persisted `include_root: true` would win without any error. This is temporal coupling with a silent, safety-relevant failure mode.

Suggested fix: persistent flags are merged into `cmd.Flags()` of the executing subcommand, so each command can read `cmd.Flags().Changed(...)` the way `playCommand.toRequest` does, for example through a single `rootFlags.toProcessOptions(flagSet)` method. That removes `mapFlags`, `PersistentPreRunE` and the shared pointer, and also covers Improvement 4.

### 10. "RunsOnly…" tests do not check the "only"

**Fixed.** Each test now asserts the other two runners were not called.

`program_test.go:69`, `:85`, `:101`

`returnFromEveryRunner` arms all three mocks, but each test asserts only that the target runner was called once. Dispatching to an extra runner would still pass.

Suggested fix: also `AssertNotCalled` on the other two runners, or arm only the target runner so an unexpected call panics.

### 11. Coverage gaps for the new CLI surface

**Partly fixed.** Added tests for root flags before the subcommand, bare `play`/`list`/`kill`, and the subcommand usage in the output. Still open: an acceptance test for bare `play`.
- A root flag placed before the subcommand (`pidshooter --include-root play x`). It works (verified manually), but every test puts root flags after the subcommand (`root_command_test.go:98`, `:117`, `:136`).
- `play` with no patterns, end to end (exit 2 plus usage on stderr). Unit-tested through `Program.Run`, but there is no acceptance test.
- `Program.Run` printing the subcommand's usage (not the root's) for an invalid-config error. `program_test.go:135` checks only for `"Usage:"`.

### 12. Bulk `kill` needs more than the root guard
`kill_command.go:31-34`, `internal/application/runner/kill.go:22-24`

When `kill` is implemented it is a non-interactive bulk `SIGKILL` driven by substring patterns of at least 3 characters (for example, `kill sys` could match many processes). Besides routing discovery through `process.Service.FindProcesses` (see Improvement 2), add a test that `kill` refuses root without `--i-am-root`, and consider a confirmation, dry-run or exact-name matching.

---

## Nits

1. **(Accepted)** **Inline table tests, no `newXTestCase()`.** `program_test.go:35-47` (`tests := [][]string{...}`) and `root_command_test.go:34` (`tests := []string{...}`) build their tables inline instead of using a `newXTestCase()` constructor.
2. **(Fixed)** **Comma splice in a doc comment.** `internal/application/config/service.go:24-25`: "AllowRoot permits running as root, is never read from the stored config, only ever set per invocation." Suggest: "AllowRoot permits running as root. It is never read from the stored config, only set per invocation."
3. **(Fixed)** **Only one field of `inbound.ProcessOptions` is commented.** `internal/application/contract/inbound/runner.go:55-59`: `AllowRoot` has a field comment, `IncludeRoot` does not. Comment both or neither.
4. **(Accepted)** **Validation inlined into `Load`.** `internal/application/config/service.go:53-63` replaced the `validate` method with an inline block plus a `// Validate ...` comment. Keeping `validate(options Options) error` as a method would keep `Load` at one level of abstraction, as it was before.
5. **(Fixed)** **`Play` doc comment uses wiring vocabulary.** `internal/application/runner/play.go:44`: "Play wires together the services required to run a game session" describes composition rather than what the type is, for example "Play plays a game session using the configuration, process, game and score services". It was carried over from `Service`, but the type was renamed and re-documented on this branch.
6. **(Fixed)** **`help <unknown>` exits 0.** `pidshooter help bogus` prints root help and exits 0, which is cobra's default. It is inconsistent with `pidshooter bogus` exiting 2. Optional: set a custom help command, or accept it.
7. **(Makefile fixed, docs left)** **Stale invocations outside the code.** `Makefile:114` (`make run ARGS=firefox`) and `docs/features.md:15-64` still use the removed top-level form (`pidshooter firefox`), which now fails with `unknown command`. Flagged only; per the owner's rule, docs are synced separately.
8. **(Accepted)** **New free function `toConfigOptions`.** `internal/application/runner/converter.go:11`, `:40` (`toFindRequest` also grew a parameter). The repo rule is no free functions. It follows the existing `converter.go` precedent, but the precedent is not a justification: its input, `inbound.PlayRequest`, lives in another package so a method cannot be attached there, and a method on `Play` would have an unused receiver. Needs an owner decision (accept converters as an exception, or have `config.Service.Load` take something the inbound DTO converts itself into). Flagged, not prescribed.
9. **(Accepted)** **Duplicate option DTOs.** `inbound.GameOptions`/`ProcessOptions` (`inbound/runner.go:301-315`) duplicate `config.GameOptions`/`ProcessOptions` (`config/service.go:12-26`) field for field, bridged by `toConfigOptions`. Deliberate decoupling, so not wrong; one more pair to keep in sync.
10. **(Accepted)** **`Kill`/`List` struct docs describe behavior they lack.** `kill.go:10`, `list.go:10` say "Kill kills the processes that match search patterns" while `Run` always returns `unimplemented`.
11. **(Fixed)** **`pidshooter -- play x` reports "unknown command: play".** `root_command.go` (`showHelp`). After `--`, `play` is positional, so the message is misleading (exit code 2 is right). Cobra's "Did you mean …?" suggestions are also lost; optionally append `c.SuggestionsFor(args[0])`.
12. **(Fixed)** **`printUsageText` has no doc comment.** `program.go`; the other unexported methods in the package have one-line comments.
13. **(Fixed)** **`TestNewCreator_CreatesErrorHandler` only asserts a field is non-nil.** `composition/creator_test.go:12`. It tests a constructor that only assigns a field; `TestCreator_Create_LogsAnErrorOfACommandOnce` already covers the handler. Consider deleting it.
14. **(Fixed)** **Stale comment grammar.** `config/result.go:37`: "regardless of which user is running." is missing its object.

---

## Out of scope, pre-existing

- The root guard (`process.ValidateRoot`) runs inside `FindProcesses`, after `config.Service.Load` may already have created `config.yaml`. As root without `--i-am-root`, the file is written to root's home before the refusal.
- `Play` struct field order (`gameSvc`, `scoreSvc`) differs from the `NewPlay` parameter order (`scoreSvc`, `gameSvc`).
