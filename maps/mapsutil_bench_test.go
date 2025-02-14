package mapsutil

import (
	"fmt"
	"testing"
)

var benchNumItems = []int{1000, 5000, 10_000, 100_000, 250_000, 500_000, 1_000_000}

func BenchmarkGet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m := make(Map[string, int], numItems)

			// Pre-populate with test data
			for i := 0; i < numItems; i++ {
				m[fmt.Sprint(i)] = i
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
	}
}

func BenchmarkSyncMapGet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			m := NewSyncLockMap[string, int]()

			// Pre-populate with test data
			for i := 0; i < numItems; i++ {
				_ = m.Set(fmt.Sprint(i), i)
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
	}
}

func BenchmarkSet(b *testing.B) {
	b.Run("seq", func(b *testing.B) {
		m := make(Map[string, int], b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Set(fmt.Sprint(i), i)
		}
	})

	// no-op: non-concurrent map is not safe for concurrent writes
	// b.Run("parallel", func(b *testing.B) {
	// 	m := make(Map[string, int], b.N)
	// 	b.ResetTimer()

	// 	b.RunParallel(func(pb *testing.PB) {
	// 		i := 0
	// 		for pb.Next() {
	// 			m.Set(fmt.Sprint(i), i)
	// 			i++
	// 		}
	// 	})
	// })
}

func BenchmarkSyncMapSet(b *testing.B) {
	for _, numItems := range benchNumItems {
		b.Run(fmt.Sprintf("items=%d", numItems), func(b *testing.B) {
			b.Run("seq", func(b *testing.B) {
				m := NewSyncLockMap[string, int]()
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = m.Set(fmt.Sprint(i), i)
				}
			})

			b.Run("parallel", func(b *testing.B) {
				m := NewSyncLockMap[string, int]()
				b.ResetTimer()

				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						_ = m.Set(fmt.Sprint(i), i)
						i++
					}
				})
			})
		})
	}
}

// func BenchmarkSyncMapConcurrent(b *testing.B) {
// 	for _, workers := range []int{2, 4, 8, 16, 32, 64} {
// 		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
// 			m := NewSyncLockMap[string, int]()
// 			var wg sync.WaitGroup

// 			b.ResetTimer()
// 			b.Run("Set", func(b *testing.B) {
// 				for i := 0; i < b.N; i++ {
// 					wg.Add(workers)
// 					for w := 0; w < workers; w++ {
// 						go func() {
// 							defer wg.Done()
// 							_ = m.Set("key", b.N)
// 						}()
// 					}
// 					wg.Wait()
// 				}
// 			})

// 			b.ResetTimer()
// 			b.Run("Get", func(b *testing.B) {
// 				for i := 0; i < b.N; i++ {
// 					wg.Add(workers)
// 					for w := 0; w < workers; w++ {
// 						go func() {
// 							defer wg.Done()
// 							_, _ = m.Get("key")
// 						}()
// 					}
// 					wg.Wait()
// 				}
// 			})
// 		})
// 	}
// }

// func BenchmarkOps(b *testing.B) {
// 	items := 1_000_000
// 	m := make(Map[string, int], items)

// 	// Pre-populate
// 	for i := 0; i < items; i++ {
// 		m[fmt.Sprint(i)] = i
// 	}

// 	b.ResetTimer()

// 	b.Run("Has", func(b *testing.B) {
// 		b.RunParallel(func(pb *testing.PB) {
// 			i := 0
// 			for pb.Next() {
// 				_ = m.Has(fmt.Sprint(i % items))
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("GetOrDefault", func(b *testing.B) {
// 		b.RunParallel(func(pb *testing.PB) {
// 			i := 0
// 			for pb.Next() {
// 				_ = m.GetOrDefault(fmt.Sprint(i%items), -1)
// 				i++
// 			}
// 		})
// 	})

// 	b.Run("GetKeys", func(b *testing.B) {
// 		keys := make([]string, 100)
// 		for i := range keys {
// 			keys[i] = fmt.Sprint(i)
// 		}
// 		b.ResetTimer()

// 		b.RunParallel(func(pb *testing.PB) {
// 			for pb.Next() {
// 				_ = m.GetKeys(keys...)
// 			}
// 		})
// 	})
// }
