package channel

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {

	var wg sync.WaitGroup

	// producer
	data := make([]int, 100)
	for i := 0; i < 100; i++ {
		data[i] = i
	}

	prod := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(prod)

		for _, i := range data {
			prod <- i
		}
	}()

	// 2 consumers
	c := Clone(prod, 2)
	wg.Add(1)
	c1 := <-c
	var cons1 []int
	go func(c <-chan int) {
		defer wg.Done()

		for i := range c {
			cons1 = append(cons1, i)
		}
	}(c1)

	wg.Add(1)
	c2 := <-c
	var cons2 []int
	go func(c <-chan int) {
		defer wg.Done()

		for i := range c {
			cons2 = append(cons2, i)
		}
	}(c2)

	wg.Wait()

	require.ElementsMatch(t, data, cons1)
	require.ElementsMatch(t, data, cons2)
}
