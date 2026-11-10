package channelutil_test

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/projectdiscovery/utils/channelutil"
)

var debug bool = false

// Helper Functions
var FillSource = func(work chan<- struct{}, count int, sg *sync.WaitGroup) {
	defer sg.Done()
	// send work
	for i := 100; i < 100+count; i++ {
		if debug {
			log.Println("sending work")
		}
		work <- struct{}{}
	}
	close(work)
}

var DrainSinkChan = func(ch <-chan struct{}, sleeptime time.Duration, chanId string, wg *sync.WaitGroup) {
	defer wg.Done()
	if ch == nil {
		fmt.Println("worker chan is nil")
	}
	for {
		val, ok := <-ch
		if !ok {
			if debug {
				log.Printf("sink channel %v closed\n", chanId)
			}
			break
		}
		// time.Sleep(sleeptime) // to simulate blocking network i/o
		if debug {
			log.Printf("completed work %v at chan %v\n", val, chanId)
		}
	}
}

func TestCloneCounter(t *testing.T) {
	// test Clone count
	for n := 1; n < 101; n++ {
		source := make(chan struct{})
		sinks := []chan<- struct{}{}
		for i := 0; i < n; i++ {
			sinks = append(sinks, make(chan struct{}))
		}
		cloneOpts := &channelutil.CloneOptions{
			MaxDrain:  3,
			Threshold: 5,
		}
		cloner := channelutil.NewCloneChannels[struct{}](cloneOpts)
		err := cloner.Clone(context.TODO(), source, sinks...)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
	}
}

func TestJoinCounter(t *testing.T) {
	// test Clone count
	for n := 1; n < 101; n++ {
		sink := make(chan struct{})
		sources := []<-chan struct{}{}
		for i := 0; i < n; i++ {
			sources = append(sources, make(chan struct{}))
		}
		jchan := channelutil.JoinChannels[struct{}]{}
		err := jchan.Join(context.TODO(), sink, sources...)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestCloneN(t *testing.T) {
	for n := 2; n < 44; n++ {
		source := make(chan struct{})
		sinks := []chan struct{}{}
		wg := &sync.WaitGroup{}
		sg := &sync.WaitGroup{}
		for i := 0; i < n; i++ {
			sinks = append(sinks, make(chan struct{}))
			wg.Add(1)
			go DrainSinkChan(sinks[i], time.Millisecond, strconv.Itoa(i), wg)
		}

		// Clone channel
		cloneOpts := &channelutil.CloneOptions{
			MaxDrain:  3,
			Threshold: 5,
		}
		recOnly := []chan<- struct{}{}
		for _, v := range sinks {
			recOnly = append(recOnly, v)
		}
		cloner := channelutil.NewCloneChannels[struct{}](cloneOpts)

		err := cloner.Clone(context.TODO(), source, recOnly...)
		if err != nil {
			t.Error(err)
		}

		sg.Add(1)
		FillSource(source, 10, sg)
		sg.Wait()
		wg.Wait()
	}
}

func TestJoinN(t *testing.T) {
	for N := 2; N < 44; N++ {
		sink := make(chan struct{})
		sg := &sync.WaitGroup{}
		// Drain Sink
		sg.Add(1)
		go DrainSinkChan(sink, time.Millisecond, "drain", sg)
		sources := []chan struct{}{}
		for i := 0; i < N; i++ {
			sources = append(sources, make(chan struct{}))
		}
		// create joiner
		joinch := channelutil.JoinChannels[struct{}]{}
		sndOnly := []<-chan struct{}{}
		for _, v := range sources {
			sndOnly = append(sndOnly, v)
		}
		err := joinch.Join(context.Background(), sink, sndOnly...)
		if err != nil {
			t.Error(err)
		}

		srcgrp := &sync.WaitGroup{}
		// Now start sending data to sources
		for _, v := range sources {
			srcgrp.Add(1)
			go FillSource(v, 10, srcgrp)
		}
		srcgrp.Wait()
		sg.Wait()
	}
}

func TestNIntegration(t *testing.T) {
	// Integration Test i.e Clone and join

	domainChan := make(chan struct{})
	sources := channelutil.CreateNChannels[struct{}](100, 0)
	task := make(chan struct{})

	controller := &sync.WaitGroup{}
	controller.Add(2)

	// send data to domainChan
	go FillSource(domainChan, 100, controller)
	// recieve tasks from all sources
	go DrainSinkChan(task, time.Millisecond, "drain", controller)

	// Clone all channels
	cloneOpts := &channelutil.CloneOptions{
		MaxDrain:  3,
		Threshold: 5,
	}
	srcs := []chan<- struct{}{}
	for _, v := range sources {
		srcs = append(srcs, v)
	}
	cloner := channelutil.NewCloneChannels[struct{}](cloneOpts)
	err := cloner.Clone(context.TODO(), domainChan, srcs...)
	if err != nil {
		t.Error(err)
	}

	// join all channels
	jchan := channelutil.JoinChannels[struct{}]{}
	jsrcs := []<-chan struct{}{}
	for _, v := range sources {
		jsrcs = append(jsrcs, v)
	}
	err = jchan.Join(context.TODO(), task, jsrcs...)
	if err != nil {
		t.Error(err)
	}
	controller.Wait()
}

// arities returns source and sink counts that cover 1 through 5 (a single
// worker padded with nil channels), the first counts past 5 (one relay level
// with a partial last group), a full second level (25) and a partial one (26).
func arities() []int {
	return []int{1, 2, 3, 4, 5, 6, 7, 9, 10, 11, 25, 26}
}

// workerFanIn is the number of channels a single worker multiplexes.
const workerFanIn = 5

const waitBound = 3 * time.Second

// waitReturns fails the test when wait does not return within waitBound.
func waitReturns(t *testing.T, name string, wait func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(waitBound):
		t.Fatalf("%s: Wait did not return within %v of cancel", name, waitBound)
	}
}

// requireClosed fails the test unless sink is closed, without blocking past
// waitBound when it is not.
func requireClosed(t *testing.T, sinkIdx int, sink <-chan int) {
	t.Helper()

	select {
	case _, open := <-sink:
		if open {
			t.Fatalf("sink %d delivered a value after cancel, want it closed", sinkIdx)
		}
	case <-time.After(waitBound):
		t.Fatalf("sink %d still open %v after cancel", sinkIdx, waitBound)
	}
}

// awaitSignal fails the test when nothing arrives on ch within waitBound.
func awaitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(waitBound):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestJoinValuesArrive(t *testing.T) {
	t.Parallel()

	for _, count := range arities() {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()
			checkJoinValuesArrive(t, count)
		})
	}
}

