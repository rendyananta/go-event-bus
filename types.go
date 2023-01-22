package bus

import "context"

type Listener func(ctx context.Context, e Event) error

type RetryOption struct {
	max          int
	count        int
	fromPosition int
}

type EmitOptions struct {
	retry *RetryOption
}

func WithEmitOptions(options ...func(options *EmitOptions)) *EmitOptions {
	opt := &EmitOptions{}
	for _, applyOption := range options {
		applyOption(opt)
	}

	return opt
}

func WithRetryOption(maxRetries int) func(option *EmitOptions) {
	return func(option *EmitOptions) {
		option.retry = &RetryOption{
			max: maxRetries,
		}
	}
}

type Event interface {
	Payload() []byte
	Name() string
}
