# pidshooter — Package & Architecture Diagram

*Updated 2026-10-13. Full hexagonal architecture: pure-core domain (game/movement/process/score),
a core/game `Input` intent gateway, an `application/input` package owning the input `Event` types
and their `Dispatcher`, contract ports split by concern (inbound/outbound subpackages, including
`ConfigStore` and `Logger` outbound ports alongside `ProcessReporter`/`ScoreReporter` and the
`ProcessManager`/`ProcessHandle` pair), a structured `apperror` package classifying every
application-layer failure and logging it through an injected `Handler`, infrastructure adapters (a
decomposed `tcellui`, a `console` reporter package, a `filestore` package persisting both the score
board and run config, a `ps`-backed `pscommand` process manager, and a dedicated `logger` adapter), a
hand-rolled (no `cobra`) entrypoint adapter, a dedicated `internal/composition` package as the actual
composition root, and an application layer split into four focused services — `config`, `game`,
`process`, `score` — each speaking its own request/result data, composed by an `application/runner`
package that implements the `inbound.Runner` port through consumer-side interfaces and imports no
core package. See `docs/hexagonal-arc.md` for the narrative behind these changes, and
`docs/improvement-finding.md` for what's still open.*

---

## Package layout

```
cmd/pidshooter/
└── main.go                             thin entrypoint — builds composition.RunnerCreator + cli.Program, maps errors to exit codes (cli.ArgumentError → 2, other error → 1), no wiring logic

internal/
├── core/                               pure domain — no infrastructure imports, no outward edges at all
│   ├── game/
│   │   ├── session.go                  lifecycle (enum: pending/running/stopped) · Config · Session · NewSession() · Start() · Stop() · IsRunning() · Update() · AvailableTargets() · PendingConfirm() · RequestConfirm() · StartTime() · Targets() · Throttle() · TimeLeft() · TimeLimit() · currentState() · initialize()
│   │   ├── timer.go                    timer · newTimer(limitSeconds, now func() time.Time) · Start() · Expired() · Remaining() · SecondsLeft() · LimitSeconds() · StartTime() · ValidateTimeLimit() · secondsToDuration()
│   │   ├── confirmation.go             confirmation · newConfirmation() · Pending() · Request() · Accept() · Cancel()
│   │   ├── roster.go                   roster · newRoster() · spawn() · move() · allDead() · isEmpty() · hitAt() · available()
│   │   ├── target.go                   Target (Info process.Info, Motion movement.Motion — named fields, not embedded) · State (enum: Alive/Killing/Fleeing/Dead) · AnimationDuration · NewTarget() · Tag() · AnimationProgress() · Update() · Kill() · Reap() · FireShot() · CeaseFire() · isAlive() · isDead() · isHitAt() · doomsdayTick() · move() · tagFor()
│   │   └── input.go                    Input · NewInput() · OnQuit() · OnYes() · OnNo() · OnSpeedUp() · OnSpeedDown() · OnClickAt()
│   ├── movement/
│   │   ├── bounds.go                   WindowSize · ChromeSize · Bounds · NewBounds() · Update() · bounce() · bounceLeft/Right/Top/Bottom()
│   │   ├── motion.go                   Motion · NewMotion() · Move() · newRandomPosition() · newRandomVelocity()
│   │   ├── vector.go                   Vector · Rounded()
│   │   ├── speed.go                    speed (unexported) · newSpeed() · Current() · Lowest() · Set()
│   │   └── throttle.go                 MinSpeed · MaxSpeed · Throttle · NewThrottle() · Speed() · LowestSpeed() · Increase() · Decrease() · ValidateSpeed()
│   ├── process/
│   │   └── info.go                     Info · NewInfo() · IsProtected() · IsKillableBy() · Find() · ValidateProcesses() · ValidateName() · ValidatePatterns() · ValidateRoot() · validatePatterns() · validatePatternLength()
│   └── score/
│       ├── board.go                    Board · NewBoard() (routes every entry through insert) · Add() · HighScore() · IsNewHighScore() · killScore() · insert() · sortByRank() (sort.SliceStable) · less()
│       └── entry.go                    Entry · beats() · isScore()
├── application/
│   ├── apperror/
│   │   ├── error.go                    Code (enum, In(err) predicate) · Severity (enum) · Error · NewError() · Error() · Unwrap() · severity()
│   │   └── error_handler.go            Handler{logger outbound.Logger} · NewHandler(logger) · Handle() · HandleAll() · logOnce()
│   ├── config/
│   │   ├── result.go                   GameResult · ProcessResult{IncludeRoot, AllowRoot} · Result{Mode outbound.Mode, Process, Game} · newResult() · apply() · fromStore() · fromOptions() · validateStore()
│   │   └── service.go                  GameOptions{ConfirmMode, Speed, TimeLimit *pointers} · ProcessOptions{IncludeRoot, AllowRoot *pointers} · Options{Game, Process} · Service{store outbound.ConfigStore, options} · NewService(store, options) · Load() · applyConfig() · validate()
│   ├── contract/
│   │   ├── inbound/
│   │   │   └── runner.go               Runner (interface, Run() error)
│   │   └── outbound/
│   │       ├── config.go               Mode (ModeGame/ModeLucky/ModeList) · GameConfig · ProcessConfig · Config{Mode, Process, Game} · ConfigStore (interface)
│   │       ├── error.go                NotFoundError · CorruptedDataError
│   │       ├── input.go                InputEventProvider (interface, Events() <-chan input.Event)
│   │       ├── logger.go               Logger (interface)
│   │       ├── process.go              ProcessInfo · ProcessHandle (interface) · ProcessManager (interface) · ProcessReporter (interface)
│   │       ├── score.go                ScoreEntry · ScoreBoard · ScoreStore (interface) · ScoreSummary · ScoreReporter (interface)
│   │       └── ui.go                   FrameViewState · TargetViewState · HUDViewState · StatusViewState · ConfirmViewState · WindowSize · ChromeSize · Renderer (interface)
│   ├── input/
│   │   ├── dispatcher.go               Dispatcher · NewDispatcher() · Dispatch()
│   │   └── event.go                    Event (interface, unexported dispatch() method) · ClickEvent · QuitEvent · ConfirmEvent · SpeedEvent
│   ├── game/
│   │   ├── service.go                  Service · PlayRequest{ConfirmMode, Speed, TimeLimit, Processes, HighScore} · ProcessRequest · PlayResult{…, Duds, Errors []error} · NewService() · Play() · runPlaySession() · validateGameConfig()
│   │   ├── play_session.go             playSession · newPlaySession() · run() · result() · frameLoop() · drainEventQueue() · registerTermSignalWatcher()
│   │   ├── in_flight_kills.go          processKiller (unexported interface) · inFlightKills · newInFlightKills() · start() · killOrReap() · close() · awaitRemaining() · applyFinished()
│   │   ├── kill_result.go              killResult · applyTo() · shouldReap()
│   │   ├── kill_tracker.go             killTracker · killScoreTracker · killFailureTracker · newKillTracker() · recordKill() · recordFailure() · recordDud() · kills() · freedMem() · highScore() · errors()
│   │   ├── kill_failure.go             killFailure · toError()
│   │   ├── kill_dud.go                 killDud · toError()
│   │   └── converter.go                toGameConfig() · toInfos() · toWindowSize() · toBounds() · toConfirmViewState() · toTargetViewState() · toTargetViewStates() · toHUDViewState() · toStatusViewState() · toFrameViewState()
│   ├── process/
│   │   ├── service.go                  Service{manager, reporter, patterns} · FindRequest{IncludeRoot, AllowRoot} · FindResult · NewService(manager, reporter, patterns) · FindProcesses() · Kill()
│   │   └── converter.go                toInfos() · toFindResults() · toProcessInfos()
│   ├── runner/
│   │   ├── service.go                  configLoader · processFinder · scoreKeeper · gamePlayer (unexported consumer-side interfaces) · Service{configSvc, processSvc, gameSvc, scoreSvc, errHandler} · NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler) · Run()   (implements inbound.Runner)
│   │   └── converter.go                toPlayRequest() · toFindRequest() · toRecordRequest() · toReportRequest()
│   └── score/
│       ├── service.go                  Service{store, reporter} · BoardEntry · LoadResult · RecordRequest · RecordResult · ReportRequest · NewService() · LoadScoreBoard() · RecordScore() · ReportResults() · mergeWithLatest()
│       └── converter.go                toEntry() · toEntries() · toBoardEntries() · toBoard() · toLoadResult() · toRecordResult() · toScoreBoard() · toScoreSummary() · toScoreEntries()
├── composition/                        the actual composition root — pure wiring, no behavior
│   └── runner_creator.go               RunnerCreator{errHandler} · NewRunnerCreator() · ErrHandler() · Create(patterns, options)
├── entrypoint/                         driving adapter — calls the application's inbound.Runner port
│   └── cli/
│       ├── program.go                  usageText · RunnerCreator (interface) · Program{creator, errHandler, parser, out, errOut} · NewProgram() · Run() · run() · prepare() · printUsageText() · help()
│       ├── flag_parser.go              argumentSplitter (interface) · flagParser · newFlagParser() · parse()
│       ├── flag_mapper.go              flagMapper · newFlagMapper() · toConfigOptions()
│       ├── flag_splitter.go            flagSplitter · newFlagSplitter() · split() · looksLikeFlag() · isBoolFlag() · flagNameAndValue()
│       └── error.go                    ArgumentError · Error() · Unwrap()
├── infrastructure/                     driven adapters — implement outbound ports
│   ├── console/
│   │   ├── score_reporter.go           ScoreReporter · NewScoreReporter()  (implements outbound.ScoreReporter)
│   │   └── process_reporter.go         ProcessReporter · NewProcessReporter()  (implements outbound.ProcessReporter)
│   ├── filestore/
│   │   ├── config_file.go              configEncoder (interface) · configFile · gameEntry · processEntry · configContent · NewConfigFile() · newConfigFileAt()  (implements outbound.ConfigStore) — schema-versioned YAML
│   │   ├── score_file.go               scoreEncoder (interface) · scoreFile · scoreEntry · scoreContent · NewScoreFile() · newScoreFileAt()  (implements outbound.ScoreStore) — schema-versioned JSON
│   │   ├── file.go                     file · newFile() · read() · write()  (I/O shared by both stores: bounded reads mapped to NotFoundError/CorruptedDataError, atomic writes)
│   │   ├── file_encoder.go             fileEncoder · encodeYAML() · encodeJSON()  (satisfies configEncoder and scoreEncoder)
│   │   ├── schema_version.go           schemaVersion · validate()
│   │   └── converter.go                toScoreBoard() · toScoreContent() · toConfig() · toConfigContent()
│   ├── fsutil/
│   │   └── fsutil.go                   ConfigDir() · WriteFileAtomic() · preserveOwner() · sudoOwner() · existingOwner()  (shared leaf helper, no port)
│   ├── logger/
│   │   └── logger.go                   Logger · NewLogger() · Warn() · Error()  (implements outbound.Logger, writes to os.Stderr)
│   ├── pscommand/
│   │   ├── manager.go                  host (interface) · manager · NewManager() · Discover() · OwnPID() · OwnUID() · LookupName() · Pin() · isZombie() · processColumnNames() · parseProcesses() · isProcessHeader()  (implements outbound.ProcessManager, `ps`-backed)
│   │   ├── handle.go                   osProcess (interface) · handle · Kill() · Release() · verify()  (implements outbound.ProcessHandle — pins a pid across verify-then-kill)
│   │   └── runner.go                   runner · run() · find() · pid() · uid() · goos() · mapError()  (real host: runs `ps`, os.FindProcess, own pid/uid)
│   └── tcellui/                        (implements outbound.Renderer via *TUI, outbound.InputEventProvider via unexported *poller)
│       ├── tui.go                      TUI · NewTUI() · Init() · Cleanup() · WindowSize() · ChromeSize() · Render() · InputEvents() · cleanup()
│       ├── poller.go                   poller · newPoller() · Events() · poll() · stop()  — TUI.InputEvents() returns TUI's own poller
│       ├── translator.go               runeBindings · controlKeyBindings · translator · newTranslator() · translate() · translateMouseEvent() · translateKeyEvent()
│       ├── renderer.go                 renderer · newRenderer() · render() · beginFrame() · drawFrame() · endFrame()
│       ├── hud.go                      bounds · hud · draw() and its text-budgeting helpers
│       ├── statusbar.go                statusBar · draw() · clear() · left() · text() · confirmPromptText() · playStatusText() · truncate()
│       ├── target.go                   target (drawing helper, distinct from core/game.Target) · draw() · isChromeRow() · styleAndTag()
│       └── animation.go                killAnimation · fleeAnimation · animation · frame()
├── testutil/                           one flat package of shared test helpers (single-consumer fakes/fixtures live in their own package's _test.go files)
│   ├── fake_logger.go                  FakeLogger  (test double for outbound.Logger)
│   └── helper.go                       IsClosed() · UnsetEnv() · ExitedProcessPID() · WarmUpSignalPackage()
└── util/
    └── util.go                          FormatBytes() · ClonePointer()
```

