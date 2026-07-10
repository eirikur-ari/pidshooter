# pidshooter — Fable AI Code Review

**Reviewed by:** Claude Fable 5  
**Date:** 2026-07-10  
**Scope:** Full codebase — functional correctness, design, security, architecture

---

## 1. Executive Summary

The codebase is small (~1,600 LOC), well-factored, and unusually well-tested for a project of its size, including regression tests for previously fixed goroutine-leak and multibyte-rune bugs. The hexagonal architecture is genuine, not cosmetic: the core game is a near-pure state machine, all I/O flows through ports, and adapters are cleanly swappable (proven by the test fakes and tcell simulation screen).

The most significant remaining problems are:

- (a) A **UI/logic contradiction in the kill-confirmation prompt** — says "(Q)uit" but Q cancels
- (b) A **`--speed=NaN` validation bypass** that breaks the game
- (c) **Kills completed in the final frame are silently lost from the score**
- (d) Lingering **console I/O in the domain and application layers** that undermines the hexagonal boundary

Security posture is strong for what the tool does (name re-verification before SIGKILL, PID ≤ 1 guard, min pattern length), with only residual TOCTOU and root-execution hardening left to consider. No critical findings.

---

## 2. Functional Issues

**F1. [moderate] Confirmation prompt says "(Q)uit" but Q cancels the confirmation instead of quitting.**  
`internal/infrastructure/tcellui/tcellui.go:128` renders `" Kill [%d %s]? (Y)es / (N)o / (Q)uit"`, but `internal/core/game/event_handler.go:26-31` treats `q` during confirmation as cancel (`g.confirming = nil`), not quit. The UI actively misleads the user; a player pressing Q to exit stays in the game. Either the prompt text or the handler must change.

**F2. [moderate] `--speed=NaN` bypasses range validation and breaks the game.**  
`internal/entrypoint/cli/cli.go:74-81`: `strconv.ParseFloat("NaN", 64)` succeeds, and both range checks (`s < 0.1 || s > 5.0`) are false for NaN, so `cfg.Speed = NaN` is accepted. NaN propagates into every `Position` via `target.go:103-104`, making `int(e.Position.X)` undefined and targets unclickable/invisible. Fix: add `math.IsNaN(s)` (or `s != s`) to the validation check. `+Inf`/`-Inf` are already rejected by the range comparison.

**F3. [moderate] Kills completing in the final frame are never recorded in the score.**  
`internal/application/service/game.go:125-134`: the loop exits as soon as `g.Running()` is false. A kill goroutine (spawned at lines 166-173) that succeeds after the final `applyKills` call sends into `s.kills` and gets the `<-done` branch when the channel isn't drained — the OS process **was killed** but the kill/freed-memory never appears in the session stats or saved score. A final drain of `s.kills` (after `signal.Stop`, before returning) would close the gap for buffered entries; fully closing it requires waiting on in-flight kill goroutines (e.g., a `sync.WaitGroup`).

**F4. [minor] Process names containing consecutive spaces can never be killed.**  
`internal/infrastructure/osprocess/osprocess.go:78` canonicalizes names as `strings.Join(strings.Fields(line)[2:], " ")` (collapses whitespace runs), while `currentName` at line 125 returns `strings.TrimSpace(raw)` (preserves interior whitespace). For a comm like `"my  app"`, the name comparison in `validateProcessName` (lines 160-165) always fails, so the kill is silently refused (`return false, nil`) and the target appears unkillable with no feedback.

**F5. [minor] Mouse drag with button held fires repeated ClickEvents — "drag-to-kill".**  
`internal/infrastructure/tcellui/tcellui.go:159-164` forwards every `EventMouse` where `Buttons() == Button1`, including motion events while the button is held. Sweeping the cursor across the screen with the button down kills every target touched with a single sustained "click". If unintended, track button-press transitions (previous button state) and only emit on press.

**F6. [minor] Zero-kill sessions pollute the top-10 board.**  
`internal/application/service/game.go:76-83` unconditionally adds an entry even if the player quit instantly with 0 kills. Until the board fills with 10 real scores, quitting immediately writes junk rows to `~/.config/pidshooter/highscores.json`.

**F7. [minor] Timer displays "Time: 0s" for up to a full second while the game is still running.**  
`internal/core/game/loop.go:51` truncates with `int(g.timeRemaining().Seconds())`; remaining = 0.9s renders as `0s`. Use `math.Ceil` on seconds for a countdown display.

