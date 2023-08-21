package circuitbreaker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorageStorage_Failure(t *testing.T) {
	t.Run("expected to increment the failure count and last error at", func(t *testing.T) {
		ms := NewMemoryStorage(Options{FailureRateThreshold: 10})

		err := ms.Failure(context.Background(), 5)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), ms.failures.Load())
		assert.Equal(t, time.Now().UTC().Minute(), ms.lastErrorAt.Load().(time.Time).Minute())
		assert.Equal(t, StateClose, state(ms.state.Load()))
		assert.Equal(t, int64(0), ms.success.Load())
	})

	t.Run("expected to increment the failure count and last error at and change state to close", func(t *testing.T) {
		ms := NewMemoryStorage(Options{FailureRateThreshold: 5})

		err := ms.Failure(context.Background(), 10)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), ms.failures.Load())
		assert.Equal(t, time.Now().UTC().Minute(), ms.lastErrorAt.Load().(time.Time).Minute())
		assert.Equal(t, StateOpen, state(ms.state.Load()))
		assert.Equal(t, int64(0), ms.success.Load())
	})
}

func TestMemoryStorageStorage_Success(t *testing.T) {
	t.Run("state is close, expect to do nothing", func(t *testing.T) {
		ms := NewMemoryStorage(Options{})

		err := ms.Success(context.Background(), 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, StateClose, state(ms.state.Load()))
		assert.Equal(t, 0, ms.lastErrorAt.Load().(time.Time).Minute())
	})

	t.Run("state is open, expect to only increment the success count", func(t *testing.T) {
		ms := NewMemoryStorage(Options{OpenWindow: 1 * time.Minute, SuccessRateThreshold: 2})
		ms.state.Store(int64(StateOpen))
		ms.lastErrorAt.Store(time.Now().UTC())

		err := ms.Success(context.Background(), 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, StateOpen, state(ms.state.Load()))
		assert.Equal(t, time.Now().UTC().Minute(), ms.lastErrorAt.Load().(time.Time).Minute())
	})

	t.Run("state is open, change state to close if last error happened befor openwindow", func(t *testing.T) {
		ms := NewMemoryStorage(Options{SuccessRateThreshold: 10, OpenWindow: 5 * time.Minute})
		ms.state.Store(int64(StateOpen))
		ms.lastErrorAt.Store(time.Now().UTC().Add(-5 * time.Minute))

		err := ms.Success(context.Background(), 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, StateClose, state(ms.state.Load()))
		assert.Equal(t, 0, ms.lastErrorAt.Load().(time.Time).Minute())
	})

	t.Run("state is open and we in the half open window, expect to change state to half open", func(t *testing.T) {
		ms := NewMemoryStorage(Options{SuccessRateThreshold: 10, OpenWindow: 6 * time.Minute, HalfOpenWindow: 3 * time.Minute})
		ms.state.Store(int64(StateOpen))
		lastErrAt := time.Now().UTC().Add(-ms.options.HalfOpenWindow)
		ms.lastErrorAt.Store(lastErrAt)

		err := ms.Success(context.Background(), 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, StateHalfOpen, state(ms.state.Load()))
		assert.Equal(t, lastErrAt, ms.lastErrorAt.Load().(time.Time))
	})
}

func TestMemorystorage_getstatus(t *testing.T) {
	t.Run("expected to return the stored state and do nothing", func(t *testing.T) {
		ms := NewMemoryStorage(Options{})
		ms.state.Store(int64(StateClose))

		cState, err := ms.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, StateClose, cState)
	})
}

func TestMemoryStorageStorage_Reset(t *testing.T) {
	t.Run("exptecte to set failure, success, last error at to default value (0)", func(t *testing.T) {
		ms := NewMemoryStorage(Options{State: StateClose})
		ms.failures.Store(20)
		ms.success.Store(10)
		ms.lastErrorAt.Store(time.Now())

		err := ms.Reset(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, int64(0), ms.success.Load())
		assert.Equal(t, int64(0), ms.failures.Load())
		assert.Equal(t, StateClose, state(ms.state.Load()))
		assert.Equal(t, time.Time{}, ms.lastErrorAt.Load().(time.Time))
	})
}
