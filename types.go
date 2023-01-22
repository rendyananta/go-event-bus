package bus

import "context"

// Event public interface. all event must implement this interface.
type Event interface {
	Payload() []byte
	Name() string
}

// State struct to save an event state.
type State struct {
	retryable     bool
	retryCount    int
	retryPosition int
}

// EventBus is a struct that hold an event with its listener
type EventBus struct {
	Event     Event
	listeners []Listener
	state     *State
}

// Listener type alias for a function that accept event and error.
type Listener func(ctx context.Context, e Event) error

// RetryOption struct to hold an event state.
type RetryOption struct {
	Max int
}

// EmitOptions struct that collects all possible value for
// event emit options
type EmitOptions struct {
	Retry *RetryOption
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
		option.Retry = &RetryOption{
			Max: maxRetries,
		}
	}
}
