# pidshooter — Improvement Findings

Everything still genuinely open, deferred to already-planned future work. Nothing resolved, settled, or accepted-as-permanent is tracked here — see `project_pidshooter_leftovers_review` memory for that history if needed.

---

### 1. No `--list`/`--dry-run` preview mode

`config.Result.Mode` (`ModeGame`/`ModeYolo`/`ModeList`) already resolves through the config pipeline, but nothing reads `cfg.Mode` anywhere yet — no list-only or dry-run behavior is wired up. Pattern matching is also undocumented case-insensitive substring matching, with no way to preview what a pattern will match before entering the game screen where a click is fatal.

**Deferred to:** the planned list/yolo CLI-modes work (`Mode` already anticipates this).
