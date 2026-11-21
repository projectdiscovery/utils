package connpool

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Bounds of the wait between failed dials in OneTimePool.Run.
const (
	minDialBackoff = 50 * time.Millisecond
	maxDialBackoff = time.Second
	backoffFactor  = 2
)

type Dialer interface {
	Dial(ctx context.Context, network, address string) (net.Conn, error)
}

// OneTimePool is a pool designed to create continous bare connections that are for one time only usage
type OneTimePool struct {
	address         string
	idleConnections chan net.Conn
	InFlightConns   *InFlightConns
	ctx             context.Context
	cancel          context.CancelFunc
	Dialer          Dialer
	mx              sync.RWMutex
}

func NewOneTimePool(ctx context.Context, address string, poolSize int, opts ...Option) (*OneTimePool, error) {
	idleConnections := make(chan net.Conn, poolSize)
	inFlightConns, err := NewInFlightConns()
	if err != nil {
		return nil, err
	}
	pool := &OneTimePool{
		address:         address,
		idleConnections: idleConnections,
		InFlightConns:   inFlightConns,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	pool.ctx, pool.cancel = context.WithCancel(ctx)
	// apply options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(pool); err != nil {
			return nil, err
		}
	}
	return pool, nil
}

// Acquire acquires an idle connection from the pool
func (p *OneTimePool) Acquire(c context.Context) (net.Conn, error) {
	// A select chooses pseudo-randomly when multiple cases are ready. Check both
	// contexts first so a buffered connection cannot win after its deadline and
	// start another potentially long I/O operation.
	if err := p.ctx.Err(); err != nil {
		return nil, err
	}
	if err := c.Err(); err != nil {
		return nil, err
	}

	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case <-c.Done():
		return nil, c.Err()
	case conn := <-p.idleConnections:
		p.InFlightConns.Remove(conn)
		return conn, nil
	}
}

// Run dials connections and keeps the pool filled until the pool is closed or
// its context is done, then returns an error wrapping the context's error
// (match it with errors.Is). Failed dials are retried with an exponentially
// growing wait, capped at maxDialBackoff and reset by the next successful
// dial, so an unreachable target does not consume a CPU core. Dial errors are
// not reported; callers observe them as Acquire blocking until its own
// context is done.
func (p *OneTimePool) Run() error {
	backoff := minDialBackoff
	for {
		if p.ctx.Err() != nil {
			return p.stopped()
		}

		conn := p.dial()
		if conn == nil {
			if !p.sleep(backoff) {
				return p.stopped()
			}

			backoff = min(backoff*backoffFactor, maxDialBackoff)

			continue
		}

		backoff = minDialBackoff

		p.InFlightConns.Add(conn)
		// Close cancels the context before it closes the in-flight
		// connections. A connection added after that sweep would never be
		// closed, so a done context here means this goroutine must close it.
		if p.ctx.Err() != nil {
			p.discard(conn)

			return p.stopped()
		}

		select {
		case <-p.ctx.Done():
			p.discard(conn)

			return p.stopped()
		case p.idleConnections <- conn:
		}
	}
}

func (p *OneTimePool) Close() error {
	p.cancel()

	// remove dialer references
	p.mx.Lock()
	p.Dialer = nil
	p.mx.Unlock()

	return p.InFlightConns.Close()
}

// stopped is the error Run returns once the pool context is done.
func (p *OneTimePool) stopped() error {
	return fmt.Errorf("connpool: run stopped: %w", p.ctx.Err())
}

// dial opens one connection and returns nil if the dial failed; Run retries
// and does not report the cause, so the error is not built into a value. The
// dialer is read under the lock but invoked outside it, so Close never waits
// for a dial in progress.
func (p *OneTimePool) dial() net.Conn {
	p.mx.RLock()
	dialer := p.Dialer
	p.mx.RUnlock()

	var (
		conn net.Conn
		err  error
	)

	if dialer != nil {
		conn, err = dialer.Dial(p.ctx, "tcp", p.address)
	} else {
		var netDialer net.Dialer

		conn, err = netDialer.DialContext(p.ctx, "tcp", p.address)
	}

	if err != nil {
		return nil
	}

	return conn
}

// sleep waits for d and reports whether the pool context is still live.
func (p *OneTimePool) sleep(d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-p.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (p *OneTimePool) discard(conn net.Conn) {
	p.InFlightConns.Remove(conn)
	_ = conn.Close()
}
