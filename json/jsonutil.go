package jsonutil

import (
	"encoding/json"
	"errors"
)

// FilterJSON filters the input JSON based on the include and exclude fields.
// - data: the original JSON data as a byte slice.
// - includeFields: list of fields to include (if empty, includes all).
// - excludeFields: list of fields to exclude (processed after include).
func FilterJSON(data []byte, includeFields, excludeFields []string) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errors.New("invalid JSON input")
	}

	filtered := make(map[string]interface{})

	includeMap := make(map[string]bool)
	excludeMap := make(map[string]bool)

	for _, field := range includeFields {
		includeMap[field] = true
	}
	for _, field := range excludeFields {
		excludeMap[field] = true
	}

	if len(includeMap) > 0 {
		for field := range includeMap {
			if val, exists := raw[field]; exists {
				filtered[field] = val
			}
		}
	} else {
		for key, val := range raw {
			filtered[key] = val
		}
	}

	for field := range excludeMap {
		delete(filtered, field)
	}

	return json.Marshal(filtered)
}

// GetJSONFields returns all the top-level fields from the given JSON data.
// - data: the original JSON data as a byte slice.
// Returns a slice of field names or an error if the JSON is invalid.
func GetJSONFields(data []byte) ([]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errors.New("invalid JSON input")
	}

	fields := make([]string, 0, len(raw))
	for key := range raw {
		fields = append(fields, key)
	}

	return fields, nil
}
