package jsonutil

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterJSON(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		includeFields []string
		excludeFields []string
		want          string
		wantErr       bool
	}{
		{
			name:          "include specific fields",
			input:         `{"name":"John","age":30,"city":"New York"}`,
			includeFields: []string{"name", "age"},
			excludeFields: []string{},
			want:          `{"age":30,"name":"John"}`,
			wantErr:       false,
		},
		{
			name:          "exclude specific fields",
			input:         `{"name":"John","age":30,"city":"New York"}`,
			includeFields: []string{},
			excludeFields: []string{"age"},
			want:          `{"city":"New York","name":"John"}`,
			wantErr:       false,
		},
		{
			name:          "include and exclude",
			input:         `{"name":"John","age":30,"city":"New York","country":"USA"}`,
			includeFields: []string{"name", "age", "city"},
			excludeFields: []string{"age"},
			want:          `{"city":"New York","name":"John"}`,
			wantErr:       false,
		},
		{
			name:          "invalid JSON",
			input:         `{"name":"John"`,
			includeFields: []string{},
			excludeFields: []string{},
			want:          "",
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FilterJSON([]byte(tc.input), tc.includeFields, tc.excludeFields)

			if (err != nil) != tc.wantErr {
				t.Errorf("FilterJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				var gotMap, wantMap map[string]interface{}
				json.Unmarshal(got, &gotMap)
				json.Unmarshal([]byte(tc.want), &wantMap)

				if !reflect.DeepEqual(gotMap, wantMap) {
					t.Errorf("FilterJSON() = %v, want %v", string(got), tc.want)
				}
			}
		})
	}
}

func TestGetJSONFields(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   `{"name":"John","age":30,"city":"New York"}`,
			want:    []string{"name", "age", "city"},
			wantErr: false,
		},
		{
			name:    "empty JSON",
			input:   `{}`,
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"name":"John"`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetJSONFields([]byte(tc.input))

			if (err != nil) != tc.wantErr {
				t.Errorf("GetJSONFields() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				// Sort both slices to ensure consistent comparison
				gotMap := make(map[string]bool)
				wantMap := make(map[string]bool)

				for _, field := range got {
					gotMap[field] = true
				}
				for _, field := range tc.want {
					wantMap[field] = true
				}

				if !reflect.DeepEqual(gotMap, wantMap) {
					t.Errorf("GetJSONFields() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}