**F8. [minor] `--time=` has no upper bound; extreme values overflow the duration math.**  
`internal/entrypoint/cli/cli.go:82-91` accepts any non-negative int. `loop.go:82` computes `time.Duration(g.timeLimit)*time.Second`; for `--time=9223372036854775807` this overflows int64 nanoseconds to a negative duration, `timeRemaining()` clamps to 0, and the game exits on the first tick. Cap at something sane (e.g., 86400 seconds).

**F9. [minor] `sort.Slice` is not stable — tie-breaking at the top-10 cutoff is arbitrary.**  
`internal/core/score/score.go:78-82` uses unstable `sort.Slice` and `beats` (lines 84-89) only tie-breaks on `FreedMem`. Two entries with equal kills and freed memory sort in unspecified order, so whether the new or old entry survives the `Scores[:maxScores]` cut is nondeterministic. Use `sort.SliceStable` or add `Date` as a final tie-break.

**F10. [minor] `capture.Output`/`capture.Stderr` leave stdout/stderr hijacked if `fn` panics.**  
`internal/testutil/capture/capture.go:16-20 and 32-36` restore the stream after `fn()` without `defer`. A panicking test leaves `os.Stdout`/`os.Stderr` pointing at a closed pipe for all subsequent tests in the package, producing confusing cascading failures. Test-only impact, but a two-line `defer` fix.

---

## 3. Design Issues

**D1. [moderate] Domain layer performs console I/O — `score.Board.PrintScores`/`PrintHighScore`.**  
`internal/core/score/score.go:50-76` — `fmt.Println` box-drawing and trophy output live in `core/`. This is the largest remaining hexagonal violation. The domain should expose data (`Entries()`, `IsNewHighScore(kills) bool`) and rendering should move to the delivery layer or a presenter adapter. Related: the table header prints "Speed" where "Time" is intended (`score.go:63`).

**D2. [moderate] Application service writes directly to stdout/stderr.**  
`internal/application/service/game.go:52, 56, 60, 86, 90-93` — `fmt.Printf`/`fmt.Fprintf(os.Stderr, ...)` in `Play`. The service is otherwise fully port-driven; these prints make it untestable without stream capture (exactly what the integration tests are forced to do via `capture.Stderr`). Introduce an outbound notifier/presenter port, or inject `io.Writer`s.

**D3. [moderate] `GameService.kills` is mutable struct state initialized inside `runLoop` — latent race and leaky testing seam.**  
`internal/application/service/game.go:26 (field), 123 (assignment)`. The channel is per-session state stored on the long-lived service; two concurrent `Play` calls would data-race on the field, and unit tests must reach into the unexported field to exercise `applyKills`. Make it a local in `runLoop` passed as a parameter.

**D4. [minor] `outbound.Process.List()` is on the port but unused by the application.**  
`internal/application/contract/outbound/process.go:9` — the service only calls `Find` and `Kill`; `List` exists for the adapter's internals. Interface segregation: drop `List` from the port.

**D5. [minor] `Game` embeds `Session`, leaking mutators through the aggregate.**  
`internal/core/game/game.go:21` — embedding exposes `g.SetHighScore`, `g.RecordKill` etc. on `Game`'s public surface. External code could call `g.RecordKill` directly, bypassing the `CompleteKill` state check at `loop.go:73-79`. Prefer a named field `session Session` with explicit read accessors.

**D6. [minor] `Board.highScore` has hidden temporal coupling: `PrintHighScore` is only meaningful after `Add`.**  
`internal/core/score/score.go:31-37, 50-54` — `highScore` is captured as a side effect of `Add` (and is zero after JSON `Load`), so `PrintHighScore` silently reports "new high score" for any kills>0 if called before `Add`. Make `Add` return `wasHighScore bool` instead.

**D7. [minor] A "pure state machine" core that contains a concurrency primitive.**  
`internal/core/game/game.go:19` — `running atomic.Bool` exists solely because the application's signal goroutine calls `g.Stop()` cross-thread. Concurrency has leaked into the "pure" core. Cleaner: the signal goroutine sets an application-level flag (or cancels a context) and the loop calls plain `g.Stop()` from the loop goroutine.

**D8. [minor] Composition root builds heavy adapters before argument parsing.**  
`cmd/pidshooter/main.go:24-37` — `osprocess.NewProcess()` (requires `ps` on `$PATH`) and `tcell.NewScreen()` run before the CLI ever parses args. Consequence: `pidshooter --help` fails on a system without `ps`, and allocates a terminal screen just to print usage. Parse args first, construct adapters only when a game will actually run.

