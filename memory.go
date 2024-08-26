package circuitbreaker

import (
	"context"
	"sync/atomic"
	"time"
)

var _ Storage = &MemoryStorage{}

// NewMemoryStorage create new instance of Memory.
func NewMemoryStorage(options ...StorageOption) *MemoryStorage {
	storage := MemoryStorage{lastErrorAt: atomic.Value{}}

	for _, op := range options {
		op(&storage.options)
	}

	storage.lastErrorAt.Store(time.Time{})

	return &storage
}

// MemoryStorage is memory based storage for circuit breaker and is concurrent safe.
// do not use single MemoryStorage for multiple service, it will override the other services state.
type MemoryStorage struct {
	options     StorageOptions
	failures    atomic.Int64
	success     atomic.Int64
	lastErrorAt atomic.Value
}

// Failure is responsible to store failures.
func (m *MemoryStorage) Failure(ctx context.Context, delta int64) error {
	m.lastErrorAt.Store(time.Now().UTC())
	m.failures.Add(delta)
	m.success.Store(0)

	return nil
}

// Success is responsible to store success.
func (m *MemoryStorage) Success(ctx context.Context, delta int64) error {
	if m.success.Add(delta) >= m.options.SuccessRateThreshold {
		return m.Reset(ctx)
	}

	return nil
}

// GetState current state.
func (m *MemoryStorage) GetState(ctx context.Context) (State, error) {
	lastErrorAt := m.lastErrorAt.Load().(time.Time)
	errorExpireTTL := lastErrorAt.Add(m.options.OpenWindow).Sub(time.Now().UTC())
	if errorExpireTTL <= 0 {
		return StateClose, m.Reset(ctx)
	}

	if errorExpireTTL <= m.options.HalfOpenWindow {
		return StateHalfOpen, nil
	}

	if m.failures.Load() >= m.options.FailureRateThreshold {
		return StateOpen, nil
	}

	return StateClose, nil
}

// Reset the state.
func (m *MemoryStorage) Reset(ctx context.Context) error {
	m.success.Store(0)
	m.failures.Store(0)
	m.lastErrorAt.Store(time.Time{})

	return nil
}