---

## Class diagram

Types, their fields/methods, and relationships across packages. Visibility follows Go conventions: `+` = exported, `-` = unexported. Where a Go identifier collides with another package's identifier of the same name, the diagram node uses a disambiguated name instead — real Go names and package/file provenance are given in the prose and the `%%` section comments, not in the diagram itself, since Mermaid's `<<...>>` class annotation only accepts a single bare keyword (`interface`, `enumeration`, …), not free text. Five distinct `Service` types are disambiguated as `ConfigService`, `GameService`, `ProcessService`, `ScoreService`, and `RunnerService` (the last is `application/runner.Service`, which implements `inbound.Runner`). `application/config.Result.Mode` is `contract/outbound.Mode` reused directly — see the note near `Result` below for why that one field doesn't follow the pointer-vs-value split the rest of `Result`/`Config` does. `core/game`'s session-level `Config` type is disambiguated as `SessionConfig`, and `contract/outbound`'s persisted-config `Config` DTO as `ConfigDTO`. `entrypoint/cli`'s `RunnerCreator` interface is disambiguated as `CliRunnerCreator`, to avoid colliding with `composition.RunnerCreator`. The `infrastructure/pscommand` types `manager`, `handle` and `runner` are shown as `PsManager`, `PsHandle` and `PsRunner`. Go's empty `struct{}` is written as bare `struct` in member lines, since a `}` there would close the Mermaid class block, and `<-chan T` as `chan T`. Purely internal rendering-detail helpers inside `infrastructure/tcellui` (`hud`, `statusBar`, `target`, `killAnimation`/`fleeAnimation`) are listed in the package layout above but not diagrammed field-by-field here — each owns drawing for one screen region and doesn't affect the port-facing architecture.

