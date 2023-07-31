package mapsutil

import (
	"errors"
	"testing"

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
		m := NewSyncLockMap[string, string](WithMap(Map[string, string]{
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
