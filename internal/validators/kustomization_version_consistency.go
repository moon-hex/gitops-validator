package validators

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
)

type KustomizationVersionConsistencyValidator struct {
	repoPath string
}

func NewKustomizationVersionConsistencyValidator(repoPath string) *KustomizationVersionConsistencyValidator {
	return &KustomizationVersionConsistencyValidator{
		repoPath: repoPath,
	}
}

func (v *KustomizationVersionConsistencyValidator) Name() string {
	return "Kustomization Version Consistency Validator"
}

// Validate implements the GraphValidator interface
func (v *KustomizationVersionConsistencyValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Kubernetes Kustomization resources from the graph
	kustomizations := ctx.Graph.GetKubernetesKustomizations()

	// Build a map of directory -> kustomization info for quick lookups
	kustomizationByDir := make(map[string]*parser.ParsedResource)
	for _, k := range kustomizations {
		dir := filepath.Dir(k.File)
		kustomizationByDir[dir] = k
	}

	// Check each kustomization's resource references
	for _, kustomization := range kustomizations {
		baseDir := filepath.Dir(kustomization.File)

		// Extract resources from the parsed content
		resources := v.extractResources(kustomization)
		for _, resourcePath := range resources {
			// Resolve the full path
			fullPath, shouldProcess := ResolvePath(baseDir, resourcePath)
			if !shouldProcess {
				continue // Skip remote resources
			}

			// Check if this resource points to another kustomization
			referencedKust := v.findKustomizationAtPath(fullPath, kustomizationByDir)
			if referencedKust == nil {
				continue // Not a kustomization reference
			}

			// Check for version mismatch
			if kustomization.APIVersion != "" && referencedKust.APIVersion != "" {
				if !v.areVersionsCompatible(kustomization.APIVersion, referencedKust.APIVersion) {
					results = append(results, types.ValidationResult{
						Type:     "kustomization-version-consistency",
						Severity: "error",
						Message: fmt.Sprintf(
							"Kustomization apiVersion mismatch: '%s' references '%s' (version: %s) but uses version %s",
							kustomization.File,
							resourcePath,
							referencedKust.APIVersion,
							kustomization.APIVersion,
						),
						File: kustomization.File,
					})
				}
			}
		}
	}

	return results, nil
}

// extractResources extracts resource paths from a parsed kustomization
func (v *KustomizationVersionConsistencyValidator) extractResources(kustomization *parser.ParsedResource) []string {
	var resources []string

	// Extract resources from the parsed content
	if resourcesInterface, exists := kustomization.Content["resources"]; exists {
		if resourcesList, ok := resourcesInterface.([]interface{}); ok {
			for _, resource := range resourcesList {
				if resourcePath, ok := resource.(string); ok {
					resources = append(resources, resourcePath)
				}
			}
		}
	}

	return resources
}

// findKustomizationAtPath checks if the given path contains or is a kustomization
func (v *KustomizationVersionConsistencyValidator) findKustomizationAtPath(
	path string,
	kustomizationByDir map[string]*parser.ParsedResource,
) *parser.ParsedResource {
	// Normalize path
	path = filepath.Clean(path)

	// Check if it's a directory
	if kust, exists := kustomizationByDir[path]; exists {
		return kust
	}

	// It's a file, check if it's in a directory with a kustomization
	dir := filepath.Dir(path)
	if kust, exists := kustomizationByDir[dir]; exists {
		return kust
	}

	return nil
}

// areVersionsCompatible checks if two kustomization apiVersions are compatible
func (v *KustomizationVersionConsistencyValidator) areVersionsCompatible(version1, version2 string) bool {
	// Versions should match exactly for consistency
	// We're checking kustomize.config.k8s.io versions
	if !strings.HasPrefix(version1, "kustomize.config.k8s.io/") {
		return true // Not a kustomize config version, skip check
	}
	if !strings.HasPrefix(version2, "kustomize.config.k8s.io/") {
		return true // Not a kustomize config version, skip check
	}

	return version1 == version2
}
