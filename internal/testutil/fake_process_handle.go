package testutil

// fakeProcessHandle is a test double for outbound.ProcessHandle.
type fakeProcessHandle struct {
	process *FakeProcessManager
	pid     int
}

func (h *fakeProcessHandle) Kill() error {
	h.process.KilledPIDs = append(h.process.KilledPIDs, h.pid)
	return h.process.KillErr
}

func (h *fakeProcessHandle) Release() error {
	h.process.ReleasedPIDs = append(h.process.ReleasedPIDs, h.pid)
	return nil
}
