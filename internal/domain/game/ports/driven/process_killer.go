package driven

// ProcessKiller is the driven port for sending a kill signal to a process.
type ProcessKiller interface {
	Kill(pid int) error
}
