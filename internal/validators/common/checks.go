package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// FileExistenceCheck validates that a file exists at the given path
func FileExistenceCheck(baseDir, filePath string) error {
	fullPath, shouldProcess := ResolvePath(baseDir, filePath)
	if !shouldProcess {
		return nil // Skip remote URLs and other non-processable paths
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist", filePath)
	}

	return nil
}

// PathValidationCheck validates that a path reference is valid
func PathValidationCheck(baseDir, path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	return FileExistenceCheck(baseDir, path)
}

// SourceValidationCheck validates that a source reference is valid
func SourceValidationCheck(ctx *context.ValidationContext, sourceName string) error {
	if sourceName == "" {
		return fmt.Errorf("source name cannot be empty")
	}

	// Check if the source exists in the repository
	// This could be enhanced to check against actual Flux sources
	// For now, we'll do a basic validation
	return nil
}

// ResourceValidationCheck validates a Kubernetes resource
func ResourceValidationCheck(resource *parser.ParsedResource) []types.ValidationResult {
	var results []types.ValidationResult

	// Basic resource validation
	if resource.APIVersion == "" {
		results = append(results, types.ValidationResult{
			Type:     "resource-validation",
			Severity: "error",
			Message:  "Resource missing apiVersion",
			File:     resource.File,
			Line:     resource.Line,
		})
	}

	if resource.Kind == "" {
		results = append(results, types.ValidationResult{
			Type:     "resource-validation",
			Severity: "error",
			Message:  "Resource missing kind",
			File:     resource.File,
			Line:     resource.Line,
		})
	}

	if resource.Name == "" {
		results = append(results, types.ValidationResult{
			Type:     "resource-validation",
			Severity: "error",
			Message:  "Resource missing metadata.name",
			File:     resource.File,
			Line:     resource.Line,
		})
	}

	return results
}

// DuplicateCheck checks for duplicate entries in a slice
func DuplicateCheck(items []string, itemType string) map[string][]int {
	duplicates := make(map[string][]int)
	seen := make(map[string]int)

	for i, item := range items {
		if prevIndex, exists := seen[item]; exists {
			if duplicates[item] == nil {
				duplicates[item] = []int{prevIndex}
			}
			duplicates[item] = append(duplicates[item], i)
		} else {
			seen[item] = i
		}
	}

	return duplicates
}

// ExtractStringFromContent extracts a string value from parsed content
func ExtractStringFromContent(content map[string]interface{}, path ...string) (string, error) {
	current := content

	for i, key := range path {
		if i == len(path)-1 {
			// Last key, return the string value
			if value, exists := current[key]; exists {
				if str, ok := value.(string); ok {
					return str, nil
				}
				return "", fmt.Errorf("value at path %v is not a string", path)
			}
			return "", fmt.Errorf("key %s not found in path %v", key, path)
		}

		// Navigate deeper
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return "", fmt.Errorf("intermediate value at path %v is not a map", path[:i+1])
			}
		} else {
			return "", fmt.Errorf("key %s not found in path %v", key, path)
		}
	}

	return "", fmt.Errorf("unexpected end of path extraction")
}

// ExtractStringSliceFromContent extracts a string slice from parsed content
func ExtractStringSliceFromContent(content map[string]interface{}, path ...string) ([]string, error) {
	current := content

	for i, key := range path {
		if i == len(path)-1 {
			// Last key, return the string slice
			if value, exists := current[key]; exists {
				if slice, ok := value.([]interface{}); ok {
					var result []string
					for _, item := range slice {
						if str, ok := item.(string); ok {
							result = append(result, str)
						}
					}
					return result, nil
				}
				return nil, fmt.Errorf("value at path %v is not a slice", path)
			}
			return nil, fmt.Errorf("key %s not found in path %v", key, path)
		}

		// Navigate deeper
		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("intermediate value at path %v is not a map", path[:i+1])
			}
		} else {
			return nil, fmt.Errorf("key %s not found in path %v", key, path)
		}
	}

	return nil, fmt.Errorf("unexpected end of path extraction")
}

// ResolvePath resolves a path relative to a base directory
func ResolvePath(baseDir, path string) (string, bool) {
	return filepath.Join(baseDir, path), true
}
