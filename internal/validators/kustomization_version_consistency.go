package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
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

type KustomizationInfo struct {
	FilePath   string
	APIVersion string
	Resources  []string // Paths to resource references
}

func (v *KustomizationVersionConsistencyValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all kustomization.yaml files and parse their info
	kustomizations, err := v.findAllKustomizations()
	if err != nil {
		return results, fmt.Errorf("failed to find kustomizations: %w", err)
	}

	// Build a map of directory -> kustomization info for quick lookups
	kustomizationByDir := make(map[string]*KustomizationInfo)
	for _, k := range kustomizations {
		dir := filepath.Dir(k.FilePath)
		kustomizationByDir[dir] = k
	}

	// Check each kustomization's resource references
	for _, kustomization := range kustomizations {
		baseDir := filepath.Dir(kustomization.FilePath)

		for _, resourcePath := range kustomization.Resources {
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
							kustomization.FilePath,
							resourcePath,
							referencedKust.APIVersion,
							kustomization.APIVersion,
						),
						File: kustomization.FilePath,
					})
				}
			}
		}
	}

	return results, nil
}

// findKustomizationAtPath checks if the given path contains or is a kustomization
func (v *KustomizationVersionConsistencyValidator) findKustomizationAtPath(
	path string,
	kustomizationByDir map[string]*KustomizationInfo,
) *KustomizationInfo {
	// Normalize path
	path = filepath.Clean(path)

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	if info.IsDir() {
		// Look for kustomization in this directory
		if kust, exists := kustomizationByDir[path]; exists {
			return kust
		}
	} else {
		// It's a file, check if it's in a directory with a kustomization
		dir := filepath.Dir(path)
		if kust, exists := kustomizationByDir[dir]; exists {
			return kust
		}
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

func (v *KustomizationVersionConsistencyValidator) findAllKustomizations() ([]*KustomizationInfo, error) {
	var kustomizations []*KustomizationInfo

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "kustomization.yaml" || info.Name() == "kustomization.yml" {
			kust, err := v.parseKustomization(path)
			if err != nil {
				// Skip files that can't be parsed
				return nil
			}
			if kust != nil {
				kustomizations = append(kustomizations, kust)
			}
		}

		return nil
	})

	return kustomizations, err
}

func (v *KustomizationVersionConsistencyValidator) parseKustomization(filePath string) (*KustomizationInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var kustomization map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&kustomization); err != nil {
		return nil, err
	}

	info := &KustomizationInfo{
		FilePath:  filePath,
		Resources: []string{},
	}

	// Extract apiVersion
	if apiVersion, ok := kustomization["apiVersion"].(string); ok {
		info.APIVersion = apiVersion
	}

	// Extract resources
	if resources, ok := kustomization["resources"].([]interface{}); ok {
		for _, resource := range resources {
			if resourcePath, ok := resource.(string); ok {
				info.Resources = append(info.Resources, resourcePath)
			}
		}
	}

	return info, nil
}

