package proxyutils

import (
	"fmt"
	"net"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	callTimeout = 5 * time.Second

	allDeadMessage = "all proxies are dead got : "
)

type aliveResult struct {
	proxy string
	err   error
}

// callGetAnyAliveProxy fails the test instead of hanging when the call does
// not return.
func callGetAnyAliveProxy(t *testing.T, proxies ...string) aliveResult {
	t.Helper()

	done := make(chan aliveResult, 1)

	go func() {
		proxy, err := GetAnyAliveProxy(1, proxies...)
		done <- aliveResult{proxy: proxy, err: err}
	}()

	select {
	case res := <-done:
		return res
	case <-time.After(callTimeout):
	}

	t.Fatalf("GetAnyAliveProxy with %d proxies did not return within %s", len(proxies), callTimeout)

	return aliveResult{proxy: "", err: nil}
}

// requireAlive fails the test unless the call returned want without error.
func requireAlive(t *testing.T, res aliveResult, want string) {
	t.Helper()

	if res.err != nil {
		t.Fatalf("GetAnyAliveProxy returned error %v, want proxy %q", res.err, want)
	}

	if res.proxy != want {
		t.Fatalf("GetAnyAliveProxy returned %q, want %q", res.proxy, want)
	}
}

// requireAllDead fails the test unless the call returned the all-dead error,
// containing every substring in wantInError, and no proxy.
func requireAllDead(t *testing.T, res aliveResult, wantInError ...string) {
	t.Helper()

	if res.proxy != "" {
		t.Fatalf("GetAnyAliveProxy returned proxy %q, want none", res.proxy)
	}

	if res.err == nil {
		t.Fatalf("GetAnyAliveProxy returned no error, want %q", allDeadMessage)
	}

	for _, want := range append([]string{allDeadMessage}, wantInError...) {
		if !strings.Contains(res.err.Error(), want) {
			t.Fatalf("GetAnyAliveProxy error %q does not contain %q", res.err, want)
		}
	}
}

func listenLoopback(t *testing.T) net.Listener {
	t.Helper()

	var lc net.ListenConfig

	l, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on loopback: %v", err)
	}

	return l
}

// loopbackProxies returns proxy URLs for aliveCount loopback listeners that
// stay open until the test ends and deadCount distinct loopback addresses
// that refuse connections. Every listener is opened before any is closed, so
// the dead addresses cannot collide with each other or the alive ones.
//
// A dead port is only reserved by being closed, so another test's listener
// could still be handed it before it is dialed; the window is tiny and
// accepted rather than solved with platform-specific socket code.
func loopbackProxies(t *testing.T, aliveCount, deadCount int) ([]string, []string) {
	t.Helper()

	listeners := make([]net.Listener, aliveCount+deadCount)
	urls := make([]string, aliveCount+deadCount)

	for i := range listeners {
		listeners[i] = listenLoopback(t)
		urls[i] = "http://" + listeners[i].Addr().String()
	}

	for _, l := range listeners[:aliveCount] {
		t.Cleanup(func() { _ = l.Close() })
	}

	for _, l := range listeners[aliveCount:] {
		err := l.Close()
		if err != nil {
			t.Fatalf("close listener %s: %v", l.Addr(), err)
		}
	}

	return urls[:aliveCount], urls[aliveCount:]
}

func invalidProxies(n int) []string {
	proxies := make([]string, n)
	for i := range proxies {
		proxies[i] = fmt.Sprintf("ftp://127.0.0.1:%d", i+1)
	}

	return proxies
}

func setProbeConcurrency(t *testing.T, n int) {
	t.Helper()

	old := ProxyProbeConcurrency
	ProxyProbeConcurrency = n

	t.Cleanup(func() { ProxyProbeConcurrency = old })
}

// requireNoLeakedGoroutines polls because goroutines exit shortly after the
// call returns. It does not use require.Eventually, whose own goroutines
// would be counted.
func requireNoLeakedGoroutines(t *testing.T, baseline int) {
	t.Helper()

	deadline := time.Now().Add(callTimeout)
	for runtime.NumGoroutine() > baseline {
		if time.Now().After(deadline) {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			t.Fatalf("goroutines left running after GetAnyAliveProxy returned: %d > %d\n%s",
				runtime.NumGoroutine(), baseline, buf)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestGetAnyAliveProxyAllDead(t *testing.T) {
	t.Parallel()

	// 8 is ProxyProbeConcurrency and 10 is the first count where the
	// producer observes a cancelled context after its semaphore wait.
	for _, n := range []int{1, 8, 9, 10, 11, 50} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			_, dead := loopbackProxies(t, 0, n)

			requireAllDead(t, callGetAnyAliveProxy(t, dead...))
		})
	}
}

