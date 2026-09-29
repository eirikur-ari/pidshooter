# pidshooter — Improvement Findings

Everything still genuinely open, deferred to already-planned future work. Nothing resolved, settled, or accepted-as-permanent is tracked here — see `project_pidshooter_leftovers_review` memory for that history if needed.

---

### 1. No `--list`/`--dry-run` preview mode

`config.Result.Mode` (`ModeGame`/`ModeYolo`/`ModeList`) already resolves through the config pipeline, but nothing reads `cfg.Mode` anywhere yet — no list-only or dry-run behavior is wired up. Pattern matching is also undocumented case-insensitive substring matching, with no way to preview what a pattern will match before entering the game screen where a click is fatal.

**Deferred to:** the planned list/yolo CLI-modes work (`Mode` already anticipates this).

### 2. Port DTOs live in `application/contract`, not just interfaces

A "port" in ports-and-adapters terms should be the interface alone — the contract an adapter implements or calls. Today `application/contract/outbound`/`inbound` hold both the interfaces (`Renderer`, `ScoreStore`, `ConfigStore`, `Runner`, …) *and* the data types those methods are expressed over (`FrameViewState`, `ProcessInfo`, `ScoreBoard`, `ConfigStoreResult`, `GameConfig`/`ProcessConfig`, `Mode`, `RunRequest`, …). The plan is to relocate all of these DTOs into the application layer proper (e.g. `application/config`, `application/game`), leaving `contract/outbound`/`contract/inbound` holding only interfaces whose method signatures reference the now-relocated types.

This isn't a boundary violation as-is — the DTOs are already owned by the application core, not by an adapter, so an application-layer package referencing them today is the core using its own vocabulary, not reaching across a layer it shouldn't. It's a tidiness/ownership move, not a bug fix.

**Why deferred rather than attempted piecemeal:** moving one type at a time hits real Go import cycles. E.g. `application/config` already imports `outbound` for `ConfigStore`/`ConfigStoreResult`/`GameConfig`/`NotFoundError`; moving just `Mode` out to `config` would require `outbound.ConfigStoreResult` (which embeds `Mode`) to import `config` right back. The DTOs form an interdependent cluster per port (`ConfigStoreResult`+`GameConfig`+`ProcessConfig`+`Mode` for `ConfigStore`; similarly for the other ports) — the whole cluster for a given port has to move together, and it's not yet decided whether the *interface* itself (`ConfigStore`, etc.) moves into the application package too, leaving `contract/outbound` with nothing for that port at all, or whether the interface stays in `contract` while only its DTOs relocate.

**How to apply when picked up:** map the full type/interface dependency graph within `contract/outbound` and `contract/inbound` first — which DTOs are only referenced by one port's interface (movable independently) versus shared across several (move together). Treat it as its own scoped piece of work, not folded into an unrelated fix. `config.Mode` was already given a local, decoupled definition in `application/config` (2026-09-28) as an interim step, matching the existing `GameResult`/`ProcessResult` pattern — not yet the full relocation, just one type that happened to be low-risk to decouple early.
