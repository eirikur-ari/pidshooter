# Code Generation Plan: pidshoot

## Unit Context
- **Unit**: pidshoot (single unit - the entire application)
- **Language**: Go
- **Target**: Single compiled binary for POSIX systems
- **Dependencies**: tcell (terminal UI library with mouse support)
- **Workspace Root**: /home/eirikur/workspace/tutorial/kiro/first-person-pid-shoot

## Requirements Coverage
- FR-01: Process search with substring matching
- FR-02: Animated terminal display with bouncing labels
- FR-03: Mouse click to kill (xterm mouse tracking via tcell)
- FR-04: SIGKILL on click, configurable --confirm flag
- FR-05: Kill confirmed animation, then remove
- FR-06: Program lifecycle (start → run → exit with terminal restore)

## Code Location
- **Application Code**: Workspace root (`/home/eirikur/workspace/tutorial/kiro/first-person-pid-shoot/`)
- **Documentation**: `aidlc-docs/construction/pidshoot/code/`

---

## Generation Steps

### Step 1: Project Structure Setup
- [x] Initialize Go module (`go.mod`)
- [x] Create directory structure:
  ```
  /
  ├── main.go          # Entry point, arg parsing, program lifecycle
  ├── process.go       # Process discovery and matching
  ├── game.go          # Game loop, animation, rendering
  ├── entity.go        # Process entity (position, velocity, state)
  ├── go.mod
  └── go.sum
  ```

### Step 2: Go Module and Dependencies
- [x] Create `go.mod` with module name `github.com/eirikur/pidshoot`
- [x] Add dependency: `github.com/gdamore/tcell/v2` (terminal UI with mouse support)
- [x] Pin exact versions in go.mod

### Step 3: Process Discovery Module (`process.go`)
- [x] Implement `FindProcesses(pattern string) []ProcessInfo` — reads /proc or uses `ps` command
- [x] Implement substring matching against process names
- [x] Filter out own PID and PID 1
- [x] Return slice of `ProcessInfo{PID int, Name string}`
- [x] Input validation on pattern (max length, no empty)
- [x] Cross-platform: use `os.FindProcess` + `/proc` on Linux, `ps` on others

### Step 4: Entity Module (`entity.go`)
- [x] Define `Entity` struct: PID, Name, X, Y, VelX, VelY, Alive, KillAnimFrame
- [x] Implement `NewEntity(pid int, name string, maxX, maxY int) *Entity`
- [x] Implement `Entity.Update(maxX, maxY int)` — move + bounce off walls
- [x] Implement `Entity.Label() string` — format as `[PID name]`
- [x] Implement `Entity.Contains(x, y int) bool` — hit detection for mouse clicks
- [x] Implement `Entity.StartKillAnim()` — trigger kill animation state

### Step 5: Game Loop Module (`game.go`)
- [x] Define `Game` struct: screen, entities, confirmMode, running
- [x] Implement `NewGame(processes []ProcessInfo, confirmMode bool) *Game`
- [x] Implement `Game.Init()` — initialize tcell screen, enable mouse
- [x] Implement `Game.Run()` — main loop (poll events + update + render at ~20 FPS)
- [x] Implement `Game.handleEvent(ev tcell.Event)` — dispatch mouse/key events
- [x] Implement `Game.handleMouseClick(x, y int)` — find clicked entity, kill or confirm
- [x] Implement `Game.handleKeyPress(key tcell.Key, r rune)` — q/Escape to quit
- [x] Implement `Game.update()` — move entities, advance kill animations
- [x] Implement `Game.render()` — draw all entities + status bar
- [x] Implement `Game.kill(entity *Entity)` — send SIGKILL, start kill animation
- [x] Implement `Game.showConfirm(entity *Entity)` — show Y/N prompt for --confirm mode
- [x] Implement `Game.Cleanup()` — restore terminal state (defer in main)

### Step 6: Main Entry Point (`main.go`)
- [x] Parse command-line args: `pidshoot <pattern> [--confirm]`
- [x] Validate arguments (SECURITY-05: input validation)
- [x] Call `FindProcesses(pattern)`
- [x] Handle no-match case (print message, exit)
- [x] Create and run Game
- [x] Handle signals (SIGINT, SIGTERM) for clean exit
- [x] Global error handler (SECURITY-15: fail-safe)

### Step 7: Code Summary Documentation
- [x] Create `aidlc-docs/construction/pidshoot/code/code-summary.md`
- [x] Document file structure and responsibilities
- [x] Document security compliance

---

## Security Compliance Plan
- **SECURITY-05**: Input validation on CLI pattern argument (max 256 chars, non-empty)
- **SECURITY-09**: No default credentials, proper error messages (no internals exposed)
- **SECURITY-10**: Exact dependency versions in go.mod, go.sum committed
- **SECURITY-15**: Global error handler, fail-safe on kill errors, terminal state always restored

## Estimated Scope
- **Total Steps**: 7
- **Files to Create**: 5 source files + go.mod + 1 documentation file
- **Lines of Code**: ~400-500 Go lines
