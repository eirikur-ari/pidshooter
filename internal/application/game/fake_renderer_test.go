package game

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

type FakeRenderer struct {
	InitErr      error
	SizeW, SizeH int
	CleanupCalls int
	LastFrame    *outbound.FrameViewState
}

func (r *FakeRenderer) Init() error { return r.InitErr }

func (r *FakeRenderer) Cleanup() { r.CleanupCalls++ }

func (r *FakeRenderer) WindowSize() outbound.WindowSize {
	if r.SizeW != 0 || r.SizeH != 0 {
		return outbound.WindowSize{Width: r.SizeW, Height: r.SizeH}
	}
	return outbound.WindowSize{Width: 80, Height: 24}
}

func (r *FakeRenderer) ChromeSize() outbound.ChromeSize {
	return outbound.ChromeSize{Top: 1, Bottom: 1}
}

func (r *FakeRenderer) Render(state outbound.FrameViewState) { r.LastFrame = &state }
