package binding

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// AbsolutePath resolves a trimmed generation path against the working directory.
func AbsolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}
	return absPath, nil
}

// OptionalAbsolutePath preserves an omitted optional output path.
func OptionalAbsolutePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	return AbsolutePath(path)
}

// SortedMapKeys provides deterministic import validation order.
func SortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
