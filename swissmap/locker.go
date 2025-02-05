package swissmap

// lock conditionally acquires the read lock if thread-safety is enabled
func (m *Map[K, V]) lock() bool {
	var locked bool

	if m.concurrent {
		m.mutex.Lock()
		locked = true
	}

	return locked
}

// unlock conditionally releases the read lock if thread-safety is enabled
func (m *Map[K, V]) unlock() {
	if m.concurrent {
		m.mutex.Unlock()
	}
}

// rLock conditionally acquires the read lock if thread-safety is enabled
func (m *Map[K, V]) rLock() bool {
	var locked bool

	if m.concurrent {
		m.mutex.RLock()
		locked = true
	}

	return locked
}

// rUnlock conditionally releases the read lock if thread-safety is enabled
func (m *Map[K, V]) rUnlock() {
	if m.concurrent {
		m.mutex.RUnlock()
	}
}
