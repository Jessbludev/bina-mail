// Package ratelimit implementa limites por ventana fija respaldados por
// Dragonfly o por memoria cuando no hay cache configurado.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter decide si una peticion con clave dada debe admitirse.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// Counter cuenta eventos dentro de una ventana; lo implementan resp.Client
// (Dragonfly) y MemoryCounter (respaldo en proceso).
type Counter interface {
	IncrWindow(ctx context.Context, key string, window time.Duration) (int64, error)
}

// FixedWindow admite hasta Limit eventos por clave y ventana.
// Un error del contador se resuelve en abierto (fail-open): la
// disponibilidad del MVP prevalece; se revisa en Fase 5.
type FixedWindow struct {
	Counter Counter
	Window  time.Duration
	Limit   int64
}

// Allow registra el evento y decide si cabe en la ventana.
func (f *FixedWindow) Allow(ctx context.Context, key string) (bool, error) {
	if f.Counter == nil || f.Limit <= 0 {
		return true, nil
	}
	n, err := f.Counter.IncrWindow(ctx, "bina:rl:"+key, f.Window)
	if err != nil {
		return true, err
	}
	return n <= f.Limit, nil
}

// MemoryCounter es un contador en proceso para entornos sin Dragonfly.
// El reloj se inyecta para determinismo en pruebas.
type MemoryCounter struct {
	mu      sync.Mutex
	entries map[string]memoryEntry
	now     func() time.Time
}

type memoryEntry struct {
	count int64
	reset time.Time
}

// NewMemoryCounter crea el contador en proceso con reloj opcional.
func NewMemoryCounter(now func() time.Time) *MemoryCounter {
	if now == nil {
		now = time.Now
	}
	return &MemoryCounter{entries: map[string]memoryEntry{}, now: now}
}

// IncrWindow cuenta un evento dentro de la ventana vigente.
func (m *MemoryCounter) IncrWindow(_ context.Context, key string, window time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	entry, ok := m.entries[key]
	if !ok || !now.Before(entry.reset) {
		entry = memoryEntry{reset: now.Add(window)}
	}
	entry.count++
	m.entries[key] = entry
	return entry.count, nil
}
