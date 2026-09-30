# pidshooter — Package & Architecture Diagram

*Updated 2026-09-30. Full hexagonal architecture: pure-core domain (game/movement/process/score),
a core/game `Input` intent gateway, an `application/input` dispatcher, contract ports split by
concern (inbound/outbound subpackages, including `ConfigStore` and `Logger` outbound ports
alongside `ProcessReporter`/`ScoreReporter` and the `ProcessManager`/`ProcessHandle` pair), a structured
`apperror` package classifying every application-layer failure and logging it through an injected
`Handler`, infrastructure adapters (a decomposed `tcellui`, a `console` reporter package, a
`filestore` package persisting both the score board and run config, and a dedicated `logger`
adapter), a hand-rolled (no `cobra`) entrypoint adapter, a dedicated `internal/composition`
package as the actual composition root, and an application layer split into four focused
services — `config`, `game`, `process`, `score` — composed by an `application/runner` package
that implements the `inbound.Runner` port. See `docs/hexagonal-arc.md` §15 for the narrative
behind this round's changes, and `docs/improvement-finding.md` for what's still open.*

---

## Package layout

```
cmd/pidshooter/
└── main.go                             thin entrypoint — builds composition.RunnerCreator + cli.Program, maps errors to exit codes, no wiring logic

internal/
├── core/                               pure domain — no infrastructure imports, no outward edges at all
│   ├── game/
│   │   ├── session.go                  lifecycle (enum: pending/running/stopped) · Config · Session · NewSession() · Start() · Update() · IsRunning() · Stop() · StartTime() · Throttle() · Targets() · AvailableTargets() · TimeLimit() · TimeLeft() · PendingConfirm() · RequestConfirm() · initialize() · currentState()
│   │   ├── timer.go                    timer · newTimer(limitSeconds, now func() time.Time) · Start() · Expired() · Remaining() · SecondsLeft() · LimitSeconds() · StartTime() · ValidateTimeLimit()
│   │   ├── confirmation.go             confirmation · newConfirmation() · Pending() · Request() · Accept() · Cancel()
│   │   ├── roster.go                   roster · newRoster() · spawn() · move() · allDead() · hitAt() · available()
│   │   ├── target.go                   Target (Info process.Info, Motion movement.Motion — named fields, not embedded) · State (enum: Alive/Killing/Fleeing/Dead) · AnimationDuration · NewTarget() · Tag() · AnimationProgress() · Update() · Kill() · Reap() · FireShot() · CeaseFire() · isAlive() · isDead() · isHitAt() · doomsdayTick() · move()
│   │   └── input.go                    Input · NewInput() · OnQuit() · OnYes() · OnNo() · OnSpeedUp() · OnSpeedDown() · OnClickAt()
│   ├── movement/
│   │   ├── bounds.go                   WindowSize · ChromeSize · Bounds · NewBounds() · Update() · bounce() · bounceLeft/Right/Top/Bottom()
│   │   ├── motion.go                   Vector · Motion · NewMotion() · Move()
│   │   ├── speed.go                    speed (unexported) · newSpeed() · Current() · Lowest() · Set()
│   │   └── throttle.go                 MinSpeed · MaxSpeed · Throttle · NewThrottle() · Speed() · LowestSpeed() · Increase() · Decrease() · ValidateSpeed()
│   ├── process/
│   │   └── info.go                     Info · NewInfo() · IsProtected() · IsKillableBy() · Find() · ValidateProcesses() · ValidateName() · ValidatePatterns() · ValidateRoot()
│   └── score/
│       ├── board.go                    Board · NewBoard() (routes every entry through Add) · HighScore() · Add() · IsNewHighScore() · killScore() · sortByRank() (sort.SliceStable) · less()
│       └── entry.go                    Entry · beats() · isScore()
├── application/
│   ├── apperror/
│   │   ├── error.go                    Code (enum, In(err) predicate) · Severity (enum) · Error · NewError() · Error() · Unwrap() · severity()
│   │   └── error_handler.go            Handler{logger outbound.Logger} · NewHandler(logger) · Handle() · logOnce()
│   ├── config/
│   │   ├── request.go                  GameRequest{ConfirmMode, Speed, TimeLimit *pointers} · ProcessRequest{IncludeRoot, AllowRoot *pointers} · Request{Game, Process} · validate()
│   │   ├── result.go                   GameResult · ProcessResult{IncludeRoot, AllowRoot} · Result{Mode outbound.Mode, Process, Game} · newResult() · apply() · fromStore() · fromRequest() · validateStore()
│   │   └── service.go                  Service{store outbound.ConfigStore} · NewService() · Load() · applyConfig()
│   ├── contract/
│   │   ├── inbound/
│   │   │   └── runner.go               RunRequest{Patterns, Config config.Request} · Runner (interface, Run(req RunRequest) error)
│   │   └── outbound/
│   │       ├── config.go               Mode (ModeGame/ModeYolo/ModeList) · GameConfig · ProcessConfig · ConfigStoreResult{Mode, Process, Game} · ConfigStore (interface)
│   │       ├── error.go                NotFoundError · CorruptedDataError
│   │       ├── input.go                InputEventProvider (interface)
│   │       ├── logger.go               Logger (interface)
│   │       ├── process.go              ProcessInfo · ProcessHandle (interface) · ProcessManager (interface) · ProcessReporter (interface)
│   │       ├── score.go                ScoreEntry · ScoreBoard · ScoreStore (interface) · ScoreSummary · ScoreReporter (interface)
│   │       └── ui.go                   FrameViewState · TargetViewState · HUDViewState · StatusViewState · ConfirmViewState · WindowSize · ChromeSize · Renderer (interface)
│   ├── input/
│   │   ├── dispatcher.go               Dispatcher · NewDispatcher() · Dispatch()
│   │   └── event.go                    EventDispatcher (interface, unexported dispatch() method) · ClickEvent · QuitEvent · ConfirmEvent · SpeedEvent
│   ├── game/
│   │   ├── service.go                  Service · PlayRequest{ConfirmMode, Speed, TimeLimit} · PlayResult · NewService() · Play() · validateGameConfig() · runLoop() · registerTermSignalWatcher() · frameLoop() · awaitOutstandingKills() · applyKillSignals() · applyKillSignal() · drainEventQueue() · killOrReap() · processKiller (unexported interface)
│   │   ├── kill_tracker.go             KillFailure · KillDud · killTracker · newKillTracker() · recordKill() · recordFailure() · recordDud()
│   │   └── converter.go                toGameConfig() · toBounds() · toConfirmViewState() · toTargetViewState() · toTargetViewStates() · toHUDViewState() · toStatusViewState() · toFrameViewState()
│   ├── process/
│   │   ├── service.go                  Service{manager, reporter} · FindRequest{IncludeRoot, AllowRoot} · NewService() · FindProcesses() · Kill()
│   │   └── converter.go                toProcessInfo() · toProcessInfos()
│   ├── runner/
│   │   ├── service.go                  Service{configSvc, processSvc, gameSvc, scoreSvc, errHandler} · NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler) · Run() · logKillFailures() · logDuds()   (implements inbound.Runner)
│   │   └── converter.go                toPlayRequest() · toFindRequest()
│   └── score/
│       ├── service.go                  Service{store, reporter} · NewService() · LoadScoreBoard() · RecordScore() · ReportResults() · mergeWithLatest()
│       └── converter.go                ToEntry() · toBoard() · toScoreBoard() · toScoreSummary() · toScoreEntries()
├── composition/                        the actual composition root — pure wiring, no behavior
│   └── runner_creator.go               RunnerCreator{errHandler} · NewRunnerCreator() · ErrHandler() · Create()
├── entrypoint/                         driving adapter — calls the application's inbound.Runner port
│   └── cli/
│       ├── program.go                  usageText · Program{creator, errHandler, out, errOut} · runnerCreator (interface) · NewProgram() · Run() · run() · prepare() · printUsageText() · help()
│       ├── flag_mapper.go              flagMapper · newFlagMapper() · toRunRequest()
│       ├── flag_splitter.go            flagSplitter · newFlagSplitter() · split() · looksLikeFlag() · isBoolFlag() · flagNameAndValue()
│       └── error.go                    ArgumentError · Error() · Unwrap()
├── infrastructure/                     driven adapters — implement outbound ports
│   ├── console/
│   │   ├── score_reporter.go           ScoreReporter · NewScoreReporter()  (implements outbound.ScoreReporter)
│   │   └── process_reporter.go         ProcessReporter · NewProcessReporter()  (implements outbound.ProcessReporter)
│   ├── filestore/
│   │   ├── config_file.go              configFile · configEntry · processEntry · configContent · NewConfigFile() · newConfigFileAt()  (implements outbound.ConfigStore) — schema-versioned YAML
│   │   ├── score_file.go               scoreFile · scoreEntry · scoreContent · NewScoreFile() · newScoreFileAt()  (implements outbound.ScoreStore) — schema-versioned JSON
│   │   ├── file.go                     file · newFile() · read() · write()  (I/O shared by both stores: bounded reads mapped to NotFoundError/CorruptedDataError, atomic writes)
│   │   ├── schema_version.go           schemaVersion · validate()
│   │   └── converter.go                toScoreBoard() · toScoreContent() · toConfigStoreResult() · toConfigContent()
│   ├── fsutil/
│   │   └── fsutil.go                   ConfigDir() · WriteFileAtomic()  (shared leaf helper, no port)
│   ├── logger/
│   │   └── logger.go                   Logger · NewLogger() · Warn() · Error()  (implements outbound.Logger, writes to os.Stderr)
│   ├── osprocess/
│   │   ├── process.go                  process · NewProcess()  (implements outbound.ProcessManager, `ps`-backed)
│   │   └── process_handle.go           processHandle  (implements outbound.ProcessHandle — pins a pid across verify-then-kill)
│   └── tcellui/                        (implements outbound.Renderer via *TUI, outbound.InputEventProvider via unexported *inputEvents)
│       ├── tui.go                      TUI · NewTUI() · Init() · Cleanup() · WindowSize() · ChromeSize() · Render() · InputEvents()
│       ├── input_events.go             inputEvents (unexported) · Events()  — shares TUI's poller
│       ├── poller.go                   poller · newPoller() · poll() · stop() · events()
│       ├── translator.go               translator · newTranslator() · translateEvent() · translateMouseEvent() · translateKeyEvent()
│       ├── renderer.go                 renderer · newRenderer() · render() · beginFrame() · drawFrame() · endFrame()
│       ├── hud.go                      hud · draw() and its text-budgeting helpers
│       ├── statusbar.go                statusBar · draw() · text() · confirmPromptText() · playStatusText() · truncate()
│       ├── target.go                   target (drawing helper, distinct from core/game.Target) · draw()
│       └── animation.go                killAnimation · fleeAnimation · animation · frame()
├── testutil/
│   ├── fake/                           ConfigStore · InputEventProvider · Logger · Process · ProcessHandle · ProcessKiller · ProcessReporter · Renderer · RunnerCreator · Runner · ScoreReporter · Store  (test doubles for every outbound port + inbound.Runner + composition.RunnerCreator)
│   ├── fixture/
│   │   ├── game.go                     Game() · ConfirmGameSession() · PendingConfirmGameSession()  (*core/game.Session builders)
│   │   └── process.go                  Process() · Processes()  (core/process.Info builders)
│   └── helper/
│       └── helper.go                   Ptr()  (generic pointer-literal helper, for inline optional-field struct literals) · UnsetEnv()
└── util/
    └── util.go                          FormatBytes() · ClonePtr()
```

