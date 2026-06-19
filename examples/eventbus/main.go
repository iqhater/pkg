package main

import (
	"fmt"

	"github.com/iqhater/pkg/async/eventbus"
)

// EventBus example
func main() {

	bus := eventbus.New()

	id1 := bus.Subscribe("user.created", func(data any) {
		fmt.Println("Handler 1: user created:", data)
	})

	id2 := bus.Subscribe("user.deleted", func(data any) {
		fmt.Println("Handler 2: user deleted:", data)
	})

	bus.Publish("user.created", "Bob")
	bus.Publish("user.deleted", "Bob")

	bus.Unsubscribe("user.created", id1)
	bus.Unsubscribe("user.deleted", id2)
}
