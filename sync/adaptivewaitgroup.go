package sync

// Extended version of https://github.com/remeh/sizedwaitgroup

import (
	"context"
	"errors"

	"github.com/remeh/sizedwaitgroup"
)

type AdaptiveGroupOption func(*AdaptiveWaitGroup) error

type AdaptiveWaitGroup struct {
	Size int

	wg sizedwaitgroup.SizedWaitGroup
}

func WithSize(size int) AdaptiveGroupOption {
	return func(wg *AdaptiveWaitGroup) error {
		if size < 0 {
			return errors.New("size must be positive")
		}
		wg.Size = size
		return nil
	}
}

func New(options ...AdaptiveGroupOption) (*AdaptiveWaitGroup, error) {
	wg := &AdaptiveWaitGroup{}
	for _, option := range options {
		if err := option(wg); err != nil {
			return nil, err
		}
	}

	wg.wg = sizedwaitgroup.New(wg.Size)
	return wg, nil
}

func (s *AdaptiveWaitGroup) Add() {
	_ = s.AddWithContext(context.Background())
}

func (s *AdaptiveWaitGroup) AddWithContext(ctx context.Context) error {
	s.wg.Add()
	return nil
}

func (s *AdaptiveWaitGroup) Done() {
	s.wg.Done()
}

func (s *AdaptiveWaitGroup) Wait() {
	s.wg.Wait()
}

func (s *AdaptiveWaitGroup) Resize(size int) {}
