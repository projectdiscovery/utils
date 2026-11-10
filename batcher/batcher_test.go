package batcher

import (
	"crypto/rand"
	"strconv"
	"sync"
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

		totalValueSize := 0
		for _, t := range ta {
			totalValueSize += len(t.Value)
		}
		if totalValueSize > maxSize {
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

	require.False(t, failedIteration)
}

// lifecycleTimeout bounds every lifecycle step so a regression fails the test
// instead of hanging the package.
const lifecycleTimeout = 5 * time.Second

// repeatedStops is how many times a scenario calls Stop in a row.
const repeatedStops = 5

// concurrentStoppers is how many goroutines call Stop at once.
const concurrentStoppers = 32

// within fails the test if step does not return within lifecycleTimeout.
func within(t *testing.T, name string, step func()) {
	t.Helper()

	finished := make(chan struct{})

	go func() {
		defer close(finished)

		step()
	}()

	select {
	case <-finished:
	case <-time.After(lifecycleTimeout):
		t.Fatalf("%s did not return within %s", name, lifecycleTimeout)
	}
}

// collector records flushed items. The flush callback only runs on the
// batcher's run goroutine, so reading counts after WaitDone is race-free.
type collector struct {
	counts map[int]int
}

func newCollector() *collector { return &collector{counts: map[int]int{}} }

func (col *collector) flush(items []int) {
	for _, item := range items {
		col.counts[item]++
	}
}

// requireExactlyOnce checks that items 0..total-1 were flushed exactly once.
func (col *collector) requireExactlyOnce(t *testing.T, total int) {
	t.Helper()
	require.Len(t, col.counts, total)

	for item := range total {
		require.Equal(t, 1, col.counts[item], "item %d", item)
	}
}

func newLifecycleBatcher(col *collector, opts ...BatcherOption[int]) *Batcher[int] {
	opts = append([]BatcherOption[int]{
		WithMaxCapacity[int](lifecycleCapacity),
		WithFlushCallback[int](col.flush),
	}, opts...)

	return New[int](opts...)
}

const (
	lifecycleCapacity = 10
	// lifecycleItems spans more than one batch and ends in a partial one.
	lifecycleItems = 25
)

// intSize is the size WithMaxSize accounts for one int item.
const intSize = strconv.IntSize / 8

// lifecycleDriver performs the Run, Stop and WaitDone calls of one scenario.
// appendItems appends the scenario's items and fails the test if that blocks.
// A driver must leave the batcher stopped.
type lifecycleDriver func(t *testing.T, bat *Batcher[int], appendItems func())

func stopRepeatedly(t *testing.T, bat *Batcher[int]) {
	t.Helper()

	within(t, "repeated Stop", func() {
		for range repeatedStops {
			bat.Stop()
		}
	})
}

func stopConcurrently(t *testing.T, bat *Batcher[int]) {
	t.Helper()

	within(t, "concurrent Stop", func() {
		var group sync.WaitGroup

		for range concurrentStoppers {
			group.Go(bat.Stop)
		}

		group.Wait()
	})
}

func driveRunAppendStopWait(_ *testing.T, bat *Batcher[int], appendItems func()) {
	bat.Run()
	appendItems()
	bat.Stop()
	bat.WaitDone()
}

func driveStopRepeatedAfterDone(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	bat.Run()
	appendItems()
	bat.Stop()
	bat.WaitDone()
	stopRepeatedly(t, bat)
}

func driveStopRepeatedBeforeDone(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	bat.Run()
	appendItems()
	stopRepeatedly(t, bat)
	bat.WaitDone()
}

func driveStopConcurrently(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	bat.Run()
	appendItems()
	stopConcurrently(t, bat)
	bat.WaitDone()
}

func driveRunTwiceStopTwice(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	bat.Run()
	bat.Run()
	appendItems()
	within(t, "Stop", func() {
		bat.Stop()
		bat.Stop()
	})
	bat.WaitDone()
}

func driveRunAfterDone(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	bat.Run()
	appendItems()
	bat.Stop()
	bat.WaitDone()
	within(t, "second Run and Stop", func() {
		bat.Run()
		bat.Stop()
	})
	bat.WaitDone()
}

// driveStopBeforeRun checks that a Run starting after Stop flushes what was
// buffered and exits.
func driveStopBeforeRun(t *testing.T, bat *Batcher[int], appendItems func()) {
	t.Helper()
	within(t, "Stop before Run", func() {
		bat.Stop()
		bat.Stop()
	})
	appendItems()
	bat.Run()
	bat.WaitDone()
}

func driveWaitDoneTwice(_ *testing.T, bat *Batcher[int], appendItems func()) {
	bat.Run()
	appendItems()
	bat.Stop()
	bat.WaitDone()
	bat.WaitDone()
}

type lifecycleScenario struct {
	name string
	// appended is how many items the driver appends. Without a run loop
	// Append can only fill the buffer, so it is smaller there.
	appended int
	drive    lifecycleDriver
}

func lifecycleScenarios() []lifecycleScenario {
	return []lifecycleScenario{
		{name: "run append stop wait", appended: lifecycleItems, drive: driveRunAppendStopWait},
		{name: "stop repeated after done", appended: lifecycleItems, drive: driveStopRepeatedAfterDone},
		{name: "stop repeated before done", appended: lifecycleItems, drive: driveStopRepeatedBeforeDone},
		{name: "stop concurrently from many goroutines", appended: lifecycleItems, drive: driveStopConcurrently},
		{name: "run twice then stop twice", appended: lifecycleItems, drive: driveRunTwiceStopTwice},
		{name: "run again after done", appended: lifecycleItems, drive: driveRunAfterDone},
		{name: "stop before run", appended: lifecycleCapacity, drive: driveStopBeforeRun},
		{name: "wait done twice", appended: lifecycleItems, drive: driveWaitDoneTwice},
	}
}

type lifecycleVariant struct {
	name string
	opts []BatcherOption[int]
}

func lifecycleVariants() []lifecycleVariant {
	return []lifecycleVariant{
		{name: "no interval", opts: nil},
		{name: "flush interval", opts: []BatcherOption[int]{WithFlushInterval[int](time.Hour)}},
		// Room for exactly lifecycleCapacity ints, so the size limit
		// triggers flushes on its own and never blocks before the buffer
		// would.
		{name: "max size", opts: []BatcherOption[int]{WithMaxSize[int](lifecycleCapacity * int32(intSize))}},
	}
}

func runLifecycleScenario(t *testing.T, variant lifecycleVariant, scenario lifecycleScenario) {
	t.Helper()

	col := newCollector()
	bat := newLifecycleBatcher(col, variant.opts...)

	appendItems := func() {
		within(t, "Append", func() {
			for item := range scenario.appended {
				bat.Append(item)
			}
		})
	}

	scenario.drive(t, bat, appendItems)

	within(t, "WaitDone", bat.WaitDone)
	col.requireExactlyOnce(t, scenario.appended)
}

func TestBatcherLifecycle(t *testing.T) {
	t.Parallel()

	for _, variant := range lifecycleVariants() {
		for _, scenario := range lifecycleScenarios() {
			t.Run(variant.name+"/"+scenario.name, func(t *testing.T) {
				t.Parallel()
				runLifecycleScenario(t, variant, scenario)
			})
		}
	}
}

func TestBatcherWaitDoneBlocksWithoutRun(t *testing.T) {
	t.Parallel()

	bat := newLifecycleBatcher(newCollector())
	returned := make(chan struct{})

	go func() {
		bat.WaitDone()
		close(returned)
	}()

	select {
	case <-returned:
		t.Fatal("WaitDone returned for a batcher that was never run")
	case <-time.After(50 * time.Millisecond):
	}

	// Release the goroutine so the test leaves nothing behind.
	bat.Run()
	bat.Stop()
	within(t, "WaitDone after Run and Stop", func() { <-returned })
}
