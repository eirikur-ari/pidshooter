# Integration Test Instructions

## Purpose
Test the full pidshooter workflow in a controlled environment to verify end-to-end behavior.

## Test Scenarios

### Scenario 1: Process Discovery + Display
- **Description**: Verify that matched processes appear as animated entities
- **Setup**: Start a known background process (e.g., `sleep 9999 &`)
- **Test Steps**:
  1. Run `./pidshooter sleep`
  2. Observe terminal — `sleep` process should appear as flying label
  3. Press 'q' to exit
- **Expected Results**: Process appears with correct PID and name, animation is smooth, terminal restores on exit
- **Cleanup**: `kill %1` (the background sleep process)

### Scenario 2: Mouse Click Kill
- **Description**: Verify clicking a process label sends SIGKILL
- **Setup**: Start a known background process (`sleep 9999 &`, note PID)
- **Test Steps**:
  1. Run `./pidshooter sleep`
  2. Click on the flying `[PID sleep]` label
  3. Observe kill animation
  4. Verify process is dead: `kill -0 <PID>` should fail
- **Expected Results**: Kill animation plays, entity removed, process is terminated
- **Cleanup**: None (process killed)

### Scenario 3: Confirm Mode
- **Description**: Verify --confirm flag prompts before killing
- **Setup**: Start `sleep 9999 &`
- **Test Steps**:
  1. Run `./pidshooter sleep --confirm`
  2. Click a process label
  3. Observe confirmation prompt in status bar
  4. Press 'y' to confirm
  5. Verify process is dead
- **Expected Results**: Prompt appears, 'y' kills, 'n' cancels
- **Cleanup**: None

### Scenario 4: No Matches
- **Description**: Verify graceful exit when no processes match
- **Test Steps**:
  1. Run `./pidshooter zzz_definitely_not_a_process`
- **Expected Results**: Prints "No processes found matching..." and exits cleanly
- **Cleanup**: None

### Scenario 5: Permission Denied
- **Description**: Verify graceful handling when attempting to kill a root-owned process
- **Setup**: Identify a root process (e.g., `./pidshooter systemd` as non-root user)
- **Test Steps**:
  1. Run `./pidshooter systemd` (or similar root process)
  2. Click the process label
- **Expected Results**: Kill animation plays (kill attempt is fire-and-forget), no crash
- **Cleanup**: None (process wasn't actually killed)

### Scenario 6: Signal Handling (Ctrl+C)
- **Description**: Verify clean terminal restore on SIGINT
- **Test Steps**:
  1. Run `./pidshooter sleep`
  2. Press Ctrl+C
- **Expected Results**: Terminal restored to normal, cursor visible, no garbage on screen
- **Cleanup**: None

## Running Integration Tests

These are manual tests. Execute each scenario in a terminal with mouse support:

```bash
# Setup
sleep 9999 &
SLEEP_PID=$!

# Run scenarios manually
./pidshooter sleep
# (interact, then exit)

# Cleanup
kill $SLEEP_PID 2>/dev/null
```
