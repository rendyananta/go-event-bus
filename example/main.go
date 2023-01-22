package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bus "github.com/rendyananta/go-event-bus"
	"log"
	"runtime"
	"time"
)

func main() {
	go func() {
		for {
			log.Printf("[main] Total current goroutine: %d", runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}
	}()

	bus.InitWithOptions(
		bus.WithOptions(
			bus.WithSuccessCallback(func(e bus.EventBus) {
				fmt.Println("success callback for event: ", e.Event.Name())
			}),
			bus.WithErrorCallback(func(e bus.EventBus, err error) {
				fmt.Println("error callback for event: ", e.Event.Name(), ", error: ", err.Error())
			}),
		),
	)

	bus.RegisterListener("order-delivered", func(ctx context.Context, e bus.Event) error {
		fmt.Println("Hello from bus: ", e.Name(), " with payload: ", string(e.Payload()))

		event := OrderDeliveredEvent{}

		_ = json.Unmarshal(e.Payload(), &event)

		if event.Total%2 != 0 {
			return errors.New("odd number")
		}

		return nil
	})

	waitChan := make(chan bool)

	time.Sleep(1 * time.Second)

	order := OrderDeliveredEvent{
		Customer: "Rendy",
		Total:    50001,
	}

	bus.EmitWithOptions(order, bus.WithEmitOptions(bus.WithRetryOption(3)))

	order2 := OrderDeliveredEvent{
		Customer: "Rendy 2",
		Total:    10000,
	}

	bus.Emit(order2)

	waitChan <- true
}

type OrderDeliveredEvent struct {
	Customer string `json:"customer"`
	Total    int    `json:"total"`
}

func (o OrderDeliveredEvent) Payload() []byte {
	payload, err := json.Marshal(o)

	if err != nil {
		return []byte("'")
	}

	return payload
}

func (o OrderDeliveredEvent) Name() string {
	return "order-delivered"
}
