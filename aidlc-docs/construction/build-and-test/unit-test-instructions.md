# Unit Test Execution

## Test Strategy

Given the nature of this application (terminal UI + process management), testing is split into:
- **Testable logic**: Argument parsing, process matching/filtering, entity physics, hit detection
- **Requires manual testing**: Terminal rendering, mouse interaction, actual process killing

## Run Unit Tests

### 1. Execute All Unit Tests
```bash
go test ./... -v
```

### 2. Run with Coverage
```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 3. Run with Race Detection
```bash
go test ./... -race
```

## Testable Components

### process.go — FindProcesses
| Test Case | Input | Expected |
|-----------|-------|----------|
| Empty pattern | `""` | Error: "pattern must not be empty" |
| Pattern too long | 257+ chars | Error: "exceeds maximum length" |
| Valid pattern | `"bash"` | Returns matching processes (non-empty on any system) |
| No matches | `"zzz_nonexistent_zzz"` | Empty slice, no error |
| Self-exclusion | own process name | Result doesn't contain own PID |
| PID 1 exclusion | `"init"` or `"systemd"` | Result doesn't contain PID 1 |

### entity.go — Entity Logic
| Test Case | Input | Expected |
|-----------|-------|----------|
| NewEntity bounds | maxX=80, maxY=24 | Entity position within bounds |
| Update bounce left | X=0, VelX=-1 | VelX becomes positive |
| Update bounce right | X=maxX-labelLen, VelX=1 | VelX becomes negative |
| Update bounce top | Y=0, VelY=-1 | VelY becomes positive |
| Update bounce bottom | Y=maxY-2, VelY=1 | VelY becomes negative |
| Contains hit | click on label position | Returns true |
| Contains miss | click outside label | Returns false |
| Contains dead | State=StateDead | Returns false |
| Kill animation | StartKillAnim() | State=StateKilling, frames advance |
| Kill animation end | KillAnimFrame >= 12 | State=StateDead |

### main.go — parseArgs
| Test Case | Input | Expected |
|-----------|-------|----------|
| No args | `[]` | Prints usage, exits 0 |
| Pattern only | `["firefox"]` | pattern="firefox", confirm=false |
| Pattern + confirm | `["node", "--confirm"]` | pattern="node", confirm=true |
| Unknown flag | `["--bad"]` | Error: "unknown flag" |
| Extra args | `["a", "b"]` | Error: "unexpected argument" |
| Help flag | `["--help"]` | Prints usage, exits 0 |
| Pattern too long | 257-char string | Error: "exceeds maximum length" |

## Expected Results
- **All tests pass**: 0 failures
- **Test Coverage**: Target 70%+ on testable logic (process.go, entity.go, parseArgs)
- **Race detection**: No data races

## Writing the Tests

Create test files alongside source:
- `process_test.go` — test FindProcesses
- `entity_test.go` — test Entity methods
- `main_test.go` — test parseArgs (needs refactoring to be testable without os.Exit)
