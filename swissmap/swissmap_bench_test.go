package swissmap

import (
	"flag"
	"fmt"
	"testing"

	mapsutil "github.com/projectdiscovery/utils/maps"
)

var (
	benchNumItems = []int{1, 1000, 100_000, 250_000, 500_000, 1_000_000}
	items         = flag.Int("items", 0, "run benchmarks with specific number of items")
)

func createMaps[K ComparableOrdered, V any](numItems int, threadSafe bool, ordered bool) *Map[K, V] {
	options := []Option[K, V]{
		WithCapacity[K, V](numItems),
	}

	if threadSafe {
		options = append(options, WithConcurrentAccess[K, V]())
	}

	if ordered {
		options = append(options, WithSortMapKeys[K, V]())
	}

	return New[K, V](options...)
}

func BenchmarkHelper(b *testing.B) {
	if *items > 0 {
		benchNumItems = []int{*items}
	}
	b.Helper()
}

func BenchmarkGet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, false, false)
			m2 := mapsutil.Map[string, int]{}

			for i := 0; i < numItems; i++ {
				m1.Set(fmt.Sprint(i), i)
				m2.Set(fmt.Sprint(i), i)
			}

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m1.Get(fmt.Sprint(i % numItems))
				}
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m2.Get(fmt.Sprint(i % numItems))
				}
			})
		})
	}
}

func BenchmarkGet_Sorted(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, false, true)
			m2 := mapsutil.NewOrderedMap[string, int]()

			for i := 0; i < numItems; i++ {
				m1.Set(fmt.Sprint(i), i)
				m2.Set(fmt.Sprint(i), i)
			}

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m1.Get(fmt.Sprint(i % numItems))
				}
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m2.Get(fmt.Sprint(i % numItems))
				}
			})
		})
	}
}

func BenchmarkGet_Sync(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, true, false)
			m2 := mapsutil.NewSyncLockMap[string, int]()

			for i := 0; i < numItems; i++ {
				m1.Set(fmt.Sprint(i), i)
				m2.Set(fmt.Sprint(i), i)
			}

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						m1.Get(fmt.Sprint(i % numItems))
						i++
					}
				})
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						m2.Get(fmt.Sprint(i % numItems))
						i++
					}
				})
			})
		})
	}
}

// -----------------------------------------------------------------------------

func BenchmarkSet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, false, false)
			m2 := mapsutil.Map[string, int]{}

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m1.Set(fmt.Sprint(i), i)
				}
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m2.Set(fmt.Sprint(i), i)
				}
			})
		})
	}
}

func BenchmarkSet_Sorted(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, false, true)
			m2 := mapsutil.NewOrderedMap[string, int]()

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m1.Set(fmt.Sprint(i), i)
				}
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m2.Set(fmt.Sprint(i), i)
				}
			})
		})
	}
}

func BenchmarkSet_Sync(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m1 := createMaps[string, int](numItems, true, false)
			m2 := mapsutil.NewSyncLockMap[string, int]()

			b.ResetTimer()
			b.Run("impl=swissmap", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						m1.Set(fmt.Sprint(i), i)
						i++
					}
				})
			})

			b.ResetTimer()
			b.Run("impl=mapsutil", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						m2.Set(fmt.Sprint(i), i)
						i++
					}
				})
			})
		})
	}
}
