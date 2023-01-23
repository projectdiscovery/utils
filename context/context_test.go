package contextutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithValues(t *testing.T) {
	type testCase struct {
		name          string
		keyValue      []string
		expectedError error
		expectedValue map[string]string
	}

	var testCases = []testCase{
		{
			name:          "even number of key-value pairs",
			keyValue:      []string{"key1", "value1", "key2", "value2"},
			expectedError: nil,
			expectedValue: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:          "odd number of key-value pairs",
			keyValue:      []string{"key1", "value1", "key2"},
			expectedError: ErrIncorrectNumberOfItems,
			expectedValue: map[string]string{},
		},
		{
			name:          "overwriting values",
			keyValue:      []string{"key1", "value1", "key1", "newValue"},
			expectedError: nil,
			expectedValue: map[string]string{"key1": "newValue"},
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newCtx, err := WithValues(ctx, tc.keyValue...)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
				require.Equal(t, ctx, newCtx, "Expected original context to be returned")
			}

			for key, expectedVal := range tc.expectedValue {
				if val := newCtx.Value(key); val != expectedVal {
					t.Errorf("Expected %s but got %v", expectedVal, val)
				}
			}
		})
	}
}
