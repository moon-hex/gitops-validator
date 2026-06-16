package checks

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// OrphanedResourceCheck identifies orphaned YAML files not referenced by any kustomization
func OrphanedResourceCheck(ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	categories := ctx.Config.GetOrphanedCategories()

	// Find entry points using the context
	entryPoints := ctx.FindEntryPoints()

	// Find orphaned resources using the context
	orphanedResources := ctx.FindOrphanedResources(entryPoints)

	// Report orphaned resources
	for _, orphaned := range orphanedResources {
		// Skip config files and other ignored files
		relPath, err := filepath.Rel(ctx.RepoPath, orphaned.File)
		if err != nil {
			continue
		}

		if ctx.Config.ShouldIgnorePath(relPath) {
			continue
		}

		results = append(results, types.ValidationResult{
			Type:     "orphaned-resource",
			Severity: "warning",
			Message:  fmt.Sprintf("File '%s' is not referenced by any kustomization and is not an entry point", filepath.Base(orphaned.File)),
			File:     orphaned.File,
			Resource: orphaned.Name,
			Category: matchOrphanedCategory(relPath, categories),
		})
	}

	return results
}

// matchOrphanedCategory returns the name of the first category whose path patterns
// match relPath (forward-slash normalised, relative to repo root), or "" if none match.
func matchOrphanedCategory(relPath string, categories []config.OrphanedResourceCategoryConfig) string {
	norm := filepath.ToSlash(relPath)
	for _, cat := range categories { // already sorted by priority
		for _, pattern := range cat.Paths {
			normPattern := filepath.ToSlash(pattern)
			if matchPathPattern(norm, normPattern) {
				return cat.Name
			}
		}
	}
	return ""
}

// matchPathPattern matches a normalised relative path against a glob pattern.
// Patterns ending with /** match any file whose path starts with the prefix.
// Other patterns are matched using filepath.Match against the full path and
// also against just the leading path components for prefix-style matches.
func matchPathPattern(path, pattern string) bool {
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}
