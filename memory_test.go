package circuitbreaker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorageStorage_Failure(t *testing.T) {
	t.Run("expected to increment the failure count and last error at", func(t *testing.T) {
		ms := NewMemoryStorage(WithFailureRateThreshold(10))

		err := ms.Failure(context.Background(), 5)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), ms.failures.Load())
		assert.Equal(t, time.Now().UTC().Minute(), ms.lastErrorAt.Load().(time.Time).Minute())
		assert.Equal(t, int64(0), ms.success.Load())
	})
}

func TestMemoryStorageStorage_Success(t *testing.T) {
	t.Run("expect to only increment the success count", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(1*time.Minute), WithSuccessRateThreshold(2))
		now := time.Now().UTC()
		ms.lastErrorAt.Store(now)

		err := ms.Success(context.Background(), 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, now, ms.lastErrorAt.Load().(time.Time))
	})

	t.Run("expect to increment the success count and reset circuit", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(1*time.Minute), WithSuccessRateThreshold(1))
		now := time.Now().UTC()
		ms.lastErrorAt.Store(now)
		ms.failures.Store(3)

		err := ms.Success(context.Background(), 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, time.Time{}, ms.lastErrorAt.Load().(time.Time))
	})
}

func TestMemorystorage_getstatus(t *testing.T) {
	t.Run("the last error is expired, expect to reset the circuit", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(10 * time.Minute))
		ms.lastErrorAt.Store(time.Now().UTC().Add(-11 * time.Minute))

		cState, err := ms.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, StateClose, cState)
	})

	t.Run("we are in the half open state based on the last error time", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(10*time.Minute), WithHalfOpenWindow(5*time.Minute))
		ms.lastErrorAt.Store(time.Now().UTC().Add(-5 * time.Minute))

		cState, err := ms.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, StateHalfOpen, cState)
	})

	t.Run("circuit is not expired and we are not in the half open state, but the error threshold is reached", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(10*time.Minute), WithHalfOpenWindow(5*time.Minute), WithFailureRateThreshold(2))
		ms.lastErrorAt.Store(time.Now().UTC().Add(1 * time.Minute))
		ms.failures.Store(2)

		cState, err := ms.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, StateOpen, cState)
	})

	t.Run("circuit is not expired and we are not in the half open state, and err threshold is not reached, so the state is close", func(t *testing.T) {
		ms := NewMemoryStorage(WithOpenWindow(10*time.Minute), WithHalfOpenWindow(5*time.Minute), WithFailureRateThreshold(2))
		ms.lastErrorAt.Store(time.Now().UTC().Add(1 * time.Minute))
		ms.failures.Store(1)

		cState, err := ms.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, StateClose, cState)
	})
}

func TestMemoryStorageStorage_Reset(t *testing.T) {
	t.Run("expected to set failure, success, last error at to default value (0)", func(t *testing.T) {
		ms := NewMemoryStorage()
		ms.failures.Store(20)
		ms.success.Store(10)
		ms.lastErrorAt.Store(time.Now())

		err := ms.Reset(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, int64(0), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, time.Time{}, ms.lastErrorAt.Load().(time.Time))
	})
}
