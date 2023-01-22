package bus

import "errors"

var (
	ErrEventAlreadyRegistered = errors.New("Event already registered")
	ErrEventNotRegistered     = errors.New("Event is not registered")
)
