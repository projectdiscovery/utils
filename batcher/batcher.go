package batcher

import (
	"time"
)

// FlushCallback is the callback function that will be called when the batcher is full or the flush interval is reached
type FlushCallback[T any] func([]T)

type Batcher[T any] struct {
	maxCapacity   int
	flushInterval time.Duration

	incomingData  chan T
	full          chan bool
	mustExit      chan bool
	done          chan bool
	flushCallback FlushCallback[T]
}

// New creates a new batcher
func New[T any](maxCapacity int, flushInterval time.Duration, fn FlushCallback[T]) *Batcher[T] {
	batcher := &Batcher[T]{
		maxCapacity:   maxCapacity,
		incomingData:  make(chan T, maxCapacity),
		full:          make(chan bool),
		flushInterval: flushInterval,
		mustExit:      make(chan bool, 1),
		done:          make(chan bool, 1),
		flushCallback: fn,
	}
	return batcher
}

// Append appends data to the batcher
func (b *Batcher[T]) Append(d ...T) {
	for _, item := range d {
		if !b.put(item) {
			// will wait until space available
			b.full <- true
			b.incomingData <- item
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

	timer := time.NewTimer(b.flushInterval)
	for {
		select {
		case <-timer.C:
			b.doCallback()
			timer.Reset(b.flushInterval)
		case <-b.full:
			if !timer.Stop() {
				<-timer.C
			}
			b.doCallback()
			timer.Reset(b.flushInterval)
		case <-b.mustExit:
			if !timer.Stop() {
				<-timer.C
			}
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
		if k >= n {
			break
		}
	}
	b.flushCallback(items)
}

// Run starts the batcher
func (b *Batcher[T]) Run() {
	go b.run()
}

// Stop stops the batcher
func (b *Batcher[T]) Stop() {
	b.mustExit <- true
}

// WaitDone waits until the batcher is done
func (b *Batcher[T]) WaitDone() {
	<-b.done
}
