# Requirements Verification Questions

Please answer the following questions to help clarify the requirements for the "First Person PID Shoot" terminal program.

## Question 1
What shell/language should this program be written in?

A) Bash script (using tput/ncurses-style escape codes)
B) Python script (using curses library)
C) Go (compiled binary with terminal UI library)
D) Rust (compiled binary with terminal UI library)
X) Other (please describe after [Answer]: tag below)

[Answer]: C

## Question 2
How should the user "click" a flying PID to kill it?

A) Mouse click support (actual terminal mouse tracking)
B) Keyboard aim — use arrow keys to move a crosshair/cursor and press Enter/Space to "shoot"
C) Type the PID number to target it
X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 3
What kill signal should be sent when the user "shoots" a process?

A) SIGTERM (graceful termination, signal 15)
B) SIGKILL (force kill, signal 9)
C) Let the user choose between SIGTERM and SIGKILL
D) Start with SIGTERM, escalate to SIGKILL if process doesn't die
X) Other (please describe after [Answer]: tag below)

[Answer]: B

## Question 4
How should fuzzy matching work for the initial process search?

A) Simple substring match (e.g., "fire" matches "firefox")
B) Fuzzy matching similar to fzf (partial character matching with scoring)
C) Regex pattern matching
X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question 5
What should happen after a process is killed?

A) Remove it from the screen and continue with remaining processes
B) Show a "kill confirmed" animation, then remove it
C) End the program after one kill
D) Keep running until all matched processes are killed or user quits
X) Other (please describe after [Answer]: tag below)

[Answer]: B

## Question 6
Should the program require confirmation before actually killing a process?

A) Yes — show a confirmation prompt before sending the kill signal
B) No — kill immediately on click/shoot (it's a game, after all)
C) Make it configurable with a flag (e.g., --confirm)
X) Other (please describe after [Answer]: tag below)

[Answer]: C

## Question 7
What platforms should this run on?

A) Linux only
B) Linux and macOS
C) Any POSIX-compatible system
X) Other (please describe after [Answer]: tag below)

[Answer]: C

## Question: Security Extensions
Should security extension rules be enforced for this project?

A) Yes — enforce all SECURITY rules as blocking constraints (recommended for production-grade applications)
B) No — skip all SECURITY rules (suitable for PoCs, prototypes, and experimental projects)
X) Other (please describe after [Answer]: tag below)

[Answer]: A

## Question: Property-Based Testing Extension
Should property-based testing (PBT) rules be enforced for this project?

A) Yes — enforce all PBT rules as blocking constraints (recommended for projects with business logic, data transformations, serialization, or stateful components)
B) Partial — enforce PBT rules only for pure functions and serialization round-trips (suitable for projects with limited algorithmic complexity)
C) No — skip all PBT rules (suitable for simple CRUD applications, UI-only projects, or thin integration layers with no significant business logic)
X) Other (please describe after [Answer]: tag below)

[Answer]: C
