# pidshooter — Package & Class Diagram

The project is split into four Go packages. The diagram maps every exported type, its fields, and its methods, plus the package-level functions. Visibility follows Go conventions: `+` = exported, `-` = unexported.

```mermaid
classDiagram
    direction TB

    %% ── package process (internal/process/process.go) ────────────────────────

    class Info {
        <<process>>
        +PID int
        +Name string
        +RSS int64
    }

    class process_pkg {
        <<process · functions>>
        +MaxPatternLength int$
        +Find(patterns []string) []Info, error
        -list() []Info, error
    }

    %% ── package score (internal/score/score.go) ──────────────────────────────

    class Entry {
        <<score>>
        +Kills int
        +FreedMem int64
        +Speed float64
        +Time int
        +Duration float64
        +Date time.Time
    }

    class Board {
        <<score>>
        +Scores []Entry
        +Save() error
        +Add(entry Entry) bool
        +HighScore() int
        +PrintScores()
    }

    class score_pkg {
        <<score · functions>>
        -filePath() string
        +Load() Board
    }

    %% ── package game / entity (internal/game/entity.go) ─────────────────────

    class EntityState {
        <<game · enumeration>>
        StateAlive
        StateKilling
        StateDead
    }

    class Entity {
        <<game · entity.go>>
        +PID int
        +Name string
        +RSS int64
        +X float64
        +Y float64
        +VelX float64
        +VelY float64
        +State EntityState
        +KillAnimFrame int
        +Label() string
        +Update(maxX int, maxY int, speed float64)
        +Contains(x int, y int) bool
        +StartKillAnim()
    }

    class entity_pkg {
        <<game · entity.go · functions>>
        +KillAnimFrames int$
        +NewEntity(pid int, name string, rss int64, maxX int, maxY int) Entity
    }

    %% ── package game / game (internal/game/game.go) ──────────────────────────

    class Game {
        <<game · game.go>>
        -screen tcell.Screen
        -entities []Entity
        -confirmMode bool
        -speed float64
        -running bool
        -confirming Entity
        -kills int
        -freedMem int64
        -timeLimit int
        -startTime time.Time
        -highScore int
        +SetHighScore(score int)
        +Kills() int
        +FreedMem() int64
        +StartTime() time.Time
        +Stop()
        +Init() error
        +PopulateEntities(processes []Info)
        +Run()
        -drainEvents(eventCh chan Event)
        -handleEvent(ev tcell.Event)
        -handleMouseClick(x int, y int)
        +HandleKeyPress(key tcell.Key, r rune)
        -update()
        -render()
        -drawStatusBar(w int, h int)
        -killEntity(e Entity)
        +Cleanup()
    }

    class game_pkg {
        <<game · game.go · functions>>
        +New(processes []Info, confirmMode bool, speed float64, timeLimit int) Game
        +FormatBytes(bytes int64) string
    }

    %% ── package main (cmd/pidshooter/main.go) ────────────────────────────────

    class main_pkg {
        <<main>>
        +main()
        -run() error
        -parseArgs(args []string) []string, bool, float64, int, error
    }

    %% ── Relationships ─────────────────────────────────────────────────────────

    %% game/entity internals
    Entity --> EntityState : has state

    %% Game owns Entities
    Game "1" *-- "0..*" Entity : owns

    %% Constructor / factory relations
    entity_pkg ..> Entity : constructs
    game_pkg ..> Game : constructs

    %% game.New takes process.Info as input
    game_pkg ..> Info : takes as input
    Game ..> Info : PopulateEntities converts to Entity

    %% score internals
    Board "1" *-- "0..*" Entry : owns
    score_pkg ..> Board : Load() produces

    %% score depends on game for FormatBytes
    Board ..> game_pkg : PrintScores calls FormatBytes

    %% process_pkg produces Info
    process_pkg ..> Info : Find() produces

    %% main orchestrates all packages
    main_pkg ..> process_pkg : calls Find()
    main_pkg ..> score_pkg : calls Load()
    main_pkg ..> game_pkg : calls New()
    main_pkg ..> Game : drives lifecycle
    main_pkg ..> Board : calls Add · Save · PrintScores
    main_pkg ..> Entry : constructs after game ends
```

