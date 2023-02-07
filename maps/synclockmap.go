package mapsutil

// SyncLock adds sync and lock capabilities to map[string]interface{}
func SyncLockMap(m map[string]interface{}) (*SyncLock, error) {
	sl := &SyncLock{}
	sl.GetCallback = func(k interface{}) (any, bool) {
		v, ok := m[k.(string)]
		return v, ok
	}
	sl.SetCallback = func(k, v interface{}) error {
		m[k.(string)] = v
		return nil
	}
	sl.IterateCallback = func(f func(k, v interface{}) error) error {
		for k, v := range m {
			if err := f(k, v); err != nil {
				return err
			}
		}
		return nil
	}
	return sl, nil
}