---

## Class diagram

Types, their fields/methods, and relationships across packages. Visibility follows Go conventions: `+` = exported, `-` = unexported. Where a Go identifier collides with another package's identifier of the same name, the diagram node uses a disambiguated name instead — real Go names and package/file provenance are given in the prose and the `%%` section comments, not in the diagram itself, since Mermaid's `<<...>>` class annotation only accepts a single bare keyword (`interface`, `enumeration`, …), not free text. Five distinct `Service` types are disambiguated as `ConfigService`, `GameService`, `ProcessService`, `ScoreService`, and `RunnerService` (the last is `application/runner.Service`, which implements `inbound.Runner`). `application/config.Result.Mode` is `contract/outbound.Mode` reused directly — see the note near `Result` below for why that one field doesn't follow the pointer-vs-value split the rest of `Result`/`ConfigStoreResult` does. `core/game`'s session-level `Config` type is disambiguated as `SessionConfig`, to avoid colliding with `contract/outbound`'s real `GameConfig` persisted-config DTO. Purely internal rendering-detail helpers inside `infrastructure/tcellui` (`hud`, `statusBar`, `target`, `killAnimation`/`fleeAnimation`) are listed in the package layout above but not diagrammed field-by-field here — each owns drawing for one screen region and doesn't affect the port-facing architecture.

