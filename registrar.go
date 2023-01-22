package bus

import "context"

// Registrar have function to hold an Event and send the emitted Event
// to the Event listener
type Registrar struct {
	channel        chan EventBus
	eventListeners map[string][]Listener
	options        *Options
}

type Options struct {
	successCallback func(e EventBus)
	errorCallback   func(e EventBus, err error)
}

func WithOptions(options ...func(options *Options)) *Options {
	o := &Options{}

	for _, applyOption := range options {
		applyOption(o)
	}

	return o
}

func WithSuccessCallback(callback func(e EventBus)) func(options *Options) {
	return func(options *Options) {
		options.successCallback = callback
	}
}

func WithErrorCallback(callback func(e EventBus, err error)) func(options *Options) {
	return func(options *Options) {
		options.errorCallback = callback
	}
}

func Init() {
	instance = &Registrar{
		channel:        make(chan EventBus),
		eventListeners: make(map[string][]Listener),
	}

	instance.listenEvents()
}

func InitWithOptions(opt *Options) {
	instance = &Registrar{
		channel:        make(chan EventBus),
		eventListeners: make(map[string][]Listener),
		options:        opt,
	}

	instance.listenEvents()
}

type EventBus struct {
	Event     Event
	listeners []Listener
	options   *EmitOptions
}

var (
	instance *Registrar
)

func emit(event EventBus) {
	instance.channel <- event
}

func (r *Registrar) listenEvents() {
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
				err := listener(context.Background(), bus.Event)

				if err == nil {
					// invoke success callback
					if instance.options != nil && instance.options.successCallback != nil {
						instance.options.successCallback(bus)
					}

					continue
				}

				// retry if err is not nil
				go retryEvent(i, bus)

				// invoke any callback
				// we may need to save in the database or log
				// or any actions preferred.
				if instance.options != nil && instance.options.errorCallback != nil {
					instance.options.errorCallback(bus, err)
				}
			}
		}
	}()
}

func Emit(e Event) {
	emit(EventBus{
		Event:     e,
		listeners: collectEventListeners(e.Name()),
	})
}

func EmitWithOptions(e Event, o *EmitOptions) {
	emit(EventBus{
		Event:     e,
		listeners: collectEventListeners(e.Name()),
		options:   o,
	})
}

func collectEventListeners(eventName string) []Listener {
	if listeners, ok := instance.eventListeners[eventName]; ok {
		return listeners
	}

	return nil
}

func RegisterListener(eventName string, handlers ...Listener) {
	instance.eventListeners[eventName] = handlers
}
