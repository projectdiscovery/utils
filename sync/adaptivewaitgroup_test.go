package sync

import (
	"context"
	"sync/atomic"
	"testing"

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
			require.False(t, swg.Current() > 4, "not the good amount of routines spawned.", swg.Current())
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
