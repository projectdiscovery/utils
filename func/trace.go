package main

import (
	"errors"
	"math"
	"runtime"
	"time"

	"github.com/projectdiscovery/utils/generic"
)

const (
	// DefaultMemorySnapshotInterval is the default interval for taking memory snapshots
	DefaultMemorySnapshotInterval = 100 * time.Millisecond
)

type MemorySnapshot struct {
	Time  time.Time
	Alloc uint64
}

type Metrics struct {
	StartTime         time.Time
	FinishTime        time.Time
	ExecutionDuration time.Duration
	Snapshots         []MemorySnapshot
	MinAllocMemory    uint64
	MaxAllocMemory    uint64
	AvgAllocMemory    uint64
}

type FunctionContext struct {
	strategy ActionStrategy
	action   func()
}

func (f *FunctionContext) Execute() {
	f.strategy.Before()
	f.action()
	f.strategy.After()
}

type ActionStrategy interface {
	Before()
	After()
	GetMetrics() *Metrics
}

type DefaultStrategy struct {
	metrics generic.Lockable[*Metrics]
	ticker  *time.Ticker
	done    chan bool
}

func (d *DefaultStrategy) Before() {
	d.metrics.Do(func(m *Metrics) {
		m.StartTime = time.Now()

		d.ticker = time.NewTicker(DefaultMemorySnapshotInterval)
		d.done = make(chan bool)
		go func() {
			for {
				select {
				case <-d.done:
					return
				case t := <-d.ticker.C:
					var mem runtime.MemStats
					runtime.ReadMemStats(&mem)
					m.Snapshots = append(m.Snapshots, MemorySnapshot{
						Time:  t,
						Alloc: mem.Alloc,
					})
				}
			}
		}()
	})
}

func (d *DefaultStrategy) After() {
	close(d.done)
	d.ticker.Stop()
	d.metrics.Do(func(m *Metrics) {
		m.FinishTime = time.Now()
		m.ExecutionDuration = m.FinishTime.Sub(m.StartTime)

		var totalMemory uint64 = 0
		if len(m.Snapshots) > 0 {
			m.MinAllocMemory = m.Snapshots[0].Alloc
			m.MaxAllocMemory = m.Snapshots[0].Alloc

			for _, s := range m.Snapshots {
				if s.Alloc < m.MinAllocMemory {
					m.MinAllocMemory = s.Alloc
				}
				m.MinAllocMemory = uint64(math.Min(float64(m.MinAllocMemory), float64(s.Alloc)))
				m.MaxAllocMemory = uint64(math.Max(float64(m.MaxAllocMemory), float64(s.Alloc)))
				totalMemory += s.Alloc
			}
			m.AvgAllocMemory = totalMemory / uint64(len(m.Snapshots))
		}
	})

}

func (d *DefaultStrategy) GetMetrics() *Metrics {
	var metrics *Metrics
	d.metrics.Do(func(m *Metrics) {
		metrics = m
	})
	return metrics
}

type TraceOptions struct {
	strategy ActionStrategy
}

type TraceOptionSetter func(opts *TraceOptions)

func WithStrategy(s ActionStrategy) TraceOptionSetter {
	return func(opts *TraceOptions) {
		opts.strategy = s
	}
}

func Trace(f func(), setter TraceOptionSetter) (*Metrics, error) {
	opts := &TraceOptions{
		strategy: &DefaultStrategy{metrics: generic.Lockable[*Metrics]{V: &Metrics{}}},
	}

	// Apply option if provided
	if setter != nil {
		setter(opts)
	}

	if opts.strategy == nil {
		return nil, errors.New("strategy should not be nil")
	}

	context := &FunctionContext{
		strategy: opts.strategy,
		action:   f,
	}

	context.Execute()
	return opts.strategy.GetMetrics(), nil
}