```mermaid
classDiagram
    direction TB

    %% ── application/config ───────────────────────────────────────────────────

    class GameRequest {
        +ConfirmMode *bool
        +Speed *float64
        +TimeLimit *int
    }

    class ProcessRequest {
        +IncludeRoot *bool
        +AllowRoot *bool
    }

    class Request {
        +Game GameRequest
        +Process ProcessRequest
        -validate() error
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
        -apply(stored, req) Result, error
        -fromStore(stored) Result, error
        -fromRequest(req) Result
    }

    %% Result.Mode is contract/outbound.Mode reused directly (not a separate
    %% config-local type) — Mode is a plain string on both sides, unlike
    %% GameResult/ProcessResult (plain-value) vs GameConfig/ProcessConfig
    %% (pointer-typed for the YAML round-trip), which do need to stay distinct.

    class ConfigService {
        -store outbound.ConfigStore
        +NewService(store) *Service
        +Load(req Request) Result, error
        -applyConfig(config, req) Result, error
    }

    %% package-level: newResult() Result · validateStore(stored) rejected []string, error

    Request *-- GameRequest : Game
    Request *-- ProcessRequest : Process
    Result *-- GameResult : Game
    Result *-- ProcessResult : Process
    Result --> Mode : Mode
    ConfigService --> ConfigStorePort : store
    ConfigService ..> Request : Load parameter
    ConfigService ..> Result  : Load returns

    %% ── application/apperror ─────────────────────────────────────────────────

    class Code {
        <<enumeration>>
        CodeUnknown
        CodeInvalidConfig
        CodeProcessDiscoveryFailed
        CodeProcessNotFound
        CodeGameFailed
        CodeKillFailed
        CodeStoreLoadFailed
        CodeStoreSaveFailed
        +In(err error) bool
    }

    class Severity {
        <<enumeration>>
        SeverityFatal
        SeverityError
        SeverityWarning
        SeverityUnknown
    }

    class AppError {
        +Code Code
        +Severity Severity
        +Message string
        +Cause error
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
        -logOnce(err, severity, appErr)
    }

    AppError --> Code     : Code
    AppError --> Severity : Severity
    Handler  --> LoggerPort : logger
    Handler  ..> AppError   : inspects via errors.As

    %% ── application/contract/inbound ─────────────────────────────────────────

    class RunRequest {
        +Patterns []string
        +Config Request
    }

    class RunnerPort {
        <<interface>>
        +Run(req RunRequest) error
    }

    RunRequest  --> Request   : Config
    RunnerPort  ..> RunRequest : Run parameter

    %% ── application/contract/outbound ────────────────────────────────────────

    class Mode {
        <<enumeration>>
        ModeGame
        ModeYolo
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

    class ConfigStoreResult {
        +Mode Mode
        +Process ProcessConfig
        +Game GameConfig
    }

    class ConfigStorePort {
        <<interface>>
        +Load() ConfigStoreResult, error
        +Save(result ConfigStoreResult) error
    }

    class LoggerPort {
        <<interface>>
        +Warn(msg string)
        +Error(msg string)
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
        +Report(count int, patterns []string)
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
        +NewHighScore bool
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
        +Events() chan EventDispatcher
    }

    ConfigStoreResult *-- GameConfig      : Game
    ConfigStoreResult *-- ProcessConfig   : Process
    ConfigStoreResult --> Mode            : Mode
    ConfigStorePort   ..> ConfigStoreResult : Load/Save
    ProcessManagerPort ..> ProcessHandlePort : Pin returns
    FrameViewState *-- "0..*" TargetViewState : Targets
    FrameViewState *-- HUDViewState           : HUD
    FrameViewState *-- StatusViewState        : StatusBar
    StatusViewState *-- "0..1" ConfirmViewState : Confirming
    ScoreBoardDTO *-- "0..*" ScoreEntryDTO   : Scores
    ScoreSummaryDTO *-- "0..*" ScoreEntryDTO : Entries

    %% ── core/movement ─────────────────────────────────────────────────────────

    class Vector {
        +X, Y float64
    }

    class Motion {
        +Position Vector
        +Velocity Vector
        +NewMotion(bounds, tagWidth) Motion
        +Move(bounds, speed, tagWidth)
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
        -bounce(pos, vel *Vector, tagWidth float64)
    }

    class speed {
        -current, lowest float64
        +newSpeed(v) speed
        +Current() float64
        +Lowest() float64
        +Set(v float64)
    }

    class Throttle {
        -speed speed
        +NewThrottle(speed) *Throttle
        +Speed() float64
        +LowestSpeed() float64
        +Increase()
        +Decrease()
    }

    %% package-level: MinSpeed, MaxSpeed (consts) · ValidateSpeed(speed float64) error

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
        +HighScore() int
        +Add(entry Entry)
        +IsNewHighScore(kills int) bool
        -killScore() int
        -sortByRank()
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
        +Update(window WindowSize)
        +IsRunning() bool
        +Stop()
        +StartTime() time.Time
        +Throttle() *Throttle
        +Targets() []*Target
        +AvailableTargets() []*Target, int
        +TimeLimit() int
        +TimeLeft() int
        +PendingConfirm() *Target
        +RequestConfirm(t *Target) *Target
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
        +Request(t *Target) *Target
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

    class EventDispatcher {
        <<interface>>
        -dispatch(input *core.Input) *Target
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
        +Dispatch(ev EventDispatcher) *Target
    }

    ClickEvent   ..|> EventDispatcher
    QuitEvent    ..|> EventDispatcher
    ConfirmEvent ..|> EventDispatcher
    SpeedEvent   ..|> EventDispatcher
    Dispatcher --> Input          : input
    Dispatcher ..> EventDispatcher : consumes
    Dispatcher ..> Target          : returns

    %% ── application/game ─────────────────────────────────────────────────────

    class processKiller {
        <<interface>>
        +Kill(pid int, name string, protected bool) shouldReap bool, err error
    }

    class PlayRequest {
        +ConfirmMode bool
        +Speed float64
        +TimeLimit int
    }

    class PlayResult {
        +Duration float64
        +LowestSpeed float64
        +Kills int
        +FreedMem int64
        +KillFailures []KillFailure
        +Duds []KillDud
    }

    class KillFailure {
        +Target string
        +PID int
        +Err error
    }

    class KillDud {
        +Target string
        +PID int
    }

    class killTracker {
        +newKillTracker(highScore) *killTracker
        +recordKill(freedMemory int64)
        +recordFailure(target, err)
        +recordDud(target)
    }

    class GameService {
        -killer processKiller
        -renderer outbound.Renderer
        -events outbound.InputEventProvider
        -killGracePeriod time.Duration
        +NewService(killer, renderer, events) *Service
        +Play(req PlayRequest, processes, highScore) PlayResult, error
        -validateGameConfig(req) error
        -runLoop(session, tracker) time.Time, error
        -frameLoop(session, tracker, dispatcher, termSignal, done) error
        -awaitOutstandingKills(wg, tracker, killSignals)
        -drainEventQueue(dispatcher, killSignals, done, wg) error
        -killOrReap(target, killSignals, done, wg)
    }

    %% package-level converter.go: toGameConfig(req PlayRequest) SessionConfig · toBounds() ·
    %% toConfirmViewState() · toTargetViewState() · toTargetViewStates() · toHUDViewState() ·
    %% toStatusViewState() · toFrameViewState()

    GameService --> processKiller          : killer
    GameService --> RendererPort           : renderer
    GameService --> InputEventProviderPort : events
    GameService ..> killTracker            : creates per Play()
    GameService ..> Session                : creates (NewSession)
    GameService ..> SessionConfig          : builds (toGameConfig)
    GameService ..> Dispatcher             : creates (runLoop)
    GameService ..> Input                  : creates (NewInput)
    GameService ..> PlayRequest            : accepts (Play req)
    GameService ..> PlayResult             : returns
    GameService ..> FrameViewState         : builds (toFrameViewState)
    GameService ..> Info                   : accepts (Play processes)
    killTracker *-- "0..*" KillFailure     : failures
    killTracker *-- "0..*" KillDud         : duds

    %% ── application/process ──────────────────────────────────────────────────

    class FindRequest {
        +IncludeRoot bool
        +AllowRoot bool
    }

    class ProcessService {
        -manager outbound.ProcessManager
        -reporter outbound.ProcessReporter
        +NewService(manager, reporter) *Service
        +FindProcesses(patterns, req FindRequest) []Info, error
        +Kill(pid, procName, protected) shouldReap bool, err error
    }

    %% package-level converter.go: toProcessInfo() · toProcessInfos()

    ProcessService --> ProcessManagerPort  : manager
    ProcessService --> ProcessReporterPort : reporter
    ProcessService ..> FindRequest         : accepts (FindProcesses)
    ProcessService ..> Info                : returns (FindProcesses)
    ProcessService ..|> processKiller      : satisfies (structural typing, no import)

    %% ── application/score ────────────────────────────────────────────────────

    class ScoreService {
        -store outbound.ScoreStore
        -reporter outbound.ScoreReporter
        +NewService(store, reporter) *Service
        +LoadScoreBoard() *Board, int, error
        +RecordScore(board, entry, err) error
        +ReportResults(duration, kills, duds, freedMem, board)
        -mergeWithLatest(entry, board) *Board
    }

    %% package-level converter.go: ToEntry(result PlayResult, cfgTimeLimit) Entry ·
    %% toBoard() · toScoreBoard() · toScoreSummary() · toScoreEntries()

    ScoreService --> ScoreStorePort    : store
    ScoreService --> ScoreReporterPort : reporter
    ScoreService ..> Board             : loads/returns
    ScoreService ..> PlayResult        : ToEntry accepts

    %% ── application/runner ───────────────────────────────────────────────────

    class RunnerService {
        -configSvc *ConfigService
        -processSvc *ProcessService
        -gameSvc *GameService
        -scoreSvc *ScoreService
        -errHandler *Handler
        +NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler) *Service
        +Run(req RunRequest) error
        -logKillFailures(failures)
        -logDuds(duds)
    }

    %% package-level converter.go: toPlayRequest(cfg GameResult) PlayRequest ·
    %% toFindRequest(cfg ProcessResult) FindRequest

    RunnerService ..|> RunnerPort     : implements
    RunnerService *-- ConfigService   : configSvc
    RunnerService *-- ProcessService  : processSvc
    RunnerService *-- GameService     : gameSvc
    RunnerService *-- ScoreService    : scoreSvc
    RunnerService --> Handler         : errHandler
    RunnerService ..> RunRequest      : Run parameter
    RunnerService ..> Info            : FindProcesses results

    %% ── infrastructure/tcellui ───────────────────────────────────────────────

    class TUI {
        -screen tcell.Screen
        -chrome outbound.ChromeSize
        -poller poller
        -initialized bool
        +NewTUI(screen) *TUI
        +Init() error
        +Cleanup()
        +WindowSize() WindowSize
        +ChromeSize() ChromeSize
        +Render(state FrameViewState)
        +InputEvents() InputEventProvider
    }

    class inputEvents {
        -poller *poller
        +Events() chan EventDispatcher
    }

    class poller {
        -screen tcell.Screen
        -eventQueue chan EventDispatcher
        -translator translator
        +newPoller(screen) poller
        +poll()
        +stop()
        +events() chan EventDispatcher
    }

    TUI ..|> RendererPort              : implements
    inputEvents ..|> InputEventProviderPort : implements
    TUI *-- poller       : poller
    TUI ..> inputEvents  : InputEvents() returns (shares TUI's poller, so Init/Cleanup on TUI still control it)
    inputEvents --> poller : poller

    %% ── infrastructure/osprocess ─────────────────────────────────────────────

    class OsProcess {
        -psPath string
        -timeout time.Duration
        +NewProcess() ProcessManager, error
        +Discover() []ProcessInfo, error
        +OwnPID() int
        +OwnUID() int
        +LookupName(pid int) string, error
        +Pin(pid int) ProcessHandle, error
    }

    class OsProcessHandle {
        -proc *os.Process
        +Kill() error
        +Release() error
    }

    OsProcess ..|> ProcessManagerPort  : implements
    OsProcess ..> OsProcessHandle      : Pin creates
    OsProcessHandle ..|> ProcessHandlePort : implements

    %% ── infrastructure/filestore ─────────────────────────────────────────────

    class SharedFile {
        -path string
        -maxSize int
        +newFile(filename, maxSize) file, error
        -read() []byte, error
        -write(data []byte) error
    }

    class ScoreFile {
        -file SharedFile
        +NewScoreFile() ScoreStore, error
        +Load() ScoreBoard, error
        +Save(board ScoreBoard) error
    }

    class ConfigFile {
        -file SharedFile
        +NewConfigFile() ConfigStore, error
        +Load() ConfigStoreResult, error
        +Save(result ConfigStoreResult) error
    }

    %% package-level: schemaVersion(v).validate(fileType, version) error (schema_version.go) ·
    %% converter.go: toScoreBoard() · toScoreContent() · toConfigStoreResult() · toConfigContent()

    ScoreFile  *-- SharedFile          : file
    ConfigFile *-- SharedFile          : file
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
        +Report(count int, patterns []string)
    }

    ConsoleScoreReporter   ..|> ScoreReporterPort   : implements
    ConsoleProcessReporter ..|> ProcessReporterPort : implements

    %% ── infrastructure/logger ────────────────────────────────────────────────

    class InfraLogger {
        +NewLogger() *Logger
        +Warn(msg string)
        +Error(msg string)
    }

    InfraLogger ..|> LoggerPort : implements

    %% ── internal/composition ─────────────────────────────────────────────────

    class RunnerCreator {
        -errHandler *apperror.Handler
        +NewRunnerCreator() RunnerCreator
        +ErrHandler() *apperror.Handler
        +Create() Runner, error
    }

    RunnerCreator ..> OsProcess              : builds
    RunnerCreator ..> ScoreFile              : builds
    RunnerCreator ..> ConfigFile             : builds
    RunnerCreator ..> TUI                    : builds
    RunnerCreator ..> ConsoleProcessReporter : builds
    RunnerCreator ..> ConsoleScoreReporter   : builds
    RunnerCreator ..> InfraLogger            : builds (via NewHandler)
    RunnerCreator ..> Handler                : builds (NewHandler)
    RunnerCreator ..> ConfigService          : builds
    RunnerCreator ..> ProcessService         : builds
    RunnerCreator ..> ScoreService           : builds
    RunnerCreator ..> GameService            : builds
    RunnerCreator ..> RunnerService          : builds & returns

    %% ── entrypoint/cli ───────────────────────────────────────────────────────

    class Program {
        -creator programRunnerCreator
        -errHandler *apperror.Handler
        -out, errOut io.Writer
        +NewProgram(creator) *Program
        +Run(args []string) error
        -run(args) bool, error
        -prepare(args) req RunRequest, showUsage bool, err error
        -printUsageText(showUsage, err)
    }

    class programRunnerCreator {
        <<interface>>
        +Create() Runner, error
        +ErrHandler() *apperror.Handler
    }

    class flagMapper {
        -flagSet *flag.FlagSet
        -confirm *bool
        -speed *float64
        -timeLimit *int
        -includeRoot *bool
        -allowRoot *bool
        +newFlagMapper(flagSet) *flagMapper
        +toRunRequest(patterns) RunRequest
    }

    class flagSplitter {
        +newFlagSplitter(flagSet) flagSplitter
        +split(args) patterns []string, flagArgs []string, err error
    }

    class ArgumentError {
        +Cause error
        +Error() string
        +Unwrap() error
    }

    Program --> programRunnerCreator : creator
    Program --> Handler              : errHandler
    Program ..> flagSplitter         : uses (prepare)
    Program ..> flagMapper           : uses (prepare)
    Program ..> RunRequest           : builds via flagMapper before calling Create()
    flagMapper ..> RunRequest        : returns (toRunRequest)
    programRunnerCreator <|.. RunnerCreator : satisfied by (structural typing)

    %% ── internal/util ────────────────────────────────────────────────────────

    %% package-level: FormatBytes(bytes int64) string · ClonePtr[T](p *T) *T  (util.go — no exported types, generic helpers only)
```

