package mock

import (
	"context"

	"github.com/mrsoftware/circuitbreaker"
	"github.com/stretchr/testify/mock"
)

type Circuit struct {
	mock.Mock
}

func (c *Circuit) Do(ctx context.Context, fn circuitbreaker.Fn) (interface{}, error) {
	args := c.Called(ctx, fn)

	return args.Get(0), args.Error(1)
}

func (c *Circuit) Done(ctx context.Context, err error) {
	c.Called(ctx, err)
}

func (c *Circuit) Is(ctx context.Context, state circuitbreaker.State) bool {
	return c.Called(ctx, state).Bool(0)
}

func (c *Circuit) IsAvailable(ctx context.Context) bool {
	return c.Called(ctx).Bool(0)
}
