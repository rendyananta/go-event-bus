package bus

import (
	"context"
	"runtime"
)

// Registrar have function to hold an Event and send the emitted Event
// to the Event listener
type Registrar struct {
	channel          chan EventBus
	eventListeners   map[string][]Listener
	eventEmitOptions map[string]*EmitOptions
	options          *Options
}

// Options define the registrar behaviour on controlling
// callback, worker, etc.
type Options struct {
	successCallback func(e EventBus, listenerPosition int)
	errorCallback   func(e EventBus, listenerPosition int, err error)
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
func WithSuccessCallback(callback func(e EventBus, listenerPosition int)) func(options *Options) {
	return func(options *Options) {
		options.successCallback = callback
	}
}

// WithErrorCallback is function that return a function to set an error callback.
// Will be executed on the registrar options builder (WithOptions).
func WithErrorCallback(callback func(e EventBus, listenerPosition int, err error)) func(options *Options) {
	return func(options *Options) {
		options.errorCallback = callback
	}
}

// Init event registrar with default options.
func Init() {
	registrar = &Registrar{
		channel:          make(chan EventBus),
		eventListeners:   make(map[string][]Listener),
		eventEmitOptions: make(map[string]*EmitOptions),
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
		channel:          make(chan EventBus),
		eventListeners:   make(map[string][]Listener),
		eventEmitOptions: make(map[string]*EmitOptions),
		options:          opt,
	}

	registrar.listen()
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

				if bus.state == nil {
					bus.state = &State{}

					o := getEventEmitOptions(bus.Event.Name())

					if o.Retry != nil && o.Retry.Max != 0 {
						bus.state.retryable = true
					}
				}

				// mark the retryPosition as not set yet by using -1.
				if bus.state.retryCount == 0 {
					bus.state.retryPosition = -1
				}

				propagateListener(bus)
			}
		}()
	}
}

func propagateListener(bus EventBus) {
	listen := func(i int, listen Listener) {
		err := listen(context.Background(), bus.Event)

		if err == nil {
			// invoke success callback
			if registrar.options != nil && registrar.options.successCallback != nil {
				registrar.options.successCallback(bus, i)
			}

			return
		}

		// Retry if err is not nil
		go retryEvent(i, bus)

		// invoke any callback
		// we may need to save in the database or log
		// or any actions preferred.
		if registrar.options != nil && registrar.options.errorCallback != nil {
			registrar.options.errorCallback(bus, i, err)
		}
	}

	for i, listener := range bus.listeners {
		if bus.state.retryPosition != -1 && bus.state.retryPosition == i {
			go listen(i, listener)
		}

		if bus.state.retryPosition == -1 {
			go listen(i, listener)
		}
	}
}

func retryEvent(i int, bus EventBus) {
	if bus.state.retryable == false {
		return
	}

	o := getEventEmitOptions(bus.Event.Name())

	bus.state.retryCount += 1
	bus.state.retryPosition = i

	if bus.state.retryCount <= o.Retry.Max {
		emit(bus)
	}
}

// Emit public function to send an event into worker queue.
func Emit(e Event) {
	emit(EventBus{
		Event:     e,
		listeners: getEventListeners(e.Name()),
	})
}

// getEventListeners is a separated function to find a Listener
// from an emitted event.
func getEventListeners(eventName string) []Listener {
	if listeners, ok := registrar.eventListeners[eventName]; ok {
		return listeners
	}

	return nil
}

// getEventListeners is a separated function to find an event EmitOptions
// from an emitted event.
func getEventEmitOptions(eventName string) *EmitOptions {
	if options, ok := registrar.eventEmitOptions[eventName]; ok {
		return options
	}

	return nil
}

// RegisterListener is function to register the event with its listener into
// the event registrar instance.
func RegisterListener(eventName string, options *EmitOptions, handlers ...Listener) {
	registrar.eventListeners[eventName] = handlers
	registrar.eventEmitOptions[eventName] = options
}
