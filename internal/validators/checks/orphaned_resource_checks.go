package checks

import (
	"fmt"
	"path/filepath"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// OrphanedResourceCheck identifies orphaned YAML files not referenced by any kustomization
func OrphanedResourceCheck(ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

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
		})
	}

	return results
}
