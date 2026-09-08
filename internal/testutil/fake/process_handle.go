package fake

// processHandle is a test double for outbound.ProcessHandle, bound to the
// Process that created it so Kill can record which pid it was called with.
type processHandle struct {
	process *Process
	pid     int
}

func (h *processHandle) Kill() error {
	h.process.KilledPIDs = append(h.process.KilledPIDs, h.pid)
	return h.process.KillErr
}