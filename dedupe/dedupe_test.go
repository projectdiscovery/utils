package dedupe

import (
	"testing"
)

func TestDedupe(t *testing.T) {
	t.Run("MapBackend", func(t *testing.T) {
		receiveCh := make(chan string, 10)
		resultCh := make(chan string, 10)
		dedupe := NewDedupe(receiveCh, resultCh, 1)

		receiveCh <- "test1"
		receiveCh <- "test2"
		receiveCh <- "test1"
		close(receiveCh)

		go dedupe.Drain()

		results := collectResults(resultCh)

		if len(results) != 2 {
			t.Fatalf("expected 2 unique items, got %d", len(results))
		}
	})

	t.Run("LevelDBBackend", func(t *testing.T) {
		receiveCh := make(chan string, 10)
		resultCh := make(chan string, 10)
		dedupe := NewDedupe(receiveCh, resultCh, MaxInMemoryDedupeSize+1)

		receiveCh <- "testA"
		receiveCh <- "testB"
		receiveCh <- "testA"
		close(receiveCh)

		go dedupe.Drain()

		results := collectResults(resultCh)

		if len(results) != 2 {
			t.Fatalf("expected 2 unique items, got %d", len(results))
		}
	})

	t.Run("Drain", func(t *testing.T) {
		receiveCh := make(chan string, 10)
		resultCh := make(chan string, 10)
		dedupe := NewDedupe(receiveCh, resultCh, 1)

		go func() {
			receiveCh <- "testX"
			receiveCh <- "testY"
			receiveCh <- "testX"
			close(receiveCh)
		}()

		dedupe.Drain()

		results := collectResults(resultCh)

		if len(results) != 2 {
			t.Fatalf("expected 2 unique items, got %d", len(results))
		}
	})
}

func collectResults(ch <-chan string) []string {
	var results []string
	for item := range ch {
		results = append(results, item)
	}
	return results
}
