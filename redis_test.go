package circuitbreaker_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/mrsoftware/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

const (
	tempkey       = "circuitBreaker:test"
	failuresField = "failures"
	successField  = "success"
	stateField    = "state"
)

var (
	serviceName = "test"
	options     = circuitbreaker.Options{
		Service:              serviceName,
		State:                circuitbreaker.StateClose,
		FailureRateThreshold: 2,
		SuccessRateThreshold: 2,
		OpenWindow:           circuitbreaker.DefaultOpenWindow,
		HalfOpenWindow:       circuitbreaker.DefaultHalfOpenWindow,
	}
)

func TestRedisStorage_GetStatus(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	rs := circuitbreaker.NewRedisStorage(redisClient, options)

	t.Run("key expired or not exits == close", func(t *testing.T) {
		mock.ExpectPTTL(tempkey).SetVal(-1)

		state, err := rs.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, circuitbreaker.StateClose, state)
	})

	t.Run("we are in halfOpen window == halfOpen", func(t *testing.T) {
		mock.ExpectPTTL(tempkey).SetVal(time.Second * 20)

		state, err := rs.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, circuitbreaker.StateHalfOpen, state)
	})

	t.Run("key exist and not in halfOpen window and errors count reached the limit == open", func(t *testing.T) {
		mock.ExpectPTTL(tempkey).SetVal(time.Second * 40)
		mock.ExpectHGet(tempkey, failuresField).SetVal(strconv.Itoa(int(options.FailureRateThreshold)))

		state, err := rs.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, circuitbreaker.StateOpen, state)
	})

	t.Run("key exist and not in halfOpen window and reachRateLimit got redis nil == close", func(t *testing.T) {
		mock.ExpectPTTL(tempkey).SetVal(time.Second * 40)
		mock.ExpectHGet(tempkey, failuresField).RedisNil()

		state, err := rs.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, circuitbreaker.StateClose, state)
	})
}

func TestRedisStorage_Failure(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	rs := circuitbreaker.NewRedisStorage(redisClient, options)

	t.Run("incr failure count to change state to open", func(t *testing.T) {

		mock.ExpectHIncrBy(tempkey, failuresField, 2).SetVal(2)
		mock.ExpectExpire(tempkey, options.OpenWindow).SetVal(true)

		err := rs.Failure(context.Background(), 2)
		assert.Nil(t, err)

		mock.ExpectPTTL(tempkey).SetVal(time.Second * 40)
		mock.ExpectHGet(tempkey, failuresField).SetVal(strconv.Itoa(int(options.FailureRateThreshold)))

		state, err := rs.GetState(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, circuitbreaker.StateOpen, state)
	})

}

func TestRedisStorage_Success(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	rs := circuitbreaker.NewRedisStorage(redisClient, options)

	t.Run("incr success count to change state to close", func(t *testing.T) {
		mock.ExpectHIncrBy(tempkey, successField, 1).SetVal(1)

		err := rs.Success(context.Background(), 1)
		assert.Nil(t, err)
	})

}
