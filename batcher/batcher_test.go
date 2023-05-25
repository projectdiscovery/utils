package batcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBatcher(t *testing.T) {
	var (
		batchSize        = 100
		wanted           = 100000
		minWantedBatches = wanted / batchSize
		got              int
		gotBatches       int
	)
	bat := New(batchSize, time.Second, func(t []int) {
		gotBatches++
		for range t {
			got++
		}
	})

	bat.Run()

	for i := 0; i < wanted; i++ {
		bat.Append(i)
	}

	bat.Stop()

	bat.WaitDone()

	require.Equal(t, wanted, got)
	require.True(t, minWantedBatches <= gotBatches)
}
