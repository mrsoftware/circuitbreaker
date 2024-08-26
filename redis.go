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
func NewRedisStorage(client *redis.Client, options ...StorageOption) *RedisStorage {
	storage := RedisStorage{client: client}

	for _, op := range options {
		op(&storage.options)
	}

	storage.serviceKey = namespace(storage.options.Service)

	return &storage
}

// RedisStorage is redis based storage for circuit breaker and is concurrent safe.
type RedisStorage struct {
	client     *redis.Client
	options    StorageOptions
	serviceKey string
}

// Failure is responsible to store failures.
func (r *RedisStorage) Failure(ctx context.Context, delta int64) error {

	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, r.serviceKey, failuresField, delta)
	pipe.HDel(ctx, r.serviceKey, successField)
	pipe.Expire(ctx, r.serviceKey, r.options.OpenWindow)

	return r.pipeExec(ctx, pipe)
}

// Success is responsible to store success.
func (r *RedisStorage) Success(ctx context.Context, delta int64) error {
	sCount, err := r.client.HIncrBy(ctx, r.serviceKey, successField, delta).Result()
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
	duration, err := r.client.PTTL(ctx, r.serviceKey).Result()
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
	fCount, err := r.client.HGet(ctx, r.serviceKey, failuresField).Int64()
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
	return r.client.Del(ctx, r.serviceKey).Err()
}
