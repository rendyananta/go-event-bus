package bus

import "context"

type Listener func(ctx context.Context, e Event) error

type RetryOption struct {
	max          int
	count        int
	fromPosition int
}

type Options struct {
	retry *RetryOption
}

func WithOptions(options ...func(options *Options)) *Options {
	opt := &Options{}
	for _, applyOption := range options {
		applyOption(opt)
	}

	return opt
}

func WithRetryOption(maxRetries int) func(option *Options) {
	return func(option *Options) {
		option.retry = &RetryOption{
			max: maxRetries,
		}
	}
}

type Event interface {
	Payload() []byte
	Name() string
}
