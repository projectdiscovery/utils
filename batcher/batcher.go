package batcher

import (
	"sync"
	"sync/atomic"
	"time"

	reflectutil "github.com/projectdiscovery/utils/reflect"
)

// FlushCallback is the callback function that will be called when the batcher is full or the flush interval is reached
type FlushCallback[T any] func([]T)

// Batcher is a batcher for any type of data
type Batcher[T any] struct {
	maxCapacity int
	maxSize     int32

	currentSize atomic.Int32

	flushInterval *time.Duration
	flushCallback FlushCallback[T]

	incomingData chan T
	full         chan bool
	// mustExit is closed by the first Stop; a closed channel is how the
	// run loop (and a Run that starts later) observes any number of Stops.
	mustExit chan struct{}
	done     chan bool

	// runOnce and stopOnce make Run and Stop idempotent. done is closed by
	// the run loop, so a second loop would close it twice and panic.
	runOnce  sync.Once
	stopOnce sync.Once
}

// BatcherOption is the option for the batcher
type BatcherOption[T any] func(*Batcher[T])

// WithMaxCapacity sets the max capacity of the batcher
func WithMaxCapacity[T any](maxCapacity int) BatcherOption[T] {
	return func(b *Batcher[T]) {
		b.maxCapacity = maxCapacity
	}
}

// WithMaxSize sets the max size of the batcher
func WithMaxSize[T any](maxSize int32) BatcherOption[T] {
	return func(b *Batcher[T]) {
		b.maxSize = maxSize
	}
}

// WithFlushInterval sets the optional flush interval of the batcher
func WithFlushInterval[T any](flushInterval time.Duration) BatcherOption[T] {
	return func(b *Batcher[T]) {
		b.flushInterval = &flushInterval
	}
}

// WithFlushCallback sets the flush callback of the batcher
func WithFlushCallback[T any](fn FlushCallback[T]) BatcherOption[T] {
	return func(b *Batcher[T]) {
		b.flushCallback = fn
	}
}

// New creates a new batcher
func New[T any](opts ...BatcherOption[T]) *Batcher[T] {
	batcher := &Batcher[T]{
		full:     make(chan bool),
		mustExit: make(chan struct{}),
		done:     make(chan bool, 1),
	}
	for _, opt := range opts {
		opt(batcher)
	}
	if batcher.maxSize > 0 {
		batcher.currentSize = atomic.Int32{}
	}
	batcher.incomingData = make(chan T, batcher.maxCapacity)
	if batcher.flushCallback == nil {
		panic("batcher: flush callback is required")
	}
	if batcher.maxCapacity <= 0 {
		panic("batcher: max capacity must be greater than 0")
	}
	return batcher
}

// Append appends data to the batcher
func (b *Batcher[T]) Append(d ...T) {
	for _, item := range d {
		var sizeofItem int

		if b.maxSize > 0 {
			sizeofItem = reflectutil.SizeOf(item)
			currentSize := b.currentSize.Load()

			if currentSize+int32(sizeofItem) > int32(b.maxSize) {
				b.full <- true
				b.incomingData <- item
				b.currentSize.Add(int32(sizeofItem))
				continue
			}
		}

		if !b.put(item) {
			// will wait until space available
			b.full <- true
			b.incomingData <- item
		}
		if b.maxSize > 0 {
			b.currentSize.Add(int32(sizeofItem))
		}
	}
}

func (b *Batcher[T]) put(d T) bool {
	// try to append the data
	select {
	case b.incomingData <- d:
		return true
	default:
		// channel is full
		return false
	}
}

func (b *Batcher[T]) run() {
	// consume all items in the queue
	defer func() {
		b.doCallback()
		close(b.done)
	}()

	var timer *time.Timer
	var flushInterval time.Duration
	if b.flushInterval != nil {
		flushInterval = *b.flushInterval
		timer = time.NewTimer(flushInterval)
		b.runWithTimer(timer, flushInterval)
		return
	}
	b.runWithoutTimer()
}

// runWithTimer runs the batcher with timer
func (b *Batcher[T]) runWithTimer(timer *time.Timer, flushInterval time.Duration) {
	for {
		select {
		case <-timer.C:
			b.doCallback()
			timer.Reset(flushInterval)
		case <-b.full:
			if !timer.Stop() {
				<-timer.C
			}
			b.doCallback()
			timer.Reset(flushInterval)
		case <-b.mustExit:
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}

// runWithoutTimer runs the batcher without timer
func (b *Batcher[T]) runWithoutTimer() {
	for {
		select {
		case <-b.full:
			b.doCallback()
		case <-b.mustExit:
			return
		}
	}
}

func (b *Batcher[T]) doCallback() {
	n := len(b.incomingData)
	if n == 0 {
		return
	}
	items := make([]T, n)

	var k int
	for item := range b.incomingData {
		items[k] = item
		k++
		if b.maxSize > 0 {
			b.currentSize.Add(-int32(reflectutil.SizeOf(item)))
		}
		if k >= n {
			break
		}
	}
	b.flushCallback(items)
}

// Run starts the batcher. Only the first call has an effect, and a batcher
// cannot be restarted once stopped.
func (b *Batcher[T]) Run() {
	b.runOnce.Do(func() { go b.run() })
}

// Stop asks the batcher to flush the remaining items and exit, without
// waiting for it (use WaitDone). It never blocks and may be called any number
// of times, from any goroutine, including before Run, in which case the
// batcher flushes and exits as soon as it runs.
//
// Append must not be called concurrently with or after Stop: once the batcher
// has exited nothing drains it, so Append may block forever (the buffer is
// full or the size limit is reached) or leave its items unflushed.
func (b *Batcher[T]) Stop() {
	b.stopOnce.Do(func() { close(b.mustExit) })
}

// WaitDone waits until the batcher is done. It blocks until the batcher has
// been run and stopped, so it never returns for a batcher that was not run.
func (b *Batcher[T]) WaitDone() {
	<-b.done
}
