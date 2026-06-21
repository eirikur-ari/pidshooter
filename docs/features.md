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
- Empty patterns are rejected
- Patterns longer than 256 characters are rejected
- The game itself and PID 1 are always excluded from results

### Confirm mode (`--confirm`)
Before a kill is executed, the status bar prompts for confirmation with the target's PID and name. The player responds with `y`/`n` or cancels with `q`. Without this flag kills are immediate on click.

```
pidshooter sleep --confirm
```

### Speed (`--speed=N`)
Sets the initial movement speed multiplier. Range: `0.1`–`5.0`, default `2.0`. Can also be adjusted live during the game with `+`/`-`.

```
pidshooter node --speed=1.5
```

### Time limit (`--time=N`)
Ends the game automatically after N seconds. `--time=0` disables the limit. Default is 30 seconds. The remaining time is shown in the status bar during play.

```
pidshooter sleep --time=60
pidshooter node --time=0
```

### Help
`--help` or `-h` prints usage and exits. Running with no arguments also prints usage.

---

## Gameplay

### Entity spawning
Each matched process becomes an entity — a label in the format `[PID name]` placed at a random position on the terminal. Entities start with randomised velocities: horizontal speed is stronger than vertical to keep labels readable.

### Bouncing physics
Entities bounce off all four walls continuously. The right and bottom bounds account for the label width and the status bar row respectively, so labels never clip out of view.

### Speed control
`+` increases speed by 0.5× per keypress (cap: 5.0×). `-` decreases by 0.5× (floor: 0.1×). Changes take effect immediately and apply to all entities uniformly.

### Killing
Clicking on an entity's label sends `SIGKILL` to that PID. A kill animation plays over 12 frames cycling through `💥 → ✦ KILLED ✦ → · · · → · → (blank)`, after which the entity transitions to dead and stops rendering.

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
- Number of alive targets remaining
- Current speed multiplier
- Time remaining (if a limit is set)
- Control hints (`Click to kill`, `+/- speed`, `q quit`)

During a confirm prompt the status bar switches to: `Kill [PID name]? (Y)es / (N)o / (Q)uit`

---

## High score system

### Persistence
Scores are stored as JSON at `~/.config/pidshooter/highscores.json`. The directory is created automatically on first save. Up to 10 entries are kept.

### Ranking
Entries are ranked by kills descending. Ties are broken by freed memory descending. The top 10 are retained; lower scores are dropped on save.

### Score record
Each entry stores: kills, freed memory (bytes), speed multiplier, time limit setting, actual game duration (seconds), and date.

### Post-game output
After the game ends the terminal is restored and a summary is printed:

```
  Game Over! Kills: 3 | Freed: 12.4 MB | Time: 18.3s
  🏆 New high score!
```

Followed by the full leaderboard table showing rank, kills, freed memory, speed, and date for all stored entries.

---

## Signal handling

The game registers handlers for `SIGINT`, `SIGTERM`, and `SIGTSTP`. Any of these signals triggers a clean shutdown: the game loop exits, the terminal is restored, and scores are saved before the process exits.

---

## Process discovery

Processes are discovered by running `ps -eo pid,rss,comm` and parsing its output. Resident Set Size (RSS) is reported in kilobytes by `ps` and converted to bytes internally. The current process and PID 1 are always filtered out.