```mermaid
classDiagram
    direction TB

    %% ── application/config ───────────────────────────────────────────────────

    class GameOptions {
        +ConfirmMode *bool
        +Speed *float64
        +TimeLimit *int
    }

    class ProcessOptions {
        +IncludeRoot *bool
        +AllowRoot *bool
    }

    class Options {
        +Game GameOptions
        +Process ProcessOptions
    }

    class GameResult {
        +ConfirmMode bool
        +Speed float64
        +TimeLimit int
    }

    class ProcessResult {
        +IncludeRoot bool
        +AllowRoot bool
    }

    class Result {
        +Mode Mode
        +Process ProcessResult
        +Game GameResult
        -apply(stored, options) Result, error
        -fromStore(stored) Result, error
        -fromOptions(options) Result
    }

    %% Result.Mode is contract/outbound.Mode reused directly (not a separate
    %% config-local type) — Mode is a plain string on both sides, unlike
    %% GameResult/ProcessResult (plain-value) vs GameConfig/ProcessConfig
    %% (pointer-typed for the YAML round-trip), which do need to stay distinct.

    class ConfigService {
        -store outbound.ConfigStore
        -options Options
        +NewService(store, options) *Service
        +Load() Result, error
        -applyConfig(storedConfig) Result, error
        -validate() error
    }

    %% package-level: newResult() Result · validateStore(stored) rejected []string, error

    Options *-- GameOptions    : Game
    Options *-- ProcessOptions : Process
    Result *-- GameResult : Game
    Result *-- ProcessResult : Process
    Result --> Mode : Mode
    ConfigService --> ConfigStorePort : store
    ConfigService *-- Options : options
    ConfigService ..> Result  : Load returns

    %% ── application/apperror ─────────────────────────────────────────────────

    class Code {
        <<enumeration>>
        CodeUnknown
        CodeInvalidConfig
        CodeProcessNotFound
        CodeGameFailed
        CodeKillFailed
        CodeStoreFailed
        +In(err error) bool
    }

    class Severity {
        <<enumeration>>
        SeverityUnknown
        SeverityFatal
        SeverityError
        SeverityWarning
    }

    class AppError {
        +Code Code
        +Severity Severity
        -message string
        -cause error
        -logged bool
        +NewError(code, severity, message, cause) *Error
        +Error() string
        +Unwrap() error
        -severity() Severity
    }

    class Handler {
        -logger outbound.Logger
        +NewHandler(logger) *Handler
        +Handle(err error) error
        +HandleAll(errs []error) error
        -logOnce(err, severity, appErr)
    }

    AppError --> Code     : Code
    AppError --> Severity : Severity
    Handler  --> LoggerPort : logger
    Handler  ..> AppError   : inspects via errors.As

    %% ── application/contract/inbound ─────────────────────────────────────────

    class RunnerPort {
        <<interface>>
        +Run() error
    }

    %% ── application/contract/outbound ────────────────────────────────────────

    class Mode {
        <<enumeration>>
        ModeGame
        ModeLucky
        ModeList
    }

    class GameConfig {
        +ConfirmMode *bool
        +Speed *float64
        +TimeLimit *int
    }

    class ProcessConfig {
        +IncludeRoot *bool
    }

    class ConfigDTO {
        +Mode Mode
        +Process ProcessConfig
        +Game GameConfig
    }

    class ConfigStorePort {
        <<interface>>
        +Load() Config, error
        +Save(config Config) error
    }

    class LoggerPort {
        <<interface>>
        +Warn(message string)
        +Error(message string)
    }

    class NotFoundError {
        +Error() string
    }

    class CorruptedDataError {
        +Message string
        +Error() string
    }

    class ProcessInfoDTO {
        +UID int
        +PID int
        +Rss int64
        +State string
        +Name string
    }

    class ProcessHandlePort {
        <<interface>>
        +Kill() error
        +Release() error
    }

    class ProcessManagerPort {
        <<interface>>
        +Discover() []ProcessInfo, error
        +OwnPID() int
        +OwnUID() int
        +LookupName(pid int) string, error
        +Pin(pid int) ProcessHandle, error
    }

    class ProcessReporterPort {
        <<interface>>
        +Report(matches []ProcessInfo, patterns []string)
    }

    class ScoreEntryDTO {
        +Kills int
        +Duds int
        +FreedMem int64
        +Speed float64
        +Time int
        +Duration float64
        +Date time.Time
    }

    class ScoreBoardDTO {
        +Scores []ScoreEntry
    }

    class ScoreStorePort {
        <<interface>>
        +Load() ScoreBoard, error
        +Save(board ScoreBoard) error
    }

    class ScoreSummaryDTO {
        +Kills int
        +Duds int
        +FreedMem int64
        +Duration float64
        +IsTopScore bool
        +Entries []ScoreEntry
    }

    class ScoreReporterPort {
        <<interface>>
        +Report(summary ScoreSummary)
    }

    class FrameViewState {
        +Targets []TargetViewState
        +HUD HUDViewState
        +StatusBar StatusViewState
    }

    class TargetViewState {
        +X, Y int
        +Tag string
        +Killing bool
        +Fleeing bool
        +AnimationProgress float64
    }

    class HUDViewState {
        +FreedMem int64
        +Kills int
        +HighScore int
    }

    class StatusViewState {
        +Alive int
        +Speed float64
        +TimeLeft *int
        +Confirming *ConfirmViewState
    }

    class ConfirmViewState {
        +PID int
        +Name string
    }

    class RendererPort {
        <<interface>>
        +Init() error
        +Cleanup()
        +WindowSize() WindowSize
        +ChromeSize() ChromeSize
        +Render(state FrameViewState)
    }

    class InputEventProviderPort {
        <<interface>>
        +Events() chan Event
    }

    ConfigDTO *-- GameConfig      : Game
    ConfigDTO *-- ProcessConfig   : Process
    ConfigDTO --> Mode            : Mode
    ConfigStorePort   ..> ConfigDTO : Load/Save
    ProcessManagerPort ..> ProcessHandlePort : Pin returns
    ProcessManagerPort ..> ProcessInfoDTO    : Discover returns
    InputEventProviderPort ..> Event         : channel element
    FrameViewState *-- "0..*" TargetViewState : Targets
    FrameViewState *-- HUDViewState           : HUD
    FrameViewState *-- StatusViewState        : StatusBar
    StatusViewState *-- "0..1" ConfirmViewState : Confirming
    ScoreBoardDTO *-- "0..*" ScoreEntryDTO   : Scores
    ScoreSummaryDTO *-- "0..*" ScoreEntryDTO : Entries

    %% ── core/movement ─────────────────────────────────────────────────────────

    class Vector {
        +X, Y float64
        +Rounded() x, y int
    }

    class Motion {
        +Position Vector
        +Velocity Vector
        +NewMotion(bounds, width) Motion
        +Move(bounds, speed, width)
    }

    class WindowSize {
        +Width, Height int
    }

    class ChromeSize {
        +Top, Bottom int
    }

    class Bounds {
        -window WindowSize
        -chrome ChromeSize
        +NewBounds(window, chrome) Bounds
        +Update(window) Bounds
        -bounce(position, velocity *Vector, width int)
    }

    class speed {
        -current, lowest float64
        +newSpeed(initial) speed
        +Current() float64
        +Lowest() float64
        +Set(value float64)
    }

    class Throttle {
        -speed speed
        +NewThrottle(initial) *Throttle
        +Speed() float64
        +LowestSpeed() float64
        +Increase()
        +Decrease()
    }

    %% package-level: MinSpeed, MaxSpeed (consts) · ValidateSpeed(value float64) error ·
    %% motion.go: newRandomPosition(bounds, width) Vector · newRandomVelocity() Vector

    Motion *-- "2" Vector : Position, Velocity
    Throttle *-- speed    : speed
    Bounds *-- WindowSize : window
    Bounds *-- ChromeSize : chrome
    Motion ..> Bounds     : bounces via

    %% ── core/process ─────────────────────────────────────────────────────────

    class Info {
        +PID int
        +Name string
        +Rss int64
        +UID int
        +NewInfo(pid, name, rss, uid) Info
        +IsProtected() bool
        +IsKillableBy(ownUID int, includeRoot bool) bool
    }

    %% package-level: Find(processes, patterns, ownPID, ownUID, includeRoot) []Info ·
    %% ValidateProcesses(processes) error · ValidateName(expected, actual) error ·
    %% ValidatePatterns(patterns) error · ValidateRoot(ownUID int, allowRoot bool) error ·
    %% minPatternLength = 3 · maxPatternLength = 256

    %% ── core/score ───────────────────────────────────────────────────────────

    class Board {
        +Scores []Entry
        -highScore int
        +NewBoard(entries) *Board
        +Add(entry Entry)
        +HighScore() int
        +IsNewHighScore(kills int) bool
        -killScore() int
        -insert(entry Entry) bool
        -sortByRank()
        -less(i, j int) bool
    }

    class Entry {
        +Kills int
        +Duds int
        +FreedMem int64
        +Speed float64
        +Time int
        +Duration float64
        +Date time.Time
        -beats(other Entry) bool
        -isScore() bool
    }

    Board "1" *-- "0..*" Entry : Scores

    %% ── core/game — lifecycle & session ──────────────────────────────────────

    class Lifecycle {
        <<enumeration>>
        pending
        running
        stopped
    }

    class SessionConfig {
        +Confirm bool
        +Speed float64
        +TimeLimit int
    }

    class Session {
        -cfg SessionConfig
        -state Lifecycle
        -timer timer
        -confirm confirmation
        -throttle *movement.Throttle
        -roster roster
        -bounds movement.Bounds
        +NewSession(processes, cfg) *Session
        +Start(bounds Bounds)
        +Stop()
        +IsRunning() bool
        +Update(window WindowSize)
        +AvailableTargets() []*Target, int
        +PendingConfirm() *Target
        +RequestConfirm(t *Target) *Target
        +StartTime() time.Time
        +Targets() []*Target
        +Throttle() *Throttle
        +TimeLeft() int
        +TimeLimit() int
    }

    class timer {
        -limit time.Duration
        -start time.Time
        -now func() time.Time
        +newTimer(limitSeconds, now) timer
        +Start()
        +Expired() bool
        +Remaining() time.Duration
        +SecondsLeft() int
        +LimitSeconds() int
        +StartTime() time.Time
    }

    class confirmation {
        -target *Target
        -confirm bool
        +newConfirmation(confirm) confirmation
        +Pending() bool
        +Request(target *Target) *Target
        +Accept() *Target
        +Cancel()
    }

    class roster {
        -processes []process.Info
        -targets []*Target
        +newRoster(processes) roster
        +spawn(bounds)
        +move(bounds, speed)
        +allDead() bool
        +isEmpty() bool
        +hitAt(x, y) *Target
        +available() []*Target, int
    }

    Session *-- timer          : timer
    Session *-- confirmation   : confirm
    Session *-- roster         : roster
    Session --> Throttle       : throttle
    Session --> Lifecycle      : state
    roster "1" *-- "0..*" Target : targets

    %% ── core/game — target & input ───────────────────────────────────────────

    class Target {
        +Info process.Info
        +Motion movement.Motion
        +State State
        +AnimationTick int
        -shotFired bool
        +NewTarget(info, bounds) *Target
        +Tag() string
        +AnimationProgress() float64
        +Update(bounds, speed float64)
        +Kill() bool
        +Reap() bool
        +FireShot()
        +CeaseFire()
        -isAlive() bool
        -isDead() bool
        -isHitAt(x, y int) bool
        -doomsdayTick()
        -move(bounds, speed float64)
    }

    class State {
        <<enumeration>>
        Alive
        Killing
        Fleeing
        Dead
    }

    class Input {
        -s *Session
        +NewInput(s *Session) *Input
        +OnQuit()
        +OnYes() *Target
        +OnNo()
        +OnSpeedUp()
        +OnSpeedDown()
        +OnClickAt(x, y int) *Target
    }

    Target --> Info    : Info (named field — not embedded, so Info's methods/fields aren't promoted)
    Target --> Motion  : Motion (named field — not embedded, so Motion.Move isn't promoted)
    Target --> State   : State
    confirmation --> Target : target
    Input --> Session  : s
    Input ..> Target   : returns

    %% ── application/input ────────────────────────────────────────────────────

    class Event {
        <<interface>>
        -dispatch(input *game.Input) *Target
    }

    class ClickEvent {
        +X, Y int
    }

    class QuitEvent {
    }

    class ConfirmEvent {
        +Accept bool
    }

    class SpeedEvent {
        +Faster bool
    }

    class Dispatcher {
        -input *Input
        +NewDispatcher(input) *Dispatcher
        +Dispatch(inputEvent Event) *Target
    }

    ClickEvent   ..|> Event
    QuitEvent    ..|> Event
    ConfirmEvent ..|> Event
    SpeedEvent   ..|> Event
    Dispatcher --> Input  : input
    Dispatcher ..> Event  : consumes
    Dispatcher ..> Target : returns

    %% ── application/game ─────────────────────────────────────────────────────

    class processKiller {
        <<interface>>
        +Kill(pid int, name string, protected bool) error
    }

    class PlayRequest {
        +ConfirmMode bool
        +Speed float64
        +TimeLimit int
        +Processes []ProcessRequest
        +HighScore int
    }

    class ProcessRequest {
        +PID int
        +Name string
        +Rss int64
        +UID int
    }

    class PlayResult {
        +Duration float64
        +LowestSpeed float64
        +Kills int
        +FreedMem int64
        +Duds int
        +Errors []error
    }

    class killFailure {
        +Name string
        +PID int
        +Err error
        -toError() *apperror.Error
    }

    class killDud {
        +Name string
        +PID int
        -toError() *apperror.Error
    }

    class killResult {
        -target *Target
        -err error
        -applyTo(tracker *killTracker)
        -shouldReap() bool
    }

    class killTracker {
        -score killScoreTracker
        -failure killFailureTracker
        -duds []killDud
        +newKillTracker(highScore) *killTracker
        -recordKill(freedMemory int64)
        -recordFailure(target, err)
        -recordDud(target)
        -kills() int
        -freedMem() int64
        -highScore() int
        -errors() []error
    }

    class killScoreTracker {
        -kills int
        -freedMem int64
        -highScore int
    }

    class killFailureTracker {
        -failures []killFailure
        -pids map[int]struct
    }

    class inFlightKills {
        -killer processKiller
        -gracePeriod time.Duration
        -tracker *killTracker
        -results chan killResult
        -waitGroup sync.WaitGroup
        -done chan struct
        +newInFlightKills(killer, gracePeriod, tracker) *inFlightKills
        -start(target *Target)
        -killOrReap(target *Target)
        -close()
        -awaitRemaining()
        -applyFinished()
    }

    class playSession {
        -renderer outbound.Renderer
        -events outbound.InputEventProvider
        -session *game.Session
        -dispatcher *input.Dispatcher
        -tracker *killTracker
        -kills *inFlightKills
        +newPlaySession(renderer, events, killer, killGracePeriod, session, tracker) *playSession, error
        -run() time.Time, error
        -result(endTime) PlayResult
        -frameLoop(termSignal) time.Time, error
        -drainEventQueue() error
    }

    class GameService {
        -killer processKiller
        -renderer outbound.Renderer
        -events outbound.InputEventProvider
        -killGracePeriod time.Duration
        +NewService(killer, renderer, events) *Service
        +Play(request PlayRequest) PlayResult, error
        -runPlaySession(session, tracker) PlayResult, error
    }

    %% package-level service.go: validateGameConfig(req PlayRequest) error ·
    %% play_session.go: registerTermSignalWatcher() termSignal, stopWatching ·
    %% converter.go: toGameConfig(req PlayRequest) SessionConfig · toInfos() · toWindowSize() · toBounds() ·
    %% toConfirmViewState() · toTargetViewState() · toTargetViewStates() · toHUDViewState() ·
    %% toStatusViewState() · toFrameViewState()

    GameService --> processKiller          : killer
    GameService --> RendererPort           : renderer
    GameService --> InputEventProviderPort : events
    GameService ..> killTracker            : creates per Play()
    GameService ..> playSession            : creates per Play()
    GameService ..> Session                : creates (NewSession)
    GameService ..> SessionConfig          : builds (toGameConfig)
    GameService ..> Info                   : builds (toInfos)
    GameService ..> PlayRequest            : accepts (Play request)
    GameService ..> PlayResult             : returns
    PlayRequest *-- "0..*" ProcessRequest  : Processes
    playSession --> RendererPort           : renderer
    playSession --> InputEventProviderPort : events
    playSession --> Session                : session
    playSession --> Dispatcher             : dispatcher (NewDispatcher(NewInput(session)))
    playSession --> killTracker            : tracker
    playSession *-- inFlightKills          : kills
    playSession ..> FrameViewState         : builds (toFrameViewState)
    inFlightKills --> processKiller        : killer
    inFlightKills --> killTracker          : tracker
    inFlightKills ..> killResult           : queues
    killResult ..> killTracker             : applyTo
    killTracker *-- killScoreTracker       : score
    killTracker *-- killFailureTracker     : failure
    killTracker *-- "0..*" killDud         : duds
    killFailureTracker *-- "0..*" killFailure : failures
    killFailure ..> AppError               : toError (CodeKillFailed)
    killDud ..> AppError                   : toError (CodeProcessNotFound)

    %% ── application/process ──────────────────────────────────────────────────

    class FindRequest {
        +IncludeRoot bool
        +AllowRoot bool
    }

    class FindResult {
        +PID int
        +Name string
        +Rss int64
        +UID int
    }

    class ProcessService {
        -manager outbound.ProcessManager
        -reporter outbound.ProcessReporter
        -patterns []string
        +NewService(manager, reporter, patterns) *Service
        +FindProcesses(request FindRequest) []FindResult, error
        +Kill(pid int, name string, protected bool) error
    }

    %% package-level converter.go: toInfos() · toFindResults() · toProcessInfos()

    ProcessService --> ProcessManagerPort  : manager
    ProcessService --> ProcessReporterPort : reporter
    ProcessService ..> FindRequest         : accepts (FindProcesses)
    ProcessService ..> FindResult          : returns (FindProcesses)
    ProcessService ..> Info                : matches via core/process.Find
    ProcessService ..|> processKiller      : satisfies (structural typing, no import)

    %% ── application/score ────────────────────────────────────────────────────

    class BoardEntry {
        +Kills int
        +Duds int
        +FreedMem int64
        +Speed float64
        +Time int
        +Duration float64
        +Date time.Time
    }

    class LoadResult {
        +Entries []BoardEntry
        +HighScore int
    }

    class RecordRequest {
        +Entries []BoardEntry
        +Kills int
        +Duds int
        +FreedMem int64
        +LowestSpeed float64
        +TimeLimit int
        +Duration float64
    }

    class RecordResult {
        +Entries []BoardEntry
        +NewHighScore bool
    }

    class ReportRequest {
        +Duration float64
        +Kills int
        +Duds int
        +FreedMem int64
        +Entries []BoardEntry
        +NewHighScore bool
    }

    class ScoreService {
        -store outbound.ScoreStore
        -reporter outbound.ScoreReporter
        +NewService(store, reporter) *Service
        +LoadScoreBoard() LoadResult, error
        +RecordScore(request RecordRequest) RecordResult, error
        +ReportResults(request ReportRequest)
        -mergeWithLatest(entry Entry) *Board, error
    }

    %% package-level converter.go: toEntry(request RecordRequest) Entry · toEntries() · toBoardEntries() ·
    %% toBoard() · toLoadResult() · toRecordResult() · toScoreBoard() · toScoreSummary() · toScoreEntries()

    LoadResult    *-- "0..*" BoardEntry : Entries
    RecordRequest *-- "0..*" BoardEntry : Entries
    RecordResult  *-- "0..*" BoardEntry : Entries
    ReportRequest *-- "0..*" BoardEntry : Entries
    ScoreService --> ScoreStorePort    : store
    ScoreService --> ScoreReporterPort : reporter
    ScoreService ..> LoadResult        : returns (LoadScoreBoard)
    ScoreService ..> RecordRequest     : accepts (RecordScore)
    ScoreService ..> RecordResult      : returns (RecordScore)
    ScoreService ..> ReportRequest     : accepts (ReportResults)
    ScoreService ..> Board             : ranks via

    %% ── application/runner ───────────────────────────────────────────────────

    class configLoader {
        <<interface>>
        +Load() Result, error
    }

    class processFinder {
        <<interface>>
        +FindProcesses(request FindRequest) []FindResult, error
    }

    class scoreKeeper {
        <<interface>>
        +LoadScoreBoard() LoadResult, error
        +RecordScore(request RecordRequest) RecordResult, error
        +ReportResults(request ReportRequest)
    }

    class gamePlayer {
        <<interface>>
        +Play(request PlayRequest) PlayResult, error
    }

    class RunnerService {
        -configSvc configLoader
        -processSvc processFinder
        -gameSvc gamePlayer
        -scoreSvc scoreKeeper
        -errHandler *Handler
        +NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler) *Service
        +Run() error
    }

    %% package-level converter.go: toPlayRequest(cfg GameResult, found []FindResult, highScore int) PlayRequest ·
    %% toFindRequest(cfg ProcessResult) FindRequest ·
    %% toRecordRequest(loaded LoadResult, result PlayResult, timeLimit int) RecordRequest ·
    %% toReportRequest(result PlayResult, recorded RecordResult) ReportRequest

    RunnerService ..|> RunnerPort      : implements (structural typing, no import)
    RunnerService --> configLoader     : configSvc
    RunnerService --> processFinder    : processSvc
    RunnerService --> gamePlayer       : gameSvc
    RunnerService --> scoreKeeper      : scoreSvc
    RunnerService --> Handler          : errHandler
    ConfigService  ..|> configLoader   : satisfies
    ProcessService ..|> processFinder  : satisfies
    ScoreService   ..|> scoreKeeper    : satisfies
    GameService    ..|> gamePlayer     : satisfies

    %% ── infrastructure/tcellui ───────────────────────────────────────────────

    class TUI {
        -screen tcell.Screen
        -chrome outbound.ChromeSize
        -poller poller
        -cleanupOnce sync.Once
        -initialized bool
        +NewTUI(screen) *TUI
        +Init() error
        +Cleanup()
        +WindowSize() WindowSize
        +ChromeSize() ChromeSize
        +Render(state FrameViewState)
        +InputEvents() InputEventProvider
        -cleanup()
    }

    class poller {
        -screen tcell.Screen
        -eventQueue chan Event
        -translator translator
        -done chan struct
        +newPoller(screen) poller
        +Events() chan Event
        -poll()
        -stop()
    }

    class translator {
        -screen tcell.Screen
        -prevButtons tcell.ButtonMask
        +newTranslator(screen) translator
        -translate(rawEvent) Event, bool
        -translateMouseEvent(mouseEvent) Event, bool
        -translateKeyEvent(keyEvent) Event, bool
    }

    TUI ..|> RendererPort                : implements
    poller ..|> InputEventProviderPort   : implements
    TUI *-- poller                       : poller (InputEvents() returns it, so Init/Cleanup on TUI still control it)
    poller *-- translator                : translator
    translator ..> Event                 : produces

    %% ── infrastructure/pscommand ─────────────────────────────────────────────

    class host {
        <<interface>>
        -run(args) []byte, error
        -find(pid int) osProcess, error
        -pid() int
        -uid() int
        -goos() string
    }

    class osProcess {
        <<interface>>
        +Signal(signal os.Signal) error
        +Release() error
    }

    class PsManager {
        -host host
        +NewManager() ProcessManager, error
        +Discover() []ProcessInfo, error
        +OwnPID() int
        +OwnUID() int
        +LookupName(pid int) string, error
        +Pin(pid int) ProcessHandle, error
    }

    class PsHandle {
        -process osProcess
        +Kill() error
        +Release() error
        -verify() error
    }

    class PsRunner {
        -path string
        -timeout time.Duration
        -run(args) []byte, error
        -find(pid int) osProcess, error
        -pid() int
        -uid() int
        -goos() string
    }

    %% package-level manager.go: isZombie() · processColumnNames() · parseProcesses() · isProcessHeader() ·
    %% runner.go: mapError(exitCode, stderr, cause) error

    PsManager ..|> ProcessManagerPort  : implements
    PsManager --> host                 : host
    PsManager ..> PsHandle             : Pin creates
    PsHandle ..|> ProcessHandlePort    : implements
    PsHandle --> osProcess             : process
    PsRunner ..|> host                 : implements
    PsRunner ..> osProcess             : find returns (*os.Process)

    %% ── infrastructure/filestore ─────────────────────────────────────────────

    class SharedFile {
        -path string
        -maxSize int
        +newFile(filename, maxSize) file, error
        -read() []byte, error
        -write(data []byte) error
    }

    class fileEncoder {
        -encodeYAML(value any) []byte, error
        -encodeJSON(value any) []byte, error
    }

    class ScoreFile {
        -file SharedFile
        -encoder scoreEncoder
        +NewScoreFile() ScoreStore, error
        +Load() ScoreBoard, error
        +Save(sb ScoreBoard) error
    }

    class ConfigFile {
        -file SharedFile
        -encoder configEncoder
        +NewConfigFile() ConfigStore, error
        +Load() Config, error
        +Save(config Config) error
    }

    %% package-level: schemaVersion(v).validate(fileType, version) error (schema_version.go) ·
    %% converter.go: toScoreBoard() · toScoreContent() · toConfig() · toConfigContent()

    ScoreFile  *-- SharedFile          : file
    ConfigFile *-- SharedFile          : file
    ScoreFile  --> fileEncoder         : encoder (as scoreEncoder)
    ConfigFile --> fileEncoder         : encoder (as configEncoder)
    ScoreFile  ..|> ScoreStorePort     : implements
    ConfigFile ..|> ConfigStorePort    : implements

    %% ── infrastructure/console ───────────────────────────────────────────────

    class ConsoleScoreReporter {
        -writer io.Writer
        +NewScoreReporter() *ScoreReporter
        +Report(summary ScoreSummary)
    }

    class ConsoleProcessReporter {
        -writer io.Writer
        +NewProcessReporter() *ProcessReporter
        +Report(matches []ProcessInfo, patterns []string)
    }

    ConsoleScoreReporter   ..|> ScoreReporterPort   : implements
    ConsoleProcessReporter ..|> ProcessReporterPort : implements

    %% ── infrastructure/logger ────────────────────────────────────────────────

    class InfraLogger {
        -writer io.Writer
        +NewLogger() *Logger
        +Warn(message string)
        +Error(message string)
    }

    InfraLogger ..|> LoggerPort : implements

    %% ── internal/composition ─────────────────────────────────────────────────

    class RunnerCreator {
        -errHandler *apperror.Handler
        +NewRunnerCreator() RunnerCreator
        +ErrHandler() *apperror.Handler
        +Create(patterns []string, options Options) Runner, error
    }

    RunnerCreator ..> PsManager              : builds
    RunnerCreator ..> ScoreFile              : builds
    RunnerCreator ..> ConfigFile             : builds
    RunnerCreator ..> TUI                    : builds
    RunnerCreator ..> ConsoleProcessReporter : builds
    RunnerCreator ..> ConsoleScoreReporter   : builds
    RunnerCreator ..> InfraLogger            : builds (via NewHandler)
    RunnerCreator ..> Handler                : builds (NewHandler)
    RunnerCreator ..> ConfigService          : builds (with options)
    RunnerCreator ..> ProcessService         : builds (with patterns)
    RunnerCreator ..> ScoreService           : builds
    RunnerCreator ..> GameService            : builds
    RunnerCreator ..> RunnerService          : builds & returns

    %% ── entrypoint/cli ───────────────────────────────────────────────────────

    class Program {
        -creator CliRunnerCreator
        -errHandler *apperror.Handler
        -parser *flagParser
        -out, errOut io.Writer
        +NewProgram(creator) *Program
        +Run(args []string) error
        -run(args) showUsage bool, err error
        -prepare(args) patterns, options Options, showUsage bool, err error
        -printUsageText(printUsage, err)
    }

    class CliRunnerCreator {
        <<interface>>
        +Create(patterns []string, options Options) Runner, error
        +ErrHandler() *apperror.Handler
    }

    class flagParser {
        -flagSet *flag.FlagSet
        -mapper *flagMapper
        -splitter argumentSplitter
        +newFlagParser() *flagParser
        -parse(args) patterns []string, options Options, err error
    }

    class argumentSplitter {
        <<interface>>
        -split(args) patterns, flagArgs []string, err error
    }

    class flagMapper {
        -flagSet *flag.FlagSet
        -confirm *bool
        -speed *float64
        -timeLimit *int
        -includeRoot *bool
        -allowRoot *bool
        +newFlagMapper(flagSet) *flagMapper
        -toConfigOptions() Options
    }

    class flagSplitter {
        -flagSet *flag.FlagSet
        +newFlagSplitter(flagSet) flagSplitter
        -split(args) patterns, flagArgs []string, err error
    }

    class ArgumentError {
        +Cause error
        +Error() string
        +Unwrap() error
    }

    Program --> CliRunnerCreator : creator
    Program --> Handler          : errHandler
    Program *-- flagParser       : parser
    Program ..> ArgumentError    : wraps CodeInvalidConfig run errors
    flagParser *-- flagMapper    : mapper
    flagParser --> argumentSplitter : splitter
    flagSplitter ..|> argumentSplitter : implements
    flagParser ..> ArgumentError : returns on bad args
    flagMapper ..> Options       : returns (toConfigOptions)
    CliRunnerCreator <|.. RunnerCreator : satisfied by (structural typing)

    %% ── internal/util ────────────────────────────────────────────────────────

    %% package-level: FormatBytes(bytes int64) string · ClonePointer[T](pointer *T) *T  (util.go — no exported types, generic helpers only)
```

