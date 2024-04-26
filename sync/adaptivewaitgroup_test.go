package sync

import (
	"testing"

	"github.com/fortytw2/leaktest"
)

func Test_AdaptiveWaitGroup_Leak(t *testing.T) {
	defer leaktest.Check(t)()

	wg, err := New(WithSize(10))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		wg.Add()
		go func(awg *AdaptiveWaitGroup) error {
			defer awg.Done()
			return nil
		}(wg)
	}
	wg.Wait()
}
