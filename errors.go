package bus

import "errors"

var (
	ErrEventAlreadyRegistered = errors.New("event already registered")
	ErrEventNotRegistered     = errors.New("event is not registered")
)
