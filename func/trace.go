package main

import (
	"errors"
	"runtime"
	"time"
)

type Metrics struct {
	StartTime         time.Time
	FinishTime        time.Time
	ExecutionDuration time.Duration
	AllocMemory       uint64
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
}

func (d *DefaultStrategy) Before() {
	d.metrics.StartTime = time.Now()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	d.metrics.AllocMemory = mem.Alloc
}

func (d *DefaultStrategy) After() {
	d.metrics.FinishTime = time.Now()
	d.metrics.ExecutionDuration = d.metrics.FinishTime.Sub(d.metrics.StartTime)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	d.metrics.AllocMemory = mem.Alloc - d.metrics.AllocMemory
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
