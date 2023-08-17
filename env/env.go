package env

import (
	"os"
	"strings"
)

// UpdateWithEnv updates string variables to their corresponding environment values.
// If the variables does not exist, they're set to empty strings.
func UpdateWithEnv(variables ...*string) {
	for _, variable := range variables {
		if variable == nil {
			continue
		}
		*variable = os.Getenv(strings.TrimPrefix(*variable, "$"))
	}
}
