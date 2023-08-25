package mock

import (
	"context"

	"github.com/mrsoftware/circuitbreaker"
	"github.com/stretchr/testify/mock"
)

type Storage struct {
	mock.Mock
}

func (s *Storage) GetState(ctx context.Context) (circuitbreaker.State, error) {
	args := s.Called(ctx)

	return args.Get(0).(circuitbreaker.State), args.Error(1)
}

func (s *Storage) Success(ctx context.Context, delta int64) error {
	return s.Called(ctx, delta).Error(0)
}

func (s *Storage) Failure(ctx context.Context, delta int64) error {
	return s.Called(ctx, delta).Error(0)
}

func (s *Storage) Reset(ctx context.Context) error {
	return s.Called(ctx).Error(0)
}
