package contextutil_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	contextutil "github.com/projectdiscovery/utils/context"
)

// When the context is cancelled before fn returns, ExecFunc* returns the
// context error but the worker goroutine keeps running fn. Once fn finishes it
// sends its result on the internal channel. If that channel is unbuffered and
// the caller has already returned via ctx.Done(), nobody ever receives, so the
// worker blocks forever on the send — one leaked goroutine per cancelled call.
//
// These tests drive many cancelled calls whose fn outlives the context, then
// assert the goroutine count returns to baseline. They fail on an unbuffered
// result channel and pass once it is buffered (cap 1) so the worker can always
// send-and-exit.

const leakCalls = 50

// assertNoGoroutineLeak polls (workers finish and exit asynchronously) until the
// live goroutine count returns to ~base, failing if it does not.
func assertNoGoroutineLeak(t *testing.T, base int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		runtime.GC()
		leaked := runtime.NumGoroutine() - base
		if leaked <= 2 { // tolerance for transient runtime/test goroutines
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutine leak: ~%d goroutines still alive after %d context-cancelled calls (base=%d, now=%d)",
				leaked, leakCalls, base, runtime.NumGoroutine())
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// baseline settles outstanding goroutines then records the count.
func baseline() int {
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	return runtime.NumGoroutine()
}

func TestExecFunc_NoGoroutineLeakOnCancel(t *testing.T) {
	base := baseline()
	for range leakCalls {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_ = contextutil.ExecFunc(ctx, func() {
			time.Sleep(40 * time.Millisecond) // outlives the context
		})
		cancel()
	}
	assertNoGoroutineLeak(t, base)
}

func TestExecFuncWithTwoReturns_NoGoroutineLeakOnCancel(t *testing.T) {
	base := baseline()
	for range leakCalls {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, _ = contextutil.ExecFuncWithTwoReturns(ctx, func() (int, error) {
			time.Sleep(40 * time.Millisecond) // outlives the context
			return 42, nil
		})
		cancel()
	}
	assertNoGoroutineLeak(t, base)
}

func TestExecFuncWithThreeReturns_NoGoroutineLeakOnCancel(t *testing.T) {
	base := baseline()
	for range leakCalls {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, _, _ = contextutil.ExecFuncWithThreeReturns(ctx, func() (int, string, error) {
			time.Sleep(40 * time.Millisecond) // outlives the context
			return 42, "hello", nil
		})
		cancel()
	}
	assertNoGoroutineLeak(t, base)
}
