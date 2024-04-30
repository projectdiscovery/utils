package sync

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/require"
)

// tests from https://github.com/remeh/sizedwaitgroup/blob/master/sizedwaitgroup_test.go

func TestWait(t *testing.T) {
	swg, err := New(WithSize(10))
	require.Nil(t, err)

	var c uint32

	for i := 0; i < 10000; i++ {
		swg.Add()
		go func(c *uint32) {
			defer swg.Done()
			atomic.AddUint32(c, 1)
		}(&c)
	}

	swg.Wait()

	if c != 10000 {
		t.Fatalf("%d, not all routines have been executed.", c)
	}
}

func TestThrottling(t *testing.T) {
	var c atomic.Uint32

	swg, err := New(WithSize(4))
	require.Nil(t, err)

	if swg.Current() != 0 {
		t.Fatalf("the SizedWaitGroup should start with zero.")
	}

	for i := 0; i < 10000; i++ {
		swg.Add()
		go func() {
			defer swg.Done()

			c.Add(1)
			require.False(t, swg.Current() > 5, "not the good amount of routines spawned.", swg.Current())
		}()
	}

	swg.Wait()
}

func TestNoThrottling(t *testing.T) {
	var c atomic.Int32
	swg, err := New(WithSize(1))
	require.Nil(t, err)

	if swg.Current() != 0 {
		t.Fatalf("the SizedWaitGroup should start with zero.")
	}
	for i := 0; i < 10000; i++ {
		swg.Add()
		go func() {
			defer swg.Done()
			c.Add(1)
		}()
	}
	swg.Wait()
	if c.Load() != 10000 {
		t.Fatalf("%d, not all routines have been executed.", c.Load())
	}
}

func TestAddWithContext(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	swg, err := New(WithSize(1))
	require.Nil(t, err)

	if err := swg.AddWithContext(ctx); err != nil {
		t.Fatalf("AddContext returned error: %v", err)
	}

	cancelFunc()
	if err := swg.AddWithContext(ctx); err != context.Canceled {
		t.Fatalf("AddContext returned non-context.Canceled error: %v", err)
	}

}

func TestMultipleResizes(t *testing.T) {
	var c atomic.Int32
	swg, err := New(WithSize(2)) // Start with a size of 2
	require.Nil(t, err)

	for i := 0; i < 10000; i++ {
		if i == 250 {
			err := swg.Resize(context.TODO(), 5) // Increase size at 2500th iteration
			require.Nil(t, err)
		}
		if i == 500 {
			err := swg.Resize(context.TODO(), 1) // Decrease size at 5000th iteration
			require.Nil(t, err)
		}
		if i == 750 {
			err := swg.Resize(context.TODO(), 3) // Increase size again at 7500th iteration
			require.Nil(t, err)
		}

		swg.Add()
		go func() {
			defer swg.Done()
			c.Add(1)
		}()
	}

	swg.Wait()
	if c.Load() != 10000 {
		t.Fatalf("%d, not all routines have been executed.", c.Load())
	}
}

func Test_AdaptiveWaitGroup_Leak(t *testing.T) {
	defer leaktest.Check(t)()

	for j := 0; j < 1000; j++ {
		wg, err := New(WithSize(10))
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 10000; i++ {
			wg.Add()
			go func(awg *AdaptiveWaitGroup) {
				defer awg.Done()
			}(wg)
		}
		wg.Wait()
	}
}

func Test_AdaptiveWaitGroup_ContinuousResizeAndCheck(t *testing.T) {
	defer leaktest.Check(t)() // Ensure no goroutines are leaked

	var c atomic.Int32

	wg, err := New(WithSize(1)) // Start with a size of 1
	if err != nil {
		t.Fatal(err)
	}

	// Perform continuous resizing and goroutine execution
	for j := 0; j < 100; j++ {
		for i := 0; i < 1000; i++ {
			wg.Add()
			go func(awg *AdaptiveWaitGroup) {
				defer awg.Done()
				c.Add(1)
			}(wg)
		}

		// Increase or decrease size
		newSize := (j % 10) + 1 // Cycle sizes between 1 and 10
		err := wg.Resize(context.TODO(), newSize)
		if err != nil {
			t.Fatalf("Resize returned error: %v", err)
		}

		wg.Wait() // Wait at each step to ensure all routines finish before resizing again
	}

	if c.Load() != 100000 {
		t.Fatalf("%d, not all routines have been executed.", c.Load())
	}
}
