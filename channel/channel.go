package channel

// Clone makes n copies of a channel
// they should be used as
// gen := Clone(in, 2)
// consumer1 := <- gen // pop consumer 1
// consumer2 := <- gen // pop consumer 2
// each consumer can be used with for x := range consumerN { ... }
func Clone[T any](in chan T, n int) <-chan <-chan T {
	ret := make(chan (<-chan T), n)
	out := make([]chan T, n)
	for i := 0; i < n; i++ {
		out[i] = make(chan T, cap(in))
		ret <- out[i]
	}

	go func() {
		for {
			msg, ok := <-in
			if ok {
				for _, ch := range out {
					ch <- msg
				}
			} else {
				for _, ch := range out {
					close(ch)
				}
				return
			}
		}
	}()

	return ret
}
