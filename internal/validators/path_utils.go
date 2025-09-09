package validators

import (
	"path/filepath"
	"strings"
)

// NormalizePath normalizes a path by handling different path formats:
// - Strips "./" prefix if present
// - Handles absolute paths (starting with "/")
// - Handles relative paths
// - Skips remote URLs (http://, https://)
func NormalizePath(path string) (string, bool) {
	// Skip remote URLs
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return "", false
	}

	// Handle absolute paths
	if strings.HasPrefix(path, "/") {
		return path, true
	}

	// Handle relative paths - strip ./ prefix if present
	cleanPath := strings.TrimPrefix(path, "./")
	return cleanPath, true
}

// ResolvePath resolves a path relative to a base directory, handling all path formats
func ResolvePath(baseDir, path string) (string, bool) {
	normalizedPath, shouldProcess := NormalizePath(path)
	if !shouldProcess {
		return "", false
	}

	return filepath.Join(baseDir, normalizedPath), true
}
