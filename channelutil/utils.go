package channelutil

import "context"

// CreateNChannels creates and returns N channels
func CreateNChannels[T any](count int, bufflen int) map[int]chan T {
	x := map[int]chan T{}

	for i := 0; i < count; i++ {
		x[i] = make(chan T, bufflen)
	}
	return x
}

// sendCtx sends v on ch and reports whether it was sent. It gives up and
// returns false once ctx is done, so a worker blocked on a consumer that
// stopped reading can still exit.
func sendCtx[T any](ctx context.Context, ch chan<- T, v T) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- v:
		return true
	}
}
