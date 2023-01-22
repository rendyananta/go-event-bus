package bus

import "context"

// Listener type alias for a function that accept event and error.
type Listener func(ctx context.Context, e Event) error

// RetryOption struct to hold an event state.
type RetryOption struct {
	max          int
	count        int
	fromPosition int
}

// EmitOptions struct that collects all possible value for
// event emit options
type EmitOptions struct {
	retry *RetryOption
}

// WithEmitOptions builder function to build an event EmitOptions struct.
func WithEmitOptions(options ...func(options *EmitOptions)) *EmitOptions {
	opt := &EmitOptions{}
	for _, applyOption := range options {
		applyOption(opt)
	}

	return opt
}

// WithRetryOption function that can be used inside WithEmitOptions function builder.
func WithRetryOption(maxRetries int) func(option *EmitOptions) {
	return func(option *EmitOptions) {
		option.retry = &RetryOption{
			max: maxRetries,
		}
	}
}

// Event public interface. all event must implement this interface.
type Event interface {
	Payload() []byte
	Name() string
}
