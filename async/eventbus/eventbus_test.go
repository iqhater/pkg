package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestSubscribeReturnsUniqueIDs(t *testing.T) {

	// Arrange
	eb := New()

	handler1 := func(data any) {}
	handler2 := func(data any) {}

	// Act
	id1 := eb.Subscribe("test", handler1)
	id2 := eb.Subscribe("test", handler2)

	// Assert
	if id1 == id2 {
		t.Error("Subscribe must return unique IDs for each subscription")
	}
}

func TestPublishCallsAllHandlers(t *testing.T) {

	// Arrange
	eb := New()

	var results []int
	var mu sync.Mutex

	handler1 := func(data any) {
		mu.Lock()
		results = append(results, data.(int))
		mu.Unlock()
	}

	handler2 := func(data any) {
		mu.Lock()
		results = append(results, data.(int)*2)
		mu.Unlock()
	}

	// Act
	eb.Subscribe("test", handler1)
	eb.Subscribe("test", handler2)

	eb.Publish("test", 5)

	// Assert
	mu.Lock()
	defer mu.Unlock()

	if len(results) != 2 {
		t.Errorf("Wait 2 results, got: %d", len(results))
	}
}

func TestUnsubscribeRemovesHandler(t *testing.T) {

	// Arrange
	eb := New()

	called := false
	handler := func(data any) {
		called = true
	}

	// Act
	id := eb.Subscribe("test", handler)
	eb.Unsubscribe("test", id)
	eb.Publish("test", nil)

	// Assert
	if called {
		t.Error("Handler should not be called after unsubscribe")
	}
}

func TestUnsubscribeNonExistentEvent(t *testing.T) {

	// Arrange
	eb := New()

	// Act/Assert
	// should not panic
	eb.Unsubscribe("nonexistent", 999)
}

func TestConcurrentAccess(t *testing.T) {

	// Arrange
	eb := New()

	var wg sync.WaitGroup
	results := make([]int, 0, 100)
	var mu sync.Mutex

	// Act
	// Concurrent subscriptions
	for i := range 10 {

		wg.Go(func() {
			eb.Subscribe("concurrent", func(data any) {
				mu.Lock()
				results = append(results, i)
				mu.Unlock()
			})
		})
	}

	// Concurrent publications
	for _ = range 10 {

		wg.Go(func() {
			eb.Publish("concurrent", 1)
		})
	}
	wg.Wait()
	time.Sleep(3 * time.Millisecond)

	// Assert
	mu.Lock()
	defer mu.Unlock()

	if len(results) == 0 {
		t.Error("No results were received, expected at least one")
	}
}
