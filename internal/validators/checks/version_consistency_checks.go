package checks

import (
	"fmt"
	"path/filepath"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

// KustomizationVersionConsistencyCheck validates API version consistency across Kustomization dependencies
func KustomizationVersionConsistencyCheck(ctx *context.ValidationContext) []types.ValidationResult {
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
		resources := extractResources(kustomization)
		for _, resourcePath := range resources {
			// Resolve the full path
			fullPath, shouldProcess := common.ResolvePath(baseDir, resourcePath)
			if !shouldProcess {
				continue // Skip remote resources
			}

			// Check if this resource points to another kustomization
			referencedKust := findKustomizationAtPath(fullPath, kustomizationByDir)
			if referencedKust == nil {
				continue // Not a kustomization reference
			}

			// Check for version mismatch
			if kustomization.APIVersion != "" && referencedKust.APIVersion != "" {
				if !areVersionsCompatible(kustomization.APIVersion, referencedKust.APIVersion) {
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

	return results
}

// extractResources extracts resource paths from a parsed kustomization
func extractResources(kustomization *parser.ParsedResource) []string {
	resources, err := common.ExtractStringSliceFromContent(kustomization.Content, "resources")
	if err != nil {
		return []string{}
	}
	return resources
}

// findKustomizationAtPath finds a kustomization at the given path
func findKustomizationAtPath(path string, kustomizationByDir map[string]*parser.ParsedResource) *parser.ParsedResource {
	// Check if there's a kustomization.yaml file in this directory
	if kustomization, exists := kustomizationByDir[path]; exists {
		return kustomization
	}

	// Check parent directories (for cases where the path points to a subdirectory)
	dir := path
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		if kustomization, exists := kustomizationByDir[parent]; exists {
			return kustomization
		}
		dir = parent
	}

	return nil
}

// areVersionsCompatible checks if two kustomization API versions are compatible
func areVersionsCompatible(version1, version2 string) bool {
	// Both versions must be the same for compatibility
	return version1 == version2
}
