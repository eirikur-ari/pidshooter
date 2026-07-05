package driven

// ProcessKiller is the driven port for sending a kill signal to a process.
// name is the expected process name as discovered at game start; Kill must
// verify the process still has that name before sending the signal.
type ProcessKiller interface {
	Kill(pid int, name string) error
}
