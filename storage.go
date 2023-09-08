package circuitbreaker

import (
	"context"
)

// Storage is what circut breaker use to store it state.
type Storage interface {
	Failure(ctx context.Context, delta int64) error
	Success(ctx context.Context, delta int64) error
	GetState(ctx context.Context) (State, error)
	Reset(ctx context.Context) error
}

// nolint
const (
	RedisStorageName  = "redis"
	MemoryStorageName = "memory"
)

const (
	storagePrefix = "circuitBreaker:"
)

func namespace(service string) string {
	return storagePrefix + service
}
