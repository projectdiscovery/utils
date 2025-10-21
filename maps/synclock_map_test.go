package mapsutil

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

const IterateCount = 100

func TestSyncLockMap(t *testing.T) {
	m := &SyncLockMap[string, string]{
		Map: Map[string, string]{
			"key1": "value1",
			"key2": "value2",
		},
	}

	t.Run("Test NewSyncLockMap with map ", func(t *testing.T) {
		m := NewSyncLockMap(WithMap(Map[string, string]{
			"key1": "value1",
			"key2": "value2",
		}))

		if !m.Has("key1") || !m.Has("key2") {
			t.Error("couldn't init SyncLockMap with NewSyncLockMap")
		}
	})

	t.Run("Test NewSyncLockMap without map", func(t *testing.T) {
		m := NewSyncLockMap[string, string]()
		_ = m.Set("key1", "value1")
		_ = m.Set("key2", "value2")

		if !m.Has("key1") || !m.Has("key2") {
			t.Error("couldn't init SyncLockMap with NewSyncLockMap")
		}
	})

	t.Run("Test lock", func(t *testing.T) {
		m.Lock()
		if m.ReadOnly.Load() != true {
			t.Error("failed to lock map")
		}
	})

	t.Run("Test unlock", func(t *testing.T) {
		m.Unlock()
		if m.ReadOnly.Load() != false {
			t.Error("failed to unlock map")
		}
	})

	t.Run("Test set", func(t *testing.T) {
		if err := m.Set("key3", "value3"); err != nil {
			t.Error("failed to set item in map")
		}
		v, ok := m.Map["key3"]
		if !ok || v != "value3" {
			t.Error("failed to set item in map")
		}
	})

	t.Run("Test delete", func(t *testing.T) {
		_ = m.Set("key5", "value5")
		if !m.Has("key5") {
			t.Error("couldn't set item to delete")
		}
		m.Delete("key5")
		if m.Has("key5") {
			t.Error("couldn't delete item")
		}
	})

	t.Run("Test set error", func(t *testing.T) {
		m.Lock()
		if err := m.Set("key4", "value4"); err != ErrReadOnly {
			t.Error("expected read only error")
		}
	})

	t.Run("Test get", func(t *testing.T) {
		v, ok := m.Get("key2")
		if !ok || v != "value2" {
			t.Error("failed to get item from map")
		}
	})

	t.Run("Test iterate", func(t *testing.T) {
		err := m.Iterate(func(k string, v string) error {
			if k != "key1" && k != "key2" && k != "key3" {
				return errors.New("invalid key")
			}
			return nil
		})
		require.Nil(t, err, "failed to iterate map")
	})
}

func TestSyncLockMapWithConcurrency(t *testing.T) {
	internalMap := Map[string, string]{
		"key1": "value1",
		"key2": "value2",
	}

	m := &SyncLockMap[string, string]{
		Map: internalMap,
	}

	t.Run("Test Clone", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			syncMap := m.Clone()
			maps.Equal(internalMap, syncMap.Map)
		}
	})

	t.Run("Test GetKeys", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			value, ok := m.Get("key1")
			require.True(t, ok, "failed to get item from map")
			require.Equal(t, "value1", value, "failed to get item from map")
		}
	})

	t.Run("Test Set", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			err := m.Set("key3", "value3")
			require.Nil(t, err, "failed to set item in map")
			value, ok := m.Get("key3")
			require.True(t, ok, "failed to get item from map")
			require.Equal(t, "value3", value, "failed to get item from map")
		}
	})

	t.Run("Test Has", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			ok := m.Has("key2")
			require.True(t, ok, "failed to check item in map")
		}
	})

	t.Run("Test Has Negative", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			ok := m.Has("key4")
			require.False(t, ok, "failed to check item in map")
		}
	})

	t.Run("Test IsEmpty", func(t *testing.T) {
		emptyMap := &SyncLockMap[string, string]{}
		for i := 0; i < IterateCount; i++ {
			ok := m.IsEmpty()
			require.False(t, ok, "failed to check item in map")
			ok = emptyMap.IsEmpty()
			require.True(t, ok, "failed to check item in map")
		}
	})

	t.Run("Test GetAll", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			values := m.GetAll()
			require.Equal(t, 3, len(values), "failed to get all items from map")
		}
	})

	t.Run("Test GetKeyWithValue", func(t *testing.T) {
		for i := 0; i < IterateCount; i++ {
			key, ok := m.GetKeyWithValue("value1")
			require.True(t, ok, "failed to get key from map")
			require.Equal(t, "key1", key, "failed to get key from map")
		}
	})
}

