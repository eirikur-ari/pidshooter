package game

import "testing"

func TestState_DefaultIsPending(t *testing.T) {
	var s state
	if s.Load() != Pending {
		t.Errorf("expected zero-value State to be Pending, got %d", s.Load())
	}
}

func TestState_Store_Load(t *testing.T) {
	var s state
	s.Store(Running)
	if s.Load() != Running {
		t.Errorf("expected Running after Store, got %d", s.Load())
	}
	s.Store(Stopped)
	if s.Load() != Stopped {
		t.Errorf("expected Stopped after Store, got %d", s.Load())
	}
}

func TestState_CompareAndSwap_Success(t *testing.T) {
	var s state
	if !s.CompareAndSwap(Pending, Running) {
		t.Error("expected CAS to succeed when old matches current")
	}
	if s.Load() != Running {
		t.Errorf("expected Running after successful CAS, got %d", s.Load())
	}
}

func TestState_CompareAndSwap_Failure(t *testing.T) {
	var s state
	s.Store(Running)
	if s.CompareAndSwap(Pending, Stopped) {
		t.Error("expected CAS to fail when old does not match current")
	}
	if s.Load() != Running {
		t.Errorf("expected state unchanged after failed CAS, got %d", s.Load())
	}
}
