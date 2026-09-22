# pidshooter — Package & Architecture Diagram

*Updated 2026-09-22. Full hexagonal architecture: pure-core domain (game/movement/process/score),
a core/game `Input` intent gateway, an `application/input` dispatcher, contract ports split by
concern (inbound/outbound subpackages, including `ProcessReporter`/`ScoreReporter` and a
`Process`/`ProcessHandle` pair replacing the old `Lister`/`Killer` split), a structured `apperror`
package classifying every application-layer failure, infrastructure adapters (including a
decomposed `tcellui` and a new `console` reporter package), a hand-rolled (no `cobra`) entrypoint
adapter, a dedicated `internal/composition` package as the actual composition root, and an
application layer split into three focused services — `game`, `process`, `score` — composed by a
top-level `Runner` that implements the `inbound.Runner` port. See `docs/hexagonal-arc.md` §14 for
the narrative behind this round's changes, and `docs/leftovers.md` for the finding-by-finding detail.*

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
│   │   └── info.go                     Info · NewInfo() · IsProtected() · IsKillableBy() · Find() · ValidateProcesses() · ValidateName() · ValidatePatterns()
│   └── score/
│       ├── board.go                    Board · NewBoard() (routes every entry through Add) · HighScore() · Add() · IsNewHighScore() · killScore() · sortByRank() (sort.SliceStable) · less()
│       └── entry.go                    Entry · beats() · isScore()
├── application/
│   ├── apperror/
│   │   └── apperror.go                 Code (enum) · Severity (enum) · Error · NewError() · Error() · Unwrap() · Handle() · logOnce() · severityOf()
│   ├── config/
│   │   └── config.go                   Config{Patterns, ConfirmMode, Speed, TimeLimit, IncludeRoot} · Validate()
│   ├── contract/
│   │   ├── inbound/
│   │   │   └── runner.go               Runner (interface, Run(cfg config.Config) error)
│   │   └── outbound/
│   │       ├── error.go                NotFoundError · CorruptedDataError
│   │       ├── input.go                InputEventProvider (interface)
│   │       ├── process.go              ProcessInfo · ProcessHandle (interface) · Process (interface) · ProcessReporter (interface)
│   │       ├── score.go                ScoreEntry · ScoreBoard · ScoreStore (interface) · ScoreSummary · ScoreReporter (interface)
│   │       └── ui.go                   FrameViewState · TargetViewState · HUDViewState · StatusViewState · ConfirmViewState · WindowSize · ChromeSize · Renderer (interface)
│   ├── input/
│   │   ├── dispatcher.go               Dispatcher · NewDispatcher() · Dispatch()
│   │   └── event.go                    EventDispatcher (interface, unexported dispatch() method) · ClickEvent · QuitEvent · ConfirmEvent · SpeedEvent
│   ├── game/
│   │   ├── service.go                  Service · NewService() · PlayResult · Play() · runLoop() · registerTermSignalWatcher() · frameLoop() · awaitOutstandingKills() · applyKillSignals() · applyKillSignal() · drainEventQueue() · killOrReap() · processKiller (unexported interface)
│   │   ├── kill_tracker.go             KillFailure · KillDud · killTracker · newKillTracker() · recordKill() · recordFailure() · recordDud()
│   │   └── converter.go                toBounds() · toConfirmViewState() · toTargetViewState() · toTargetViewStates() · toHUDViewState() · toStatusViewState() · toFrameViewState()
│   ├── process/
│   │   ├── service.go                  Service{proc, reporter} · NewService() · FindProcesses() · Kill()
│   │   └── converter.go                toProcessInfo() · toProcessInfos()
│   ├── score/
│   │   ├── service.go                  Service{store, reporter} · NewService() · LoadScoreBoard() · RecordScore() · ReportResults() · mergeWithLatest()
│   │   └── converter.go                ToEntry() · toBoard() · toScoreBoard() · toScoreSummary() · toScoreEntries()
│   └── runner.go                       Runner{processSvc, gameSvc, scoreSvc} · NewRunner(processSvc, scoreSvc, gameSvc) · Run() · logKillFailures() · logDuds()   (application package — implements inbound.Runner)
├── composition/                        the actual composition root — pure wiring, no behavior
│   └── runner_creator.go               RunnerCreator · NewRunnerCreator() · Create()
├── entrypoint/                         driving adapter — calls the application's inbound.Runner port
│   └── cli/
│       ├── program.go                  usageText · Program · runnerCreator (interface) · NewProgram() · Run() · run() · printUsageText() · help()
│       ├── flag_splitter.go            flagSplitter · newFlagSplitter() · split() · looksLikeFlag() · isBoolFlag() · flagNameAndValue()
│       └── error.go                    ArgumentError · Error() · Unwrap()
├── infrastructure/                     driven adapters — implement outbound ports
│   ├── console/
│   │   ├── score_reporter.go           ScoreReporter · NewScoreReporter()  (implements outbound.ScoreReporter)
│   │   └── process_reporter.go         ProcessReporter · NewProcessReporter()  (implements outbound.ProcessReporter)
│   ├── filescore/
│   │   └── file_score.go               fileScore · fileEntry · fileContent · NewFileScore() · newFileScoreAt()  (implements outbound.ScoreStore) — schema-versioned JSON, atomic writes
│   ├── fsutil/
│   │   └── fsutil.go                   ConfigDir() · WriteFileAtomic()  (shared leaf helper, no port)
│   ├── osprocess/
│   │   ├── process.go                  process · NewProcess()  (implements outbound.Process, `ps`-backed)
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
│   ├── fake/                           InputEventProvider · Process · ProcessHandle · ProcessKiller · ProcessReporter · Renderer · RunnerCreator · Runner · ScoreReporter · Store  (test doubles for every outbound port + inbound.Runner + composition.RunnerCreator)
│   ├── fixture/
│   │   ├── game.go                     Game() · ConfirmGameSession() · PendingConfirmGameSession()  (*core/game.Session builders)
│   │   └── process.go                  Process() · Processes()  (core/process.Info builders)
│   └── helper/
│       └── helper.go                   UnsetEnv()
└── util/
    ├── logger.go                        Logger · NewLogger() · Warn() · Error()  (called directly by apperror — not an adapter, no port)
    └── util.go                          FormatBytes()
