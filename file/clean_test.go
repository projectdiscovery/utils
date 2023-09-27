package fileutil

import (
	"path/filepath"
	"testing"
)

func FuzzCleanPath(f *testing.F) {
	// // Define your custom payloads here
	// customPayloads := []string{
	// 	"../../etc/passwd",
	// 	"/absolute/path/to/file",
	// 	"./relative/path",
	// 	// Add more payloads as needed
	// }

	// // Use each custom payload for fuzzing
	// for _, payload := range customPayloads {
	// 	f.Add(payload)
	// }

	f.Fuzz(func(t *testing.T, inputPath string) {
		result, err := CleanPath(inputPath)
		if err != nil {
			t.Fatal(err)
		}

		// You can add more assertions here based on your requirements
		// For example, you might want to check if the returned path is absolute
		if !filepath.IsAbs(result) {
			t.Errorf("CleanPath(%q) returned a non-absolute path", result)
		}
	})
}