**D9. [minor] Pattern validation duplicated in two layers with drift risk.**  
`internal/entrypoint/cli/cli.go:97-102` and `internal/infrastructure/osprocess/osprocess.go:128-141` implement the same min/max length rules. Consider a shared `process.ValidatePattern(p string) error` in the domain package that owns the constants.

**D10. [minor] `UI.Cleanup` panics if called twice.**  
`internal/infrastructure/tcellui/tcellui.go:45-48` — `close(a.done)` on an already-closed channel panics, and the `Renderer` port doesn't document one-shot semantics. Guard with `sync.Once`.

---

## 4. Security Concerns

**S1. [moderate] Residual TOCTOU window between name verification and SIGKILL.**  
`internal/infrastructure/osprocess/osprocess.go:97-111` — `currentName(pid)` (a `ps` subprocess, milliseconds of latency) then `proc.Signal(SIGKILL)`. If the PID dies and is recycled in that window, an unrelated process is killed. The previous fix narrowed the window dramatically but didn't eliminate it — that's inherent to signal-by-PID. On Linux, `pidfd_open` + `pidfd_send_signal` closes it completely; on macOS there is no perfect primitive. This residual risk should be documented in the `Kill` contract (`outbound/process.go:6-7`).

**S2. [moderate] No safeguard against running as root.**  
Nothing in `main.go` or `osprocess.go` checks `os.Geteuid()`. As root, every process on the machine (except PID ≤ 1 and self) becomes a one-click SIGKILL target in a game where mis-clicks are the core mechanic — and F5 (drag-to-kill) amplifies this significantly. Recommend refusing to run as root, or requiring an explicit `--i-am-root` override flag.

