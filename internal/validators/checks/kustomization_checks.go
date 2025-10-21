package checks

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

// KustomizationResourceCheck validates resource references in Kubernetes Kustomizations
func KustomizationResourceCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract resources list
	resources, err := common.ExtractStringSliceFromContent(kustomization.Content, "resources")
	if err != nil {
		// Resources is optional, so this is not an error
		return results
	}

	// Check for duplicates
	duplicates := common.DuplicateCheck(resources, "resource")
	for resourcePath, indices := range duplicates {
		results = append(results, types.ValidationResult{
			Type:     "kustomization-resource",
			Severity: "error",
			Message:  fmt.Sprintf("Duplicate resource reference: '%s' (appears at indices: %v)", resourcePath, indices),
			File:     kustomization.File,
			Resource: kustomization.Name,
		})
	}

	// Validate each resource exists
	baseDir := ctx.RepoPath
	for _, resourcePath := range resources {
		if err := common.FileExistenceCheck(baseDir, resourcePath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kustomization-resource",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid resource reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}
	}

	return results
}

// KustomizationPatchCheck validates patch references in Kubernetes Kustomizations
func KustomizationPatchCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract patches list
	patches, err := common.ExtractStringSliceFromContent(kustomization.Content, "patches")
	if err != nil {
		// Patches is optional, so this is not an error
		return results
	}

	// Check for duplicates
	duplicates := common.DuplicateCheck(patches, "patch")
	for patchPath, indices := range duplicates {
		results = append(results, types.ValidationResult{
			Type:     "kustomization-patch",
			Severity: "error",
			Message:  fmt.Sprintf("Duplicate patch reference: '%s' (appears at indices: %v)", patchPath, indices),
			File:     kustomization.File,
			Resource: kustomization.Name,
		})
	}

	// Validate each patch exists
	baseDir := ctx.RepoPath
	for _, patchPath := range patches {
		if err := common.FileExistenceCheck(baseDir, patchPath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kustomization-patch",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid patch reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}
	}

	return results
}

// KustomizationStrategicMergeCheck validates strategic merge patch references in Kubernetes Kustomizations
func KustomizationStrategicMergeCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract patchesStrategicMerge list
	patches, err := common.ExtractStringSliceFromContent(kustomization.Content, "patchesStrategicMerge")
	if err != nil {
		// patchesStrategicMerge is optional, so this is not an error
		return results
	}

	// Validate each strategic merge patch exists
	baseDir := ctx.RepoPath
	for _, patchPath := range patches {
		if err := common.FileExistenceCheck(baseDir, patchPath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kustomization-strategic-merge",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid strategic merge patch reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}
	}

	return results
}
