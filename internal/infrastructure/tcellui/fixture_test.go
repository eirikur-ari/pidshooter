package tcellui

import (
	"github.com/gdamore/tcell/v2"
)

func newScreenFixture(width, height int) tcell.SimulationScreen {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		panic(err)
	}
	screen.SetSize(width, height)
	return screen
}

func newUninitializedScreenFixture() tcell.SimulationScreen {
	return tcell.NewSimulationScreen("")
}
