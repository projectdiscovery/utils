package osutils

import (
	"os"
	"strings"
)

// UpdateWithEnv replaces a string variable with its corresponding environment variable value.
// If the environment variable does not exist, it remains unchanged.
func UpdateWithEnv(variable *string) {
	if variable == nil {
		return
	}
	*variable = os.Getenv(strings.TrimPrefix(*variable, "$"))
}
