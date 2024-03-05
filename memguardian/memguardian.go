package memguardian

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type MemGuardianOption func(*MemGuardian) error

func WitInterval(d time.Duration) MemGuardianOption {
	return func(mg *MemGuardian) error {
		mg.t = time.NewTicker(d)
		return nil
	}
}

func WithCallback(f func()) MemGuardianOption {
	return func(mg *MemGuardian) error {
		mg.f = f
		return nil
	}
}

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

func (mg *MemGuardian) Close() {
	mg.cancel()
	mg.t.Stop()
}

func UsedRamRatio() (float64, error) {
	vms, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return vms.UsedPercent, nil
}
