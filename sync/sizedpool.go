package sync

import (
	"context"
	"errors"
	"math"
	"sync"

	"golang.org/x/sync/semaphore"
)

type PoolOption[T any] func(*SizedPool[T]) error

func WithSize[T any](size int64) PoolOption[T] {
	return func(sz *SizedPool[T]) error {
		if size <= 0 {
			return errors.New("size must be positive")
		}
		sz.initialSize = size
		sz.maxSize = math.MaxInt64
		sz.sem = semaphore.NewWeighted(sz.maxSize)
		sz.sem.Acquire(context.Background(), sz.maxSize-sz.initialSize)
		return nil
	}
}

func WithPool[T any](p *sync.Pool) PoolOption[T] {
	return func(sz *SizedPool[T]) error {
		sz.pool = p
		return nil
	}
}

type SizedPool[T any] struct {
	sem         *semaphore.Weighted
	initialSize int64
	maxSize     int64
	pool        *sync.Pool
}

func New[T any](options ...PoolOption[T]) (*SizedPool[T], error) {
	sz := &SizedPool[T]{}
	for _, option := range options {
		if err := option(sz); err != nil {
			return nil, err
		}
	}
	return sz, nil
}

func (sz *SizedPool[T]) Get(ctx context.Context) (T, error) {
	if sz.sem != nil {
		if err := sz.sem.Acquire(ctx, 1); err != nil {
			var t T
			return t, err
		}
	}
	return sz.pool.Get().(T), nil
}

func (sz *SizedPool[T]) Put(x T) {
	if sz.sem != nil {
		sz.sem.Release(1)
	}
	sz.pool.Put(x)
}

// Vary capacity by x - it's internally qneuqued as a normal Acquire/Release operation as other Get/Put
// but tokens are held internally
func (sz *SizedPool[T]) Vary(ctx context.Context, x int64) error {
	switch {
	case x > 0:
		sz.sem.Release(x)
		return nil
	case x < 0:
		sz.sem.Acquire(ctx, x)
		return nil
	default:
		return errors.New("x is zero")
	}
}
