# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `--version` flag that prints the version of pidshooter.
- `--help` for each command, with the config file and the default values of `--speed` and `--time` documented in the help of `play`.
- A suggestion of the commands an unknown command resembles (`did you mean play?`).
- Exit status `2` when the command line is wrong (an unknown flag or command, a missing pattern, or a value that fails validation), with the usage text printed to standard error.

### Changed
- **Breaking:** the game now runs through the `play` command: `pidshooter play <pattern>...` instead of `pidshooter <pattern>...`.
- A missing pattern is now reported before anything is set up, so nothing is written to disk.

## [0.2.0] - 2026-10-14

### Added
- A config file at `~/.config/pidshooter/config.yaml` (or `$XDG_CONFIG_HOME/pidshooter/config.yaml`) that persists defaults for `--confirm`, `--speed`, `--time` and `--include-root`. A flag overrides it for that run, and an invalid persisted value is dropped with a warning.
- `--include-root` to also target root-owned processes.
- `--i-am-root` to allow running as root, which is refused by default.
- `--` to end flag processing, so a pattern can start with `-`.
- Duds: a target that exits before it is clicked flees instead of counting as a kill. Duds are shown in the post-game summary and the score table.
- The process is re-verified just before the `SIGKILL` is sent, so a recycled PID is never killed.

### Changed
- By default only processes owned by the current user are targets.
- A search pattern must be 3 to 256 characters (was up to 256).
- `--speed` accepts `0.5` to `5.0` (was `0.1` to `5.0`).
- `--time` accepts `1` to `300`, and `0` still disables the limit (was any value from `0`).
- Finding no matching process is now an error with a non-zero exit status (was a message and exit status `0`).
- Zombie processes are no longer targets.

## [0.1.0] - 2026-06-21

### Added
- Initial proof of concept: processes matching the given patterns bounce around the terminal as labels, and clicking one sends it `SIGKILL`.
- `--confirm` to ask before each kill, `--speed=N` and `--time=N`.
- `+` and `-` to change the speed during play, and `q` or Escape to quit.
- High scores stored in `~/.config/pidshooter/highscores.json`.

[Unreleased]: https://github.com/eirikur-ari/pidshooter/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/eirikur-ari/pidshooter/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/eirikur-ari/pidshooter/releases/tag/v0.1.0
