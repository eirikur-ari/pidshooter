package fake

// Killer adapts a plain function to the game.Service killer interface for tests.
type Killer func(pid int, name string, protected bool) (killed, shouldReap bool, err error)

func (f Killer) Kill(pid int, name string, protected bool) (killed, shouldReap bool, err error) {
	return f(pid, name, protected)
}
