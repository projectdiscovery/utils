package batcher

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBatcherStandard(t *testing.T) {
	var (
		batchSize        = 100
		wanted           = 100000
		minWantedBatches = wanted / batchSize
		got              int
		gotBatches       int
	)
	callback := func(t []int) {
		gotBatches++
		for range t {
			got++
		}
	}
	bat := New[int](
		WithMaxCapacity[int](batchSize),
		WithFlushCallback[int](callback),
	)

	bat.Run()

	for i := 0; i < wanted; i++ {
		bat.Append(i)
	}

	bat.Stop()

	bat.WaitDone()

	require.Equal(t, wanted, got)
	require.True(t, minWantedBatches <= gotBatches)
}

func TestBatcherWithInterval(t *testing.T) {
	var (
		batchSize        = 200
		wanted           = 1000
		minWantedBatches = 10
		got              int
		gotBatches       int
	)
	callback := func(t []int) {
		gotBatches++
		for range t {
			got++
		}
	}
	bat := New[int](
		WithMaxCapacity[int](batchSize),
		WithFlushCallback[int](callback),
		WithFlushInterval[int](10*time.Millisecond),
	)

	bat.Run()

	for i := 0; i < wanted; i++ {
		time.Sleep(2 * time.Millisecond)
		bat.Append(i)
	}

	bat.Stop()

	bat.WaitDone()

	require.Equal(t, wanted, got)
	require.True(t, minWantedBatches <= gotBatches)
}

type exampleBatcherStruct struct {
	Value []byte
}

func TestBatcherWithSizeLimit(t *testing.T) {
	var (
		batchSize  = 100
		maxSize    = 1000
		wanted     = 10
		gotBatches int
	)
	var failedIteration bool

	callback := func(ta []exampleBatcherStruct) {
		gotBatches++

		if len(ta) != 5 {
			failedIteration = true
		}
	}
	bat := New[exampleBatcherStruct](
		WithMaxCapacity[exampleBatcherStruct](batchSize),
		WithMaxSize[exampleBatcherStruct](int32(maxSize)),
		WithFlushCallback[exampleBatcherStruct](callback),
	)

	bat.Run()

	for i := 0; i < wanted; i++ {
		randData := make([]byte, 200)
		_, _ = rand.Read(randData)
		bat.Append(exampleBatcherStruct{Value: randData})
	}

	bat.Stop()

	bat.WaitDone()

	require.Equal(t, 2, gotBatches)
	require.False(t, failedIteration)
}
