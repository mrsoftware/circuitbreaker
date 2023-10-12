package circuitbreaker

import (
	"os"
	"time"
)

const (
	// DefaultOpenWindow is time window of open state.
	DefaultOpenWindow = time.Second * 60

	// DefaultHalfOpenWindow is time widow of half open state.
	DefaultHalfOpenWindow = DefaultOpenWindow / 2

	// DefaultFailureRateThreshold is circuit max Failure Rate Threshold, to open circuit.
	DefaultFailureRateThreshold int64 = 50

	// DefaultSuccessRateThreshold is circuit max Success Rate Threshold, to close circuit.
	DefaultSuccessRateThreshold int64 = 10

	// DefaultState is state that used fallback state in case of internal failure.
	DefaultState State = StateClose
)

// Options is circuit breaker options.
type Options struct {
	Storage Storage
	Logger  Logger
	State   State
}

type StorageOptions struct {
	Service string
	// FailureRateThreshold haw many error to consider circuit as open
	FailureRateThreshold int64
	// SuccessRateThreshold how much success to consider circuit as full close
	// if its 0, then success counter will not change state to close and only timeBased solution will do it
	SuccessRateThreshold int64
	// OpenWindow is the duration of circuit open state will last
	OpenWindow time.Duration
	// HalfOpenWindow is the duration of circuit halfOpen state will last
	HalfOpenWindow time.Duration
}

func StorageWithDefaultOptions() StorageOption {
	return func(o *StorageOptions) {
		o.OpenWindow = DefaultHalfOpenWindow
		o.HalfOpenWindow = DefaultHalfOpenWindow
		o.FailureRateThreshold = DefaultFailureRateThreshold
		o.SuccessRateThreshold = DefaultSuccessRateThreshold
	}
}

func WithDefaultOptions() Option {
	return func(o *Options) {
		o.State = DefaultState
		o.Storage = NewMemoryStorage(StorageWithDefaultOptions())
		o.Logger = NewIOLogger(os.Stdout, OutPutTypeSimple)
	}
}

type StorageOption func(*StorageOptions)
type Option func(*Options)

// WithServiceName configures the storage option with a specific service name.
// It allows you to set a unique name or identifier for the service that the circuit
// breaker protects. This information is valuable for tracking and identifying
// circuit breakers when monitoring multiple services in a distributed system.
func WithServiceName(service string) StorageOption {
	return func(o *StorageOptions) {
		o.Service = service
	}
}

// WithStorage configures the circuit breaker with a specific storage mechanism.
// It lets you specify the storage component that will be used to persist the
// circuit breaker's state. This is useful for retaining the state across
// application restarts and ensuring that the circuit breaker maintains its
// state even in cases of application crashes or restarts.
func WithStorage(storage Storage) Option {
	return func(o *Options) {
		o.Storage = storage
	}
}

// WithLogger configures the circuit breaker with a custom logger.
// It allows you to integrate logging capabilities into the circuit breaker, making
// it possible to log events and important information related to the circuit
// breaker's operation. This feature is valuable for monitoring and debugging the
// circuit breaker's behavior.
func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithFallbackState sets the default state for the circuit breaker. This state
// will be used as a fallback in case of an internal failure or malfunction of
// the circuit breaker itself. It ensures that the circuit breaker can continue
// to operate even when it encounters errors or issues with its own functionality.
func WithFallbackState(state State) Option {
	return func(o *Options) {
		o.State = state
	}
}

// WithFailureRateThreshold sets the threshold for the failure rate that triggers
// the circuit breaker to transition from a closed to an open state. It allows you
// to define the number of failed requests that will lead to the
// circuit being opened to protect the service from further requests.
func WithFailureRateThreshold(rate int64) StorageOption {
	return func(o *StorageOptions) {
		o.FailureRateThreshold = rate
	}
}

// WithSuccessRateThreshold sets the threshold for the success rate that determines
// when the circuit breaker should transition from an open state to a closed.
// It helps decide when the service has sufficiently recovered
// and can resume normal operation.
func WithSuccessRateThreshold(rate int64) StorageOption {
	return func(o *StorageOptions) {
		o.SuccessRateThreshold = rate
	}
}

// WithOpenWindow sets the duration of the "open" state window in the circuit breaker. During this period,
// incoming requests are blocked to protect the service from continued damage in response to consecutive failures.
// The duration specified with this option determines how long the circuit breaker maintains the "open" state
// before transitioning to "close" state.
func WithOpenWindow(duration time.Duration) StorageOption {
	return func(o *StorageOptions) {
		o.OpenWindow = duration
	}
}

// WithHalfOpenWindow configures the time window during which the circuit breaker
// transitions into a "half-open" state after being in an open state. During this
// period, the circuit allows a limited number of test requests to pass through
// to assess the service's stability before fully reopening the circuit.
func WithHalfOpenWindow(duration time.Duration) StorageOption {
	return func(o *StorageOptions) {
		o.HalfOpenWindow = duration
	}
}