---

## Package dependency graph

```
cmd/pidshooter
    ├─ internal/composition                   (RunnerCreator, NewRunnerCreator)
    └─ entrypoint/cli                         (Program, NewProgram, ArgumentError)

internal/composition
    ├─ application/apperror                   (Handler, NewHandler)
    ├─ application/config                     (Options, NewService)
    ├─ application/contract/inbound           (Runner interface)
    ├─ application/game                       (NewService)
    ├─ application/process                    (NewService)
    ├─ application/runner                     (NewService)
    ├─ application/score                      (NewService)
    ├─ infrastructure/console                 (NewScoreReporter, NewProcessReporter)
    ├─ infrastructure/filestore               (NewScoreFile, NewConfigFile)
    ├─ infrastructure/logger                  (NewLogger)
    ├─ infrastructure/pscommand               (NewManager)
    └─ infrastructure/tcellui                 (NewTUI)

entrypoint/cli
    ├─ application/apperror                   (Handler, Code.In)
    ├─ application/config                     (Options, GameOptions, ProcessOptions — what the flags map to)
    ├─ application/contract/inbound           (Runner interface, for the RunnerCreator return type)
    └─ internal/util                          (ClonePointer — flag_mapper.go only)

application/runner  (Service, wires configSvc/processSvc/scoreSvc/gameSvc — implements inbound.Runner)
    ├─ application/apperror                   (Handler)
    ├─ application/config                     (Result, GameResult, ProcessResult)
    ├─ application/game                       (PlayRequest, ProcessRequest, PlayResult)
    ├─ application/process                    (FindRequest, FindResult)
    └─ application/score                      (LoadResult, RecordRequest, RecordResult, ReportRequest)
    (no core package and no contract package: every collaborator is reached through runner's own
     consumer-side interfaces — configLoader, processFinder, scoreKeeper, gamePlayer — and inbound.Runner
     is satisfied structurally)

application/config
    ├─ application/apperror
    ├─ application/contract/outbound          (ConfigStore, Config, Mode, NotFoundError)
    ├─ core/game                              (ValidateTimeLimit)
    └─ core/movement                          (ValidateSpeed)

application/contract/inbound
    (no internal package imports)

application/contract/outbound
    └─ application/input                      (InputEventProvider's channel element type)

application/game
    ├─ application/apperror
    ├─ application/contract/outbound          (Renderer, InputEventProvider, WindowSize, ChromeSize, FrameViewState, TargetViewState, HUDViewState, StatusViewState, ConfirmViewState)
    ├─ application/input                      (Dispatcher, NewDispatcher)
    ├─ core/game                              (Session, NewSession, Config, Target, Input, NewInput, ValidateTimeLimit)
    ├─ core/movement                          (WindowSize, ChromeSize, Bounds, NewBounds, ValidateSpeed)
    └─ core/process                           (Info, NewInfo, ValidateProcesses)

application/process
    ├─ application/apperror
    ├─ application/contract/outbound          (ProcessInfo, ProcessManager, ProcessReporter, NotFoundError)
    └─ core/process                           (Info, NewInfo, Find, ValidateRoot, ValidatePatterns, ValidateProcesses, ValidateName)

application/score
    ├─ application/apperror
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, ScoreSummary, ScoreReporter, NotFoundError, CorruptedDataError)
    └─ core/score                             (Board, Entry, NewBoard)

application/input
    └─ core/game                              (Input, Target)

application/apperror
    └─ application/contract/outbound          (Logger — error_handler.go only)

infrastructure/tcellui
    ├─ application/contract/outbound          (WindowSize, ChromeSize, FrameViewState and its parts, Renderer, InputEventProvider)
    ├─ application/input                      (Event and its concrete event types)
    └─ internal/util                          (FormatBytes — hud.go only)

infrastructure/pscommand
    └─ application/contract/outbound          (ProcessInfo, ProcessManager, ProcessHandle, NotFoundError)

infrastructure/filestore
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, Config, Mode, GameConfig, ProcessConfig, ConfigStore, NotFoundError, CorruptedDataError)
    ├─ infrastructure/fsutil                  (ConfigDir, WriteFileAtomic)
    └─ internal/util                          (ClonePointer — converter.go only)

infrastructure/console
    ├─ application/contract/outbound          (ProcessInfo, ScoreSummary)
    └─ internal/util                          (FormatBytes — score_reporter.go only)

infrastructure/logger
    (no internal package imports — implements application/contract/outbound.Logger structurally)

core/game
    ├─ core/movement                          (Bounds, Motion, Throttle, WindowSize)
    └─ core/process                           (Info)

core/* imports nothing from application, infrastructure, entrypoint, or composition — confirmed
zero outward edges from core/game, core/movement, core/process, or core/score, not even to
internal/util.

application/config, application/game, application/process, and application/score share no
imports of one another — each speaks its own request/result data (config.Options/Result,
process.FindRequest/FindResult, game.PlayRequest/PlayResult, score.LoadResult/RecordRequest/
RecordResult/ReportRequest). application/runner is the only place all four application
sub-packages are wired together: its converter.go translates one service's result into the next
service's request, and it holds already-constructed collaborators
(NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)) behind its own consumer-side
interfaces rather than raw ports.

application/game depends on application/process only through the unexported `processKiller`
interface it declares itself — process.Service satisfies it structurally, with no import of
application/process required in application/game.

internal/composition is the only package that imports every infrastructure adapter and every
application service directly — that's its entire job. main.go imports only composition and cli.

Infrastructure packages satisfy contract interfaces via Go structural typing — no import of
application/contract required beyond the shared data/port types (infrastructure/logger needs
none at all).
```

