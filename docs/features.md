# pidshooter — Feature Reference

## Overview

pidshooter is a terminal-based game that turns process management into a first-person shooter. Running processes matching a given name pattern appear as bouncing labels on the terminal screen. The player clicks on them to send `SIGKILL`.

---

## Command-line interface

### Process pattern matching
One or more substring patterns are passed as positional arguments. Any running process whose name contains any of the patterns (case-insensitive) is spawned as a target. Multiple patterns are ORed together.

```
pidshooter firefox
pidshooter chrome firefox node
```

Patterns are validated at startup:
- At least one pattern is required
- Each pattern must be between 3 and 256 characters (this also rejects empty patterns; patterns shorter than 3 characters match too broadly and risk surfacing critical system processes)
- Unknown flags (any `-` prefixed argument that is not a recognised flag) are rejected with a usage hint
- The game itself and any process with PID ≤ 1 are always excluded from results

A pattern that itself starts with `-` can't be written directly (it would be parsed as a flag); precede it with `--` to end flag processing, after which every remaining argument is treated as a pattern:

```
pidshooter -- -process-name
```

### Target ownership and `--include-root`
By default, only processes owned by the current user are valid targets — a matching process owned by another user (including root-owned processes) is silently excluded from results, not just refused at kill time. Pass `--include-root` to also include root-owned processes as targets. Running pidshooter itself as root always has full access regardless of this flag, matching normal Unix permission semantics — a non-root player has to opt in explicitly, and even then a kill attempt against a root-owned process is still subject to the OS's real permission check and can fail.

```
pidshooter sshd --include-root
```

### Confirm mode (`--confirm`)
Before a kill is executed, the status bar prompts for confirmation with the target's PID and name. The player responds with `y`/`n` or cancels with `q`. Without this flag kills are immediate on click.

```
pidshooter sleep --confirm
```

### Speed (`--speed=N`)
Sets the initial movement speed multiplier. Range: `0.5`–`5.0`, default `2.0`. Can also be adjusted live during the game with `+`/`-`. Expects a plain decimal number.

```
pidshooter node --speed=1.5
```

### Time limit (`--time=N`)
Ends the game automatically after N seconds. `--time=0` disables the limit. Range: `1`–`300` (5 minutes), default `30`. The remaining time is shown in the status bar during play. Expects a plain decimal number.

```
pidshooter sleep --time=60
pidshooter node --time=0
```

### Help
`--help` or `-h` prints usage and exits. Running with no arguments also prints usage.

---

## Startup

Before entering the game loop the application:
1. Searches for processes matching the given patterns.
2. If no processes are found, prints `No processes found matching [patterns]` and exits cleanly.
3. If processes are found, prints `Found N process(es) matching [patterns]. Starting game...` and launches the game.

---

## Gameplay

### Entity spawning
Each matched process becomes an entity — a label in the format `[PID name]` placed at a random position on the terminal. Entities start with randomised velocities: horizontal speed is stronger than vertical to keep labels readable. Specifically, horizontal velocity magnitudes are in the range 0.5–1.0, vertical in the range 0.25–0.5.

### Bouncing physics
Entities bounce off all four walls continuously. The right and bottom bounds account for the label width and the status bar row respectively, so labels never clip out of view.

### Speed control
`+` increases speed by 0.5× per keypress (cap: 5.0×). `-` decreases by 0.5× (floor: 0.5×). Changes take effect immediately and apply to all entities uniformly.

### Killing
Left-clicking on an entity's label sends `SIGKILL` to that PID. A kill animation plays over 12 frames cycling through `💥 → ✦ KILLED ✦ → · · · → · → (blank)`, after which the entity transitions to dead and stops rendering.

### Visual style
- Alive targets are rendered in **green bold** text.
- Targets playing the kill animation are rendered in **red bold** text.
- The status bar uses black text on a white background.

### Terminal resize
Terminal resize events are handled: the screen is synchronised and the new dimensions are picked up on the next game tick.

### Game loop
The game runs at 20 FPS. Each tick drains pending input events, updates entity positions and states, then renders a frame.

### Win condition
The game ends automatically when all entities have been killed (all reach the dead state). It also ends when the time limit expires, or when the player quits.

### Quit
`q`, `Q`, or Escape quits at any time. During a confirm prompt, `q` cancels the prompt instead of quitting — a second `q` (outside a prompt) exits. `Ctrl-C`, `Ctrl-Z`, and OS signals `SIGINT`/`SIGTERM`/`SIGTSTP` also terminate the game cleanly.

---

## HUD

The top row of the terminal displays three live counters:

| Position | Content | Colour |
|---|---|---|
| Left | `FREED: <memory>` | Cyan |
| Centre | `Highscore: <n>` | Purple |
| Right | `KILLS: <n>` | Yellow |

Memory is formatted as bytes, KB, MB, or GB automatically.

The bottom row is a persistent status bar showing:

```
 Targets: N | Speed: X.Xx | Time: Ns | Click to kill | +/- speed | 'q' quit
```

The `Time:` segment is omitted when no time limit is set.

During a confirm prompt the status bar switches to: `Kill [PID name]? (Y)es / (N)o / (Q)uit`

---

## High score system

### Persistence
Scores are stored as JSON at `~/.config/pidshooter/highscores.json`. The directory is created automatically on first save. Up to 10 entries are kept.

### Ranking
Entries are ranked by kills descending. Ties are broken, in order, by speed descending, then duration ascending (faster clears win), then freed memory descending. The top 10 are retained; lower scores are dropped on save.

### Score record
Each entry stores: kills, freed memory (bytes), speed multiplier, time limit setting, actual game duration (seconds), and date.

### Post-game output
After the game ends the terminal is restored and a summary is printed:

```
  Game Over! Kills: 3 | Freed: 12.4 MB | Time: 18.3s
  🏆 New high score!
```

The "New high score!" line appears when the session's kill count is greater than zero and equals or exceeds the previous top score (a tie also qualifies).

Followed by the full leaderboard table showing rank, kills, speed, duration, freed memory, and date for all stored entries.

---

## Signal handling

The game registers handlers for `SIGINT`, `SIGTERM`, and `SIGTSTP`. Any of these signals triggers a clean shutdown: the game loop exits, the terminal is restored, and scores are saved before the process exits.

---

## Process discovery

Processes are discovered by running `ps` and parsing its output:
- On **macOS** the command is `ps -ceo pid,rss,comm` (the `-c` flag returns the short executable name without the full path).
- On **Linux and other platforms** the command is `ps -eo pid,rss,comm`.

Resident Set Size (RSS) is reported in kilobytes by `ps` and converted to bytes internally. The current process and any process with PID ≤ 1 are always filtered out.
