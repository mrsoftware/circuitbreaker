package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrsoftware/circuitbreaker"
	"github.com/mrsoftware/circuitbreaker/mock"
	"github.com/stretchr/testify/assert"
	mockPkg "github.com/stretchr/testify/mock"
)

func TestCircuitbreaker_IS(t *testing.T) {
	storage := &mock.Storage{}
	logger := &mock.Logger{}

	breaker := circuitbreaker.NewCircuit(
		circuitbreaker.WithStorage(storage),
		circuitbreaker.WithLogger(logger),
		circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
	)

	t.Run("expect storage to not fails and state is what asked", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once()

		isOK := breaker.Is(context.Background(), circuitbreaker.StateClose)
		assert.True(t, isOK)

		storage.AssertExpectations(t)
	})

	t.Run("expect storage to not fails and state is not what asked", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once()

		isOK := breaker.Is(context.Background(), circuitbreaker.StateOpen)
		assert.False(t, isOK)

		storage.AssertExpectations(t)
	})

	t.Run("storage fails to provide state, expect to use default state", func(t *testing.T) {
		expectedErr := errors.New("some error")

		logger.On("Error", mockPkg.MatchedBy(func(err error) bool { return errors.Is(err, expectedErr) })).Once()
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, expectedErr).Once()

		isOK := breaker.Is(context.Background(), circuitbreaker.StateOpen)
		assert.False(t, isOK)

		storage.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestCircuitbreaker_Done(t *testing.T) {
	storage := &mock.Storage{}
	logger := &mock.Logger{}

	breaker := circuitbreaker.NewCircuit(
		circuitbreaker.WithStorage(storage),
		circuitbreaker.WithLogger(logger),
		circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
	)

	t.Run("done with error, expect to increase failure", func(t *testing.T) {
		storage.On("Failure", context.Background(), int64(1)).Return(nil).Once()

		breaker.Done(context.Background(), errors.New("some error"))

		storage.AssertExpectations(t)
	})

	t.Run("done with error, expect storage to fail to increase failure", func(t *testing.T) {
		expectedErr := errors.New("some error")

		storage.On("Failure", context.Background(), int64(1)).Return(expectedErr).Once()
		logger.On("Error", mockPkg.MatchedBy(func(err error) bool { return errors.Is(err, expectedErr) })).Once()

		breaker.Done(context.Background(), errors.New("some error"))

		logger.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("done without error, expect to increase success", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once()

		breaker.Done(context.Background(), nil)

		storage.AssertExpectations(t)

	})

	t.Run("done without error, expect getState to fail in getting state, use default state", func(t *testing.T) {
		expectedErr := errors.New("some error")

		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, expectedErr).Once()
		logger.On("Error", mockPkg.MatchedBy(func(err error) bool { return errors.Is(err, expectedErr) })).Once()

		breaker.Done(context.Background(), nil)

		storage.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("done without error, expect getState to return the open state and increase success", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateOpen, nil).Once()
		storage.On("Success", context.Background(), int64(1)).Return(nil).Once()

		breaker.Done(context.Background(), nil)

		storage.AssertExpectations(t)
	})

	t.Run("done without error, expect getState to return the open state but incresing success faild", func(t *testing.T) {
		expectedErr := errors.New("some error")

		storage.On("GetState", context.Background()).Return(circuitbreaker.StateOpen, nil).Once()
		storage.On("Success", context.Background(), int64(1)).Return(expectedErr).Once()
		logger.On("Error", mockPkg.MatchedBy(func(err error) bool { return errors.Is(err, expectedErr) })).Once()

		breaker.Done(context.Background(), nil)

		storage.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func TestCircuitBreaker_Do(t *testing.T) {
	storage := &mock.Storage{}

	breaker := circuitbreaker.NewCircuit(
		circuitbreaker.WithStorage(storage),
		circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
	)

	t.Run("circuit is close/available, expect to work and get expected result", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once() // is called in is available
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once() // is called in success

		fn := func() (interface{}, error) { return "response", nil }
		expectedRes, expectedErr := fn()

		response, err := breaker.Do(context.Background(), fn)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedRes, response)

		storage.AssertExpectations(t)
	})

	t.Run("circuit is close/available, but service call is failed, expect to record error", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Once() // is called in is available
		storage.On("Failure", context.Background(), int64(1)).Return(nil).Once()

		fn := func() (interface{}, error) { return nil, errors.New("service faild") }
		expectedRes, expectedErr := fn()

		response, err := breaker.Do(context.Background(), fn)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedRes, response)

		storage.AssertExpectations(t)
	})

	t.Run("circuit is not available, expect to get ErrIsOpen Error", func(t *testing.T) {
		storage.On("GetState", context.Background()).Return(circuitbreaker.StateOpen, nil).Once() // is called in is available

		response, err := breaker.Do(context.Background(), nil)
		assert.Nil(t, response)
		assert.Equal(t, circuitbreaker.ErrIsOpen, err)

		storage.AssertExpectations(t)
	})
}

func TestCircuitBreaker_Stat(t *testing.T) {
	t.Run("expect success to have value", func(t *testing.T) {
		storage := &mock.Storage{}

		breaker := circuitbreaker.NewCircuit(
			circuitbreaker.WithStorage(storage),
			circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
		)

		storage.On("GetState", context.Background()).Return(circuitbreaker.StateClose, nil).Twice()

		breaker.Done(context.Background(), nil)

		stat := breaker.Stat(context.Background())
		assert.Equal(t, circuitbreaker.Stat{Success: 1, Failure: 0, State: circuitbreaker.StateClose}, stat)

		storage.AssertExpectations(t)
	})

	t.Run("expect failure to have value", func(t *testing.T) {
		storage := &mock.Storage{}

		breaker := circuitbreaker.NewCircuit(
			circuitbreaker.WithStorage(storage),
			circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
		)

		storage.On("GetState", context.Background()).Return(circuitbreaker.StateOpen, nil).Once()
		storage.On("Failure", context.Background(), int64(1)).Return(nil).Once()

		breaker.Done(context.Background(), errors.New("single error to increase failure"))

		stat := breaker.Stat(context.Background())
		assert.Equal(t, circuitbreaker.Stat{Failure: 1, Success: 0, State: circuitbreaker.StateOpen}, stat)

		storage.AssertExpectations(t)
	})

	t.Run("expect state have open as its value", func(t *testing.T) {
		storage := &mock.Storage{}

		breaker := circuitbreaker.NewCircuit(
			circuitbreaker.WithStorage(storage),
			circuitbreaker.WithFallbackState(circuitbreaker.StateClose),
		)

		storage.On("GetState", context.Background()).Return(circuitbreaker.StateHalfOpen, nil).Once()

		stat := breaker.Stat(context.Background())
		assert.Equal(t, circuitbreaker.Stat{State: circuitbreaker.StateHalfOpen}, stat)

		storage.AssertExpectations(t)
	})
}
