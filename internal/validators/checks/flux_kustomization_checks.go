package checks

import (
	"fmt"
	"strings"

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

	// spec.path is relative to the source repository named in sourceRef, not this
	// repo. When the source is an external GitRepository/OCIRepository we cannot
	// check the path against the local filesystem.
	if isExternalSourceRef(kustomization, ctx) {
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

// isExternalSourceRef returns true when the Flux Kustomization's sourceRef resolves
// to a GitRepository or OCIRepository with a remote URL. In that case spec.path is
// relative to the remote source and cannot be validated against the local filesystem.
func isExternalSourceRef(kustomization *parser.ParsedResource, ctx *context.ValidationContext) bool {
	sourceRefKind, _ := common.ExtractStringFromContent(kustomization.Content, "spec", "sourceRef", "kind")
	if sourceRefKind != "GitRepository" && sourceRefKind != "OCIRepository" {
		return false
	}

	sourceRefName, err := common.ExtractStringFromContent(kustomization.Content, "spec", "sourceRef", "name")
	if err != nil || sourceRefName == "" {
		return false
	}

	// Look up by kind+name to avoid matching a same-named Namespace or other
	// cluster-scoped resource whose key collides in the Resources map.
	source := findSourceByKindAndName(ctx, sourceRefKind, sourceRefName)
	if source == nil {
		// Source not found locally — likely defined in another repo. Be conservative
		// and skip path validation to avoid false positives.
		return true
	}

	url, err := common.ExtractStringFromContent(source.Content, "spec", "url")
	if err != nil || url == "" {
		return false
	}

	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "ssh://") ||
		strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "git://")
}

// findSourceByKindAndName returns the first resource matching both kind and name.
// Using GetResource(name) alone can return a wrong resource when an unrelated
// cluster-scoped resource (e.g. a Namespace) shares the same name as the
// GitRepository/OCIRepository being looked up.
func findSourceByKindAndName(ctx *context.ValidationContext, kind, name string) *parser.ParsedResource {
	for _, r := range ctx.Graph.GetResourcesByKind(kind) {
		if r.Name == name {
			return r
		}
	}
	return nil
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
