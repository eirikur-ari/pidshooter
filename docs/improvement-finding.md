# pidshooter — Improvement Findings

Everything still genuinely open, deferred to already-planned future work. Nothing resolved, settled, or accepted-as-permanent is tracked here — see `project_pidshooter_leftovers_review` memory for that history if needed.

---

### 1. No `--list`/`--dry-run` preview mode

`config.Result.Mode` (`ModeGame`/`ModeYolo`/`ModeList`) already resolves through the config pipeline, but nothing reads `cfg.Mode` anywhere yet — no list-only or dry-run behavior is wired up. Pattern matching is also undocumented case-insensitive substring matching, with no way to preview what a pattern will match before entering the game screen where a click is fatal.

**Deferred to:** the planned list/yolo CLI-modes work (`Mode` already anticipates this).

---

### 2. `Board.highScore` is hidden state that lags one `Add` behind

`score.Board` keeps a `highScore` field that `Add` sets to the best kill count *before* the new entry is inserted, so `IsNewHighScore(kills)` only makes sense when asked right after that entry was added. `application/score.Service.RecordScore` adds the entry first and `ReportResults` asks the question afterwards, which is the only reason the snapshot exists. The method therefore answers "is this above the previous high score?", not "is this a record on the board right now?" — e.g. on a board holding 3 and 7, `IsNewHighScore(5)` is true.

Planned change: drop the `highScore` field and let `Board.IsNewHighScore(kills)` compare against the current best (`kills > 0 && kills > killScore()`), asked *before* the entry is added. `RecordScore` (the score service owns the rule, not the runner) evaluates it before `board.Add(entry)` and returns it with the error — `RecordScore(board, entry, err) (newHighScore bool, saveErr error)`. The runner only forwards the flag into `ReportResults`, and `toScoreSummary` takes it as a parameter instead of querying the board afterwards. A service field remembering the flag between the two calls was rejected, since it recreates the same hidden coupling.

Touches `core/score/board.go` (field, `Add`, `NewBoard` seeding, tests built around the snapshot), `application/score` (`RecordScore`, `ReportResults`, `toScoreSummary` and their tests) and `application/runner/service.go`.

**Deferred to:** a follow-up after the testing audit; the current behavior is correct for the runner's call order.

---

### 3. A negative flag value given as a separate argument reports a misleading error

`flagSplitter.split` treats any argument starting with `-` (other than a lone `-`) as a flag. `--speed -3` therefore fails with "flag needs an argument: -speed" even though a value was supplied, while `--speed=-3` is accepted by the splitter and rejected later by range validation. The `flag` package itself would take `-3` as the value. Both inputs are out of range today, so the only effect is the wrong message, but a future flag with a legitimately negative value would be unusable in the separate-argument form.

**Deferred to:** whenever a flag that accepts negative values is added, or the splitter is next reworked.

---

### 4. A failed play session discards the kills already made

When `game.Service.Play` fails mid-session (for example the input event channel closes), `playSession.frameLoop` stops the session and returns the error without waiting for in-flight kills, and `Play` returns an empty `PlayResult{}` with a `CodeGameFailed` error. Processes that were already killed (or whose kill is still in flight) are therefore reported nowhere: no kills, freed memory, duds or failures reach the runner, the score board or the log, even though the processes are really gone.

Waiting for in-flight kills before returning the error would not help on its own, since the result is dropped either way. The fix is to return the partial `PlayResult` alongside the error and let the runner decide what to log and whether to record a score for an aborted session, which also needs a decision on whether an aborted session counts as a score at all.

Touches `application/game/service.go` (`Play`), `application/game/play_session.go` (`run`, `frameLoop`, `result`) and `application/runner/service.go`.

**Deferred to:** whenever aborted-session handling is designed; today the only abort path is an input source that died unexpectedly.
