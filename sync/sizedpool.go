package sync

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/semaphore"
)

type PoolOption[T any] func(*SizedPool[T]) error

func WithMaxCapacity[T any](maxCapacity int) PoolOption[T] {
	return func(sz *SizedPool[T]) error {
		if maxCapacity <= 0 {
			return errors.New("capacity must be positive")
		}
		sz.maxCapacity = maxCapacity
		sz.sem = semaphore.NewWeighted(int64(maxCapacity))
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
	maxCapacity int
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
