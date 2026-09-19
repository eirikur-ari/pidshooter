# Application Configurator — Deferring Adapter Construction Until the Mode Is Known

*Design proposal, 2026-09-15. Not yet implemented. Scope: extracting adapter construction out of `cmd/pidshooter/main.go` into a per-mode configurator, sized to plausibly support two future non-interactive modes (list, lucky) without implementing them; and (§11) layering a config file between code defaults and CLI flags — for both which mode runs by default and each mode's own flag defaults — so the composition root's lazy, per-mode construction still works once "which mode" is no longer hardcoded. Companion documents: [`hexagonal-arc.md`](hexagonal-arc.md) (the architecture rules this proposal must not violate) and [`features.md`](features.md) (the CLI surface it must not silently break).*

*Every code sketch below was checked against the live tree at commit `66ddc02` (branch `wrapup-infra-layer`). Appendix A logs what was read and what was found; Appendix B lists the amendments `hexagonal-arc.md` needs once this lands, because that document has drifted from the code in several places this proposal touches.*

---

## Contents

1. [Problem statement](#1-problem-statement)
2. [Goals and non-goals](#2-goals-and-non-goals)
3. [Where mode is decided](#3-where-mode-is-decided)
4. [What the configurator is, and where it lives](#4-what-the-configurator-is-and-where-it-lives)
5. [Per-mode adapter needs](#5-per-mode-adapter-needs)
6. [Impact on the inbound contract](#6-impact-on-the-inbound-contract)
7. [Package and file layout](#7-package-and-file-layout)
8. [Migration plan](#8-migration-plan)
9. [Testing strategy](#9-testing-strategy)
10. [Trade-offs, decisions, and open questions](#10-trade-offs-decisions-and-open-questions)
11. [Layered configuration: default mode and per-mode defaults](#11-layered-configuration-default-mode-and-per-mode-defaults)
- [Appendix A — Verification log](#appendix-a--verification-log)
- [Appendix B — Amendments required to hexagonal-arc.md](#appendix-b--amendments-required-to-hexagonal-arcmd)

---

## 1. Problem statement

### 1.1 The current composition root

`cmd/pidshooter/main.go` is 53 lines and reads cleanly. Reproduced verbatim, because the rest of this document is a set of edits to it:

```go
func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := stderr.NewLogger()

	// These two adapters are constructed before application.NewRunner, so
	// there's no apperror.Handler yet to route their failures through —
	// they're the one place in the program that logs directly.
	proc, err := osprocess.NewProcess()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	store, err := filescore.NewFileScore()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		err = fmt.Errorf("failed to create screen: %w", err)
		logger.Error(err.Error())
		return err
	}
	ui := tcellui.NewUI(screen)
	reporter := stdout.NewScoreReporter()

	runner := application.NewRunner(proc, store, reporter, ui, ui, logger)
	return cli.NewCLI(runner, logger).Run(os.Args[1:])
}
```

Line count is not the complaint. The complaint is the ordering this encodes.

### 1.2 Adapters are built before the mode is known

`cli.NewCLI(runner, logger).Run(os.Args[1:])` is the call that parses argv. It is the *last* statement in `run()`. Every adapter above it — including `tcell.NewScreen()`, which opens the controlling terminal and puts it into raw mode — is constructed before a single argument has been looked at.

This is invisible today because there is exactly one mode. `internal/entrypoint/cli/cli.go` builds a single root `cobra.Command` with `RunE: c.play` and no subcommands, so whatever `run()` builds is, by definition, what the one and only invocation path needs.

That invariant breaks the moment a second mode exists:

- `pidshooter list firefox` must not call `tcell.NewScreen()`. It never renders a frame, and `NewScreen` on a non-tty stdin (a pipe, a CI runner, `pidshooter list foo | grep`) fails. A mode that only prints would abort for a reason that has nothing to do with what it was asked to do.
- `pidshooter lucky firefox` needs neither the score store nor the score reporter. `filescore.NewFileScore()` fails when the user's home directory cannot be resolved; there is no reason for that to sink a mode that never touches a score board.
- `run()` currently has no branch point at all. Adding a mode means either polluting `run()` with conditionals that decide, after the fact, which of the eagerly-built adapters to actually use, or restructuring so construction happens after dispatch. Only the second actually solves the problem.

### 1.3 The `run() error` / `os.Exit(1)` split, and an error-reporting wart it currently forces

`main()` discards the error value and only sets an exit code; the message has already been printed by whoever was holding the logger. Note the comment quoted above: the two fallible pre-`NewRunner` adapters are *the one place in the program that logs directly*, because there is no `apperror.Handler` yet.

That carve-out is a consequence of the eager ordering, not a design choice worth preserving. Once construction moves inside cobra's `RunE`, a construction failure returns through `cmd.Execute()` into `CLI.report`, which already logs exactly the errors that have not been logged elsewhere:

```go
func (c *CLI) report(err error) error {
	var appErr *apperror.Error
	if err != nil && !errors.As(err, &appErr) {
		c.logger.Error(err.Error())
	}
	return err
}
```

A plain (non-`*apperror.Error`) construction failure is logged there, once. So this refactor lets the special case be deleted rather than moved. That is a small but real secondary win, and §8 schedules it.

### 1.4 One `Runner` shape does not fit three modes

`inbound.Runner` is a single method, `Run(cfg Config) error`, and `inbound.Config` is:

```go
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}
```

Every field after `Patterns` is a game concept. `ConfirmMode` drives a full-screen status-bar prompt (`features.md` §Confirm mode). `Speed` is validated against `movement.MinSpeed`/`MaxSpeed` (0.5–5.0). `TimeLimit` seeds `core/game`'s session timer. List mode has no animation, no session clock, and no confirm prompt; lucky mode needs only `Patterns`.

Pushing three modes through one `Config` produces either a kitchen-sink struct whose fields are meaningful in one mode out of three, or overloaded fields. Both erode the type safety the outbound side already has — the outbound ports in `application/contract/outbound` are narrow and single-purpose (`Renderer`, `InputSource`, `ScoreStore`, `ScoreReporter`, `Logger`, `Process`), and the inbound side should mirror that.

### 1.5 Root cause

The composition root conflates two questions:

1. *What does the user want to do?* — answered by parsing argv.
2. *What resources does that require?* — answered by constructing adapters.

`run()` answers (2) unconditionally, with no dependency on (1). The fix is to make adapter construction a function of the resolved mode, invoked strictly after mode resolution.

---

## 2. Goals and non-goals

### 2.1 Goals

- **G1 — Defer construction.** No adapter is constructed until the requested mode is known from argv. In particular, `tcell.NewScreen()` is reachable only from the play path.
- **G2 — One seam.** A single, named place turns "resolved mode + parsed flags" into "concrete adapters + a runnable application service".
- **G3 — Preserve the dependency rule as `hexagonal-arc.md` actually states it.** `internal/core` keeps zero imports of application, infrastructure, or entrypoint (§8 "Key property", §12 "Clear dependency rule"). `internal/application` keeps zero imports of infrastructure or entrypoint — a rule the document does not spell out in prose but does enforce in its §8 dependency graph, where the only arrows into `infrastructure` originate at `main`.
- **G4 — Plausibility for list and lucky, not implementation.** The seam must credibly extend to both modes without their application logic being written now.
- **G5 — Fail fast with one reporting path.** Construction failures surface as clear errors on the existing `run() error` → `os.Exit(1)` path, logged exactly once, via `CLI.report` rather than via a hand-rolled carve-out.
- **G6 — Testability without new environmental surface.** The laziness property ("list does not construct the screen") must be provable by a plain unit test, with no tty, no `ps`, and no writable `$HOME`.
- **G7 — Incremental migration.** Each step compiles, keeps `make test` green, and is separately revertible.

### 2.2 Non-goals

- **NG1.** Designing or implementing list/lucky application-layer logic or any new core domain types.
- **NG2.** Changing the internals of `application/game.Service`, `application/process.Service`, or `application/score.Service` — with one narrow, forced exception (§5.3, the stdout write inside `FindProcesses`).
- **NG3.** Finalizing the list/lucky user experience. §5.4 sketches what a "pick which to kill" step would need, without committing to it.
- **NG4.** Changing any existing outbound port interface.

---

## 3. Where mode is decided

### 3.1 Subcommands, not a `--mode` flag

**Recommendation: cobra subcommands.** `pidshooter play [patterns...]`, `pidshooter list [patterns...]`, `pidshooter lucky [patterns...]`, with the bare root invocation (`pidshooter firefox`) continuing to behave as `play`.

The mechanical reason is that cobra resolves the target command *before* invoking any `RunE`. If construction is invoked from inside a subcommand's `RunE`, an interactive-only adapter is only ever reachable through the closure the `play` command owns. A `--mode=list` flag would work too, but it parses into a variable that some later branch must switch on — the branch point ends up hand-written, whereas with subcommands cobra is the branch point, and the guarantee is structural rather than a condition somebody has to remember to write.

The secondary reason is help output: `pidshooter --help` would list three commands with their own flag sets, instead of one command whose flags are silently inert in two of three modes.

### 3.2 Backward compatibility for the bare invocation

`cobra.Command` supports a root with both a `RunE` and subcommands: an argument that does not match a registered subcommand falls through to the root's `RunE`. So `pidshooter firefox` keeps working unchanged, and `pidshooter play firefox` is the explicit spelling.

Two consequences to accept knowingly:

- **A pattern that collides with a subcommand name changes meaning.** `pidshooter list` today searches for processes whose name contains "list"; after this change it invokes list mode with no patterns. Same for `play` and `lucky`. The escape hatch is the standard `--` separator: `pidshooter -- list`. This is a user-visible behavior change and belongs in `features.md`.
- **Flags must be registered on both the root and `play`.** Today `buildCommand` binds flags to fields on the `CLI` struct:

  ```go
  cmd.Flags().BoolVar(&c.confirm, "confirm", false, "Ask for confirmation before killing")
  cmd.Flags().Float64Var(&c.speed, "speed", 2.0, "Speed multiplier (range: 0.1-5.0)")
  cmd.Flags().IntVar(&c.timeLimit, "time", 30, "Time limit in seconds (0 = no limit)")
  ```

  Binding the same pointers on two `*cobra.Command`s is safe — pflag writes each default at registration time, and only the executed command's flag set is parsed — but it should go through one helper so the defaults cannot drift apart:

  ```go
  func (c *CLI) registerPlayFlags(cmd *cobra.Command) {
  	cmd.Flags().BoolVar(&c.confirm, "confirm", false, "Ask for confirmation before killing")
  	cmd.Flags().Float64Var(&c.speed, "speed", 2.0, "Speed multiplier (range: 0.5-5.0)")
  	cmd.Flags().IntVar(&c.timeLimit, "time", 30, "Time limit in seconds (0 = no limit)")
  }
  ```

  The defaults above (`2.0` and `30`) are today's real values, confirmed by `cli_test.go`'s `TestRunBasicPattern`, which asserts `2.0` and `30` on a bare `pidshooter firefox`. The help string is corrected from `0.1-5.0` to `0.5-5.0` at the same time: `movement.MinSpeed` is `0.5`, and `features.md` §Speed already documents `0.5`–`5.0`. The current help text is simply wrong and this is a free moment to fix it.

  "Bare invocation means `play`" and "`--speed` defaults to `2.0`" are both, at this point in the document, hardcoded — the code default in a three-tier precedence (code default → config file → CLI flag) this proposal's owner also wants. §11 adds the config file and revises both this subsection and §7.3's `buildCommand` sketch accordingly; nothing here needs to be re-read once it does, since §11 states its changes as edits to specific lines rather than a competing design.

### 3.3 The no-args behavior must survive

`CLI.play` currently opens with:

```go
if len(args) == 0 && cmd.Flags().NFlag() == 0 {
	return cmd.Help()
}
```

`cli_test.go` pins three related behaviors: bare `Run([]string{})` prints help and returns nil; `--help`/`-h` print help and return nil; and `Run([]string{"--confirm"})` with no patterns *does* forward to the service with empty `Patterns` (`TestRunNoArgsWithFlagsForwardsToService`). That last one matters here — it means the no-args guard is not a blanket "no patterns, no work" rule, and the provider for play must therefore not be invoked before the guard runs, or `pidshooter` with no arguments would construct a screen just to print help. §4.4 places the provider call after the guard for exactly this reason.

---

## 4. What the configurator is, and where it lives

### 4.1 Placement: the composition root grows files, not a new package

`hexagonal-arc.md` is unusually direct about this, and the wording matters. §9, verbatim:

> In hexagonal architecture `main.go` is not logic — it is wiring. All wiring happens here and nowhere else.

and the package tree in §5 annotates the file as:

> `main.go    # Composition root only — no logic`

So the document already has a term for this code — **composition root** — and it already assigns it a location. It has no separate notion of "composition-root-adjacent code" living elsewhere, and no precedent for a package whose job is to import every layer.

It does, however, have a precedent for *placement reasoning*, in the §3 paragraph about `fsutil`:

> It's kept under `infrastructure/` rather than promoted to a top-level package like `internal/util`, specifically *because* it touches the filesystem — `internal/util` is safe for `core` to import (§3 names it explicitly) only because it's pure Go with no OS access; giving a filesystem-touching helper the same top-level standing would blur that signal and invite a future core→filesystem leak.

The principle is: put code where its dirtiest dependency is already visible; do not create a package that launders it into looking cleaner than it is.

Applying both: **keep the configurator in `cmd/pidshooter`, in `package main`, split across files.** A new `internal/composition` package would contradict §9's "nowhere else" and would add a node to the §8 dependency graph that imports core, application, infrastructure, *and* entrypoint — the single most privileged package in the tree, sitting under `internal/` where everything else is layered. `cmd/pidshooter` already has exactly that privilege and is already documented as having it. Growing it from one file to four adds no new edge to the dependency graph at all.

Concretely:

```
cmd/pidshooter/
├── main.go        # main() · run() — logger, providers, dispatch
├── wire_play.go   # newPlayRunner()
├── wire_list.go   # newLister()
└── wire_lucky.go  # newLuckyShooter()
```

The objection to `package main` is testability. It does not hold: `go test ./cmd/pidshooter/` works, and a test file in `package main` can call `newPlayRunner` directly. What `package main` genuinely costs is that nothing can *import* it — which is the point (§4.3), and which also means a hypothetical second entrypoint could not share the wiring. There is one entrypoint and no plan for a second; if one ever appears, promoting these three files into a package is a mechanical move. §10.1 records this as the reversible half of the decision.

### 4.2 Shape: three provider functions, not a factory type

Three named functions, one per mode, each in its own file, each constructing exactly the adapters that mode needs and returning the application service that implements the mode's inbound port:

```go
// wire_play.go
package main

// newPlayRunner constructs every adapter the interactive game needs and wires
// them into the application.Runner that implements inbound.Runner. It is
// called from the play subcommand's RunE, so none of this runs when another
// mode was requested — in particular tcell.NewScreen, which opens the
// controlling terminal, is unreachable from list and lucky.
func newPlayRunner(logger outbound.Logger) (inbound.Runner, error) {
	proc, err := osprocess.NewProcess()
	if err != nil {
		return nil, err
	}

	store, err := filescore.NewFileScore()
	if err != nil {
		return nil, err
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}
	ui := tcellui.NewUI(screen)

	// ui is passed twice: *tcellui.UI satisfies both Renderer and InputSource.
	return application.NewRunner(proc, store, stdout.NewScoreReporter(), ui, ui, logger), nil
}
```

This is today's `run()` body with the `logger.Error(...)` calls dropped (§1.3 — `CLI.report` logs them now) and a return value instead of locals. Nothing else changes; `application.NewRunner`'s six-parameter signature `(proc, store, reporter, renderer, events, logger)` is unchanged.

The list and lucky files have the same shape with shorter bodies — for list, today, only `osprocess.NewProcess()` plus a result printer.

There is deliberately no generic factory type, no registry map, and no `Built{Run, Cleanup}` wrapper:

- **No registry.** Exactly one of these functions executes per process invocation, so there is nothing to share, cache, or hoist. A map from mode name to constructor would add a lookup that cobra already performs.
- **No `Cleanup` in the wrapper**, because the renderer's lifecycle is not the wiring layer's business — see §4.5, which is a verified statement about the live code rather than an assumption.

### 4.3 Keeping `entrypoint/cli` free of infrastructure knowledge

This is the decision with the largest blast radius, so it is stated as a decision rather than an option.

**Decision: `cli` never learns that infrastructure exists. It receives providers — functions returning inbound ports — from `main`, and calls them from inside `RunE`.**

The rejected alternative is `cli` importing the wiring directly and calling `newPlayRunner` itself. That would make `entrypoint/cli` transitively import `infrastructure/tcellui`, `infrastructure/osprocess`, and `infrastructure/filescore`, adding three arrows to the §8 dependency graph out of an adapter that today imports exactly two application packages (`contract/inbound` and `apperror`) plus cobra. It would also make `cli`'s own tests drag in tcell. With the wiring in `package main` (§4.1) the alternative is not merely undesirable, it is impossible — `package main` cannot be imported. Choosing the placement and choosing this option are the same decision, made once, which is part of why the placement is worth the `package main` awkwardness.

The provider types are declared by `cli`, the consumer, following the convention `hexagonal-arc.md` §7 states ("In Go, the consumer defines the interfaces it needs") and which `cli.go` already follows for its logger:

```go
// logger is the minimal reporting capability this adapter needs.
type logger interface {
	Error(msg string)
}
```

So:

```go
// Providers supplies the application services this adapter can drive. Each
// field is called at most once, from the RunE of the subcommand that needs
// it, so the adapters behind a mode are constructed only when that mode is
// the one the user asked for. A nil field means the mode is not available in
// this build; Run reports that rather than panicking.
type Providers struct {
	Play  func() (inbound.Runner, error)
	List  func() (inbound.Lister, error)
	Lucky func() (inbound.LuckyShooter, error)
}

// NewCLI returns a CLI adapter that obtains its application services from providers.
func NewCLI(providers Providers, logger logger) *CLI {
	return &CLI{providers: providers, logger: logger, out: os.Stdout, errOut: os.Stderr}
}
```

(§11.9 adds a third parameter, `cfg config.File`, stored as a `cfg` field alongside `providers` and `logger` above — not shown here since the config file is a separate concern from §11 onward; the shape of this constructor does not otherwise change.)

Function fields rather than an interface with three methods, because the whole point is that only one is ever called, and a struct of funcs makes "was this one called?" trivially observable in a test (§9.2) without a mock framework.

`NewCLI`'s signature changes from `(inbound.Runner, logger)` to `(Providers, logger)`. That is a breaking change to one call site (`main.go`) and to `cli_test.go`'s `newTestCLI` helper; §8 sequences it.

### 4.4 The `RunE` bodies

```go
func (c *CLI) play(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && cmd.Flags().NFlag() == 0 {
		return cmd.Help()
	}

	runner, err := c.providers.Play()
	if err != nil {
		return err
	}

	return runner.Run(inbound.PlayConfig{
		Patterns:    args,
		ConfirmMode: c.confirm,
		Speed:       c.speed,
		TimeLimit:   c.timeLimit,
	})
}
```

The provider call sits *after* the help guard (§3.3) and after flag parsing, which cobra has already completed by the time `RunE` runs. `list` and `lucky` get structurally identical bodies against their own configs.

### 4.5 Renderer lifecycle is not the configurator's problem — verified

`application/game/service.go` owns it, in `runLoop`:

```go
func (s *Service) runLoop(session *game.Session, tracker *killTracker) (time.Time, error) {
	if err := s.renderer.Init(); err != nil {
		return time.Time{}, fmt.Errorf("renderer initialization failed: %w", err)
	}
	defer s.renderer.Cleanup()
	...
}
```

This matches the contract documented on the port itself in `contract/outbound/ui.go` ("Init must be called and succeed before Size or Render is used… Cleanup releases what Init acquired and is safe to call more than once").

So the split is: **the configurator constructs the renderer; the service that uses it drives its lifecycle.** `newPlayRunner` calls `tcell.NewScreen()` and `tcellui.NewUI(screen)` and stops there. It does not call `Init`, does not `defer Cleanup`, and does not need to return a teardown function. The `Built{Run, Cleanup}` wrapper shape is unnecessary, and adding one would create a second, competing owner of terminal teardown — the exact bug where a `defer` in the wiring races the `defer` in `runLoop` on a signal-driven exit.

The one gap worth naming: `tcell.NewScreen()` allocates a screen object but does not claim the terminal (that happens in `UI.Init`, which calls `screen.Init`). So if construction succeeds and the application then fails *before* `Play` runs — say `FindProcesses` returns "no processes found" — the screen is dropped without `Init` ever having been called, which is harmless. There is no leak to close here.

### 4.6 The logger: built before dispatch, and this is forced, not a carve-out

**Decision: `stderr.NewLogger()` stays in `run()`, constructed before dispatch. No other adapter gets this treatment.**

This is sometimes argued as a judgment call to be defended on "it is cheap and needed everywhere" grounds. It is stronger than that — the existing code leaves no alternative:

1. `cli.NewCLI(service, logger)` already takes a logger as a **second parameter, independent of the inbound port**. `CLI.report` uses it to log any error that is not an `*apperror.Error`, including errors that never reach the application layer at all (an unknown flag, per `TestRunUnknownFlag`). The CLI adapter therefore needs a logger *before* any mode is resolved, in every design. Making the logger lazy would mean the CLI could not report a flag-parse error.
2. `stderr.NewLogger()` is infallible — `func NewLogger() *Logger { return &Logger{w: os.Stderr} }`. It cannot produce a pre-dispatch failure that would itself need a logger to report. The bootstrap regress that motivates "build nothing before dispatch" does not exist here.
3. It acquires no resource: it captures `os.Stderr`, which the process already holds. Unlike `osprocess.NewProcess()` (which shells out to `exec.LookPath("ps")`), `filescore.NewFileScore()` (which resolves the user's config directory), or `tcell.NewScreen()` (which opens the tty), constructing it has no observable effect and cannot fail for environmental reasons. It is not the class of thing §1.2 is about.

The stricter alternative — build nothing before dispatch, and report pre-dispatch errors with a bare `fmt.Fprintln(os.Stderr, ...)` — is therefore worse on its own terms: it would give the program two different error-reporting paths with two different prefixes (the `Logger` writes `error: %s`), for no gain, and would require reworking `CLI.report` and its four tests to accommodate a logger that might not exist yet.

The right way to state the rule, then, is not "the logger is an exception to laziness" but: **laziness applies to adapters that acquire a resource. The logger acquires nothing.** That formulation happens to generate exactly one pre-dispatch construction today, and gives a clear test for any future addition.

---

## 5. Per-mode adapter needs

### 5.1 The real port and adapter inventory

Verified against `internal/application/contract/outbound/` and `internal/infrastructure/`. Note this differs from `hexagonal-arc.md` §3, which is stale (Appendix B).

| Outbound port | Adapter | Constructor | Fallible? |
|---|---|---|---|
| `outbound.Process` | `infrastructure/osprocess` | `NewProcess() (outbound.Process, error)` | yes — `exec.LookPath("ps")` |
| `outbound.Logger` | `infrastructure/stderr` | `NewLogger() *Logger` | no |
| `outbound.ScoreStore` | `infrastructure/filescore` | `NewFileScore() (outbound.ScoreStore, error)` | yes — home dir resolution |
| `outbound.ScoreReporter` | `infrastructure/stdout` | `NewScoreReporter() *ScoreReporter` | no |
| `outbound.Renderer` + `outbound.InputSource` | `infrastructure/tcellui` | `NewUI(tcell.Screen) *UI` | no — but `tcell.NewScreen()` is |

`*tcellui.UI` satisfies both `Renderer` and `InputSource`, which is why `application.NewRunner(proc, store, reporter, ui, ui, logger)` passes it twice.

### 5.2 Which mode needs what

| Port | play | list | lucky | Notes |
|---|---|---|---|---|
| `outbound.Process` | ✅ | ✅ | ✅ | All three discover; lucky and play also kill |
| `outbound.Logger` | ✅ | ✅ | ✅ | Built pre-dispatch (§4.6) |
| `outbound.ScoreStore` | ✅ | ❌ | ❌ | Score board is a game concept |
| `outbound.ScoreReporter` | ✅ | ❌ | ❌ | Prints the game-over summary and leaderboard |
| `outbound.Renderer` | ✅ | ❌ | ❌ | **The one that must not be constructed off the play path** |
| `outbound.InputSource` | ✅ | ❌ | ❌ | Same adapter value as `Renderer` |
| `outbound.ResultPrinter` *(new)* | ❌ | ✅ | ✅ | §5.3 |

The payoff is concentrated in two cells. List and lucky do not construct a screen, so they work over a pipe and in CI. Lucky does not construct `filescore`, so a broken or unwritable `$HOME` cannot stop it killing processes — §9.3 turns that into a test.

### 5.3 A blocker found in the current code: `FindProcesses` writes to stdout

`application/process/service.go`:

```go
func (s *Service) FindProcesses(patterns []string) ([]process.Info, error) {
	processes, err := s.proc.Discover()
	if err != nil {
		return nil, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityFatal, "process discovery failed", err)
	}

	matches := process.Find(toProcessInfos(processes), patterns, s.proc.OwnPID(), s.proc.OwnUID())
	if err := process.ValidateProcesses(matches); err != nil {
		return nil, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(matches), patterns)

	return matches, nil
}
```

That `fmt.Printf` is a direct stdout write from the application layer, with game-specific wording, bypassing every outbound port. It is documented user-facing behavior — `features.md` §Startup step 3 quotes the string.

This matters because list and lucky both want `FindProcesses` and neither is starting a game. Reusing it as-is makes `pidshooter list firefox` print "Starting game…" above a list it will never play, and makes the output of both new modes un-redirectable and untestable through the existing fake conventions.

**Recommendation: lift the write out of `FindProcesses` as the first step of the migration**, into the new `outbound.ResultPrinter` port, dropping the mode-specific suffix:

```go
// contract/outbound/result.go
package outbound

// DiscoveryResult is the view representation of a pattern search's outcome.
type DiscoveryResult struct {
	Patterns []string
	Matches  []ProcessInfo
}

// KillOutcome is the view representation of one attempted kill.
type KillOutcome struct {
	PID  int
	Name string
	Err  error // nil on success
}

// ResultPrinter is the outbound port for non-interactive textual output.
type ResultPrinter interface {
	// PrintDiscovery reports which processes matched the search patterns.
	PrintDiscovery(result DiscoveryResult)
	// PrintKillOutcomes reports the result of each attempted kill.
	PrintKillOutcomes(outcomes []KillOutcome)
}
```

Implemented in the existing `infrastructure/stdout` package (which already owns "write structured results to stdout" and already imports `internal/util` for `FormatBytes`), called by `application.Runner.Run` — which already owns the find → play → record → report orchestration — rather than from inside `process.Service`.

This is the one violation of NG2, and it is forced: without it, the reuse claim in §5.5 is false. It is a user-visible output change (`"Starting game..."` disappears) and needs a `features.md` edit in the same commit.

### 5.4 The "user picks which to kill" step is deliberately unresolved

List mode is described as "print found PIDs, user picks which to kill". Two readings:

- **Flag-driven, two invocations.** `pidshooter list firefox` prints a numbered table and exits; the user then runs `pidshooter lucky firefox` or a narrower pattern. Needs no new port beyond `ResultPrinter`.
- **Interactive, one invocation.** List prints, then reads a selection from stdin. Needs a new inbound-adjacent port — call it `outbound.Prompter` — and a `infrastructure/stdin` adapter, plus a decision about behavior on a non-tty stdin.

**Recommendation: ship the flag-driven form first.** It needs one new port instead of two, it keeps list mode pipe-safe (which was half the motivation in §1.2), and a stdin prompt in a "just tell me the PIDs" mode is the kind of thing that is easy to add later and awkward to remove.

The important structural point is that this choice **does not affect the configurator's design either way**. An interactive picker is one more adapter in the list column of §5.2's table and one more line in `newLister`. That independence is the concrete evidence for G4.

### 5.5 Is `process.Service.Kill` reusable by lucky mode? Yes — verified

```go
func (s *Service) Kill(pid int, procName string, protected bool) (shouldReap bool, err error)
```

It takes three scalars, not a `*game.Target`. `application/process/service.go` imports only `apperror`, `contract/outbound`, and `core/process` — **no `core/game`, no `inbound`**. Everything it does (refuse protected PIDs, `Pin`, `LookupName`, `ValidateName` against the expected name, `handle.Kill()`, map `NotFoundError` to `shouldReap`) is mode-agnostic, and the PID-recycling protection it provides is exactly what an auto-kill-everything mode most needs.

A lucky-mode service can call it in a loop over `FindProcesses` results with no changes:

```go
for _, p := range matches {
	shouldReap, err := s.processSvc.Kill(p.PID, p.Name, p.IsProtected())
	outcomes = append(outcomes, outbound.KillOutcome{PID: p.PID, Name: p.Name, Err: err})
	_ = shouldReap // already gone counts as success for reporting purposes
}
```

Worth recording that this is *better* than `hexagonal-arc.md` §13 documents. That section describes the seam as `Kill(target *game.Target) (killed, shouldReap bool, err error)`, which would have coupled every caller to `core/game`. The live `processKiller` interface in `application/game/service.go` is:

```go
type processKiller interface {
	Kill(pid int, name string, protected bool) (shouldReap bool, err error)
}
```

The narrowing already happened; the document just never caught up (Appendix B).

---

## 6. Impact on the inbound contract

### 6.1 One interface and one config per mode

Mirroring the outbound side's segregation:

```go
// contract/inbound/runner.go
type Runner interface {
	Run(cfg PlayConfig) error
}

// contract/inbound/lister.go
type Lister interface {
	List(cfg ListConfig) error
}

// contract/inbound/lucky.go
type LuckyShooter interface {
	Shoot(cfg LuckyConfig) error
}
```

```go
// contract/inbound/config.go
type PlayConfig struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

type ListConfig struct {
	Patterns []string
}

type LuckyConfig struct {
	Patterns []string
}
```

`ListConfig` and `LuckyConfig` being single-field structs rather than a bare `[]string` parameter is intentional: it leaves room for the flags those modes will grow (`--json`, `--dry-run`, `--kill`) without another signature change, and it keeps the three inbound ports shaped alike.

Rejected alternatives: a kitchen-sink `Config` with mode-irrelevant fields (§1.4); a generic `Runner[C any]`, which buys nothing because the three configs share no useful constraint and would force type parameters through `cli.Providers` and every fake.

### 6.2 Renaming `Config` → `PlayConfig`, but not `Runner`

`hexagonal-arc.md` §13 records why `GamePlayConfig` was shortened to `Config`:

> `GamePlayConfig` became plain `Config` — the `inbound` package qualifier (`inbound.Config`) already disambiguates it from `core/game.Config`, so the prefix was redundant.

The premise is that the qualifier disambiguates. With three configs in the package it stops disambiguating, so the same reasoning that justified the shortening now justifies undoing it. `PlayConfig` it is.

The interface keeps its name. `Runner`/`Run` is admittedly generic once three things "run", and `Player`/`Play` would read better in isolation — but `inbound.Runner` is implemented by `application.Runner`, the top-level composing type that `hexagonal-arc.md` §13 names repeatedly, and renaming the interface would drag `application.Runner`, `runner.go`, `runner_test.go`, `runner_integration_test.go`, and `fake.Runner` with it for a readability gain that the distinct verbs (`Run`/`List`/`Shoot`) mostly already deliver. Recorded as considered and declined.

The `PlayConfig` rename alone touches `contract/inbound/config.go`, `application/runner.go`, `application/validation.go`, `application/game/service.go` (`Play(cfg inbound.Config, …)`), `entrypoint/cli/cli.go`, `testutil/fake/runner.go`, and three test files. Mechanical, but it gets its own commit (§8 step 2).

### 6.3 `validateConfig` is play-specific and should say so

`application/validation.go` is in the top-level `application` package and validates all four `Config` fields:

```go
func validateConfig(cfg inbound.Config) error {
	if err := process.ValidatePatterns(cfg.Patterns); err != nil {
		return err
	}
	if err := movement.ValidateSpeed(cfg.Speed); err != nil {
		return err
	}
	return game.ValidateTimeLimit(cfg.TimeLimit)
}
```

Two of the three checks (`ValidateSpeed`, `ValidateTimeLimit`) are meaningless for list and lucky; only `ValidatePatterns` is shared. When the new modes land, this becomes `validatePlayConfig` plus a shared pattern check, following the precedent in `hexagonal-arc.md` §13 where `application/service/validation.go` was moved wholesale into `application/process/` because "validation was always process-specific". Not part of this refactor — noted so it is not rediscovered as a surprise.

---

## 7. Package and file layout

### 7.1 Target tree (changed entries only)

```
cmd/pidshooter/
├── main.go                       # CHANGED — shrinks to logger + config + Providers + dispatch
├── wire_play.go                  # NEW — newPlayRunner()   (today's run() body)
├── wire_list.go                  # NEW — newLister()
├── wire_lucky.go                 # NEW — newLuckyShooter()
└── wire_test.go                  # NEW — package main, constructor smoke tests

internal/
├── config/
│   └── file.go                   # NEW (§11) — File, PlayDefaults: pure data, no I/O, no YAML import
├── application/
│   ├── contract/
│   │   ├── inbound/
│   │   │   ├── config.go         # CHANGED — Config → PlayConfig; + ListConfig, LuckyConfig
│   │   │   ├── runner.go         # CHANGED — Run(PlayConfig)
│   │   │   ├── lister.go         # NEW  (deferred to list-mode work)
│   │   │   └── lucky.go          # NEW  (deferred to lucky-mode work)
│   │   └── outbound/
│   │       └── result.go         # NEW — ResultPrinter, DiscoveryResult, KillOutcome
│   ├── process/service.go        # CHANGED — the fmt.Printf comes out (§5.3)
│   ├── runner.go                 # CHANGED — PlayConfig; calls ResultPrinter
│   ├── list.go                   # NEW  (deferred — implements inbound.Lister)
│   └── lucky.go                  # NEW  (deferred — implements inbound.LuckyShooter)
├── entrypoint/cli/cli.go         # CHANGED — Providers, subcommand tree, config.File-aware flag defaults
├── infrastructure/stdout/
│   └── result_printer.go         # NEW — implements outbound.ResultPrinter
├── infrastructure/fileconfig/
│   └── fileconfig.go             # NEW (§11) — Load(): reads + parses ~/.config/pidshooter/config.yaml
└── testutil/fake/
    ├── result_printer.go         # NEW — follows fake.ScoreReporter's shape
    ├── lister.go                 # NEW  (deferred)
    └── lucky_shooter.go          # NEW  (deferred)
```

No new top-level package under `internal/application`, `internal/infrastructure`, or `internal/entrypoint` — `internal/config` is a new top-level package, but it sits at the same privilege level as `internal/util` (§11.4 justifies this in detail: it is pure data, no I/O, importable by `core` in principle though nothing there needs it today). `cmd/pidshooter` keeps the privileges it already had, and `fileconfig` is one more infrastructure adapter next to `filescore`, `osprocess`, and the rest — not a second composition root.

### 7.2 Target `main.go`

```go
package main

import (
	"os"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/entrypoint/cli"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/stderr"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	// The logger is the only adapter built before the mode is known: it
	// acquires nothing (it captures os.Stderr), cannot fail, and the CLI
	// adapter needs it to report flag-parse errors that never reach the
	// application layer at all.
	logger := stderr.NewLogger()

	providers := cli.Providers{
		Play:  func() (inbound.Runner, error) { return newPlayRunner(logger) },
		List:  func() (inbound.Lister, error) { return newLister(logger) },
		Lucky: func() (inbound.LuckyShooter, error) { return newLuckyShooter(logger) },
	}

	return cli.NewCLI(providers, logger).Run(os.Args[1:])
}
```

Note what is gone: `tcell`, `tcellui`, `osprocess`, `filescore`, `stdout`, `application`, and `fmt` are no longer imported by `main.go` — they moved to the `wire_*.go` files, where each import is adjacent to the one mode that justifies it.

### 7.3 Target `buildCommand`

```go
func (c *CLI) buildCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "pidshooter <pattern> [pattern2] [pattern3...]",
		Short: "Process ID Shooter",
		Long:  `Hunt running processes by name and kill them in a terminal shooter game.`,
		Example: `  pidshooter firefox
  pidshooter chrome firefox node
  pidshooter firefox --confirm
  pidshooter node --speed=2.5 --time=60
  pidshooter list firefox
  pidshooter lucky firefox`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          c.play, // bare invocation stays play mode
	}

	root.SetOut(c.out)
	root.SetErr(c.errOut)
	c.registerPlayFlags(root)

	playCmd := &cobra.Command{
		Use:   "play <pattern> [pattern2...]",
		Short: "Play the shooter game against matching processes",
		RunE:  c.play,
	}
	c.registerPlayFlags(playCmd)

	root.AddCommand(
		playCmd,
		&cobra.Command{
			Use:   "list <pattern> [pattern2...]",
			Short: "Print matching processes without killing anything",
			RunE:  c.list,
		},
		&cobra.Command{
			Use:   "lucky <pattern> [pattern2...]",
			Short: "Kill every matching process, then report what happened",
			RunE:  c.lucky,
		},
	)

	return root
}
```

`SetOut`/`SetErr` on the root propagate to subcommands, so `cli_test.go`'s `io.Discard` redirection keeps working unchanged.

This is the tree as of migration Step 6 (§8), before the config file exists. §11.8 shows the version where `root`'s flags and `RunE` are resolved from config instead of hardcoded to `play`.

---

## 8. Migration plan

Nine steps. Each compiles and leaves `make test` (which runs `go test -tags integration ./...`) green. Steps 0–7 are this proposal; step 8 is the work it exists to enable. (The original six-step cut of this plan under-counted its own last two steps — "step 6 is the work it exists to enable" described what is actually step 7; fixed here rather than left to drift further as steps are added.)

**Step 0 — Confirm the baseline net.** No code change. `cli_test.go` already covers no-args help, `--help`/`-h`, single and multiple patterns, `--confirm`, `--speed` (valid/min/max/invalid), `--time` (valid/zero/invalid), unknown flags, flags-without-patterns, and all four `report` logging paths. That is a sufficient regression net for a signature change to `NewCLI`; nothing needs adding first.

**Step 1 — Lift the stdout write out of `FindProcesses` (§5.3).** Add `outbound.ResultPrinter` and `infrastructure/stdout`'s implementation, add `fake.ResultPrinter`, thread it through `application.NewRunner`, delete the `fmt.Printf` from `process.Service.FindProcesses`, and call `PrintDiscovery` from `application.Runner.Run`. Update `features.md` §Startup step 3. This is the only step with user-visible output change, and it is deliberately first so the rest of the migration is behavior-preserving.

**Step 2 — Rename `inbound.Config` → `inbound.PlayConfig` (§6.2).** Pure mechanical rename across seven files. Own commit, so the later diffs stay readable.

**Step 3 — Extract `newPlayRunner` (§4.2).** Move today's `run()` body into `cmd/pidshooter/wire_play.go`, drop the two `logger.Error` calls and the comment that explains them, and have `run()` call `newPlayRunner(logger)` and pass the result to the *existing* `NewCLI(runner, logger)`. **Still eager** — construction still happens before `cli.Run`. No behavior change except that construction errors are now logged by `CLI.report` instead of inline… which they are not yet, because `run()` still returns them directly. So: no behavior change at all. Verify by hand that `pidshooter firefox`, `pidshooter --help`, and `pidshooter` still behave identically.

**Step 4 — Subcommand tree, still eager.** Add `play` as an explicit subcommand aliasing the root's `RunE`, extract `registerPlayFlags`, fix the `--speed` help range. Add `cli_test.go` cases for `pidshooter play firefox` matching `pidshooter firefox`, and for `pidshooter -- list` treating `list` as a pattern. No construction changes.

**Step 5 — Make it lazy.** The crux. Change `NewCLI` to take `cli.Providers`; move the `newPlayRunner` call into `CLI.play` after the help guard; update `newTestCLI` in `cli_test.go`. Add the laziness regression test (§9.2) using a stub `List` provider. After this step, `pidshooter --help` no longer touches the terminal, `ps`, or `$HOME` — which is directly checkable by hand: `pidshooter --help | cat` should be clean.

**Step 6 — Mode skeletons.** Add `inbound.Lister`/`inbound.LuckyShooter`, `newLister`/`newLuckyShooter` returning minimal services, and the `list`/`lucky` subcommands. Even a service whose `List` only calls `FindProcesses` and `PrintDiscovery` is enough to make the laziness tests meaningful against a real non-play path. Update `features.md`.

**Step 7 — Layered configuration (§11).** Add `internal/config` (the pure `File`/`PlayDefaults` shape), `internal/infrastructure/fileconfig` (the YAML-reading adapter), and `newFileConfig()` called from `main.go` alongside the logger. Change `registerPlayFlags` to take a `config.PlayDefaults` and use it as each flag's default instead of a literal. Change `buildCommand` to resolve `root`'s flags and `RunE` from `cfg.DefaultMode` (§11.8) instead of hardcoding `play`. Add tests: config absent behaves identically to today (regression net for every existing test in `cli_test.go`); a config setting `default_mode: lucky` routes bare `pidshooter firefox` to the lucky provider; a config setting `play.speed: 3.0` changes the bare invocation's default while `pidshooter firefox --speed=1.5` still wins; a malformed config file logs a warning and falls back to code defaults rather than erroring. This step depends on Step 6 only for `list`/`lucky` to be valid `default_mode` values to test against — the mechanism itself would work with `play` alone.

**Step 8 — Real list/lucky logic.** Out of scope here (NG1).

---

## 9. Testing strategy

Grounded in the conventions the repo already has, which are: `stretchr/testify` (`assert` + `require`), hand-written fakes in `internal/testutil/fake` (one struct per port, exported fields for inputs, captured values for assertions), fixtures in `internal/testutil/fixture`, white-box tests in the package under test where unexported fields need setting, and integration tests behind `//go:build integration` with a `TestIntegration` name prefix (Makefile: `go test -tags integration -run TestIntegration`).

### 9.1 New fakes, following the existing shape

`fake.Runner` is the template — it captures the config it was called with and returns a preset error:

```go
type Runner struct {
	Err error
	Cfg inbound.Config
}

func (r *Runner) Run(cfg inbound.Config) error {
	r.Cfg = cfg
	return r.Err
}
```

`fake.Lister` and `fake.LuckyShooter` copy it verbatim against their own configs. `fake.ResultPrinter` follows `fake.ScoreReporter` instead, which captures the last call by pointer so "never called" is distinguishable from "called with a zero value":

```go
type ScoreReporter struct {
	Reported *outbound.ScoreSummary // nil if Report was never called
}
```

No mock framework, no codegen — matching what is there.

### 9.2 The laziness test, which needs no terminal at all

This is the test the whole refactor exists to make possible, and the key realization is that it does **not** require constructing or faking tcell. Because `cli.Providers` is a struct of plain functions, "was the play provider called?" is directly observable from a white-box test in `package cli`:

```go
func TestRunListDoesNotInvokePlayProvider(t *testing.T) {
	playCalled := false
	p := Providers{
		Play: func() (inbound.Runner, error) {
			playCalled = true
			return &fake.Runner{}, nil
		},
		List: func() (inbound.Lister, error) { return &fake.Lister{}, nil },
	}

	require.NoError(t, newTestCLIWith(p).Run([]string{"list", "firefox"}))
	assert.False(t, playCalled, "list mode must not construct the interactive game's adapters")
}
```

Plus the mirror cases: `lucky` does not invoke `Play`; `--help` and a bare `pidshooter` invoke *nothing*; `play`/bare-with-patterns invoke only `Play`; and a provider returning an error propagates it out of `Run` and gets logged exactly once by `report` (asserting on `fake.Logger.Errors`, as `TestRunLogsUnrecognizedError` already does).

These run in milliseconds, need no tty, no `ps`, no writable `$HOME`, and no build tag. The architectural property is enforced at the seam where it is decided, rather than inferred from the absence of a side effect somewhere downstream.

### 9.3 Constructor smoke tests in `package main`

A small `cmd/pidshooter/wire_test.go`, kept deliberately thin:

- `newLister(logger)` and `newLuckyShooter(logger)` **succeed with `$HOME` unset**, proving they do not construct `filescore`. `internal/testutil/helper.UnsetEnv(t, "HOME")` exists for exactly this and is already used against the filesystem adapters — it unsets rather than setting to `""`, so the genuinely-absent case is exercised.
- `newPlayRunner(logger)` is **not** given a success-path test. It calls `tcell.NewScreen()`, which fails without a controlling terminal and would make the test pass or fail based on how it was invoked. A test whose result depends on whether `go test` was run from a terminal is worse than no test.

### 9.4 What `tcell.SimulationScreen` can and cannot do here

`internal/infrastructure/tcellui/tcellui_test.go` does use it — seven call sites, including a `newUI(t) (*tcellui.UI, tcell.SimulationScreen)` helper — so the pattern is established and the package's rendering behavior is well covered without a real terminal. But the substitution happens at `tcellui.NewUI(screen)`, with the test constructing the screen:

```go
screen := tcell.NewSimulationScreen("")
ui := tcellui.NewUI(screen)
```

The configurator's job is the line *above* that — `tcell.NewScreen()` — which is precisely what a simulation screen cannot stand in for. Making it substitutable would mean threading a `func() (tcell.Screen, error)` factory through `newPlayRunner` purely for the benefit of a test.

**Recommendation: do not add that seam.** `tcell.NewScreen()` is a single library call with no branching; there is nothing in `newPlayRunner` worth testing that a compile does not already prove. The property that actually matters — that it is unreachable from list and lucky — is fully covered by §9.2 without it. This is the correction to the instinct that the new modes need heavyweight environment-sensitive integration tests: they need one cheap unit test at the right seam.

### 9.5 Making the dependency rule mechanical

`hexagonal-arc.md` §12 offers this as the enforcement story:

> **Clear dependency rule** — `grep -r "tcell" internal/core` must return nothing. The build enforces the architecture.

That is a manual grep, and it only covers core→tcell. There is no CI in this repo (no `.github/`, no `.golangci.yml`); the automation surface is the `Makefile`. So the right place for a real check is a make target, in the same idiom:

```make
## arch: Check the hexagonal dependency rule
arch:
	@go list -deps ./internal/core/... | grep -q 'pidshooter/internal/\(application\|infrastructure\|entrypoint\)' \
		&& { echo "FAIL: core imports an outer layer"; exit 1; } || true
	@go list -deps ./internal/application/... | grep -q 'pidshooter/internal/\(infrastructure\|entrypoint\)' \
		&& { echo "FAIL: application imports an adapter"; exit 1; } || true
	@go list -deps ./internal/entrypoint/... | grep -q 'pidshooter/internal/infrastructure' \
		&& { echo "FAIL: entrypoint imports infrastructure"; exit 1; } || true
	@echo "architecture OK"
```

The third clause is new with this proposal and is what keeps §4.3's decision honest: it fails the build if anyone later wires `entrypoint/cli` straight to an adapter. Wiring it into the `all` target makes the rule enforced rather than remembered.

---

## 10. Trade-offs, decisions, and open questions

**10.1 `package main` costs discoverability, and that is the accepted price.** Wiring in `cmd/pidshooter` cannot be imported, does not appear in `go doc`, and could not be shared with a second entrypoint. The first two are irrelevant (nothing should import it; it has no API). The third is real but hypothetical, and the fix — promote three files into a package — is mechanical if a second entrypoint ever exists. The alternative, a new `internal/composition` package, would need `hexagonal-arc.md` §9 amended from "all wiring happens here and nowhere else" to something weaker, and would create the most privileged package in the tree. The decision is reversible; the documentation change is not free. *Decided: `cmd/pidshooter`.*

**10.2 The inbound port surface roughly triples.** Three interfaces, three configs, three fakes, where there is one of each today. For a project of this size that is real overhead, and §12 of `hexagonal-arc.md` already lists "more packages" as an accepted cost of the existing structure. The alternative — one `Config` with fields meaningful in one mode out of three — trades compile-time clarity for less typing, in a codebase whose documented values run the other way (the outbound side already has six single-purpose ports rather than one `Environment`).

**10.3 The subcommand names collide with pattern names.** `pidshooter list` changes meaning. Mitigated by `--`, documented in `features.md`, and judged acceptable because a three-character-minimum pattern of exactly `play`, `list`, or `lucky` is an unlikely thing to hunt. Alternative considered: prefix the subcommands (`pidshooter --list`). Rejected because it reintroduces the hand-written branch point §3.1 is trying to avoid.

**10.4 Renderer lifecycle stays with `game.Service` — verified, not assumed.** `runLoop` calls `s.renderer.Init()` and `defer s.renderer.Cleanup()`. The configurator constructs and stops. No `Cleanup` returned from provider functions; adding one would create a second owner of terminal teardown racing the existing `defer` on a signal-driven exit.

**10.5 Naming is provisional.** `Providers`, `newPlayRunner`/`newLister`/`newLuckyShooter`, `wire_*.go`, `LuckyShooter.Shoot`, `ResultPrinter`. `hexagonal-arc.md` §13 shows this project renames deliberately and late; none of these should be treated as settled before the code exists.

**10.6 `ResultPrinter` is the one speculative port, and §5.3 forces it early.** It has to land in step 1 to unblock reuse of `FindProcesses`, before either consuming mode is written — so its method set is a guess informed by one known caller. `PrintDiscovery` is certain (it replaces existing behavior); `PrintKillOutcomes` is a sketch for lucky mode and may well change shape when that mode is actually implemented. Splitting it into two single-method ports later is cheap; over-designing it now is not.

**10.7 Interactive PID picking is unresolved by design (§5.4).** Recommending the flag-driven form, noting that it does not constrain the configurator either way.

**10.8 `hexagonal-arc.md` is stale in several places this proposal touches.** The design record still refers to `scorefilestore`, `stderrlog`, `outbound.ProcessManager`, `Lister`/`Killer`, `testutil/capture`, and a `killer` interface signature taking `*game.Target` — none of which exist. That is not this proposal's problem to fix wholesale, but it does mean the architecture rules had to be read from §§1, 8, 9, and 12 (which are still accurate, being about structure rather than names) rather than from the package inventory in §§3, 5, and 7. Appendix B lists what needs updating.

**10.9 The config-file load is a second, structurally different exception to "nothing is constructed before the mode is known" (§11.5).** §4.6 justified the logger's pre-dispatch construction because it is infallible and acquires no resource — `fileconfig.Load()` is neither: it opens and reads a file, and a malformed one is a real error. The justification here is different in kind, not a stretch of the same one: default-mode resolution is part of *deciding* the mode, not part of *acting on* it, so the config file has to be read before dispatch for the same reason argv does. §11.5 states the generalized rule so a third exception, if one is ever proposed, has to clear the same bar rather than being waved through by precedent creep.

**10.10 `gopkg.in/yaml.v3` needs no new dependency to add — verified.** `go mod why gopkg.in/yaml.v3` shows it already required, indirectly, via `stretchr/testify/assert/yaml`; it is in `go.sum` today. Giving `internal/infrastructure/fileconfig` a real import of it only moves the `require` line from indirect to direct on the next `go mod tidy` — no new module is fetched, no new supply-chain surface is added. If `yaml.v3` is rejected regardless (e.g. on the theory that a test-only transitive dependency could be dropped later, breaking this assumption), the same design works with `encoding/json` and a `.json` config file at zero dependency cost, at the price of comments and the trailing-comma friendliness YAML gives a hand-edited preferences file — the schema in §11.7 does not otherwise change.

**10.11 The config-driven `default_mode` does not add a second collision surface.** §3.2/§10.3 already accept that a pattern spelled exactly `list`, `play`, or `lucky` needs `--` to be treated as a pattern rather than a subcommand. `default_mode` does not change this: an explicit subcommand name in argv always wins outright over whatever the config file says (§11.1), so the collision remains exactly the one §10.3 already describes, not a new one layered on top of it.

**10.12 Config-file errors are non-fatal by decision, matching an existing precedent.** A missing file, an unreadable one, a malformed one, or an unrecognized `default_mode` value all resolve to "warn and fall back to code defaults" rather than aborting the program (§11.6). This mirrors `score.Service`'s own leniency toward a missing or corrupted score board (`outbound.NotFoundError`/`CorruptedDataError` are both absorbed, not fatal) — an optional preferences file failing to parse is judged the same kind of problem: real, worth surfacing, not worth blocking the user's actual request over.

---

## 11. Layered configuration: default mode and per-mode defaults

Everything above answers "which adapters get built for a given mode." This section answers a related but separate question the owner raised after reading it: **which mode, and with what flag defaults, when the user does not say?** Today that answer is hardcoded twice — `play` is hardcoded as what the bare invocation means (§3.2), and `2.0`/`30`/`false` are hardcoded as `--speed`/`--time`/`--confirm`'s defaults (§3.2's `registerPlayFlags`). The request is to make both overridable by a persisted config file, with three-tier precedence: **code default → config file → CLI flag**, the CLI flag always winning when the user actually types one.

### 11.1 Two independent precedence chains, not one

"Mode" and "a mode's own flag values" are resolved by different mechanisms in this design (§3.1: mode is a subcommand, not a flag), so they need separate precedence rules rather than one generic "config overrides code, flags override config" statement that would gloss over how each actually resolves:

- **Mode selection.** An explicit subcommand name in argv (`play`, `list`, `lucky`) wins outright — there is no way to "half-specify" a subcommand, so this tier is binary, not a value to merge. Absent one, the config file's `default_mode` applies if set to a recognized name. Absent that, the code default is `play` — today's only behavior, preserved as the floor.
- **Per-mode flag values.** Within whichever mode runs, each flag resolves independently: the CLI flag wins if the user passed it; otherwise the config file's value for that flag, if set; otherwise cobra's built-in default (the literals in today's `registerPlayFlags`).

The reason these are separable is that cobra itself already implements the second chain's bottom two tiers for free, once the "config file" tier is expressed as the flag's registered default instead of a hardcoded literal — this is the whole mechanism, and it is smaller than it sounds. §11.2 shows it.

### 11.2 Per-mode flag defaults: feed the config value in as the flag's default, don't merge after parsing

The tempting but more complex approach is to parse flags against their hardcoded defaults as today, then afterward check `cmd.Flags().Changed("speed")` and overwrite `c.speed` with the config value when the flag was not explicitly passed. That is a second data path a reader has to trace next to the one cobra already provides, and it has to be repeated per flag, per mode.

The simpler mechanism: resolve the config-sourced default *before* calling `cmd.Flags().Float64Var(...)`, and hand cobra that value instead of the literal. cobra's own precedence — a flag's bound variable ends up at whichever value pflag last wrote, and pflag only writes the user's value when the flag was actually present on the command line — then *is* the merge, with no `Changed()` check anywhere in this package's code:

```go
// registerPlayFlags binds play's flags, using defaults resolved from cfg
// (code defaults where cfg leaves a field unset) as each flag's own
// default. A user-supplied --speed/--time/--confirm always overrides
// this value, because that is what a cobra flag default means; nothing
// here has to ask cobra whether the user typed it.
func (c *CLI) registerPlayFlags(cmd *cobra.Command, cfg config.PlayDefaults) {
	cmd.Flags().BoolVar(&c.confirm, "confirm", cfg.Confirm.OrElse(false), "Ask for confirmation before killing")
	cmd.Flags().Float64Var(&c.speed, "speed", cfg.Speed.OrElse(2.0), "Speed multiplier (range: 0.5-5.0)")
	cmd.Flags().IntVar(&c.timeLimit, "time", cfg.Time.OrElse(30), "Time limit in seconds (0 = no limit)")
}
```

`config.PlayDefaults`'s fields are optional (§11.7 — `Optional[float64]`, not a bare `float64`), because `0` is a legitimate configured value for `--time` (no limit) and `false` is a legitimate configured value for `--confirm`, and a bare zero-value field could not be told apart from "the user's config file does not mention this key." `OrElse` is the one method that type needs.

This also means `application.validateConfig` (§6.3) needs no new code: a config-sourced `speed: 999` reaches `movement.ValidateSpeed` exactly the way a CLI-sourced `--speed=999` does, because by the time either reaches `inbound.PlayConfig` they are the same value read from the same struct field — the config file does not get a separate, weaker validation path.

### 11.3 Mode selection: resolve `default_mode` before `buildCommand` runs, then build root to match

Cobra's command tree is ordinary Go values built once, before `Execute()` parses anything (§3.1 already relies on this — it's why subcommand dispatch is structural). Because the config file is loaded once, in `main()`, before `cli.NewCLI(...)` is even called (§11.5 justifies why this one load is allowed to happen that early), `buildCommand` has the resolved default mode available at tree-construction time, not just at run time. So instead of `root`'s `RunE` being hardcoded to `c.play` (§3.2's sketch), it is set to whichever mode's `RunE` the config resolved, and `root`'s flags are registered by calling that mode's own flag-registration helper — the same one the named subcommand also calls, so the two can never drift apart. §11.8 shows the full `buildCommand`.

An unrecognized `default_mode` value (a typo, a value from a future pidshooter version's config schema) does not fail the whole program — see §11.6 — it resolves to `play`, logged as a warning, identically to a missing key.

### 11.4 Where this lives: a pure data package, plus one more infrastructure adapter — not a fourth exception to §4.1

§4.1 already settled where infrastructure-touching wiring code lives (`cmd/pidshooter`, `package main`) and §4.3 already settled that `entrypoint/cli` never imports infrastructure. Config loading needs to respect both, and does, by splitting into two pieces with different privileges — the same split `hexagonal-arc.md` §3 already uses to justify keeping `fsutil` out of `internal/util`, applied in the opposite direction this time:

- **`internal/config`** — a new top-level package holding only the *shape* of a loaded config file: `File`, `PlayDefaults`, `Optional[T]`. No file I/O, no YAML import, no error type of its own. It is pure Go with no OS access, which is the exact test `hexagonal-arc.md` §3 uses to justify `internal/util`'s standing (`internal/util` "is safe for `core` to import... only because it's pure Go with no OS access") — `internal/config` clears the same bar, so it sits beside `util`, not beside `fsutil`. `internal/entrypoint/cli` imports it (to read `cfg.DefaultMode` and pass `cfg.Play` into `registerPlayFlags`); nothing under `internal/core` or `internal/application` needs to import it today, but nothing would be structurally wrong if a future mode's application-layer defaults needed to.
- **`internal/infrastructure/fileconfig`** — a new adapter, the same shape as `filescore`: `Load() (config.File, error)` resolves `fsutil.ConfigDir("pidshooter")` (the exact call `filescore.defaultPath` already makes — confirmed by reading it, §Appendix A) joined with `config.yaml`, reads it, and unmarshals YAML into `config.File`. This is constructed exactly once, from `cmd/pidshooter`, next to `osprocess.NewProcess()` and `filescore.NewFileScore()` — it is one more infrastructure package `main` knows about, not a new privileged layer.

`cli` therefore imports `internal/config` (a data shape, no different in kind from importing `inbound.PlayConfig`) but never `internal/infrastructure/fileconfig` (the adapter that produces one) — `main` calls `fileconfig.Load()` once and hands `cli.NewCLI` the resulting `config.File` value, the same way it already hands `cli.NewCLI` a `Providers` struct of closures rather than letting `cli` construct application services itself. §4.3's decision is not reopened; it is applied to a second kind of dependency the same way.

### 11.5 The generalized pre-dispatch rule (revising §4.6)

§4.6 asked "is this adapter forced to exist before the mode is known?" and answered yes for the logger, on two grounds: it is infallible, and it acquires no resource. `fileconfig.Load()` fails neither test cleanly — it opens a file (a resource) and can fail (bad YAML) — so it cannot be waved through on the same reasoning without weakening what "forced, not a carve-out" was supposed to mean. It needs its own, honestly different justification:

> An adapter may run before the mode is known only if constructing it is itself part of *deciding* the mode (not part of acting on a mode already chosen), or it is needed identically, at negligible and mode-independent cost, by every mode that could be chosen — and even then, a failure must degrade to the code default rather than abort the program, since nothing this early has a mode-specific way to report it.

`fileconfig.Load()` clears this on the first branch directly: §11.3 cannot resolve `root`'s `RunE` without knowing `default_mode`, and the only two sources of that value are the code default and the file. It also clears the second branch independently: unlike `tcell.NewScreen()` (the adapter this whole document exists to keep out of non-`play` modes), reading one small file claims no terminal, blocks on no user input, and writes nothing — there is no mode for which doing this eagerly is wrong the way constructing a screen is wrong for `list`/`lucky`. Both the logger and the file config satisfy this rule; they satisfy it for different reasons, which is exactly why §4.6 needed revising rather than merely citing as precedent. No other adapter in this document's scope satisfies either branch — `osprocess`, `filescore`, and `tcellui` all stay lazy.

### 11.6 Failure handling

None of the following abort the program; `main` has no `apperror.Handler` this early (the existing comment on today's `osprocess`/`filescore` calls already says as much — this joins that same category, a third pre-`Handler` caller of `logger.Error`/`logger.Warn` directly):

- **File absent.** The common case (nobody has created `~/.config/pidshooter/config.yaml`). `fileconfig.Load()` returns a zero-value `config.File{}` and a nil error — not `outbound.NotFoundError`, because there is no `outbound` port here to satisfy (§11.4) and nothing downstream needs to distinguish "absent" from "present but empty." Every `Optional[T]` field's zero value means "unset," so this is silently identical to today's behavior.
- **File present but malformed (bad YAML).** `logger.Warn("config file unreadable, using defaults: " + err.Error())`, then proceed with `config.File{}`. A broken preferences file should not block the user's actual request.
- **`default_mode` set to something other than `play`/`list`/`lucky`.** Same treatment: `logger.Warn(...)`, fall back to code default `play`.
- **A per-mode value that is syntactically valid but semantically out of range** (`play.speed: 999`) is deliberately *not* caught here — §11.2 already routes it through the same validation the CLI flag path uses, so catching it twice would be redundant, and duplicating `movement.ValidateSpeed`'s range logic here recreates exactly the two-diverging-copies problem §3.2's `registerPlayFlags` extraction was written to avoid for `play`/root's flag literals.

### 11.7 Schema

```yaml
# ~/.config/pidshooter/config.yaml — every key optional; absent means "use the code default"
default_mode: lucky   # play | list | lucky

play:
  speed: 3.0
  time: 60
  confirm: true

# list: {}    # reserved — no per-mode fields exist yet (NG3, §5.4)
# lucky: {}   # reserved — no per-mode fields exist yet (NG3)
```

```go
// internal/config/file.go
package config

// Optional distinguishes "the config file did not set this" from "the
// config file set this to the zero value" — both true and 0 are
// legitimate configured values for Confirm and Time.
type Optional[T any] struct {
	Value T
	Set   bool
}

func (o Optional[T]) OrElse(fallback T) T {
	if o.Set {
		return o.Value
	}
	return fallback
}

// File is the parsed shape of config.yaml. Every field is optional;
// a zero-value File (no config file present) is a valid input meaning
// "use every code default," which is today's only behavior.
type File struct {
	DefaultMode string       `yaml:"default_mode"`
	Play        PlayDefaults `yaml:"play"`
}

// PlayDefaults overrides play mode's own flag defaults. Field names
// mirror inbound.PlayConfig, not the CLI flag names, since a config key
// is closer to "a persisted Config value" than to "a flag."
type PlayDefaults struct {
	Speed   Optional[float64] `yaml:"speed"`
	Time    Optional[int]     `yaml:"time"`
	Confirm Optional[bool]    `yaml:"confirm"`
}
```

```go
// internal/infrastructure/fileconfig/fileconfig.go
package fileconfig

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/config"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/fsutil"
)

// Load reads and parses the per-user config file. A missing file is not
// an error: it returns a zero-value config.File, identical in effect to
// every field being explicitly unset. A malformed file is returned as an
// error for the caller to log and recover from (§11.6) — Load itself
// never falls back silently, so a caller cannot forget to report it.
func Load() (config.File, error) {
	dir, err := fsutil.ConfigDir("pidshooter")
	if err != nil {
		return config.File{}, err
	}
	path := filepath.Join(dir, "config.yaml")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return config.File{}, nil
	}
	if err != nil {
		return config.File{}, err
	}

	var file config.File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return config.File{}, err
	}
	return file, nil
}
```

`fsutil.ConfigDir("pidshooter")` is the exact call `filescore.defaultPath` already makes for `highscores.json` (confirmed by reading it — Appendix A); `config.yaml` lands in the same directory as a sibling file, not a new location the user has to discover.

### 11.8 `buildCommand`, revised to resolve root from config instead of hardcoding `play`

Supersedes §7.3's sketch, which stands as the correct Step 6 state (§8) — this is what Step 7 changes it into:

`buildCommand` keeps its existing no-argument signature (§7.3) — `cfg` is a field on `CLI` (`NewCLI(providers, logger, cfg)` stores it, the same way it already stores `providers` and `logger`), not a new parameter threaded through, so this stays consistent with how `confirm`/`speed`/`timeLimit` are already struct fields rather than values passed around between methods:

```go
func (c *CLI) buildCommand() *cobra.Command {
	playCmd := &cobra.Command{
		Use:   "play <pattern> [pattern2...]",
		Short: "Play the shooter game against matching processes",
		RunE:  c.play,
	}
	c.registerPlayFlags(playCmd, c.cfg.Play)

	listCmd := &cobra.Command{
		Use:   "list <pattern> [pattern2...]",
		Short: "Print matching processes without killing anything",
		RunE:  c.list,
	}
	luckyCmd := &cobra.Command{
		Use:   "lucky <pattern> [pattern2...]",
		Short: "Kill every matching process, then report what happened",
		RunE:  c.lucky,
	}

	root := &cobra.Command{
		Use:           "pidshooter <pattern> [pattern2] [pattern3...]",
		Short:         "Process ID Shooter",
		Long:          `Hunt running processes by name and kill them in a terminal shooter game.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.bindDefaultMode(root, playCmd, listCmd, luckyCmd, c.cfg.DefaultMode)

	root.SetOut(c.out)
	root.SetErr(c.errOut)
	root.AddCommand(playCmd, listCmd, luckyCmd)
	return root
}

// bindDefaultMode makes root's own flags and RunE match whichever
// subcommand defaultMode names, falling back to play for an empty or
// unrecognized value — the same fallback for "key absent" and "key
// present but not one of play/list/lucky" (§11.6). Copying the resolved
// command's already-registered FlagSet onto root (rather than root
// calling a second, mode-specific "register flags" function itself)
// is what keeps this function correct as list/lucky gain real flags
// later without editing it again: whatever registerPlayFlags,
// registerListFlags, etc. bound onto their own command is exactly what
// root should expose, by construction, with zero duplication.
func (c *CLI) bindDefaultMode(root, play, list, lucky *cobra.Command, defaultMode string) {
	resolved := play
	switch defaultMode {
	case "list":
		resolved = list
	case "lucky":
		resolved = lucky
	case "", "play":
		// falls through with resolved already set to play
	default:
		c.logger.Warn(fmt.Sprintf("config: unrecognized default_mode %q, using play", defaultMode))
	}
	root.RunE = resolved.RunE
	root.Flags().AddFlagSet(resolved.Flags())
}
```

Today `list.Flags()` and `lucky.Flags()` are empty sets (§5.4, NG3), so `AddFlagSet` on those two branches is currently a no-op — correct, not a stub, since there is nothing yet for the bare invocation to inherit from either mode. The function does not need to change when `list`/`lucky` gain their own flags later; whatever a future `registerListFlags` binds onto `listCmd` is what `AddFlagSet` will pick up.

### 11.9 What changes in `main.go`

```go
func run() error {
	logger := stderr.NewLogger()

	cfg, err := fileconfig.Load()
	if err != nil {
		logger.Warn("config file unreadable, using defaults: " + err.Error())
		cfg = config.File{}
	}

	providers := cli.Providers{
		Play:  func() (inbound.Runner, error) { return newPlayRunner(logger) },
		List:  func() (inbound.Lister, error) { return newLister(logger) },
		Lucky: func() (inbound.LuckyShooter, error) { return newLuckyShooter(logger) },
	}

	return cli.NewCLI(providers, logger, cfg).Run(os.Args[1:])
}
```

One new adapter call (`fileconfig.Load()`), handled inline exactly as `osprocess`/`filescore` are handled today when there is no `Handler` yet to route through (§11.6) — except this one degrades instead of returning, since a broken preferences file is not fatal the way a missing `ps` binary is.

---

## Appendix A — Verification log

Everything below was read at commit `66ddc02`.

| Checked | Finding |
|---|---|
| `cmd/pidshooter/main.go` | 53 lines; `run()` builds logger, `osprocess`, `filescore`, `tcell.NewScreen`, `tcellui.NewUI`, `stdout.NewScoreReporter`, then `application.NewRunner(proc, store, reporter, ui, ui, logger)`, then `cli.NewCLI(runner, logger).Run(os.Args[1:])`. `main()` discards the error and only calls `os.Exit(1)` — it does **not** print, contrary to `hexagonal-arc.md` §9's sketch. |
| `internal/application/game/service.go` | **Confirms the renderer-lifecycle assumption.** `runLoop` opens with `if err := s.renderer.Init(); err != nil { … }` then `defer s.renderer.Cleanup()`. The composition layer needs no Init/Cleanup responsibility. |
| `internal/application/game/service.go` | `processKiller` is `Kill(pid int, name string, protected bool) (shouldReap bool, err error)` — scalars, not `*game.Target`. |
| `internal/infrastructure/tcellui/tcellui_test.go` | **Confirms `SimulationScreen` is established practice** — 7 call sites, plus a `newUI(t) (*tcellui.UI, tcell.SimulationScreen)` helper at line 338. But every one injects the screen into `tcellui.NewUI`; none substitutes for `tcell.NewScreen()`, which is what the configurator calls (§9.4). |
| `internal/entrypoint/cli/cli.go` | Single root command, `RunE: c.play`, no subcommands. Flags bound to `CLI` struct fields. **Defaults confirmed: `--speed` 2.0, `--time` 30, `--confirm` false.** Help text says `range: 0.1-5.0`, which is **wrong** — `movement.MinSpeed` is 0.5 and `features.md` documents 0.5–5.0. `NewCLI(service inbound.Runner, logger logger)` — the logger is already a separate parameter, which decides §4.6. |
| `internal/entrypoint/cli/cli_test.go` | White-box (`package cli`), 14 tests, `newTestCLI` sets `c.out`/`c.errOut` to `io.Discard`. `TestRunBasicPattern` pins 2.0/30. A sufficient baseline net already exists. |
| `internal/application/contract/inbound/*.go` | `Runner{Run(Config) error}`; `Config{Patterns, ConfirmMode, Speed, TimeLimit}`. Field names as used in the document. |
| `internal/application/contract/outbound/*.go` | Six ports: `Process` (`Discover`/`OwnPID`/`OwnUID`/`LookupName`/`Pin`) + `ProcessHandle`, `Logger` (`Warn`/`Error`), `ScoreStore`, `ScoreReporter`, `Renderer`, `InputSource`. **No `ProcessManager`, no `Lister`, no `Killer`** — `hexagonal-arc.md` §§3/7/13 are stale. `ui.go` documents the Init-before-use / idempotent-Cleanup contract. |
| `internal/application/process/service.go` | **`Kill` is genuinely mode-agnostic** (§5.5): scalar params, imports only `apperror`/`outbound`/`core/process`. **But `FindProcesses` contains `fmt.Printf("Found %d process(es) matching %v. Starting game...\n", …)`** — a raw stdout write in the application layer with game-specific wording (§5.3). |
| `internal/application/runner.go` | `NewRunner(proc, store, reporter, renderer, events, logger)`; `Run` does validate → find → load → play → log kill failures → record → report. |
| `internal/infrastructure/*` | Package names are `filescore`, `stderr`, `stdout`, `osprocess`, `tcellui`, `fsutil` — **not** the `scorefilestore`/`stderrlog` of the design record. `stderr.NewLogger()` is infallible and captures `os.Stderr` (decides §4.6). `osprocess.NewProcess()` fails on `exec.LookPath("ps")`; `filescore.NewFileScore()` fails on home-dir resolution. |
| `internal/testutil/fake/*.go` | Nine fakes: `InputSource`, `Logger`, `Process` (+ `processHandle`), `ProcessKiller` (a func adapter), `Renderer`, `Runner`, `ScoreReporter`, `Store`. Conventions as described in §9.1. **There is no `testutil/capture` package** despite `hexagonal-arc.md` §5 listing one; there is a `testutil/helper` with `UnsetEnv`. |
| `Makefile`, repo root | No CI configuration anywhere (`.github/` absent, no `.golangci.yml`). `make test` = `go test -v -tags integration -count=1 ./...`; `make test-integration` = `-run TestIntegration`. Integration tests live behind `//go:build integration` in four files. This is why §9.5 proposes a make target rather than a CI job. |
| `docs/features.md` | §Speed: 0.5–5.0, default 2.0. §Time limit: default 30, `0` disables. §Startup step 3 quotes the exact `Found N process(es) matching [patterns]. Starting game...` string that §5.3 proposes changing. |
| `docs/hexagonal-arc.md` | The dependency rule is stated **core-first**: §8's "Key property" and §12's "Clear dependency rule" are both about `internal/core`. "Application never imports infrastructure" is enforced by the §8 graph (the only arrows into `infrastructure` start at `main`) but is never written as prose. The rigid, explicitly-worded rule that this proposal actually had to reckon with is §9's "All wiring happens here and nowhere else" — which is what settles §4.1. |
| `internal/infrastructure/filescore/file_score.go`, `defaultPath` (§11.4/§11.7) | **Confirms the path-resolution pattern `fileconfig` reuses.** `defaultPath()` is exactly `fsutil.ConfigDir("pidshooter")` joined with a filename (`highscores.json`); `fileconfig.Load` does the identical call joined with `config.yaml` instead — same directory, sibling file, no new resolution logic invented. |
| `go.mod`, `go mod why gopkg.in/yaml.v3` (§10.10) | **Confirms no new dependency is needed for §11.** `gopkg.in/yaml.v3 v3.0.1` is already listed `// indirect` in `go.mod`, pulled in via `github.com/stretchr/testify/assert/yaml`, and is present in `go.sum`. A real import from `internal/infrastructure/fileconfig` only promotes the existing `require` line to direct on the next `go mod tidy`. |

---

## Appendix B — Amendments required to `hexagonal-arc.md`

Not part of this proposal's implementation, but needed for the design record to describe the codebase again. The first four are pre-existing drift; the rest are consequences of this proposal.

1. **§3, §5, §7, §13 — adapter package names.** `scorefilestore` → `filescore` (file `file_score.go`, constructor `NewFileScore`), `stderrlog` → `stderr`. Add `infrastructure/stdout` (`ScoreReporter`), which the document does not mention at all.
2. **§3, §7, §13 — the process port.** `ProcessManager`/`Lister`/`Killer` no longer exist. The live port is `outbound.Process` with `Discover`/`OwnPID`/`OwnUID`/`LookupName`/`Pin`, plus a separate `ProcessHandle` interface (`Kill`/`Release`) implementing pinned, recycle-safe termination — a design the document predates entirely.
3. **§13 — the `killer` seam.** Documented as `Kill(target *game.Target) (killed, shouldReap bool, err error)`; actually `processKiller` with `Kill(pid int, name string, protected bool) (shouldReap bool, err error)`. The live version is the better one and its reusability is load-bearing for §5.5 above.
4. **§5, §7, §9 — smaller drift.** `testutil/capture` does not exist (`testutil/helper` does); `core/process/process.go` is `info.go`; `NewRunner` now takes six parameters including a reporter and a logger; `main()` does not print before exiting; `application/score.Service` has `ReportResults`, not a package-level `PrintResults`; `apperror` is not mentioned anywhere.
5. **§5 and §8 — composition root.** `cmd/pidshooter` becomes four files. The annotation `# Composition root only — no logic` still holds and should be kept; the §8 graph gains no new node, but `main` gains an arrow to `contract/inbound` (it names the provider return types).
6. **§9 — add the laziness rule.** The invariant introduced here deserves a sentence next to "all wiring happens here and nowhere else": *wiring happens here, and each mode's wiring runs only when that mode is the one requested; the only adapters constructed before the mode is known are the logger and the config file, neither of which is mode-differentiated (§11.5).*
7. **§3, §5 — a new top-level package.** `internal/config` (§11.4) joins `internal/util` as pure Go with no OS access, safe for `core` to import even though nothing there needs it yet. It should be listed next to `util` wherever §3/§5 enumerate what core is allowed to depend on, so a future contributor does not have to rediscover the "pure Go, no OS access" test from `fsutil`'s placement reasoning (§3) to work out why `config` is exempt from the "everything under `internal/` other than `core`/`application`/`infrastructure`/`entrypoint` needs justifying" pattern the rest of the tree follows.
7. **§12 — enforcement.** Replace the `grep -r "tcell" internal/core` line with a reference to `make arch` (§9.5), covering core→outer, application→adapter, and entrypoint→infrastructure.
