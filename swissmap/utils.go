package swissmap

import "github.com/bytedance/sonic"

// getDefaultSonicConfig provides the default configuration for the Sonic
// library.
//
// This function returns a [sonic.Config] instance with standard `encoding/json`
// settings but with unsorted map keys. You may want to use the [WithSortMapKeys]
// option to enable sorting of map keys.
func getDefaultSonicConfig() sonic.Config {
	return sonic.Config{
		EscapeHTML:       true,
		CompactMarshaler: true,
		CopyString:       true,
		ValidateString:   true,
	}
}
