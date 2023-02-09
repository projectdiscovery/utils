package mapsutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncLockMap(t *testing.T) {
	m := &SyncLockMap[string, string]{
		Map: Map[string, string]{
			"key1": "value1",
			"key2": "value2",
		},
	}

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
