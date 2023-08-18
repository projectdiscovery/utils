package env

import (
	"os"
	"testing"
)

func TestExpandWithEnv(t *testing.T) {
	testEnvVar := "TEST_VAR"
	testEnvValue := "TestValue"
	os.Setenv(testEnvVar, testEnvValue)
	defer os.Unsetenv(testEnvVar)

	tests := []struct {
		input    string
		expected string
		name     string
	}{
		{"$" + testEnvVar, testEnvValue, "Existing env variable"},
		{"$NON_EXISTENT_VAR", "", "Non-existent env variable"},
		{"NOT_AN_ENV_VAR", "", "Not prefixed with $"},
		{"", "", "Empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExpandWithEnv(&tt.input)
			if tt.input != tt.expected {
				t.Errorf("got %q, want %q", tt.input, tt.expected)
			}
		})
	}
}

func TestExpandWithEnvNilInput(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked with %v", r)
		}
	}()

	var nilVar *string = nil
	ExpandWithEnv(nilVar)
}
