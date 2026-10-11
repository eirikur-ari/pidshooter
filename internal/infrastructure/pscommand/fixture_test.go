package pscommand

import (
	"strconv"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

const psHeader = "  UID   PID    RSS STAT COMM\n"

func newHostFixture(goos string) *MockHost {
	host := &MockHost{}
	host.On("goos").Return(goos)

	return host
}

func newDiscoverArgumentsFixture() []string {
	return []string{"-eo", "uid,pid,rss,stat,comm"}
}

func newLookupArgumentsFixture(pid int) []string {
	return []string{"-p", strconv.Itoa(pid), "-o", "uid,pid,rss,stat,comm"}
}

func newPsOutputFixture() []byte {
	return newPsOutputWithStateFixture("S")
}

func newPsOutputWithStateFixture(state string) []byte {
	return []byte(psHeader + " 1000   123   4096 " + state + "    sleep\n")
}

func newProcessInfoFixture() outbound.ProcessInfo {
	return outbound.ProcessInfo{PID: 123, Name: "sleep", Rss: 4096 * 1024, UID: 1000, State: "S"}
}
