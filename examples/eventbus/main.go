package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/iqhater/pkg/async/eventbus"
)

// EventBus examples
func main() {

	basicUsage()
	concurrencyUsage()
	subscriptionInGoroutine()
}

func basicUsage() {

	fmt.Printf("Basic Example Usage #1\n\n")

	bus := eventbus.New()

	unsubscribe1 := bus.Subscribe("user.created", func(data any) {
		fmt.Println("Handler 1: user created:", data)
	})
	defer unsubscribe1()

	unsubscribe2 := bus.Subscribe("user.deleted", func(data any) {
		fmt.Println("Handler 2: user deleted:", data)
	})
	defer unsubscribe2()

	bus.Publish("user.created", "Bob")
	bus.Publish("user.deleted", "Bob")

	fmt.Println()
}

func concurrencyUsage() {

	fmt.Printf("Concurrency Example Usage #2\n\n")

	bus := eventbus.New()

	unsubscribe1 := bus.Subscribe("user.created", func(data any) {
		fmt.Println("Data received from user.created:", data)
	})
	defer unsubscribe1()

	unsubscribe2 := bus.Subscribe("cart.added", func(data any) {
		fmt.Println("Data received from cart.added:", data)
	})
	defer unsubscribe2()

	var wg sync.WaitGroup

	wg.Go(func() {
		time.Sleep(time.Millisecond * 250)
		bus.Publish("user.created", "Event data from 1 goroutine...")
	})

	wg.Go(func() {
		time.Sleep(time.Millisecond * 500)
		bus.Publish("cart.added", "Event data from 2 goroutine...")
	})
	wg.Wait()

	fmt.Println()
}

func subscriptionInGoroutine() {

	fmt.Printf("Subscription in Goroutine Example Usage #3\n\n")

	bus := eventbus.New()

	var wg sync.WaitGroup
	ch := make(chan func(), 1)

	wg.Go(func() {

		unsubscribe3 := bus.Subscribe("user.created", func(data any) {
			fmt.Println("Data received from user.created::", data)
		})
		time.Sleep(time.Millisecond * 250)
		ch <- unsubscribe3
	})
	wg.Wait()

	unsubscribe3 := <-ch
	close(ch)

	bus.Publish("user.created", "Event data into goroutine...")

	unsubscribe3()

	fmt.Println()
}
