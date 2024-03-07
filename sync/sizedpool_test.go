package sync

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testStruct struct{}

func TestSizedPool(t *testing.T) {
	p := &sync.Pool{
		New: func() any {
			return &testStruct{}
		},
	}

	// Create a new SizedPool with a max capacity of 2
	pool, err := New[*testStruct](
		WithMaxCapacity[*testStruct](2),
		WithPool[*testStruct](p),
	)
	if err != nil {
		t.Errorf("Error creating pool: %v", err)
	}

	// Test Get and Put operations
	ctx := context.Background()
	obj1, _ := pool.Get(ctx)
	obj2, _ := pool.Get(ctx)

	go func() {
		time.Sleep(3 * time.Second)
		pool.Put(obj1)
		time.Sleep(1 * time.Second)
		pool.Put(obj2)
	}()

	start := time.Now()
	obj3, _ := pool.Get(ctx)
	require.WithinDuration(t, start.Add(3*time.Second), time.Now(), 500*time.Millisecond)
	require.NotNil(t, obj3)
}
