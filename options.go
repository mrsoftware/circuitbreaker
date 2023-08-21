package circuitbreaker

import "time"

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
	DefaultState state = StateClose
)

// Options is circuit breaker options.
type Options struct {
	Storage Storage
	Logger  Logger
	Service string
	State   state
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

// Logger for circuit breaker.
type Logger interface {
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

type Option func(*Options)

func WithServiceName(service string) Option {
	return func(o *Options) {
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

func WithDefaultState(state state) Option {
	return func(o *Options) {
		o.State = state
	}
}

func WithFailureRateThreshold(rate int64) Option {
	return func(o *Options) {
		o.FailureRateThreshold = rate
	}
}

func WithSuccessRateThreshold(rate int64) Option {
	return func(o *Options) {
		o.SuccessRateThreshold = rate
	}
}

func WithOpenWindow(duration time.Duration) Option {
	return func(o *Options) {
		o.OpenWindow = duration
	}
}

func WithHalfOpenWindow(duration time.Duration) Option {
	return func(o *Options) {
		o.HalfOpenWindow = duration
	}
}