**S3. [low] `ps` output parsing trusts process-name content.**  
`internal/infrastructure/osprocess/osprocess.go:55-85` — the parser splits raw `ps` output on lines/fields. A process that names itself with embedded newlines could (on platforms where `ps` doesn't escape control characters) inject a fake row with an arbitrary PID. Exploitability is largely neutralized by the kill-time name re-verification, but the parser should reject rows with implausible field counts.

**S4. [low] SIGKILL only — no graceful termination option.**  
`osprocess.go:108` sends `SIGKILL` unconditionally: no flush, no cleanup handlers, guaranteed data loss for the victim. A `--sigterm` mode (SIGTERM, escalating to SIGKILL after a grace period) would make the tool safer to use as a process manager.

**S5. [low] Score file written non-atomically.**  
`internal/infrastructure/scorefilestore/score_file_store.go:52` — `os.WriteFile` truncates in place; a crash mid-write corrupts `highscores.json` permanently (triggering the load-error/skip-save path on every subsequent run). Write to a temp file in the same directory and `os.Rename` to replace atomically. File permissions (0600/0700) are already correct.

**S6. [low] Unbounded kill-goroutine spawning per input event.**  
`internal/application/service/game.go:166` — every accepted kill request spawns a goroutine that runs a `ps` subprocess. Combined with F5 (drag events), a single mouse sweep can spawn dozens of concurrent `ps` invocations. A small worker pool or per-target in-flight dedup (a `Killing`-pending flag checked in `HandleClick`) would bound resource consumption.

---

## 5. Architecture Observations & Proposals

**Current state.** The hexagon is genuine: `core/` has zero outward dependencies; ports live in `application/contract/{inbound,outbound}`; `entrypoint/cli` is a thin driver adapter; `infrastructure/*` are driven adapters wired only in `cmd/pidshooter/main.go`. The `Game`-as-state-machine + `KillRequest` design (core decides *what* to kill, application performs the side effect, `CompleteKill` confirms) is a textbook-clean way to keep OS calls out of the domain.

**Proposals:**

**P1. Presenter port** — Introduce `outbound.Presenter` (or `Notifier`) covering pre-game messages, warnings, game-over summary, and score table. This resolves D1+D2 in one move, makes `Play` fully deterministic under test, and removes the `testutil/capture` dependency from service tests.

**P2. Select-based game loop** — `runLoop` (`service/game.go:125-134`) currently busy-drains two channels with `default:` each frame. A single `select { case <-ticker.C: ...; case ev := <-events: ...; case t := <-kills: ...; case <-ctx.Done(): }` loop is simpler, lower-latency for input, and naturally handles the final-frame kill drain (F3).

**P3. Context propagation** — There is no `context.Context` anywhere. `signal.NotifyContext` in `main`/`Play` would replace the hand-rolled `sigCh`/`done` plumbing (`service/game.go:107-118`) and give kill goroutines a proper cancellation story.

**P4. Kill executor abstraction** — The goroutine-per-kill logic embedded in `drainEvents` (lines 164-174) is the hardest part of the service to test (no test covers the click→kill→CompleteKill round trip through `drainEvents`). Extract an async `killExecutor` with an injectable completion callback; the fake becomes synchronous and this path becomes unit-testable.

**P5. Process discovery via `/proc` or `gopsutil`** — The `ps` subprocess dependency causes D8 (help needs `ps`), F4 (name canonicalization drift), S3 (parse trust), platform-specific flags (`osprocess.go:44-47`), and the Linux 15-char comm truncation issue. Reading `/proc/<pid>/{comm,statm}` on Linux (keeping `ps` for darwin) removes a whole class of issues.

**P6. Stdlib flag parsing** — `parseArgs` hand-rolls `--flag=value` parsing (`cli.go:61-112`); `--speed 2.5` (space-separated) is silently treated as a pattern. The stdlib `flag` package would fix that, and F2's NaN gap gets a natural home in a custom `flag.Value` type.

---

## 6. Positive Observations

1. **Test discipline is exceptional for the project size**: regression tests pinned to fixed issue numbers (`tcellui_test.go:55, 83, 142`), goroutine-leak tests for both the signal goroutine and the poll goroutine, multibyte-rune bug tests (`target_test.go:178-239`), and a build-tag-separated integration suite with Makefile targets including `test-race`.

2. **Kill-safety layering is thoughtful**: PID ≤ 1 refusal (`osprocess.go:94`), name re-verification before signaling, own-PID and PID-0/1 exclusion in `filter` (`osprocess.go:145-148`), 3-char minimum pattern with a documented rationale (`core/process/process.go:4-8`), and `ps` resolved by absolute path at construction (`osprocess.go:23-29`) to avoid runtime `$PATH` hijacking.

3. **The `Frame`/`TargetView` snapshot pattern** (`core/game/frame.go`) cleanly decouples domain state from rendering — the renderer receives plain data, never domain objects.

4. **Correct channel hygiene**: the `done`-guarded send in kill goroutines (`service/game.go:168-171`) and in `tcellui.poll` (`tcellui.go:173-177`) prevent the goroutine leaks that plagued earlier revisions.

5. **Score file security**: stored with 0600/0700 permissions in a dedicated config dir; corrupt-file handling deliberately skips save to preserve the file for recovery, with a test pinning that behavior.

6. **Living documentation**: `docs/code-review.md` and `docs/security-issues.md` track issues to resolution with fix descriptions appended — a rare and valuable practice.

---

## Summary Table

| ID  | Severity | Category    | Short Description                                      |
|-----|----------|-------------|--------------------------------------------------------|
| F1  | moderate | Functional  | "(Q)uit" in prompt does cancel, not quit               |
| F2  | moderate | Functional  | `--speed=NaN` bypasses validation                      |
| F3  | moderate | Functional  | Final-frame kills lost from score                      |
| F4  | minor    | Functional  | Multi-space process names never killable               |
| F5  | minor    | Functional  | Drag-to-kill via held mouse button                     |
| F6  | minor    | Functional  | Zero-kill sessions pollute scoreboard                  |
| F7  | minor    | Functional  | Timer shows "0s" for sub-second remaining              |
| F8  | minor    | Functional  | `--time=` integer overflow on extreme values           |
| F9  | minor    | Functional  | Unstable sort causes nondeterministic score cutoff     |
| F10 | minor    | Functional  | `capture` helpers don't use `defer` for restore        |
| D1  | moderate | Design      | Domain layer performs console I/O                      |
| D2  | moderate | Design      | Application service writes directly to stdout/stderr   |
| D3  | moderate | Design      | `kills` channel on service struct is a latent race     |
| D4  | minor    | Design      | `List()` on process port violates ISP                  |
| D5  | minor    | Design      | Embedded `Session` leaks mutators on `Game`            |
| D6  | minor    | Design      | Temporal coupling on `Board.highScore`                 |
| D7  | minor    | Design      | Atomic in "pure" core leaks concurrency concerns       |
| D8  | minor    | Design      | Adapters constructed before arg parsing                |
| D9  | minor    | Design      | Pattern validation duplicated across layers            |
| D10 | minor    | Design      | `UI.Cleanup` panics if called twice                    |
| S1  | moderate | Security    | Residual TOCTOU window in kill flow                    |
| S2  | moderate | Security    | No guard against running as root                       |
| S3  | low      | Security    | `ps` output parsing trusts process-name content        |
| S4  | low      | Security    | SIGKILL only — no graceful shutdown option             |
| S5  | low      | Security    | Score file written non-atomically                      |
| S6  | low      | Security    | Unbounded kill-goroutine spawning                      |
