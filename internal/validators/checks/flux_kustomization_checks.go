package checks

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

// FluxKustomizationPathCheck validates path references in Flux Kustomizations
func FluxKustomizationPathCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract path from the parsed resource
	path, err := common.ExtractStringFromContent(kustomization.Content, "spec", "path")
	if err != nil {
		results = append(results, types.ValidationResult{
			Type:     "flux-kustomization-path",
			Severity: "error",
			Message:  fmt.Sprintf("Invalid path specification: %s", err.Error()),
			File:     kustomization.File,
			Resource: kustomization.Name,
		})
		return results
	}

	// Validate path exists
	baseDir := ctx.RepoPath
	if err := common.PathValidationCheck(baseDir, path); err != nil {
		results = append(results, types.ValidationResult{
			Type:     "flux-kustomization-path",
			Severity: "error",
			Message:  fmt.Sprintf("Invalid path reference: %s", err.Error()),
			File:     kustomization.File,
			Resource: kustomization.Name,
		})
	}

	return results
}

// FluxKustomizationSourceCheck validates source references in Flux Kustomizations
func FluxKustomizationSourceCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract source reference
	sourceRef, err := common.ExtractStringFromContent(kustomization.Content, "spec", "sourceRef", "name")
	if err != nil {
		// SourceRef is optional, so this is not an error
		return results
	}

	// Validate source reference
	if err := common.SourceValidationCheck(ctx, sourceRef); err != nil {
		results = append(results, types.ValidationResult{
			Type:     "flux-kustomization-source",
			Severity: "error",
			Message:  fmt.Sprintf("Invalid source reference: %s", err.Error()),
			File:     kustomization.File,
			Resource: kustomization.Name,
		})
	}

	return results
}