---

## Call flow

```
main()
└─ run()
     ├─ composition.NewRunnerCreator()          → composition.RunnerCreator{errHandler: apperror.NewHandler(logger.NewLogger())}
     └─ cli.NewProgram(creator)                  → *cli.Program{creator, errHandler: creator.ErrHandler(), parser: newFlagParser(), out, errOut}
          └─ program.Run(os.Args[1:])
               ├─ p.run(args)   → showUsage bool, err error
               │    ├─ p.prepare(args)   → patterns []string, options config.Options, showUsage bool, err error
               │    │    ├─ no args / help(args)   [-h/--help short-circuits to showUsage; no cobra]
               │    │    └─ p.parser.parse(args)
               │    │         ├─ splitter.split(args)        → patterns, flagArgs   (an unknown flag or missing value → ArgumentError)
               │    │         ├─ flagSet.Parse(flagArgs)     (a leftover positional argument → ArgumentError)
               │    │         └─ mapper.toConfigOptions()    → config.Options{Game, Process}   (only explicitly-passed flags populate its pointer fields)
               │    ├─ creator.Create(patterns, options)                      [composition.RunnerCreator.Create()]
               │    │    ├─ pscommand.NewManager()                    → outbound.ProcessManager  (looks up `ps` on PATH)
               │    │    ├─ filestore.NewScoreFile()                  → outbound.ScoreStore  (~/.config/pidshooter/highscores.json)
               │    │    ├─ filestore.NewConfigFile()                 → outbound.ConfigStore (~/.config/pidshooter/config.yaml)
               │    │    ├─ tcell.NewScreen()
               │    │    ├─ tcellui.NewTUI(screen)                    → *tcellui.TUI  (outbound.Renderer)
               │    │    ├─ config.NewService(configStore, options)                          → *config.Service
               │    │    ├─ process.NewService(manager, console.NewProcessReporter(), patterns) → *process.Service
               │    │    ├─ score.NewService(scoreStore, console.NewScoreReporter())         → *score.Service
               │    │    ├─ game.NewService(processSvc, ui, ui.InputEvents())                → *game.Service   (processSvc satisfies game's processKiller interface; ui.InputEvents() is ui's own poller)
               │    │    └─ runner.NewService(configSvc, processSvc, scoreSvc, gameSvc, c.errHandler) → *runner.Service
               │    └─ runner.Run()     [runner.Service implements inbound.Runner — every step below routes failures through errHandler.Handle, which logs once and returns only Fatal/Unknown errors]
               │         ├─ configSvc.Load()
               │         │     ├─ s.validate()                            [core/movement.ValidateSpeed, core/game.ValidateTimeLimit — only for explicitly-passed flags; failure → Fatal CodeInvalidConfig]
               │         │     ├─ store.Load()                            [outbound.ConfigStore → filestore; NotFoundError → defaults + options; other failure → Warning CodeStoreFailed]
               │         │     └─ newResult().apply(stored, options)      merges: domain defaults < config file < options, options always win; a stored field failing validation is skipped and reported as a Warning CodeInvalidConfig
               │         ├─ processSvc.FindProcesses(toFindRequest(cfg.Process))
               │         │     ├─ process.ValidateRoot(manager.OwnUID(), request.AllowRoot)   refuses to run as root without --i-am-root
               │         │     ├─ process.ValidatePatterns(patterns)
               │         │     ├─ manager.Discover()                      [outbound.ProcessManager → pscommand: `ps -eo …`; failure → Fatal CodeProcessNotFound]
               │         │     ├─ process.Find(toInfos(processes), patterns, ownPID, ownUID, includeRoot)   → []core/process.Info
               │         │     ├─ process.ValidateProcesses(matches)
               │         │     ├─ reporter.Report(toProcessInfos(matches), patterns)  [outbound.ProcessReporter → console]
               │         │     └─ returns toFindResults(matches)          → []process.FindResult
               │         ├─ scoreSvc.LoadScoreBoard()                     [ScoreStore → filestore]  → LoadResult{Entries, HighScore}   (failure → Warning CodeStoreFailed, play continues)
               │         ├─ gameSvc.Play(toPlayRequest(cfg.Game, found, loaded.HighScore))
               │         │     ├─ validateGameConfig(request)             [core/movement.ValidateSpeed, core/game.ValidateTimeLimit]
               │         │     ├─ process.ValidateProcesses(toInfos(request.Processes))
               │         │     ├─ game.NewSession(processes, toGameConfig(request))   → *core/game.Session
               │         │     ├─ newKillTracker(request.HighScore)
               │         │     └─ runPlaySession(session, tracker)
               │         │          ├─ newPlaySession(renderer, events, killer, killGracePeriod, session, tracker)
               │         │          │     ├─ input.NewDispatcher(game.NewInput(session))
               │         │          │     └─ newInFlightKills(killer, killGracePeriod, tracker)
               │         │          ├─ play.run()
               │         │          │     ├─ renderer.Init()                         [Renderer → tcellui: screen.Init + poller.poll goroutine]
               │         │          │     ├─ session.Start(toBounds(window, chrome)) → pending → running, spawns targets
               │         │          │     ├─ registerTermSignalWatcher()             [signal.NotifyContext SIGINT/SIGTERM/SIGTSTP → a channel, no goroutine of ours]
               │         │          │     └─ frameLoop(termSignal) ─────────────────────────────────────────────────────────┐
               │         │          │          ├─ kills.applyFinished()               drains buffered async kill results     │
               │         │          │          │     └─ killResult.applyTo(tracker)   CeaseFire; CodeProcessNotFound → Reap → recordDud; other error → recordFailure; nil → Kill → recordKill
               │         │          │          ├─ drainEventQueue()                                                            │
               │         │          │          │     ├─ events.Events()              [InputEventProvider → tcellui.poller]     │
               │         │          │          │     ├─ dispatcher.Dispatch(ev)      → input.On*(...) → *Target or nil        │
               │         │          │          │     └─ hit: kills.start(target)     target.FireShot(); waitGroup.Go(killOrReap)
               │         │          │          │           └─ killer.Kill(pid, name, protected)   [processKiller → process.Service]
               │         │          │          │                 ├─ protected → Warning CodeKillFailed
               │         │          │          │                 ├─ manager.Pin(pid)          → ProcessHandle  [os.FindProcess + signal-0 verify; pidfd-pinned on Linux 5.3+]
               │         │          │          │                 ├─ manager.LookupName(pid)   [re-verify — the pinned handle survives PID reuse]
               │         │          │          │                 ├─ process.ValidateName(expected, actual)
               │         │          │          │                 ├─ handle.Kill()              [outbound.ProcessHandle → pscommand: SIGKILL]
               │         │          │          │                 │     (process gone / name mismatch → Warning CodeProcessNotFound; other failure → Warning CodeKillFailed)
               │         │          │          │                 └─ results <- killResult{target, err}                        │
               │         │          │          ├─ session.Update(toWindowSize(renderer.WindowSize()))  → timer.Expired / roster.move + target bounce
               │         │          │          ├─ renderer.Render(toFrameViewState(session, tracker))  [FrameViewState snapshot → tcellui]
               │         │          │          └─ select { ticker.C | termSignal → session.Stop() }   — same goroutine, no concurrent Session access
               │         │          │     └─ (loop exits) kills.awaitRemaining()   bounded by killGracePeriod (default 5s)
               │         │          │     └─ kills.close() / stopWatching() / renderer.Cleanup()  [deferred]
               │         │          └─ play.result(endTime)
               │         │     └─ returns PlayResult{Duration, LowestSpeed, Kills, FreedMem, Duds, Errors: tracker.errors()}   (each killFailure/killDud → *apperror.Error via toError)
               │         ├─ errHandler.HandleAll(result.Errors)            logs every kill failure / dud as a Warning
               │         ├─ scoreSvc.RecordScore(toRecordRequest(loaded, result, cfg.Game.TimeLimit))
               │         │     ├─ score.NewBoard(toEntries(request.Entries)).Add(toEntry(request))   [rejects non-scores: Entry.isScore() — Kills>0 && FreedMem>=0]
               │         │     ├─ toRecordResult(board, kills)            → RecordResult{Entries, NewHighScore}
               │         │     ├─ mergeWithLatest(entry)                  reloads the persisted board (NotFound/Corrupted → empty) and adds the entry to it
               │         │     └─ store.Save(...)                         [ScoreStore → filestore, atomic write via fsutil; failure → Warning CodeStoreFailed]
               │         └─ scoreSvc.ReportResults(toReportRequest(result, recorded))
               │               └─ reporter.Report(toScoreSummary(request))   [ScoreReporter → console]
               │    (a remaining error from runner.Run is a Fatal apperror.Error already logged once by errHandler above;
               │     p.run remaps a CodeInvalidConfig error specifically into cli.ArgumentError, so it's shown with usage text)
               ├─ p.errHandler.Handle(err)        [idempotent — an error already logged inside Run is not logged twice]
               └─ p.printUsageText(showUsage, err)  usage on missing args / --help / any ArgumentError
```

