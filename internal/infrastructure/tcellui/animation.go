package tcellui

var (
	killAnimation = animation{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
	fleeAnimation = animation{"🏃💨", "↝ RAN AWAY ↝", "· · ·", "  ·  ", "     "}
)

// animation is an ordered sequence of frames drawn in place of a target's
// tag while it plays out.
type animation []string

// frame returns the frame corresponding to progress, a fraction from 0 to 1
// through the animation.
func (a animation) frame(progress float64) string {
	idx := int(progress * float64(len(a)))
	if idx >= len(a) {
		idx = len(a) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return a[idx]
}
