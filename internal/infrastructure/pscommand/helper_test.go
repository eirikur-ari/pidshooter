package pscommand

func hostMockSetup(goos string) *MockHost {
	host := &MockHost{}
	host.On("goos").Return(goos)

	return host
}
