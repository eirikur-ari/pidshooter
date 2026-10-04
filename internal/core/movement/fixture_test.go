package movement

func newBoundsFixture() Bounds {
	return NewBounds(
		windowSizeFixture(),
		chromeSizeFixture(),
	)
}

func windowSizeFixture() WindowSize {
	return WindowSize{Width: 80, Height: 24}
}

func chromeSizeFixture() ChromeSize {
	return ChromeSize{Top: 1, Bottom: 1}
}
