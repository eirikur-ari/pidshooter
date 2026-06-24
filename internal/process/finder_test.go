package process

import "testing"

func TestFind_EmptyPatterns(t *testing.T) {
	_, err := finder{}.Find([]string{})
	if err == nil {
		t.Error("expected error for empty patterns")
	}
}

func TestFind_EmptyTerm(t *testing.T) {
	_, err := finder{}.Find([]string{""})
	if err == nil {
		t.Error("expected error for empty term")
	}
}