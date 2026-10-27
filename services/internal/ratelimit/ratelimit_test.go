package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPermiteHastaLimite(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	limiter := &FixedWindow{
		Counter: NewMemoryCounter(func() time.Time { return now }),
		Window:  time.Minute,
		Limit:   2,
	}
	if ok, _ := limiter.Allow(context.Background(), "u1"); !ok {
		t.Fatal("la primera peticion debia admitirse")
	}
	if ok, _ := limiter.Allow(context.Background(), "u1"); !ok {
		t.Fatal("la segunda peticion debia admitirse")
	}
	if ok, _ := limiter.Allow(context.Background(), "u1"); ok {
		t.Fatal("la tercera peticion debia denegarse")
	}
}

func TestClavesIndependientes(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	limiter := &FixedWindow{Counter: NewMemoryCounter(func() time.Time { return now }), Window: time.Minute, Limit: 1}
	if ok, _ := limiter.Allow(context.Background(), "u1"); !ok {
		t.Fatal("u1 debia admitirse")
	}
	if ok, _ := limiter.Allow(context.Background(), "u2"); !ok {
		t.Fatal("u2 debia admitirse de forma independiente")
	}
}

func TestLaVentanaSeReinicia(t *testing.T) {
	base := time.Unix(1700000000, 0).UTC()
	current := base
	limiter := &FixedWindow{Counter: NewMemoryCounter(func() time.Time { return current }), Window: time.Minute, Limit: 1}
	if ok, _ := limiter.Allow(context.Background(), "u1"); !ok {
		t.Fatal("la primera peticion debia admitirse")
	}
	if ok, _ := limiter.Allow(context.Background(), "u1"); ok {
		t.Fatal("la segunda peticion debia denegarse")
	}
	current = base.Add(2 * time.Minute)
	if ok, _ := limiter.Allow(context.Background(), "u1"); !ok {
		t.Fatal("la ventana nueva debia admitir la peticion")
	}
}

type failingCounter struct{}

func (failingCounter) IncrWindow(context.Context, string, time.Duration) (int64, error) {
	return 0, errors.New("caida del contador")
}

func TestFailOpenAnteError(t *testing.T) {
	limiter := &FixedWindow{Counter: failingCounter{}, Window: time.Minute, Limit: 1}
	ok, err := limiter.Allow(context.Background(), "u1")
	if !ok || err == nil {
		t.Fatalf("se esperaba fail-open con error: (%v, %v)", ok, err)
	}
}