---

## Exported surface per package

| Package | Exported identifiers |
|---|---|
| `application/apperror` | `Code`, `CodeUnknown`, `CodeInvalidConfig`, `CodeProcessNotFound`, `CodeGameFailed`, `CodeKillFailed`, `CodeStoreFailed`, `(Code).In`, `Severity`, `SeverityUnknown`, `SeverityFatal`, `SeverityError`, `SeverityWarning`, `Error`, `NewError`, `Handler`, `NewHandler`, `(*Handler).Handle`, `(*Handler).HandleAll` |
| `application/config` | `Options`, `GameOptions`, `ProcessOptions`, `Result`, `GameResult`, `ProcessResult`, `Service`, `NewService` |
| `application/contract/inbound` | `Runner` |
| `application/contract/outbound` | `Mode`, `ModeGame`, `ModeLucky`, `ModeList`, `GameConfig`, `ProcessConfig`, `Config`, `ConfigStore`, `NotFoundError`, `CorruptedDataError`, `InputEventProvider`, `Logger`, `ProcessInfo`, `ProcessHandle`, `ProcessManager`, `ProcessReporter`, `ScoreEntry`, `ScoreBoard`, `ScoreStore`, `ScoreSummary`, `ScoreReporter`, `FrameViewState`, `TargetViewState`, `HUDViewState`, `StatusViewState`, `ConfirmViewState`, `WindowSize`, `ChromeSize`, `Renderer` |
| `application/input` | `Dispatcher`, `NewDispatcher`, `Event`, `ClickEvent`, `QuitEvent`, `ConfirmEvent`, `SpeedEvent` |
| `application/game` | `Service`, `NewService`, `PlayRequest`, `ProcessRequest`, `PlayResult` |
| `application/process` | `Service`, `NewService`, `FindRequest`, `FindResult` |
| `application/score` | `Service`, `NewService`, `BoardEntry`, `LoadResult`, `RecordRequest`, `RecordResult`, `ReportRequest` |
| `application/runner` | `Service`, `NewService` |
| `composition` | `RunnerCreator`, `NewRunnerCreator` |
| `core/game` | `Config`, `Session`, `NewSession`, `Target`, `NewTarget`, `State`, `Alive`, `Killing`, `Fleeing`, `Dead`, `AnimationDuration`, `Input`, `NewInput`, `ValidateTimeLimit` |
| `core/movement` | `Bounds`, `NewBounds`, `WindowSize`, `ChromeSize`, `Vector`, `Motion`, `NewMotion`, `Throttle`, `NewThrottle`, `MinSpeed`, `MaxSpeed`, `ValidateSpeed` |
| `core/process` | `Info`, `NewInfo`, `Find`, `ValidateProcesses`, `ValidateName`, `ValidatePatterns`, `ValidateRoot` |
| `core/score` | `Board`, `NewBoard`, `Entry` |
| `entrypoint/cli` | `Program`, `NewProgram`, `RunnerCreator`, `ArgumentError` |
| `infrastructure/tcellui` | `TUI`, `NewTUI` |
| `infrastructure/pscommand` | `NewManager` (the `manager` type itself is unexported — callers only ever see the `outbound.ProcessManager` it returns) |
| `infrastructure/filestore` | `NewScoreFile`, `NewConfigFile` (the `scoreFile`/`configFile` types themselves are unexported) |
| `infrastructure/console` | `ScoreReporter`, `NewScoreReporter`, `ProcessReporter`, `NewProcessReporter` |
| `infrastructure/fsutil` | `ConfigDir`, `WriteFileAtomic` |
| `infrastructure/logger` | `Logger`, `NewLogger` |
| `internal/testutil` | `FakeLogger`, `IsClosed`, `UnsetEnv`, `ExitedProcessPID`, `WarmUpSignalPackage` |
| `internal/util` | `FormatBytes`, `ClonePointer` |