// An alive proxy behind dead ones must be found. 3 dead proxies fit in the
// default concurrency; 20 and 50 make the producer wait for probe slots while
// earlier dead results are already coming in.
func TestGetAnyAliveProxyAliveAfterDead(t *testing.T) {
	t.Parallel()

	for _, deadBefore := range []int{3, 20, 50} {
		t.Run(strconv.Itoa(deadBefore), func(t *testing.T) {
			t.Parallel()

			alive, dead := loopbackProxies(t, 1, deadBefore+5)
			proxies := slices.Concat(dead[:deadBefore], alive, dead[deadBefore:])

			requireAlive(t, callGetAnyAliveProxy(t, proxies...), alive[0])
		})
	}
}

func TestGetAnyAliveProxyFirstAlive(t *testing.T) {
	t.Parallel()

	alive, _ := loopbackProxies(t, 1, 0)

	requireAlive(t, callGetAnyAliveProxy(t, alive...), alive[0])
}

func TestGetAnyAliveProxyInvalidURLs(t *testing.T) {
	t.Parallel()

	t.Run("only invalid", func(t *testing.T) {
		t.Parallel()

		requireAllDead(t, callGetAnyAliveProxy(t, invalidProxies(30)...), "invalid proxy format")
	})

	t.Run("invalid then alive", func(t *testing.T) {
		t.Parallel()

		alive, _ := loopbackProxies(t, 1, 0)
		proxies := slices.Concat(invalidProxies(30), alive)

		requireAlive(t, callGetAnyAliveProxy(t, proxies...), alive[0])
	})

	t.Run("invalid and dead mixed", func(t *testing.T) {
		t.Parallel()

		_, dead := loopbackProxies(t, 0, 15)
		proxies := slices.Concat(invalidProxies(15), dead)

		requireAllDead(t, callGetAnyAliveProxy(t, proxies...))
	})
}

func TestGetAnyAliveProxyEmptyInput(t *testing.T) {
	t.Parallel()

	requireAllDead(t, callGetAnyAliveProxy(t))
}

// Sequential: it writes ProxyProbeConcurrency and compares goroutine counts,
// which other running tests would disturb.
//
// Only "all dead" reproduces a defect of the previous implementation (it
// hangs). The "alive first" cases are regression guards: they check that every
// goroutine eventually exits after an early alive return, which includes a
// sender blocked on a full result channel. They cannot detect a producer that
// keeps probing after the return (no cancel, or Add instead of AddWithContext),
// because loopback probes finish immediately and it exits either way.
func TestGetAnyAliveProxyNoLeakedGoroutines(t *testing.T) {
	cases := map[string]func(t *testing.T) ([]string, string){
		"all dead": func(t *testing.T) ([]string, string) {
			t.Helper()

			_, dead := loopbackProxies(t, 0, 30)

			return dead, ""
		},
		// With concurrency 1 the producer is parked on the semaphore behind
		// the dead proxy when the alive one answers, then sends the invalid
		// URLs' errors after the receiver has returned.
		"alive first, dead and invalid after": func(t *testing.T) ([]string, string) {
			t.Helper()

			alive, dead := loopbackProxies(t, 1, 1)

			return slices.Concat(alive, dead, invalidProxies(40)), alive[0]
		},
		"alive first, dead after": func(t *testing.T) ([]string, string) {
			t.Helper()

			alive, dead := loopbackProxies(t, 1, 40)

			return slices.Concat(alive, dead), alive[0]
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			// Concurrency 1 keeps the producer waiting on the semaphore
			// when the alive proxy answers.
			setProbeConcurrency(t, 1)

			proxies, wantAlive := build(t)
			baseline := runtime.NumGoroutine()

			res := callGetAnyAliveProxy(t, proxies...)

			if wantAlive != "" {
				requireAlive(t, res, wantAlive)
			} else {
				requireAllDead(t, res)
			}

			requireNoLeakedGoroutines(t, baseline)
		})
	}
}