func checkJoinValuesArrive(t *testing.T, sourceCount int) {
	t.Helper()

	const perSource = 7

	sink := make(chan int)
	sources := make([]<-chan int, sourceCount)

	for idx := range sources {
		source := make(chan int)
		sources[idx] = source

		go func() {
			defer close(source)

			for j := range perSource {
				source <- idx*perSource + j
			}
		}()
	}

	joiner := channelutil.JoinChannels[int]{Log: nil}

	err := joiner.Join(context.Background(), sink, sources...)
	if err != nil {
		t.Fatal(err)
	}

	seen := drainCounting(t, sink)

	waitReturns(t, "JoinChannels", joiner.Wait)

	if len(seen) != sourceCount*perSource {
		t.Fatalf("got %d distinct values, want %d", len(seen), sourceCount*perSource)
	}

	for value, times := range seen {
		if times != 1 {
			t.Fatalf("value %d delivered %d times", value, times)
		}
	}
}

// drainCounting reads sink until it is closed and counts each value.
func drainCounting(t *testing.T, sink <-chan int) map[int]int {
	t.Helper()

	seen := map[int]int{}
	timeout := time.After(10 * time.Second)

	for {
		select {
		case value, ok := <-sink:
			if !ok {
				return seen
			}

			seen[value]++
		case <-timeout:
			t.Fatal("timed out reading joined values")
		}
	}
}

func TestCloneValuesArrive(t *testing.T) {
	t.Parallel()

	for _, count := range arities() {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()
			checkCloneValuesArrive(t, count)
		})
	}
}

func checkCloneValuesArrive(t *testing.T, sinkCount int) {
	t.Helper()

	const values = 20

	source := make(chan int)
	sinks := make([]chan<- int, sinkCount)
	got := make([][]int, sinkCount)

	var readers sync.WaitGroup

	for idx := range sinks {
		sink := make(chan int)
		sinks[idx] = sink

		readers.Go(func() {
			for value := range sink {
				got[idx] = append(got[idx], value)
			}
		})
	}

	cloner := channelutil.NewCloneChannels[int](nil)

	err := cloner.Clone(context.Background(), source, sinks...)
	if err != nil {
		t.Fatal(err)
	}

	for value := range values {
		select {
		case source <- value:
		case <-time.After(10 * time.Second):
			t.Fatal("timed out feeding source")
		}
	}

	close(source)
	waitReturns(t, "CloneChannels", cloner.Wait)
	readers.Wait()

	for idx, received := range got {
		requireSequence(t, idx, received, values)
	}
}

