package bus

import "context"

// Registrar have function to hold an event and send the emitted event
// to the event listener
type Registrar struct {
	channel        chan EventBus
	quit           chan bool
	eventListeners map[string][]Listener
}

type EventBus struct {
	event     Event
	listeners []Listener
	options   *Options
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
				err := listener(context.Background(), bus.event)

				if err != nil {
					go retryEvent(i, bus)

					// invoke any callback
					// we may need to save in the database or log
					// or any actions preferred.
				}
			}
		}
	}()
}

func InitRegistrar() {
	instance = &Registrar{
		channel:        make(chan EventBus),
		eventListeners: make(map[string][]Listener),
	}

	instance.listenEvents()
}

func Emit(e Event) {
	emit(EventBus{
		event:     e,
		listeners: collectEventListeners(e.Name()),
	})
}

func EmitWithOptions(e Event, o *Options) {
	emit(EventBus{
		event:     e,
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
