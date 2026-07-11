package service

import (
	"errors"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestGameService_FinderError(t *testing.T) {
	svc := NewGameService(
		&fake.Process{FindErr: errors.New("ps failed")},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGameService_ApplyKills_CompletesPendingKill(t *testing.T) {
	info := process.Info{Pid: 100, Name: "target", Rss: 4096}
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{info}, game.Config{Speed: 2.0})

	target := game.NewTarget(info, 80, 24)
	svc.kills <- target

	svc.applyKills(g)

	if target.State != game.Killing {
		t.Errorf("expected target state Killing, got %v", target.State)
	}
	if g.Kills() != 1 {
		t.Errorf("expected 1 kill recorded, got %d", g.Kills())
	}
}

func TestGameService_ApplyKills_EmptyChannelNoOps(t *testing.T) {
	svc := NewGameService(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
	svc.kills = make(chan *game.Target, 1)

	g := game.New([]process.Info{}, game.Config{Speed: 2.0})

	svc.applyKills(g) // must not block

	if g.Kills() != 0 {
		t.Errorf("expected 0 kills, got %d", g.Kills())
	}
}

func TestGameService_NoProcesses(t *testing.T) {
	svc := NewGameService(
		&fake.Process{},
		&fake.Store{},
		&fake.Renderer{},
		fake.NewInputSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	if err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}
}
