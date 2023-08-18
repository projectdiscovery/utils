package env

import (
	"os"
	"strings"
)

var (
	TLS_VERIFY = os.Getenv("TLS_VERIFY") == "true"
	DEBUG      = os.Getenv("DEBUG") == "true"
)

// ExpandWithEnv updates string variables to their corresponding environment values.
// If the variables does not exist, they're set to empty strings.
func ExpandWithEnv(variables ...*string) {
	for _, variable := range variables {
		if variable == nil {
			continue
		}
		*variable = os.Getenv(strings.TrimPrefix(*variable, "$"))
	}
}
