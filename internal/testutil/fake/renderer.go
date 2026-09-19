package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// Renderer is a test double for outbound.Renderer.
type Renderer struct {
	InitErr error
	// SizeW, SizeH are returned by WindowSize, overriding the default 80x24
	// when either is nonzero.
	SizeW, SizeH int
	// CleanupCalls counts how many times Cleanup was called.
	CleanupCalls int
	// LastFrame captures the state passed to the most recent Render call,
	// nil if Render was never called.
	LastFrame *outbound.FrameState
}

func (r *Renderer) Init() error { return r.InitErr }

func (r *Renderer) Cleanup() { r.CleanupCalls++ }

func (r *Renderer) WindowSize() outbound.WindowSize {
	if r.SizeW != 0 || r.SizeH != 0 {
		return outbound.WindowSize{Width: r.SizeW, Height: r.SizeH}
	}
	return outbound.WindowSize{Width: 80, Height: 24}
}

func (r *Renderer) ChromeSize() outbound.ChromeSize {
	return outbound.ChromeSize{Top: 1, Bottom: 1}
}

func (r *Renderer) Render(state outbound.FrameState) { r.LastFrame = &state }
