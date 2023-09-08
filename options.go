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

	// DefaultState is state that used and default.
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
	// TODO 03.04.22 mrsoftware: do we need error weight?
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

func WithServiceName(service string) StorageOption {
	return func(o *StorageOptions) {
		o.Service = service
	}
}

func WithStorage(storage Storage) Option {
	return func(o *Options) {
		o.Storage = storage
	}
}

func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

func WithDefaultState(state State) Option {
	return func(o *Options) {
		o.State = state
	}
}

func WithFailureRateThreshold(rate int64) StorageOption {
	return func(o *StorageOptions) {
		o.FailureRateThreshold = rate
	}
}

func WithSuccessRateThreshold(rate int64) StorageOption {
	return func(o *StorageOptions) {
		o.SuccessRateThreshold = rate
	}
}

func WithOpenWindow(duration time.Duration) StorageOption {
	return func(o *StorageOptions) {
		o.OpenWindow = duration
	}
}

func WithHalfOpenWindow(duration time.Duration) StorageOption {
	return func(o *StorageOptions) {
		o.HalfOpenWindow = duration
	}
}
