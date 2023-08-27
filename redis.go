package circuitbreaker

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

const (
	failuresField = "failures"
	successField  = "success"
)

var _ Storage = &RedisStorage{}

// NewRedisStorage create new instance of RedisStorage.
func NewRedisStorage(client *redis.Client, options Options) *RedisStorage {
	return &RedisStorage{client: client, options: options}
}

// RedisStorage is redis based storage for circuit breaker and is concurrent safe.
type RedisStorage struct {
	client  *redis.Client
	options Options
}

// Failure is responsible to store failures.
func (r *RedisStorage) Failure(ctx context.Context, delta int64) error {
	keyName := namespace(r.options.Service)

	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, keyName, failuresField, delta)
	pipe.Expire(ctx, keyName, r.options.OpenWindow)

	return r.pipeExec(ctx, pipe)
}

// Success is responsible to store success.
func (r *RedisStorage) Success(ctx context.Context, delta int64) error {
	sCount, err := r.client.HIncrBy(ctx, namespace(r.options.Service), successField, delta).Result()
	if err != nil {
		return err
	}

	if sCount >= r.options.SuccessRateThreshold {
		return r.Reset(ctx)
	}

	return nil
}

func (r *RedisStorage) pipeExec(ctx context.Context, pipe redis.Pipeliner) error {
	cmdErrs, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	for _, cmdErr := range cmdErrs {
		if er := cmdErr.Err(); er != nil {
			return err
		}
	}

	return nil
}

// GetState current state.
// if key expired or not exits == close
// if we are in halfOpen window == halfOpen
// if key exist and not in halfOpen window and errors count reached the limit == open.
func (r *RedisStorage) GetState(ctx context.Context) (State, error) {
	duration, err := r.client.PTTL(ctx, namespace(r.options.Service)).Result()
	if err != nil {
		return StateClose, err
	}

	// -1, -2 means no expire and key not exist and
	if duration < 0 {
		return StateClose, nil
	}

	if duration <= r.options.HalfOpenWindow {
		return StateHalfOpen, nil
	}

	reachRateLimit, err := r.reachRateLimit(ctx)
	if err != nil {
		return StateClose, err
	}

	if reachRateLimit {
		return StateOpen, nil
	}

	return StateClose, nil
}

func (r *RedisStorage) reachRateLimit(ctx context.Context) (bool, error) {
	fCount, err := r.client.HGet(ctx, namespace(r.options.Service), failuresField).Int64()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return fCount >= r.options.FailureRateThreshold, nil
}

// Reset storage.
func (r *RedisStorage) Reset(ctx context.Context) error {
	return r.client.Del(ctx, namespace(r.options.Service)).Err()
}
