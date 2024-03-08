package semaphore

import (
	"context"
	"errors"
	"math"

	"golang.org/x/sync/semaphore"
)

type Semaphore struct {
	sem         *semaphore.Weighted
	initialSize int64
	maxSize     int64
}

func New(size int64) (*Semaphore, error) {
	maxSize := int64(math.MaxInt64)
	s := &Semaphore{
		initialSize: size,
		maxSize:     maxSize,
		sem:         semaphore.NewWeighted(maxSize),
	}
	err := s.sem.Acquire(context.Background(), s.maxSize-s.initialSize)
	return s, err
}

func (s *Semaphore) Acquire(ctx context.Context, n int64) error {
	return s.sem.Acquire(ctx, n)
}

func (s *Semaphore) Release(n int64) {
	s.sem.Release(n)
}

// Vary capacity by x - it's internally enqueued as a normal Acquire/Release operation as other Get/Put
// but tokens are held internally
func (s *Semaphore) Vary(ctx context.Context, x int64) error {
	switch {
	case x > 0:
		s.sem.Release(x)
		return nil
	case x < 0:
		return s.sem.Acquire(ctx, x)
	default:
		return errors.New("x is zero")
	}
}
