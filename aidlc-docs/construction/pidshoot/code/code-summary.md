# Code Summary: pidshoot

## File Structure

| File | Lines | Responsibility |
|------|-------|---------------|
| `main.go` | ~95 | Entry point, argument parsing, signal handling, usage display |
| `process.go` | ~90 | Process discovery via `ps`, substring matching, PID filtering |
| `entity.go` | ~145 | Entity struct, movement/physics, bounce logic, hit detection, kill animation |
| `game.go` | ~230 | Game loop, tcell screen management, rendering, mouse/key events, kill logic |
| `go.mod` | ~14 | Module definition with pinned dependencies |

## Architecture

```
main.go (entry point)
  ├── parseArgs() → pattern, confirmMode
  ├── FindProcesses(pattern) → []ProcessInfo    [process.go]
  ├── NewGame(processes, confirmMode)            [game.go]
  │     ├── Init() → tcell screen + mouse
  │     ├── PopulateEntities() → []*Entity      [entity.go]
  │     └── Run() → game loop at 20 FPS
  └── signal handler → clean exit
```

## Key Design Decisions

1. **tcell/v2**: Mature terminal library with cross-platform mouse support
2. **Float64 positions**: Smooth sub-cell movement with integer rendering
3. **Event channel pattern**: Non-blocking event polling alongside ticker-based frame updates
4. **POSIX ps command**: Cross-platform process listing (Linux, macOS, BSDs)
5. **Fail-safe cleanup**: defer + signal handler ensures terminal state restored

## Security Compliance

| Rule | Status | Notes |
|------|--------|-------|
| SECURITY-05 | ✅ Compliant | Pattern validated: non-empty, max 256 chars, unknown flags rejected |
| SECURITY-09 | ✅ Compliant | No default credentials, generic error messages |
| SECURITY-10 | ✅ Compliant | Exact versions in go.mod, go.sum committed |
| SECURITY-15 | ✅ Compliant | Global error handler in run(), defer Cleanup(), signal handler |
| SECURITY-01–04 | N/A | No data stores, network intermediaries, or web endpoints |
| SECURITY-06–08 | N/A | No IAM policies, network config, or auth endpoints |
| SECURITY-11 | ✅ Compliant | Security logic isolated (input validation in parseArgs), process kill safety checks |
| SECURITY-12 | N/A | No authentication |
| SECURITY-13 | N/A | No deserialization of untrusted data, no CDN resources |
| SECURITY-14 | N/A | CLI tool, no centralized logging/alerting infrastructure |
