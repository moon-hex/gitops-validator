package validators

import (
	"fmt"
	"path/filepath"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

type OrphanedResourceValidator struct {
	repoPath string
}

func NewOrphanedResourceValidator(repoPath string) *OrphanedResourceValidator {
	return &OrphanedResourceValidator{
		repoPath: repoPath,
	}
}

func (v *OrphanedResourceValidator) Name() string {
	return "Orphaned Resource Validator"
}

// Validate implements the GraphValidator interface
func (v *OrphanedResourceValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
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

	return results, nil
}