```

---

## Class diagram

Types, their fields/methods, and relationships across packages. Visibility follows Go conventions: `+` = exported, `-` = unexported. Where a Go identifier collides with another package's identifier of the same name (e.g. three `Service` types, or `Process`/`ProcessInfo`/`ProcessHandle` appearing both as ports and as concrete adapter types), the diagram node uses a disambiguated name instead — real Go names and package/file provenance are given in the prose and the `%%` section comments, not in the diagram itself, since Mermaid's `<<...>>` class annotation only accepts a single bare keyword (`interface`, `enumeration`, …), not free text. Purely internal rendering-detail helpers inside `infrastructure/tcellui` (`hud`, `statusBar`, `target`, `killAnimation`/`fleeAnimation`) are listed in the package layout above but not diagrammed field-by-field here — each owns drawing for one screen region and doesn't affect the port-facing architecture.

```mermaid
classDiagram
    direction TB

    %% ── application/config ───────────────────────────────────────────────────

    class Config {
        +Patterns []string
        +ConfirmMode bool
        +Speed float64
        +TimeLimit int
        +IncludeRoot bool
        +Validate() error
    }

    %% ── application/apperror ─────────────────────────────────────────────────

    class Code {
        <<enumeration>>
        CodeUnknown
        CodeInvalidConfig
        CodeProcessDiscoveryFailed
        CodeProcessNotFound
        CodeGameFailed
        CodeScoreLoadFailed
        CodeScoreSaveFailed
        CodeKillFailed
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
    }

    %% package-level: Handle(err error) error — the one place severity decides log-and-swallow vs. return

    AppError --> Code     : Code
    AppError --> Severity : Severity

    %% ── application/contract/inbound ─────────────────────────────────────────

    class RunnerPort {
        <<interface>>
        +Run(cfg config.Config) error
    }

    RunnerPort ..> Config : Run parameter

    %% ── application/contract/outbound ────────────────────────────────────────

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

    class ProcessPort {
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

    ProcessPort ..> ProcessHandlePort : Pin returns
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
    %% ValidatePatterns(patterns) error · minPatternLength = 3 · maxPatternLength = 256

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

    class GameConfig {
        +Confirm bool
        +Speed float64
        +TimeLimit int
    }

    class Session {
        -cfg GameConfig
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
        +Play(cfg, processes, highScore) PlayResult, error
        -runLoop(session, tracker) time.Time, error
        -frameLoop(session, tracker, dispatcher, termSignal, done) error
        -awaitOutstandingKills(wg, tracker, killSignals)
        -drainEventQueue(dispatcher, killSignals, done, wg) error
        -killOrReap(target, killSignals, done, wg)
    }

    %% package-level converter.go: toBounds() · toConfirmViewState() · toTargetViewState() ·
    %% toTargetViewStates() · toHUDViewState() · toStatusViewState() · toFrameViewState()

    GameService --> processKiller          : killer
    GameService --> RendererPort           : renderer
    GameService --> InputEventProviderPort : events
    GameService ..> killTracker            : creates per Play()
    GameService ..> Session                : creates (NewSession)
    GameService ..> Dispatcher             : creates (runLoop)
    GameService ..> Input                  : creates (NewInput)
    GameService ..> PlayResult             : returns
    GameService ..> FrameViewState         : builds (toFrameViewState)
    GameService ..> Config                 : accepts (Play cfg)
    GameService ..> Info                   : accepts (Play processes)
    killTracker *-- "0..*" KillFailure     : failures
    killTracker *-- "0..*" KillDud         : duds

    %% ── application/process ──────────────────────────────────────────────────

    class ProcessService {
        -proc outbound.Process
        -reporter outbound.ProcessReporter
        +NewService(proc, reporter) *Service
        +FindProcesses(patterns, includeRoot) []Info, error
        +Kill(pid, procName, protected) shouldReap bool, err error
    }

    %% package-level converter.go: toProcessInfo() · toProcessInfos()

    ProcessService --> ProcessPort         : proc
    ProcessService --> ProcessReporterPort : reporter
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
    %% toBoard() · toScoreBoard() · toScoreSummary()

    ScoreService --> ScoreStorePort    : store
    ScoreService --> ScoreReporterPort : reporter
    ScoreService ..> Board             : loads/returns
    ScoreService ..> PlayResult        : ToEntry accepts

    %% ── application (top-level) ──────────────────────────────────────────────

    class Runner {
        -processSvc *ProcessService
        -gameSvc *GameService
        -scoreSvc *ScoreService
        +NewRunner(processSvc, scoreSvc, gameSvc) *Runner
        +Run(cfg Config) error
        -logKillFailures(failures)
        -logDuds(duds)
    }

    Runner ..|> RunnerPort     : implements
    Runner *-- ProcessService  : processSvc
    Runner *-- GameService     : gameSvc
    Runner *-- ScoreService    : scoreSvc
    Runner ..> AppError        : routes every failure through apperror.Handle
    Runner ..> Info            : FindProcesses results

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
        +NewProcess() Process, error
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

    OsProcess ..|> ProcessPort         : implements
    OsProcess ..> OsProcessHandle      : Pin creates
    OsProcessHandle ..|> ProcessHandlePort : implements

    %% ── infrastructure/filescore ─────────────────────────────────────────────

    class FileScore {
        -path string
        +NewFileScore() *fileScore, error
        +Load() ScoreBoard, error
        +Save(board ScoreBoard) error
    }

    FileScore ..|> ScoreStorePort : implements

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

    %% ── internal/composition ─────────────────────────────────────────────────

    class RunnerCreator {
        +NewRunnerCreator() RunnerCreator
        +Create() Runner, error
    }

    RunnerCreator ..> OsProcess              : builds
    RunnerCreator ..> FileScore              : builds
    RunnerCreator ..> TUI                    : builds
    RunnerCreator ..> ConsoleProcessReporter : builds
    RunnerCreator ..> ConsoleScoreReporter   : builds
    RunnerCreator ..> ProcessService         : builds
    RunnerCreator ..> ScoreService           : builds
    RunnerCreator ..> GameService            : builds
    RunnerCreator ..> Runner                 : builds & returns

    %% ── entrypoint/cli ───────────────────────────────────────────────────────

    class Program {
        -creator runnerCreator
        -out, errOut io.Writer
        +NewProgram(creator) *Program
        +Run(args []string) error
        -run(args) bool, error
        -printUsageText(showUsage, err)
    }

    class programRunnerCreator {
        <<interface>>
        +Create() Runner, error
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
    Program ..> flagSplitter         : uses (run)
    Program ..> Config               : builds & validates before calling Create()
    programRunnerCreator <|.. RunnerCreator : satisfied by (structural typing)

    %% ── internal/util ────────────────────────────────────────────────────────

    class Logger {
        +NewLogger() *Logger
        +Warn(msg string)
        +Error(msg string)
    }

    %% package-level: FormatBytes(bytes int64) string (util.go)

    AppError ..> Logger : Handle() calls Warn/Error directly — a deliberate, unported exception
```

---

## Package dependency graph

```
cmd/pidshooter
    ├─ internal/composition                   (RunnerCreator, NewRunnerCreator)
    └─ entrypoint/cli                         (Program, NewProgram, ArgumentError)

internal/composition
    ├─ application                            (Runner, NewRunner)
    ├─ application/contract/inbound           (Runner interface)
    ├─ application/game                       (NewService)
    ├─ application/process                    (NewService)
    ├─ application/score                      (NewService)
    ├─ infrastructure/console                 (NewScoreReporter, NewProcessReporter)
    ├─ infrastructure/filescore                (NewFileScore)
    ├─ infrastructure/osprocess                (NewProcess)
    └─ infrastructure/tcellui                  (NewTUI)

entrypoint/cli
    ├─ application/apperror                   (Handle)
    ├─ application/config                     (Config, Validate)
    └─ application/contract/inbound           (Runner interface, for the runnerCreator return type)

application  (top-level Runner, wires the three services)
    ├─ application/apperror
    ├─ application/config
    ├─ application/game
    ├─ application/process
    └─ application/score

application/config
    ├─ application/apperror
    ├─ core/game                              (TimeLimit validation range)
    ├─ core/movement                          (MinSpeed, MaxSpeed)
    └─ core/process                           (pattern validation)

application/contract/inbound
    └─ application/config                     (Runner.Run's parameter type)

application/contract/outbound
    └─ application/input                      (InputEventProvider's channel element type)

application/game
    ├─ application/apperror
    ├─ application/config
    ├─ application/contract/outbound          (Renderer, InputEventProvider, FrameViewState, HUDViewState, StatusViewState, ConfirmViewState)
    ├─ application/input                      (Dispatcher, NewDispatcher, EventDispatcher)
    ├─ core/game                              (Session, NewSession, Config, Target, Input, NewInput)
    ├─ core/movement                          (WindowSize)
    └─ core/process                           (Info)

application/process
    ├─ application/apperror
    ├─ application/contract/outbound          (ProcessInfo, Process, ProcessHandle, ProcessReporter)
    └─ core/process                           (Info, Find, Validate*)

application/score
    ├─ application/apperror
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, ScoreSummary, ScoreReporter)
    ├─ application/game                       (PlayResult, for ToEntry)
    └─ core/score                             (Board, Entry, NewBoard)

application/input
    └─ core/game                              (Input, Target)

application/apperror
    └─ internal/util                          (Logger)

infrastructure/tcellui
    ├─ application/contract/outbound          (WindowSize, ChromeSize, FrameViewState, Renderer, InputEventProvider)
    └─ application/input                      (EventDispatcher and its concrete event types)

infrastructure/osprocess
    └─ application/contract/outbound          (ProcessInfo, Process, ProcessHandle — satisfied structurally)

infrastructure/filescore
    ├─ application/contract/outbound          (ScoreBoard, ScoreEntry, ScoreStore, NotFoundError, CorruptedDataError)
    └─ infrastructure/fsutil                  (ConfigDir, WriteFileAtomic)

infrastructure/console
    ├─ application/contract/outbound          (ScoreSummary, ScoreReporter, ProcessReporter)
    └─ internal/util                          (FormatBytes — score_reporter.go only)

core/game
    ├─ core/movement                          (Bounds, Motion, Throttle, WindowSize, ChromeSize)
    └─ core/process                           (Info)

core/* imports nothing from application, infrastructure, entrypoint, or composition — confirmed
zero outward edges from core/game, core/movement, core/process, or core/score, not even to
internal/util.

application/game, application/process, and application/score share no imports of one another
(application/score depends on application/game only for the PlayResult *type*, used by its own
ToEntry converter — it still never touches application/process). `application` (top-level Runner)
is the only place all three application sub-packages are wired together, and it does so with
already-constructed collaborators (NewRunner(processSvc, scoreSvc, gameSvc)) rather than raw ports.

application/game depends on application/process only through the unexported `processKiller`
interface it declares itself — process.Service satisfies it structurally, with no import of
application/process required in application/game.

internal/composition is the only package that imports every infrastructure adapter and every
application service directly — that's its entire job. main.go imports only composition and cli.

Infrastructure packages satisfy contract interfaces via Go structural typing — no import of
application/contract required beyond the shared data/port types.
```

---

## Call flow

```
main()
└─ run()
     ├─ composition.NewRunnerCreator()          → composition.RunnerCreator{}
     └─ cli.NewProgram(creator)                  → *cli.Program
          └─ program.Run(os.Args[1:])
               └─ p.run(args)
                    ├─ help(args) / flagSplitter.split(args) / flag.FlagSet.Parse   [no cobra]
                    ├─ config.Config{Patterns, ConfirmMode, Speed, TimeLimit, IncludeRoot}
                    ├─ cfg.Validate()
                    └─ creator.Create()                                    [composition.RunnerCreator.Create()]
                         ├─ osprocess.NewProcess()                    → outbound.Process
                         ├─ filescore.NewFileScore()                  → outbound.ScoreStore  (~/.config/pidshooter/highscores.json)
                         ├─ tcell.NewScreen()
                         ├─ tcellui.NewTUI(screen)                    → *tcellui.TUI  (outbound.Renderer)
                         ├─ process.NewService(proc, console.NewProcessReporter())  → *process.Service
                         ├─ score.NewService(store, console.NewScoreReporter())     → *score.Service
                         ├─ game.NewService(processSvc, ui, ui.InputEvents())       → *game.Service   (processSvc satisfies game's processKiller interface; ui.InputEvents() shares ui's poller)
                         └─ application.NewRunner(processSvc, scoreSvc, gameSvc)    → *application.Runner
                    └─ runner.Run(cfg)     [Runner implements inbound.Runner — every step below routes failures through apperror.Handle]
                         ├─ cfg.Validate()
                         ├─ processSvc.FindProcesses(patterns, includeRoot)
                         │     ├─ proc.Discover()                          [outbound.Process → osprocess]
                         │     ├─ process.Find(infos, patterns, ownPID, ownUID, includeRoot)   → []core/process.Info
                         │     ├─ process.ValidateProcesses(matches)
                         │     └─ reporter.Report(len(matches), patterns)  [outbound.ProcessReporter → console]
                         ├─ scoreSvc.LoadScoreBoard()                 [ScoreStore → filescore]  → board, highScore, loadErr
                         ├─ gameSvc.Play(cfg, processes, highScore)
                         │     ├─ game.NewSession(processes, gameConfig)   → *core/game.Session
                         │     ├─ newKillTracker(highScore)
                         │     └─ runLoop(session, tracker)
                         │          ├─ renderer.Init()                         [Renderer → tcellui: screen.Init + poller.poll goroutine]
                         │          ├─ session.Start(bounds)                   → pending → running, spawns targets
                         │          ├─ registerTermSignalWatcher()             [signal.NotifyContext SIGINT/SIGTERM/SIGTSTP → a channel, no goroutine of ours]
                         │          ├─ input.NewDispatcher(game.NewInput(session))
                         │          └─ frameLoop(session, tracker, dispatcher, termSignal, done) ─────────────────────────┐
                         │               ├─ applyKillSignals(tracker, killSignals)   drains buffered async kill results  │
                         │               ├─ drainEventQueue(dispatcher, killSignals, done, wg)                           │
                         │               │     ├─ events.Events()              [InputEventProvider → tcellui.inputEvents]│
                         │               │     ├─ dispatcher.Dispatch(ev)      → input.On*(...) → *Target or nil        │
                         │               │     └─ hit: target.FireShot(); go killOrReap(target, ...)                    │
                         │               │           └─ killer.Kill(pid, name, protected)   [processKiller → process.Service]
                         │               │                 ├─ proc.Pin(pid)             → ProcessHandle  [pidfd-pinned on Linux 5.3+]
                         │               │                 ├─ proc.LookupName(pid)      [re-verify — the pinned handle survives PID reuse]
                         │               │                 ├─ process.ValidateName(expected, actual)
                         │               │                 ├─ handle.Kill()              [outbound.ProcessHandle → osprocess]
                         │               │                 └─ killSignals <- killSignal{target, shouldReap, err}        │
                         │               ├─ session.Update(window)             → timer.Expired / roster.move + target bounce
                         │               ├─ renderer.Render(toFrameViewState(session, tracker))  [FrameViewState snapshot → tcellui]
                         │               └─ select { ticker.C | termSignal → session.Stop() }   — same goroutine, no concurrent Session access
                         │          └─ (loop exits) awaitOutstandingKills(wg, tracker, killSignals)   bounded by killGracePeriod (default 5s)
                         │          └─ renderer.Cleanup()  [deferred]
                         │     └─ returns PlayResult{Duration, LowestSpeed, Kills, FreedMem, KillFailures, Duds}
                         ├─ logKillFailures(result.KillFailures) / logDuds(result.Duds)     [via apperror.Handle]
                         ├─ score.ToEntry(result, cfg.TimeLimit)                → core/score.Entry
                         ├─ scoreSvc.RecordScore(board, entry, loadErr)
                         │     ├─ board.Add(entry)              [rejects non-scores: Entry.isScore() — Kills>0 && FreedMem>=0]
                         │     └─ store.Save(...)               [ScoreStore → filescore, atomic write via fsutil]
                         └─ scoreSvc.ReportResults(duration, kills, duds, freedMem, board)
                               └─ reporter.Report(summary)      [ScoreReporter → console]
```

---

## Exported surface per package

| Package | Exported identifiers |
|---|---|
| `application/apperror` | `Code`, `CodeUnknown`, `CodeInvalidConfig`, `CodeProcessDiscoveryFailed`, `CodeProcessNotFound`, `CodeGameFailed`, `CodeScoreLoadFailed`, `CodeScoreSaveFailed`, `CodeKillFailed`, `Severity`, `SeverityFatal`, `SeverityError`, `SeverityWarning`, `SeverityUnknown`, `Error`, `NewError`, `Handle` |
| `application/config` | `Config` |
| `application/contract/inbound` | `Runner` |
| `application/contract/outbound` | `NotFoundError`, `CorruptedDataError`, `InputEventProvider`, `ProcessInfo`, `ProcessHandle`, `Process`, `ProcessReporter`, `ScoreEntry`, `ScoreBoard`, `ScoreStore`, `ScoreSummary`, `ScoreReporter`, `FrameViewState`, `TargetViewState`, `HUDViewState`, `StatusViewState`, `ConfirmViewState`, `WindowSize`, `ChromeSize`, `Renderer` |
| `application/input` | `Dispatcher`, `NewDispatcher`, `EventDispatcher`, `ClickEvent`, `QuitEvent`, `ConfirmEvent`, `SpeedEvent` |
| `application/game` | `Service`, `NewService`, `PlayResult`, `KillFailure`, `KillDud` |
| `application/process` | `Service`, `NewService` |
| `application/score` | `Service`, `NewService`, `ToEntry` |
| `application` | `Runner`, `NewRunner` |
| `composition` | `RunnerCreator`, `NewRunnerCreator` |
| `core/game` | `Config`, `Session`, `NewSession`, `Target`, `NewTarget`, `State`, `Alive`, `Killing`, `Fleeing`, `Dead`, `AnimationDuration`, `Input`, `NewInput`, `ValidateTimeLimit` |
| `core/movement` | `Bounds`, `NewBounds`, `WindowSize`, `ChromeSize`, `Vector`, `Motion`, `NewMotion`, `Throttle`, `NewThrottle`, `MinSpeed`, `MaxSpeed`, `ValidateSpeed` |
| `core/process` | `Info`, `NewInfo`, `Find`, `ValidateProcesses`, `ValidateName`, `ValidatePatterns` |
| `core/score` | `Board`, `NewBoard`, `Entry` |
| `entrypoint/cli` | `Program`, `NewProgram`, `ArgumentError` |
| `infrastructure/tcellui` | `TUI`, `NewTUI` |
| `infrastructure/osprocess` | `NewProcess` (the `process` type itself is unexported — callers only ever see the `outbound.Process` it returns) |
| `infrastructure/filescore` | `NewFileScore` (the `fileScore` type itself is unexported) |
| `infrastructure/console` | `ScoreReporter`, `NewScoreReporter`, `ProcessReporter`, `NewProcessReporter` |
| `internal/util` | `Logger`, `NewLogger`, `FormatBytes` |
