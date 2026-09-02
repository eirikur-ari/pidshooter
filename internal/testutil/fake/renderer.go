package fake

import "github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"

// Renderer is a test double for outbound.Renderer.
type Renderer struct {
	InitErr error
}

func (r *Renderer) Init() error                  { return r.InitErr }
func (r *Renderer) Cleanup()                     {}
func (r *Renderer) Size() (int, int)             { return 80, 24 }
func (r *Renderer) Render(_ outbound.FrameState) {}
