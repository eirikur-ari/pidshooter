package process

func pidsOf(processes []Info) []int {
	var pids []int
	for _, process := range processes {
		pids = append(pids, process.PID)
	}
	return pids
}
