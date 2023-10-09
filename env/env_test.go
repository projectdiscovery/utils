package env

import (
	"os"
	"testing"
	"time"
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

func TestGetEnvOrDefault(t *testing.T) {
	// Test for string
	os.Setenv("TEST_STRING", "test")
	resultString := GetEnvOrDefault("TEST_STRING", "default")
	if resultString != "test" {
		t.Errorf("Expected 'test', got %s", resultString)
	}

	// Test for int
	os.Setenv("TEST_INT", "123")
	resultInt := GetEnvOrDefault("TEST_INT", 0)
	if resultInt != 123 {
		t.Errorf("Expected 123, got %d", resultInt)
	}

	// Test for bool
	os.Setenv("TEST_BOOL", "true")
	resultBool := GetEnvOrDefault("TEST_BOOL", false)
	if resultBool != true {
		t.Errorf("Expected true, got %t", resultBool)
	}

	// Test for float64
	os.Setenv("TEST_FLOAT", "1.23")
	resultFloat := GetEnvOrDefault("TEST_FLOAT", 0.0)
	if resultFloat != 1.23 {
		t.Errorf("Expected 1.23, got %f", resultFloat)
	}

	// Test for time.Duration
	os.Setenv("TEST_DURATION", "1h")
	resultDuration := GetEnvOrDefault("TEST_DURATION", time.Duration(0))
	if resultDuration != time.Hour {
		t.Errorf("Expected 1h, got %s", resultDuration)
	}

	// Test for default value
	resultDefault := GetEnvOrDefault("NON_EXISTING", "default")
	if resultDefault != "default" {
		t.Errorf("Expected 'default', got %s", resultDefault)
	}
}
