package circuitbreaker

import (
	"context"
	"fmt"
)

// Storage is what circut breaker use to store it state.
type Storage interface {
	Failure(ctx context.Context, delta int64) error
	Success(ctx context.Context, delta int64) error
	GetState(ctx context.Context) (state, error)
	Reset(ctx context.Context) error
}

// nolint
const (
	RedisStorageName  = "redis"
	MemoryStorageName = "memory"
)

const (
	storagePrefix   = "circuitBreaker"
	namespaceFormat = "%s:%s" // storagePrefix , operation
)

func namespace(operation string) string {
	return fmt.Sprintf(namespaceFormat, storagePrefix, operation)
}