---

## Package dependency graph

```
cmd/pidshooter
    ├─ internal/composition                   (RunnerCreator, NewRunnerCreator)
    └─ entrypoint/cli                         (Program, NewProgram, ArgumentError)

internal/composition
    ├─ application/apperror                   (Handler, NewHandler)
    ├─ application/config                     (NewService)
    ├─ application/contract/inbound           (Runner interface)
    ├─ application/game                       (NewService)
    ├─ application/process                    (NewService)
    ├─ application/runner                     (NewService)
    ├─ application/score                      (NewService)
    ├─ infrastructure/console                 (NewScoreReporter, NewProcessReporter)
    ├─ infrastructure/filestore               (NewScoreFile, NewConfigFile)
    ├─ infrastructure/logger                  (NewLogger)
    ├─ infrastructure/osprocess               (NewProcess)
    └─ infrastructure/tcellui                 (NewTUI)

entrypoint/cli
    ├─ application/apperror                   (Handler, Code.In)
    ├─ application/contract/inbound           (Runner interface, RunRequest, for the runnerCreator return type)
    └─ internal/util                          (ClonePtr — flag_mapper.go only)

application/runner  (Service, wires configSvc/processSvc/scoreSvc/gameSvc — implements inbound.Runner)
    ├─ application/apperror
    ├─ application/config
    ├─ application/contract/inbound
    ├─ application/game
    └─ application/process
    (application/score is reached only through the score.ToEntry call, not a Service field import beyond what application/game already provides)

application/config
    ├─ application/apperror
    ├─ application/contract/outbound          (ConfigStore, ConfigStoreResult, NotFoundError)
    ├─ core/game                              (TimeLimit validation range)
    └─ core/movement                          (MinSpeed, MaxSpeed)

application/contract/inbound
    └─ application/config                     (Request, embedded in RunRequest.Config)

application/contract/outbound
    └─ application/input                      (InputEventProvider's channel element type)

application/game
    ├─ application/apperror
    ├─ application/contract/outbound          (Renderer, InputEventProvider, FrameViewState, HUDViewState, StatusViewState, ConfirmViewState)
    ├─ application/input                      (Dispatcher, NewDispatcher, EventDispatcher)
    ├─ core/game                              (Session, NewSession, Config, Target, Input, NewInput)
    ├─ core/movement                          (WindowSize)
    └─ core/process                           (Info)

application/process
    ├─ application/apperror
    ├─ application/contract/outbound          (ProcessInfo, ProcessManager, ProcessHandle, ProcessReporter)
    └─ core/process                           (Info, Find, Validate*, ValidateRoot)

application/score
    ├─ application/apperror
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, ScoreSummary, ScoreReporter)
    ├─ application/game                       (PlayResult, for ToEntry)
    └─ core/score                             (Board, Entry, NewBoard)

application/input
    └─ core/game                              (Input, Target)

application/apperror
    └─ application/contract/outbound          (Logger — error_handler.go only)

infrastructure/tcellui
    ├─ application/contract/outbound          (WindowSize, ChromeSize, FrameViewState, Renderer, InputEventProvider)
    └─ application/input                      (EventDispatcher and its concrete event types)

infrastructure/osprocess
    └─ application/contract/outbound          (ProcessInfo, ProcessManager, ProcessHandle — satisfied structurally)

infrastructure/filestore
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, ConfigStoreResult, ConfigStore, NotFoundError, CorruptedDataError)
    ├─ infrastructure/fsutil                  (ConfigDir, WriteFileAtomic)
    └─ internal/util                          (ClonePtr — converter.go only)

infrastructure/console
    ├─ application/contract/outbound          (ScoreSummary, ScoreReporter, ProcessReporter)
    └─ internal/util                          (FormatBytes — score_reporter.go only)

infrastructure/logger
    (no internal package imports — implements application/contract/outbound.Logger structurally)

core/game
    ├─ core/movement                          (Bounds, Motion, Throttle, WindowSize, ChromeSize)
    └─ core/process                           (Info)

core/* imports nothing from application, infrastructure, entrypoint, or composition — confirmed
zero outward edges from core/game, core/movement, core/process, or core/score, not even to
internal/util.

application/config, application/game, application/process, and application/score share no
imports of one another (application/score depends on application/game only for the PlayResult
*type*, used by its own ToEntry converter — it still never touches application/process or
application/config). application/runner is the only place all four application sub-packages are
wired together, and it does so with already-constructed collaborators
(NewService(configSvc, processSvc, scoreSvc, gameSvc, errHandler)) rather than raw ports.

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
     └─ cli.NewProgram(creator)                  → *cli.Program{creator, errHandler: creator.ErrHandler(), out, errOut}
          └─ program.Run(os.Args[1:])
               ├─ p.run(args)   → showUsage bool, err error
               │    ├─ p.prepare(args)   → req inbound.RunRequest, showUsage bool, err error
               │    │    ├─ help(args) / flagSplitter.split(args) / flag.FlagSet.Parse   [no cobra; -h/--help or no args short-circuits to showUsage]
               │    │    └─ flagMapper.toRunRequest(patterns)   → inbound.RunRequest{Patterns, Config: config.Request{Game, Process}}   (only explicitly-passed flags populate Config's pointer fields)
               │    ├─ creator.Create()                                    [composition.RunnerCreator.Create()]
               │    │    ├─ osprocess.NewProcess()                    → outbound.ProcessManager
               │    │    ├─ filestore.NewScoreFile()                  → outbound.ScoreStore  (~/.config/pidshooter/highscores.json)
               │    │    ├─ filestore.NewConfigFile()                 → outbound.ConfigStore (~/.config/pidshooter/config.yaml)
               │    │    ├─ tcell.NewScreen()
               │    │    ├─ tcellui.NewTUI(screen)                    → *tcellui.TUI  (outbound.Renderer)
               │    │    ├─ config.NewService(configStore)                        → *config.Service
               │    │    ├─ process.NewService(manager, console.NewProcessReporter())  → *process.Service
               │    │    ├─ score.NewService(store, console.NewScoreReporter())     → *score.Service
               │    │    ├─ game.NewService(processSvc, ui, ui.InputEvents())       → *game.Service   (processSvc satisfies game's processKiller interface; ui.InputEvents() shares ui's poller)
               │    │    └─ runner.NewService(configSvc, processSvc, scoreSvc, gameSvc, c.errHandler)    → *runner.Service
               │    └─ runner.Run(req)     [runner.Service implements inbound.Runner — every step below routes failures through errHandler.Handle before returning]
               │         ├─ configSvc.Load(req.Config)
               │         │     ├─ req.validate()                          [core/movement.ValidateSpeed, core/game.ValidateTimeLimit — only for explicitly-passed flags]
               │         │     ├─ configStore.Load()                      [outbound.ConfigStore → filestore]
               │         │     └─ newResult().apply(stored, req)          merges: domain defaults < config file < req, req always wins; a stored field failing validation is skipped and reported as a warning
               │         ├─ processSvc.FindProcesses(req.Patterns, toFindRequest(cfg.Process))
               │         │     ├─ process.ValidateRoot(manager.OwnUID(), req.AllowRoot)   refuses to run as root without --i-am-root
               │         │     ├─ process.ValidatePatterns(patterns)
               │         │     ├─ manager.Discover()                      [outbound.ProcessManager → osprocess]
               │         │     ├─ process.Find(infos, patterns, ownPID, ownUID, includeRoot)   → []core/process.Info
               │         │     ├─ process.ValidateProcesses(matches)
               │         │     └─ reporter.Report(len(matches), patterns)  [outbound.ProcessReporter → console]
               │         ├─ scoreSvc.LoadScoreBoard()                 [ScoreStore → filestore]  → board, highScore, loadErr
               │         ├─ gameSvc.Play(toPlayRequest(cfg.Game), processes, highScore)
               │         │     ├─ validateGameConfig(req)                 [core/movement.ValidateSpeed, core/game.ValidateTimeLimit]
               │         │     ├─ game.NewSession(processes, toGameConfig(req))   → *core/game.Session
               │         │     ├─ newKillTracker(highScore)
               │         │     └─ runLoop(session, tracker)
               │         │          ├─ renderer.Init()                         [Renderer → tcellui: screen.Init + poller.poll goroutine]
               │         │          ├─ session.Start(bounds)                   → pending → running, spawns targets
               │         │          ├─ registerTermSignalWatcher()             [signal.NotifyContext SIGINT/SIGTERM/SIGTSTP → a channel, no goroutine of ours]
               │         │          ├─ input.NewDispatcher(game.NewInput(session))
               │         │          └─ frameLoop(session, tracker, dispatcher, termSignal, done) ─────────────────────────┐
               │         │               ├─ applyKillSignals(tracker, killSignals)   drains buffered async kill results  │
               │         │               ├─ drainEventQueue(dispatcher, killSignals, done, wg)                           │
               │         │               │     ├─ events.Events()              [InputEventProvider → tcellui.inputEvents]│
               │         │               │     ├─ dispatcher.Dispatch(ev)      → input.On*(...) → *Target or nil        │
               │         │               │     └─ hit: target.FireShot(); go killOrReap(target, ...)                    │
               │         │               │           └─ killer.Kill(pid, name, protected)   [processKiller → process.Service]
               │         │               │                 ├─ manager.Pin(pid)          → ProcessHandle  [pidfd-pinned on Linux 5.3+]
               │         │               │                 ├─ manager.LookupName(pid)   [re-verify — the pinned handle survives PID reuse]
               │         │               │                 ├─ process.ValidateName(expected, actual)
               │         │               │                 ├─ handle.Kill()              [outbound.ProcessHandle → osprocess]
               │         │               │                 └─ killSignals <- killSignal{target, shouldReap, err}        │
               │         │               ├─ session.Update(window)             → timer.Expired / roster.move + target bounce
               │         │               ├─ renderer.Render(toFrameViewState(session, tracker))  [FrameViewState snapshot → tcellui]
               │         │               └─ select { ticker.C | termSignal → session.Stop() }   — same goroutine, no concurrent Session access
               │         │          └─ (loop exits) awaitOutstandingKills(wg, tracker, killSignals)   bounded by killGracePeriod (default 5s)
               │         │          └─ renderer.Cleanup()  [deferred]
               │         │     └─ returns PlayResult{Duration, LowestSpeed, Kills, FreedMem, KillFailures, Duds}
               │         ├─ logKillFailures(result.KillFailures) / logDuds(result.Duds)     [via errHandler.Handle]
               │         ├─ score.ToEntry(result, cfg.Game.TimeLimit)                → core/score.Entry
               │         ├─ scoreSvc.RecordScore(board, entry, loadErr)
               │         │     ├─ board.Add(entry)              [rejects non-scores: Entry.isScore() — Kills>0 && FreedMem>=0]
               │         │     └─ store.Save(...)               [ScoreStore → filestore, atomic write via fsutil]
               │         └─ scoreSvc.ReportResults(duration, kills, duds, freedMem, board)
               │               └─ reporter.Report(summary)      [ScoreReporter → console]
               │    (a remaining error from runner.Run is a Fatal apperror.Error already logged once by errHandler above;
               │     p.run remaps a CodeInvalidConfig error specifically into cli.ArgumentError, so it's shown with usage text)
               ├─ p.errHandler.Handle(err)        [idempotent — an error already logged inside Run is not logged twice]
               └─ p.printUsageText(showUsage, err)  usage on missing args / --help / any ArgumentError
```

