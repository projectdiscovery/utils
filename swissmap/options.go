package swissmap

// Option represents a configuration option for the Map
type Option[K, V comparable] func(*Map[K, V])

// WithCapacity sets the initial capacity of the map
func WithCapacity[K, V comparable](capacity int) Option[K, V] {
	return func(m *Map[K, V]) {
		m.data.Init(capacity)
	}
}

// WithThreadSafety enables thread-safety for the map
func WithThreadSafety[K, V comparable]() Option[K, V] {
	return func(m *Map[K, V]) {
		m.threadSafe = true
	}
}

// WithSortMapKeys enables sorting of map keys
func WithSortMapKeys[K, V comparable]() Option[K, V] {
	cfg := getDefaultSonicConfig()
	cfg.SortMapKeys = true

	return func(m *Map[K, V]) {
		m.sorted = true
		m.api = cfg.Froze()
	}
}
