package memguardian

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type MemGuardianOption func(*MemGuardian) error

// WithInterval defines the ticker interval of the memory monitor
func WitInterval(d time.Duration) MemGuardianOption {
	return func(mg *MemGuardian) error {
		mg.t = time.NewTicker(d)
		return nil
	}
}

// WithCallback defines an optional callback if the warning ration is exceeded
func WithCallback(f func()) MemGuardianOption {
	return func(mg *MemGuardian) error {
		mg.f = f
		return nil
	}
}

// WithRatioWarning defines the threshold of the warning state (and optional callback invocation)
func WithRatioWarning(ratio float64) MemGuardianOption {
	return func(mg *MemGuardian) error {
		if ratio == 0 || ratio > 100 {
			return errors.New("ratio must be between 1 and 100")
		}
		mg.ratio = ratio
		return nil
	}
}

type MemGuardian struct {
	t       *time.Ticker
	f       func()
	ctx     context.Context
	cancel  context.CancelFunc
	Warning atomic.Bool
	ratio   float64
}

// New mem guadian instance with user defined options
func New(options ...MemGuardianOption) (*MemGuardian, error) {
	mg := &MemGuardian{}
	for _, option := range options {
		if err := option(mg); err != nil {
			return nil, err
		}
	}

	mg.ctx, mg.cancel = context.WithCancel(context.TODO())

	return mg, nil
}

// Run the instance monitor (cancel using the Stop method or context parameter)
func (mg *MemGuardian) Run(ctx context.Context) error {
	for {
		select {
		case <-mg.ctx.Done():
			mg.Close()
			return nil
		case <-ctx.Done():
			mg.Close()
			return nil
		case <-mg.t.C:
			usedRatio, err := UsedRamRatio()
			if err != nil {
				return err
			}

			if usedRatio >= mg.ratio {
				mg.Warning.Store(true)
				if mg.f != nil {
					mg.f()
				}
			} else {
				mg.Warning.Store(false)
			}
		}
	}
}

// Close and stops the instance
func (mg *MemGuardian) Close() {
	mg.cancel()
	mg.t.Stop()
}

// Calculate the system absolute ratio of used RAM vs total available (as of now doesn't consider swap)
func UsedRamRatio() (float64, error) {
	vms, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return vms.UsedPercent, nil
}