---

## Exported surface per package

| Package | Exported identifiers |
|---|---|
| `application/apperror` | `Code`, `CodeUnknown`, `CodeInvalidConfig`, `CodeProcessDiscoveryFailed`, `CodeProcessNotFound`, `CodeGameFailed`, `CodeKillFailed`, `CodeStoreLoadFailed`, `CodeStoreSaveFailed`, `(Code).In`, `Severity`, `SeverityFatal`, `SeverityError`, `SeverityWarning`, `SeverityUnknown`, `Error`, `NewError`, `Handler`, `NewHandler`, `(*Handler).Handle` |
| `application/config` | `Request`, `GameRequest`, `ProcessRequest`, `Result`, `Mode`, `ModeGame`, `ModeYolo`, `ModeList`, `GameResult`, `ProcessResult`, `Service`, `NewService` |
| `application/contract/inbound` | `RunRequest`, `Runner` |
| `application/contract/outbound` | `Mode`, `ModeGame`, `ModeYolo`, `ModeList`, `GameConfig`, `ProcessConfig`, `ConfigStoreResult`, `ConfigStore`, `NotFoundError`, `CorruptedDataError`, `InputEventProvider`, `Logger`, `ProcessInfo`, `ProcessHandle`, `ProcessManager`, `ProcessReporter`, `ScoreEntry`, `ScoreBoard`, `ScoreStore`, `ScoreSummary`, `ScoreReporter`, `FrameViewState`, `TargetViewState`, `HUDViewState`, `StatusViewState`, `ConfirmViewState`, `WindowSize`, `ChromeSize`, `Renderer` |
| `application/input` | `Dispatcher`, `NewDispatcher`, `EventDispatcher`, `ClickEvent`, `QuitEvent`, `ConfirmEvent`, `SpeedEvent` |
| `application/game` | `Service`, `NewService`, `PlayRequest`, `PlayResult`, `KillFailure`, `KillDud` |
| `application/process` | `Service`, `NewService`, `FindRequest` |
| `application/score` | `Service`, `NewService`, `ToEntry` |
| `application/runner` | `Service`, `NewService` |
| `composition` | `RunnerCreator`, `NewRunnerCreator` |
| `core/game` | `Config`, `Session`, `NewSession`, `Target`, `NewTarget`, `State`, `Alive`, `Killing`, `Fleeing`, `Dead`, `AnimationDuration`, `Input`, `NewInput`, `ValidateTimeLimit` |
| `core/movement` | `Bounds`, `NewBounds`, `WindowSize`, `ChromeSize`, `Vector`, `Motion`, `NewMotion`, `Throttle`, `NewThrottle`, `MinSpeed`, `MaxSpeed`, `ValidateSpeed` |
| `core/process` | `Info`, `NewInfo`, `Find`, `ValidateProcesses`, `ValidateName`, `ValidatePatterns`, `ValidateRoot` |
| `core/score` | `Board`, `NewBoard`, `Entry` |
| `entrypoint/cli` | `Program`, `NewProgram`, `ArgumentError` |
| `infrastructure/tcellui` | `TUI`, `NewTUI` |
| `infrastructure/osprocess` | `NewProcess` (the `process` type itself is unexported — callers only ever see the `outbound.ProcessManager` it returns) |
| `infrastructure/filestore` | `NewScoreFile`, `NewConfigFile` (the `scoreFile`/`configFile` types themselves are unexported) |
| `infrastructure/console` | `ScoreReporter`, `NewScoreReporter`, `ProcessReporter`, `NewProcessReporter` |
| `infrastructure/logger` | `Logger`, `NewLogger` |
| `internal/util` | `FormatBytes`, `ClonePtr` |
