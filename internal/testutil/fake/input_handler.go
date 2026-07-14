package fake

import "github.com/eirikur-ari/pidshooter/internal/core/game"

// InputHandler is a test double for event.InputHandler.
type InputHandler struct {
	QuitCalled      int
	YesCalled       int
	NoCalled        int
	SpeedUpCalled   int
	SpeedDownCalled int
	ClickX, ClickY  int
	ClickCalled     int
	YesResult       *game.Target
	ClickResult     *game.Target
}

func (h *InputHandler) OnQuit()                          { h.QuitCalled++ }
func (h *InputHandler) OnYes() *game.Target              { h.YesCalled++; return h.YesResult }
func (h *InputHandler) OnNo()                            { h.NoCalled++ }
func (h *InputHandler) OnSpeedUp()                       { h.SpeedUpCalled++ }
func (h *InputHandler) OnSpeedDown()                     { h.SpeedDownCalled++ }
func (h *InputHandler) OnClickAt(x, y int) *game.Target {
	h.ClickCalled++
	h.ClickX = x
	h.ClickY = y
	return h.ClickResult
}