---

## Package dependency graph

```
cmd/pidshooter (main)
    ├── internal/game
    │       └── internal/process   (game.PopulateEntities takes process.Info)
    ├── internal/process
    └── internal/score
            └── internal/game      (score.Board.PrintScores calls game.FormatBytes)
```

`score` importing `game` for `FormatBytes` is a design smell — score is a leaf package that should not depend on game. Moving `FormatBytes` to a shared `internal/format` package, or inlining the formatting in `score`, would break the cycle.

---

## Relationships explained

| From | To | Kind | Description |
|---|---|---|---|
| `Entity` | `EntityState` | association | Each entity holds one state value |
| `Game` | `Entity` | composition | Game owns and manages the entity slice |
| `Game` | `Info` | dependency | `PopulateEntities` converts `process.Info` records into `Entity` objects |
| `Board` | `Entry` | composition | Board owns the ordered list of score records |
| `Board` | `game_pkg` | dependency | `PrintScores` calls `game.FormatBytes` to render memory strings |
| `entity_pkg` | `Entity` | factory | `NewEntity` constructs an Entity with random position and velocity |
| `game_pkg` | `Game` | factory | `New` constructs a Game from process list and config |
| `process_pkg` | `Info` | producer | `Find` / `list` discover and return process records |
| `score_pkg` | `Board` | producer | `Load` deserialises the on-disk JSON into a Board |
| `main_pkg` | `process_pkg` | calls | `run()` calls `Find()` to get the initial process list |
| `main_pkg` | `game_pkg` | calls | `run()` calls `New()` then drives the full game lifecycle |
| `main_pkg` | `score_pkg` | calls | `run()` calls `Load()` before the game |
| `main_pkg` | `Board` | calls | `run()` calls `Add`, `Save`, `PrintScores` after the game |

---

## Call flow — `run()` in main.go

```
main()
  └─ run()
       ├─ parseArgs(os.Args[1:])               → []patterns, confirmMode, speed, timeLimit
       ├─ process.Find(patterns)               → []process.Info
       ├─ score.Load()                         → *score.Board
       ├─ game.New(processes, ...)             → *game.Game
       ├─ g.SetHighScore(board.HighScore())
       ├─ g.Init()                             creates tcell.Screen
       ├─ defer g.Cleanup()
       ├─ g.PopulateEntities(processes)        []process.Info → []game.Entity
       ├─ go signal handler → g.Stop()         goroutine: SIGINT/SIGTERM/SIGTSTP
       ├─ g.Run()  ◄── blocks at 20 FPS ──────────────────────────────────────────┐
       │     ├─ goroutine: screen.PollEvent() → eventCh                           │
       │     └─ loop while g.running:                                             │
       │           ├─ drainEvents(eventCh)                                        │
       │           │     └─ handleEvent()                                         │
       │           │           ├─ handleMouseClick() → killEntity()               │
       │           │           └─ HandleKeyPress()   → killEntity() / g.Stop()   │
       │           ├─ update()      moves entities, checks time limit / all-dead  │
       │           └─ render()      draws entities + HUD, calls FormatBytes()     │
       │                                                                          ─┘
       ├─ g.Cleanup()                          restores terminal (explicit call)
       ├─ score.Entry{Kills, FreedMem, ...}
       ├─ board.Add(entry)
       ├─ board.Save()
       └─ board.PrintScores()                  calls game.FormatBytes internally
```

---

## Key exported surface per package

| Package | Exported types | Exported functions / constructors |
|---|---|---|
| `process` | `Info` | `Find()` |
| `score` | `Entry`, `Board` | `Load()` |
| `game` | `Entity`, `EntityState`, `Game` | `New()`, `FormatBytes()`, `NewEntity()` |
| `main` | — | `main()` |
