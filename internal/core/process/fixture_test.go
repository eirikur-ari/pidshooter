package process

func newInfoFixture() Info {
	return NewInfo(100, "myapp", 1024, 1000)
}

func newInfoFixtureFor(pid int, name string) Info {
	info := newInfoFixture()
	info.PID = pid
	info.Name = name
	return info
}
