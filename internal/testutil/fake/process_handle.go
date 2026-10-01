package fake

// processHandle is a test double for outbound.ProcessHandle, bound to the
// ProcessManager that created it so Kill can record which pid it was called
// with.
type processHandle struct {
	process *ProcessManager
	pid     int
}

func (h *processHandle) Kill() error {
	h.process.KilledPIDs = append(h.process.KilledPIDs, h.pid)
	return h.process.KillErr
}

func (h *processHandle) Release() error {
	h.process.ReleasedPIDs = append(h.process.ReleasedPIDs, h.pid)
	return nil
}