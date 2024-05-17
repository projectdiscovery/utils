package sliceutil

import (
	"sync"
	"testing"
	"time"
)

func TestSimpleUsage(t *testing.T) {
	ss := NewSyncSlice[int]()
	expected := 10
	for i := 0; i < expected; i++ {
		ss.Append(i)
	}
	value, ok := ss.Get(5)
	if !ok {
		t.Errorf("Failed to get value at index 5")
	} else if value != 5 {
		t.Errorf("Expected value 5 at index 5, got %d", value)
	}

	success := ss.Put(5, 20)
	if !success {
		t.Errorf("Failed to put value at index 5")
	}

	value, ok = ss.Get(5)
	if !ok {
		t.Errorf("Failed to get value at index 5 after put")
	} else if value != 20 {
		t.Errorf("Expected value 20 at index 5 after put, got %d", value)
	}
	if ss.Len() != expected {
		t.Errorf("Expected slice length %d, got %d", expected, ss.Len())
	}
	ss.Empty()
	if ss.Len() != 0 {
		t.Errorf("Expected slice length 0 after emptying, got %d", ss.Len())
	}
}

func TestConcurrentAppend(t *testing.T) {
	ss := NewSyncSlice[int]()
	var wg sync.WaitGroup
	count := 1000

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ss.Append(val)

			if val%10 == 0 {
				ss.Put(val, val*2) // Double the value at positions that are multiples of 10
			}
			if val%5 == 0 {
				retrievedVal, _ := ss.Get(val) // Attempt to get the value at positions that are multiples of 5
				_ = retrievedVal               // Use the retrieved value to ensure it's not optimized away
			}
		}(i)
	}
	wg.Wait()

	if ss.Len() != count {
		t.Errorf("Expected slice length %d after concurrent append, got %d", count, ss.Len())
	}
}

func TestConcurrentReadWriteAndIteration(t *testing.T) {
	ss := NewSyncSlice[int]()
	var wg sync.WaitGroup
	readWriteCount := 1000

	wg.Add(3) // Adding three groups: writer, reader, iterator

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < readWriteCount; i++ {
			ss.Append(i) // Write
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()

		time.Sleep(250 * time.Millisecond)

		for i := 0; i < readWriteCount; i++ {
			if value, ok := ss.Get(i % ss.Len()); !ok {
				t.Errorf("Failed to get value at index %d", i%ss.Len())
			} else {
				_ = value // Use the value to ensure it's not optimized away
			}
		}
	}()

	// Iterator goroutine
	go func() {
		defer wg.Done()
		for repeat := 0; repeat < 1000; repeat++ { // Repeat the iteration 1000 times
			ss.Each(func(index int, value int) error {
				// Simulate some processing
				_ = index
				_ = value
				return nil
			})
		}
	}()

	wg.Wait()

	if ss.Len() != readWriteCount {
		t.Errorf("Expected slice length %d after concurrent read/write, got %d", readWriteCount, ss.Len())
	}
}
