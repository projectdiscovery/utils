package sync

// Extended version of https://github.com/remeh/sizedwaitgroup

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/eapache/channels"
	"github.com/projectdiscovery/gologger"
)

type AdaptiveGroupOption func(*AdaptiveWaitGroup) error

type AdaptiveWaitGroup struct {
	caller string
	Size   int

	current *channels.ResizableChannel
	wg      sync.WaitGroup
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
	_, file, no, ok := runtime.Caller(1)
	var caller string
	if ok {
		caller = fmt.Sprintf("called from %s#%d\n", file, no)
	}
	gologger.Info().Msgf("New AdaptiveWaitGroup %s\n", caller)

	wg := &AdaptiveWaitGroup{caller: caller}
	for _, option := range options {
		if err := option(wg); err != nil {
			return nil, err
		}
	}

	wg.current = channels.NewResizableChannel()
	wg.current.Resize(channels.BufferCap(wg.Size))
	wg.wg = sync.WaitGroup{}
	return wg, nil
}

func (s *AdaptiveWaitGroup) Add() {
	_ = s.AddWithContext(context.Background())
}

func (s *AdaptiveWaitGroup) AddWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.current.In() <- struct{}{}:
		break
	}
	s.wg.Add(1)
	return nil
}

func (s *AdaptiveWaitGroup) Done() {
	<-s.current.Out()
	s.wg.Done()
}

func (s *AdaptiveWaitGroup) Wait() {
	s.wg.Wait()
	s.current.Close()

	gologger.Info().Msgf("Wait AdaptiveWaitGroup %s\n", s.caller)
}

func (s *AdaptiveWaitGroup) Resize(size int) {
	s.current.Resize(channels.BufferCap(size))
	s.Size = int(s.current.Cap())
}
