# Requirements: First Person PID Shoot

## Intent Analysis

- **User Request**: Create a shell program that lists process IDs with process names from fuzzy matching input and has them fly around the terminal, with the goal for the user to point and click the PID to kill it.
- **Request Type**: New Project (greenfield)
- **Scope Estimate**: Single Component — one self-contained CLI application
- **Complexity Estimate**: Moderate — terminal UI with mouse tracking, animation loop, process management

---

## Functional Requirements

### FR-01: Process Search with Fuzzy Matching
- The program accepts a search term as a command-line argument
- Processes are matched using simple substring matching against process names (e.g., "fire" matches "firefox")
- All matching processes (PID + name) are collected at startup

### FR-02: Animated Terminal Display
- Matched processes are displayed as floating labels (PID + process name) that move around the terminal
- Each process label moves in a random direction, bouncing off terminal boundaries
- The display updates at a smooth frame rate (targeting ~15-30 FPS)
- Terminal resizing is handled gracefully

### FR-03: Mouse Click to Kill
- The program enables terminal mouse tracking (xterm mouse protocol)
- The user can click on a flying process label to "shoot" it
- Click detection identifies which process label (if any) was under the mouse cursor

### FR-04: Kill Signal Behavior
- When a process is shot, SIGKILL (signal 9) is sent immediately
- Configurable confirmation mode via `--confirm` flag:
  - Without `--confirm`: kill immediately on click
  - With `--confirm`: show a confirmation prompt before killing

### FR-05: Kill Animation and Continuation
- After a successful kill, a "kill confirmed" animation is displayed at the process's last position
- The killed process is then removed from the screen
- The program continues running with remaining processes
- The program exits when all processes are killed or the user presses 'q'/Escape

### FR-06: Program Lifecycle
- Start: Parse arguments, search processes, display matched PIDs
- Running: Animate processes, handle mouse clicks, process kills
- Exit: User presses 'q' or Escape, or all processes are eliminated
- On exit: Restore terminal state (disable mouse tracking, restore cursor, clear alternate screen)

---

## Non-Functional Requirements

### NFR-01: Platform Compatibility
- Must run on any POSIX-compatible system (Linux, macOS, BSDs)
- Relies only on POSIX process APIs and standard terminal capabilities

### NFR-02: Language and Build
- Written in Go
- Compiled as a single static binary with no runtime dependencies
- Uses a terminal UI library for rendering and mouse support

### NFR-03: Performance
- Smooth animation at 15-30 FPS with minimal CPU usage
- Responsive mouse click detection (< 100ms latency)

### NFR-04: Terminal Compatibility
- Works in terminals supporting xterm mouse protocol (most modern terminals)
- Graceful degradation if mouse support is unavailable (show error and exit)
- Properly restores terminal state on exit (normal exit, Ctrl+C, panic)

### NFR-05: Security (SECURITY extension enabled)
- Input validation on command-line arguments (FR-05 input validation per SECURITY-05)
- No hardcoded credentials (SECURITY-12 — N/A for this app, but enforced as practice)
- Proper error handling with fail-safe defaults (SECURITY-15)
- Dependency pinning with go.sum lock file (SECURITY-10)

---

## User Interface

### Display Layout
```
+------------------------------------------+
|                                           |
|     [1234 firefox]  -->                   |
|                                           |
|          <--  [5678 chrome]               |
|                                           |
|   [9012 node]  -->                        |
|                                           |
|              [3456 python] -->            |
|                                           |
|  Press 'q' to quit | Click PID to kill   |
+------------------------------------------+
```

### Usage
```
pidshoot <search-term> [--confirm]
```

---

## Constraints

- Must not kill the pidshoot process itself
- Must not kill PID 1 (init/systemd)
- Should warn if no processes match the search term
- Should handle permission errors gracefully (attempting to kill processes owned by other users)