func TestSyncLockMapWithEviction(t *testing.T) {
	t.Run("Test WithEviction option", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](100 * time.Millisecond))
		defer m.StopEviction()

		require.NotNil(t, m.evictionMap, "eviction map should be initialized")
		require.Equal(t, 100*time.Millisecond, m.inactivityDuration, "inactivity duration should be set")
	})

	t.Run("Test eviction after inactivity", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](50 * time.Millisecond))
		defer m.StopEviction()

		// Add some items
		err := m.Set("key1", "value1")
		require.NoError(t, err)
		err = m.Set("key2", "value2")
		require.NoError(t, err)

		// Verify items exist
		require.True(t, m.Has("key1"))
		require.True(t, m.Has("key2"))

		// Wait for eviction (wait longer to ensure eviction happens)
		time.Sleep(200 * time.Millisecond)

		// Items should be evicted
		require.False(t, m.Has("key1"))
		require.False(t, m.Has("key2"))
	})

	t.Run("Test access resets eviction timer", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](100 * time.Millisecond))
		defer m.StopEviction()

		// Add an item
		err := m.Set("key1", "value1")
		require.NoError(t, err)

		// Access the item before eviction
		time.Sleep(60 * time.Millisecond)
		_, ok := m.Get("key1")
		require.True(t, ok, "item should still exist after access")

		// Wait a bit more but not enough for eviction
		time.Sleep(60 * time.Millisecond)
		_, ok = m.Get("key1")
		require.True(t, ok, "item should still exist after recent access")

		// Now wait for eviction
		time.Sleep(150 * time.Millisecond)
		_, ok = m.Get("key1")
		require.False(t, ok, "item should be evicted after inactivity")
	})

	t.Run("Test Set updates access time", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](100 * time.Millisecond))
		defer m.StopEviction()

		// Add an item
		err := m.Set("key1", "value1")
		require.NoError(t, err)

		// Update the item before eviction
		time.Sleep(60 * time.Millisecond)
		err = m.Set("key1", "value1_updated")
		require.NoError(t, err)

		// Wait a bit more but not enough for eviction
		time.Sleep(60 * time.Millisecond)
		value, ok := m.Get("key1")
		require.True(t, ok, "item should still exist after update")
		require.Equal(t, "value1_updated", value, "value should be updated")

		// Now wait for eviction
		time.Sleep(150 * time.Millisecond)
		_, ok = m.Get("key1")
		require.False(t, ok, "item should be evicted after inactivity")
	})

	t.Run("Test Delete removes from eviction tracking", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](50 * time.Millisecond))
		defer m.StopEviction()

		// Add an item
		err := m.Set("key1", "value1")
		require.NoError(t, err)

		// Delete the item
		m.Delete("key1")
		require.False(t, m.Has("key1"))

		// Wait and verify it's not in eviction map
		time.Sleep(100 * time.Millisecond)
		require.False(t, m.Has("key1"))
	})

	t.Run("Test Clone with eviction", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](100 * time.Millisecond))
		defer m.StopEviction()

		// Add some items
		err := m.Set("key1", "value1")
		require.NoError(t, err)
		err = m.Set("key2", "value2")
		require.NoError(t, err)

		// Clone the map
		cloned := m.Clone()
		defer cloned.StopEviction()

		// Verify cloned map has the same items
		require.True(t, cloned.Has("key1"))
		require.True(t, cloned.Has("key2"))

		// Verify cloned map has eviction enabled
		require.Equal(t, m.inactivityDuration, cloned.inactivityDuration)
		require.NotNil(t, cloned.evictionMap)
	})

	t.Run("Test StopEviction", func(t *testing.T) {
		m := NewSyncLockMap(WithEviction[string, string](50 * time.Millisecond))

		// Add an item
		err := m.Set("key1", "value1")
		require.NoError(t, err)

		// Stop eviction
		m.StopEviction()

		// Wait for what would normally be eviction time
		time.Sleep(100 * time.Millisecond)

		// Item should still exist since eviction is stopped
		require.True(t, m.Has("key1"))
	})
}
