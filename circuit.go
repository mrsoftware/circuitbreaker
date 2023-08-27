package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
)

var _ Manager = &Circuit{}

// ErrIsOpen meant circuit is open and can not accept any new request.
var ErrIsOpen = errors.New("CircuitBreaker: external service dont accept new request")

// Fn is type of callable than Do and DoWithFallback accept.
type Fn func() (interface{}, error)

// Circuit is a Circuit manager.
type Circuit struct {
	ops Options
}

// NewCircuit breaker.
func NewCircuit(options ...Option) Circuit {
	circuit := Circuit{ops: Options{}}

	for _, op := range options {
		op(&circuit.ops)
	}

	return circuit
}

// Manager is Circuit Breaker manager.
type Manager interface {
	Is(ctx context.Context, state State) bool
	IsAvailable(ctx context.Context) bool
	Done(ctx context.Context, err error)
	Do(ctx context.Context, fn Fn) (interface{}, error)
}

// IsAvailable checks if the service is available.
func (s *Circuit) IsAvailable(ctx context.Context) bool {
	return !s.Is(ctx, StateOpen)
}

// Is compare current state with requested state.
func (s *Circuit) Is(ctx context.Context, state State) (is bool) {
	currentState, err := s.ops.Storage.GetState(ctx)
	if err != nil {
		s.ops.Logger.Error(fmt.Errorf("checking service status: %w", err))

		return state == s.ops.State
	}

	return currentState == state
}

// Done call when operation is done/failed.
func (s *Circuit) Done(ctx context.Context, err error) {
	if err != nil {
		s.doneWithError(ctx)

		return
	}

	s.doneWithoutError(ctx)
}

func (s *Circuit) doneWithError(ctx context.Context) {
	if err := s.ops.Storage.Failure(ctx, 1); err != nil {
		s.ops.Logger.Error(fmt.Errorf("storing service failure: %w", err))
	}
}

func (s *Circuit) doneWithoutError(ctx context.Context) {
	state, err := s.ops.Storage.GetState(ctx)
	if err != nil {
		s.ops.Logger.Error(fmt.Errorf("getting service status: %w", err))

		state = s.ops.State
	}

	if state == StateClose {
		return
	}

	if err := s.ops.Storage.Success(ctx, 1); err != nil {
		s.ops.Logger.Error(fmt.Errorf("storing service success: %w", err))
	}
}

// Do check circuit state and call fn is not open.
func (s *Circuit) Do(ctx context.Context, fn Fn) (res interface{}, err error) {
	if !s.IsAvailable(ctx) {
		return nil, ErrIsOpen
	}
	defer s.Done(ctx, err)

	return fn()
}
