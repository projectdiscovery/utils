package swissmap

// lock conditionally acquires the read lock if thread-safety is enabled
func (m *Map[K, V]) lock() {
	if m.threadSafe {
		m.mutex.Lock()
	}
}

// unlock conditionally releases the read lock if thread-safety is enabled
func (m *Map[K, V]) unlock() {
	if m.threadSafe {
		m.mutex.Unlock()
	}
}

// rLock conditionally acquires the read lock if thread-safety is enabled
func (m *Map[K, V]) rLock() {
	if m.threadSafe {
		m.mutex.RLock()
	}
}

// rUnlock conditionally releases the read lock if thread-safety is enabled
func (m *Map[K, V]) rUnlock() {
	if m.threadSafe {
		m.mutex.RUnlock()
	}
}
