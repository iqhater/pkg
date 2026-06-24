package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestSubscribeReturnsUniqueIDs(t *testing.T) {

	// Arrange
	eb := New()

	handler1 := func(data any) {}
	handler2 := func(data any) {}

	// Act
	unsub1 := eb.Subscribe("test", handler1)
	id1 := eb.nextID
	unsub2 := eb.Subscribe("test", handler2)
	id2 := eb.nextID

	unsub1()
	unsub2()

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
	unsub1 := eb.Subscribe("test", handler1)
	unsub2 := eb.Subscribe("test", handler2)

	eb.Publish("test", 5)

	// Assert
	mu.Lock()
	defer mu.Unlock()

	unsub1()
	unsub2()

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
	unsubscribe := eb.Subscribe("test", handler)
	unsubscribe()
	eb.Publish("test", nil)

	// Assert
	if called {
		t.Error("Handler should not be called after unsubscribe")
	}
}

func TestConcurrentAccess(t *testing.T) {

	synctest.Test(t, func(t *testing.T) {

		// Arrange
		eb := New()

		var wg sync.WaitGroup
		results := make(chan int, 100)

		var calls atomic.Int64

		// Act
		// Concurrent subscriptions
		for _ = range 10 {

			wg.Go(func() {
				_ = eb.Subscribe("concurrent", func(data any) {
					time.Sleep(100 * time.Millisecond)
					results <- data.(int)
					calls.Add(1)
				})
			})
		}

		// Concurrent publications
		for i := range 10 {

			wg.Go(func() {
				time.Sleep(50 * time.Millisecond)
				eb.Publish("concurrent", i)
			})
		}
		wg.Wait()
		synctest.Wait()

		close(results)

		sum := 0
		count := 0

		for v := range results {
			sum += v
			count++
		}

		// Assert
		if got := calls.Load(); got != 100 {
			t.Fatalf("expected 100 calls, got %d", got)
		}

		if count != 100 {
			t.Fatalf("expected 100 results, got %d", count)
		}

		// 10 subscribers sum = 450
		if sum != 450 {
			t.Fatalf("expected sum 450, got %d", sum)
		}
	})
}

func TestSubscribeRaceWithPublish(t *testing.T) {

	synctest.Test(t, func(t *testing.T) {

		// Arrange
		eb := New()
		var got atomic.Int64
		var wg sync.WaitGroup

		// Act
		wg.Go(func() {
			for range 100 {
				eb.Subscribe("x", func(data any) {
					got.Add(1)
				})
			}
		})
		wg.Wait()
		synctest.Wait()

		for range 100 {
			eb.Publish("x", nil)
		}
		time.Sleep(2 * time.Second)

		// Assert
		if got.Load() == 0 {
			t.Fatal("expected some calls")
		}
	})
}

func TestConcurrentPublishAndUnsubscribe(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		eb := New()

		var wg sync.WaitGroup

		unsub := eb.Subscribe("x", func(any) {
			time.Sleep(100 * time.Millisecond)
		})

		wg.Go(func() {
			eb.Publish("x", nil)
		})

		wg.Go(func() {
			unsub()
		})

		wg.Wait()
		synctest.Wait()
	})
}

func TestDoubleUnsubscribeIsSafe(t *testing.T) {

	// Arrange
	eb := New()

	// Act/Assert
	unsub := eb.Subscribe("x", func(any) {})

	unsub()
	unsub() // should not panic
}

func TestEventIsolation(t *testing.T) {

	// Arrange
	eb := New()
	var a, b atomic.Int64

	// Act
	eb.Subscribe("a", func(any) { a.Add(1) })
	eb.Subscribe("b", func(any) { b.Add(1) })

	eb.Publish("a", nil)

	// Assert
	if b.Load() != 0 {
		t.Fatal("event leakage between topics")
	}
}

func TestUnsubscribeRemovesFromMemory(t *testing.T) {

	// Arrange
	eb := New()

	// Act
	unsub := eb.Subscribe("x", func(any) {})
	unsub()

	// Assert
	if len(eb.subscribers["x"]) != 0 {
		t.Fatal("subscriber was not removed")
	}
}