// requireSequence fails the test unless received holds exactly 0..want-1.
func requireSequence(t *testing.T, sinkIdx int, received []int, want int) {
	t.Helper()

	if len(received) != want {
		t.Fatalf("sink %d got %d values, want %d", sinkIdx, len(received), want)
	}

	sort.Ints(received)

	for expected, value := range received {
		if value != expected {
			t.Fatalf("sink %d missing or duplicated value near %d: %v", sinkIdx, expected, received)
		}
	}
}

// Cancelling while workers are blocked sending into a sink nobody reads must
// stop every worker and close the sink. Every library goroutine is paired with
// a WaitGroup Add and Done, so a worker that fails to exit shows up as a Wait
// that does not return.
func TestJoinCancelUnblocksWait(t *testing.T) {
	t.Parallel()

	for _, count := range arities() {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()
			checkJoinCancel(t, count)
		})
	}
}

func checkJoinCancel(t *testing.T, sourceCount int) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	// One feeder per source offers a single value and reports once a worker
	// took it. A worker takes a value before it can block on its downstream
	// send, so once any report has arrived at least one worker holds a value
	// it cannot deliver because the sink is never read, which is enough to
	// hang Wait if sends ignore ctx. The loop below waits for one report per
	// leaf worker, but those may all come from the same leaf.
	sources := make([]<-chan int, sourceCount)
	consumed := make(chan struct{}, sourceCount)

	var feeders sync.WaitGroup

	for idx := range sources {
		source := make(chan int)
		sources[idx] = source

		feeders.Go(func() {
			select {
			case source <- idx:
				consumed <- struct{}{}
			case <-ctx.Done():
			}
		})
	}

	sink := make(chan int)
	joiner := channelutil.JoinChannels[int]{Log: nil}

	err := joiner.Join(ctx, sink, sources...)
	if err != nil {
		t.Fatal(err)
	}

	for range (sourceCount + workerFanIn - 1) / workerFanIn {
		awaitSignal(t, consumed, "a worker to take a source value")
	}

	cancel()
	waitReturns(t, "JoinChannels", joiner.Wait)
	requireClosed(t, 0, sink)
	waitReturns(t, "feeders", feeders.Wait)
}

// Cancelling while the source is idle must stop every worker and close every
// sink.
func TestCloneCancelWhileIdle(t *testing.T) {
	t.Parallel()

	for _, count := range arities() {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())

			defer cancel()

			sinks, sendOnly := makeSinks(count)
			cloner := channelutil.NewCloneChannels[int](nil)

			err := cloner.Clone(ctx, make(chan int), sendOnly...)
			if err != nil {
				t.Fatal(err)
			}

			// Nothing is ever sent, so each worker waits on its source
			// whether or not it has been scheduled yet.
			cancel()
			waitReturns(t, "CloneChannels", cloner.Wait)

			for idx, sink := range sinks {
				requireClosed(t, idx, sink)
			}
		})
	}
}

// makeSinks returns count unbuffered channels, both as receivable and as
// send-only.
func makeSinks(count int) ([]chan int, []chan<- int) {
	sinks := make([]chan int, count)
	sendOnly := make([]chan<- int, count)

	for idx := range sinks {
		sinks[idx] = make(chan int)
		sendOnly[idx] = sinks[idx]
	}

	return sinks, sendOnly
}

// Cancelling while a worker is draining into sinks nobody reads must stop it.
// Only counts that fit one worker are covered: with relays, a value can end up
// in a branch whose sink is never read, so no signal tells that the reading
// worker has reached its drain.
func TestCloneCancelWhileDraining(t *testing.T) {
	t.Parallel()

	for count := 2; count <= workerFanIn; count++ {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())

			defer cancel()

			source := make(chan int)
			sinks, sendOnly := makeSinks(count)

			// With a threshold of 1, the first value sent to the one sink
			// that is read puts every other sink into the backlog, and
			// draining it blocks on a sink nobody reads.
			cloner := channelutil.NewCloneChannels[int](&channelutil.CloneOptions{MaxDrain: 3, Threshold: 1})

			err := cloner.Clone(ctx, source, sendOnly...)
			if err != nil {
				t.Fatal(err)
			}

			received := make(chan struct{})

			go func() {
				<-sinks[0]
				close(received)
			}()

			select {
			case source <- 1:
			case <-time.After(waitBound):
				t.Fatal("timed out feeding source")
			}

			awaitSignal(t, received, "the read sink to get the value")
			cancel()
			waitReturns(t, "CloneChannels", cloner.Wait)

			for idx, sink := range sinks {
				requireClosed(t, idx, sink)
			}
		})
	}
}
