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
	bus.InitRegistrar()

	bus.RegisterListener("order-delivered", func(ctx context.Context, e bus.Event) error {
		fmt.Println("Hello from bus: ", e.Name(), " with payload: ", string(e.Payload()))

		event := OrderDeliveredEvent{}

		_ = json.Unmarshal(e.Payload(), &event)

		if event.Total%2 != 0 {
			return errors.New("odd number")
		}

		return nil
	}, func(ctx context.Context, e bus.Event) error {
		return nil
	})

	waitChan := make(chan bool)

	go func() {
		for {
			log.Printf("[main] Total current goroutine: %d", runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}
	}()

	order := OrderDeliveredEvent{
		Customer: "Rendy",
		Total:    50001,
	}

	bus.EmitWithOptions(order, bus.WithOptions(bus.WithRetryOption(3)))

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