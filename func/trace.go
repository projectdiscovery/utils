package main

import (
	"errors"
	"math"
	"runtime"
	"time"
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
	metrics *Metrics
	ticker  *time.Ticker
	done    chan bool
}

func (d *DefaultStrategy) Before() {
	d.metrics.StartTime = time.Now()

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
				d.metrics.Snapshots = append(d.metrics.Snapshots, MemorySnapshot{
					Time:  t,
					Alloc: mem.Alloc,
				})
			}
		}
	}()
}

func (d *DefaultStrategy) After() {
	close(d.done)
	d.ticker.Stop()

	d.metrics.FinishTime = time.Now()
	d.metrics.ExecutionDuration = d.metrics.FinishTime.Sub(d.metrics.StartTime)

	var totalMemory uint64 = 0
	if len(d.metrics.Snapshots) > 0 {
		d.metrics.MinAllocMemory = d.metrics.Snapshots[0].Alloc
		d.metrics.MaxAllocMemory = d.metrics.Snapshots[0].Alloc

		for _, s := range d.metrics.Snapshots {
			if s.Alloc < d.metrics.MinAllocMemory {
				d.metrics.MinAllocMemory = s.Alloc
			}
			d.metrics.MinAllocMemory = uint64(math.Min(float64(d.metrics.MinAllocMemory), float64(s.Alloc)))
			d.metrics.MaxAllocMemory = uint64(math.Max(float64(d.metrics.MaxAllocMemory), float64(s.Alloc)))
			totalMemory += s.Alloc
		}
		d.metrics.AvgAllocMemory = totalMemory / uint64(len(d.metrics.Snapshots))
	}
}

func (d *DefaultStrategy) GetMetrics() *Metrics {
	return d.metrics
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
		strategy: &DefaultStrategy{metrics: &Metrics{}},
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
