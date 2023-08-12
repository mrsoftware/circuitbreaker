package circuitbreaker

type Options struct {
	Service string
}

type Option func(*Options)

func WithServiceName(service string) Option {
	return func(o *Options) {
		o.Service = service
	}
}
