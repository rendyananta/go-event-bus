package bus

import (
	"context"
	"runtime"
)

// Registrar have function to hold an Event and send the emitted Event
// to the Event listener
type Registrar struct {
	channel        chan EventBus
	eventListeners map[string][]Listener
	options        *Options
}

// Options define the registrar behaviour on controlling
// callback, worker, etc.
type Options struct {
	successCallback func(e EventBus)
	errorCallback   func(e EventBus, err error)
	maxWorker       int
}

// WithOptions is function to build event registrar options.
func WithOptions(options ...func(options *Options)) *Options {
	o := &Options{}

	for _, applyOption := range options {
		applyOption(o)
	}

	return o
}

// WithSuccessCallback is function that return a function to set a success callback.
// Will be executed on the registrar options builder (WithOptions).
func WithSuccessCallback(callback func(e EventBus)) func(options *Options) {
	return func(options *Options) {
		options.successCallback = callback
	}
}

// WithErrorCallback is function that return a function to set an error callback.
// Will be executed on the registrar options builder (WithOptions).
func WithErrorCallback(callback func(e EventBus, err error)) func(options *Options) {
	return func(options *Options) {
		options.errorCallback = callback
	}
}

// Init event registrar with default options.
func Init() {
	registrar = &Registrar{
		channel:        make(chan EventBus),
		eventListeners: make(map[string][]Listener),
		options: &Options{
			maxWorker: runtime.NumCPU(),
		},
	}

	registrar.listen()
}

// InitWithOptions is a function to initialize event registrar
// with provided options available.
func InitWithOptions(opt *Options) {
	if opt.maxWorker == 0 {
		opt.maxWorker = runtime.NumCPU()
	}

	registrar = &Registrar{
		channel:        make(chan EventBus),
		eventListeners: make(map[string][]Listener),
		options:        opt,
	}

	registrar.listen()
}

// EventBus is a struct that hold an event with its listener
type EventBus struct {
	Event     Event
	listeners []Listener
	options   *EmitOptions
}

// registrar is an event registrar singleton instance.
var registrar *Registrar

// emit function accept an EventBus to send a signal towards event channel.
func emit(event EventBus) {
	registrar.channel <- event
}

// listen is the main function and logic to watch the emitted event
// it will propagate the listener to ran independently using goroutines.
func (r *Registrar) listen() {
	for i := 0; i < r.options.maxWorker; i++ {
		go func() {
			for bus := range r.channel {
				o := bus.options

				if o != nil && o.retry != nil {
					if o.retry.count >= o.retry.max {
						return
					}
				}

				retryEvent := func(position int, bus EventBus) {
					// mark as not retryable if the config was not set.
					if o.retry == nil && o.retry.max == 0 {
						return
					}

					o.retry.count += 1
					o.retry.fromPosition = position

					emit(bus)
				}

				for i, listener := range bus.listeners {
					go func(i int, eventListener Listener) {
						err := eventListener(context.Background(), bus.Event)

						if err == nil {
							// invoke success callback
							if registrar.options != nil && registrar.options.successCallback != nil {
								registrar.options.successCallback(bus)
							}

							return
						}

						// retry if err is not nil
						go retryEvent(i, bus)

						// invoke any callback
						// we may need to save in the database or log
						// or any actions preferred.
						if registrar.options != nil && registrar.options.errorCallback != nil {
							registrar.options.errorCallback(bus, err)
						}
					}(i, listener)
				}
			}
		}()
	}
}

// Emit public function to send an event into worker queue.
func Emit(e Event) {
	emit(EventBus{
		Event:     e,
		listeners: collectEventListeners(e.Name()),
	})
}

// EmitWithOptions public function to send an event into worker queue, but with a specific options.
func EmitWithOptions(e Event, o *EmitOptions) {
	emit(EventBus{
		Event:     e,
		listeners: collectEventListeners(e.Name()),
		options:   o,
	})
}

// collectEventListeners is a separated function to find a listener
// from an emitted event.
func collectEventListeners(eventName string) []Listener {
	if listeners, ok := registrar.eventListeners[eventName]; ok {
		return listeners
	}

	return nil
}

// RegisterListener is function to register the event with its listener into
// the event registrar instance.
func RegisterListener(eventName string, handlers ...Listener) {
	registrar.eventListeners[eventName] = handlers
}
