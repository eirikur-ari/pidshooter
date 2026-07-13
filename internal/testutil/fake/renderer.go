package fake

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// Renderer is a test double for outbound.Renderer.
type Renderer struct{}

func (r *Renderer) Init() error         { return nil }
func (r *Renderer) Cleanup()            {}
func (r *Renderer) Size() (int, int)    { return 80, 24 }
func (r *Renderer) Render(_ game.FrameState) {}
