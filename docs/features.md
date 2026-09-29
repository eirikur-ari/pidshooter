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
By default, only processes owned by the current user are valid targets — a matching process owned by another user (including root-owned processes) is silently excluded from results, not just refused at kill time. Pass `--include-root` to also include root-owned processes as targets. This only matters for a non-root player: running pidshooter itself as root already sees and can target every process regardless of this flag (see "Running as root" below for whether it's even allowed to start).

```
pidshooter sshd --include-root
```

### Running as root and `--i-am-root`
Running pidshooter itself as root is refused by default — as root, every process on the machine (not just root-owned ones) becomes a one-click `SIGKILL` target, which is a worse default for a click-to-kill game than requiring an explicit acknowledgment. Pass `--i-am-root` to allow it anyway. This is a per-invocation flag only — it is never read from or written to the config file described below, so it can't be silently left on.

```
sudo pidshooter sshd --i-am-root
```

### Confirm mode (`--confirm`)
Before a kill is executed, the status bar prompts for confirmation with the target's PID and name. The player responds with `y`/`n` or cancels with `q`. Without this flag kills are immediate on click. Configurable via the config file (see below).

```
pidshooter sleep --confirm
```

### Speed (`--speed=N`)
Sets the initial movement speed multiplier. Range: `0.5`–`5.0`, default `2.0`. Can also be adjusted live during the game with `+`/`-`. Expects a plain decimal number. Configurable via the config file (see below).

```
pidshooter node --speed=1.5
```

### Time limit (`--time=N`)
Ends the game automatically after N seconds. `--time=0` disables the limit. Range: `1`–`300` (5 minutes), default `30`. The remaining time is shown in the status bar during play. Expects a plain decimal number. Configurable via the config file (see below).

```
pidshooter sleep --time=60
pidshooter node --time=0
```

### Help
`--help` or `-h` prints usage and exits. Running with no arguments also prints usage.

### Config file
Defaults for `--confirm`, `--speed`, `--time`, and `--include-root` can be persisted at `~/.config/pidshooter/config.yaml` (or `$XDG_CONFIG_HOME/pidshooter/config.yaml`, if set), so they don't need to be typed on every run. Values are resolved in priority order: a hardcoded application default, then the config file if present, then a flag passed on the command line — a flag always wins for that run. A persisted value that fails validation (e.g. a corrupted or hand-edited `speed: 99`) is dropped with a warning rather than failing the whole run; the rest of the file's values still apply.

```yaml
game:
  confirm_mode: true
  speed: 3.0
  time_limit: 60
process:
  include_root: false
```

`--i-am-root` is deliberately **not** part of this file — see "Running as root" above.

---

## Startup

Before entering the game loop the application:
1. Searches for processes matching the given patterns.
2. If no processes are found, prints `error: no processes found` to stderr and exits with a non-zero status.
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
Left-clicking on an entity's label attempts to send `SIGKILL` to that PID. Before the signal is sent, the process's identity is re-verified against the same PID pinned at click time — this closes the window where the OS could have already recycled that PID to an unrelated process between when it was first discovered and when you clicked it. A kill animation then plays over 36 ticks (about 1.8 seconds at 20 FPS), cycling through `💥` → `✦ KILLED ✦` → `· · ·` → `·` → blank, after which the entity transitions to dead and stops rendering.

### Fleeing (duds)
If the re-verification above finds the process already gone — it exited on its own between being discovered and being clicked — the target isn't killed; it *flees* instead. A flee animation plays over the same 36 ticks, cycling through `🏃💨` → `↝ RAN AWAY ↝` → `· · ·` → `·` → blank, after which the target disappears the same way a killed one does. A fled target counts as a **dud**: it doesn't add to your kill count, but it is tracked and shown in the post-game summary and score table.

### Visual style
- Alive targets are rendered in **green bold** text.
- Targets playing the kill animation are rendered in **red bold** text.
- Targets playing the flee animation are rendered in **orange bold** text.
- The status bar uses black text on a white background.

### Terminal resize
Terminal resize events are handled: the screen is synchronised and the new dimensions are picked up on the next game tick.

### Game loop
The game runs at 20 FPS. Each tick drains pending input events, updates entity positions and states, then renders a frame.

### Win condition
The game ends automatically when every entity has reached the dead state — whether by being killed or by fleeing. It also ends when the time limit expires, or when the player quits.

### Quit
`q`, `Q`, or Escape quits at any time. During a confirm prompt, `q` cancels the prompt instead of quitting — a second `q` (outside a prompt) exits. `Ctrl-C`, `Ctrl-Z`, and OS signals `SIGINT`/`SIGTERM`/`SIGTSTP` also terminate the game cleanly. If a kill was still being verified at the moment you quit, the game waits briefly (up to 5 seconds) for it to resolve before exiting, so it isn't lost from your final score.

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
Entries are ranked by kills descending. Ties are broken, in order, by duds descending (more duds means more session activity on an equal kill count), then speed descending, then duration ascending (faster clears win), then freed memory descending. The top 10 are retained; lower scores are dropped on save. A session with zero kills is never recorded, even if it's the only thing played that day — an instant quit doesn't pollute the leaderboard.

### Score record
Each entry stores: kills, duds, freed memory (bytes), speed multiplier, time limit setting, actual game duration (seconds), and date.

### Post-game output
After the game ends the terminal is restored and a summary is printed:

```
  Game Over! Kills: 3 | Duds: 1 | Freed: 12.4 MB | Time: 18.3s
  🏆 New high score!
```

The "New high score!" line appears when the session's kill count is greater than zero and equals or exceeds the previous top score (a tie also qualifies).

Followed by the full leaderboard table showing rank, kills, duds, speed, duration, freed memory, and date for all stored entries.

---

## Signal handling

The game registers handlers for `SIGINT`, `SIGTERM`, and `SIGTSTP`. Any of these signals triggers a clean shutdown: the game loop exits, the terminal is restored, and scores are saved before the process exits.

---

## Process discovery

Processes are discovered by running `ps` and parsing its output:
- On **macOS** the command is `ps -eo uid,pid,rss,stat,ucomm` (`ucomm` returns the kernel-owned short executable name, which — unlike the process's own `argv[0]` — a process can't spoof).
- On **Linux and other platforms** the command is `ps -eo uid,pid,rss,stat,comm`.

The `uid` column is used for the ownership/`--include-root` filtering described above; the `stat` column is used to detect and skip zombie processes, which exist as a PID slot but can no longer be usefully signaled. Resident Set Size (RSS) is reported in kilobytes by `ps` and converted to bytes internally. The current process and any process with PID ≤ 1 are always filtered out.
