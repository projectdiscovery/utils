package swissmap

import (
	"fmt"
	"testing"
)

var benchNumItems = []int{1000, 5000, 10_000, 100_000, 250_000, 500_000, 1_000_000}

func createMaps[K ComparableOrdered, V any](numItems int, threadSafe bool) *Map[K, V] {
	options := []Option[K, V]{
		WithCapacity[K, V](numItems),
	}

	if threadSafe {
		options = append(options, WithThreadSafety[K, V]())
	}

	return New[K, V](options...)
}

func BenchmarkGet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m := createMaps[string, int](numItems, false)
			for i := 0; i < numItems; i++ {
				m.Set(fmt.Sprint(i), i)
			}

			b.ResetTimer()
			b.Run("seq", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m.Get(fmt.Sprint(i % numItems))
				}
			})

			b.ResetTimer()
			b.Run("parallel", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						_, _ = m.Get(fmt.Sprint(i % numItems))
						i++
					}
				})
			})

			b.Run("WithThreadSafety", func(b *testing.B) {
				m := createMaps[string, int](numItems, true)
				for i := 0; i < numItems; i++ {
					m.Set(fmt.Sprint(i), i)
				}

				b.ResetTimer()
				b.Run("seq", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, _ = m.Get(fmt.Sprint(i % numItems))
					}
				})

				b.ResetTimer()
				b.Run("parallel", func(b *testing.B) {
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							_, _ = m.Get(fmt.Sprint(i % numItems))
							i++
						}
					})
				})
			})
		})
	}
}

func BenchmarkSet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m := createMaps[string, int](numItems, false)

			b.ResetTimer()
			b.Run("seq", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m.Set(fmt.Sprint(i), i)
				}
			})

			m.Clear()

			b.ResetTimer()
			b.Run("parallel", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						m.Set(fmt.Sprint(i), i)
						i++
					}
				})
			})

			b.Run("WithThreadSafety", func(b *testing.B) {
				m := createMaps[string, int](numItems, true)

				b.ResetTimer()
				b.Run("seq", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						m.Set(fmt.Sprint(i), i)
					}
				})

				m.Clear()

				b.ResetTimer()
				b.Run("parallel", func(b *testing.B) {
					b.RunParallel(func(pb *testing.PB) {
						i := 0
						for pb.Next() {
							m.Set(fmt.Sprint(i), i)
							i++
						}
					})
				})
			})

		})
	}
}
