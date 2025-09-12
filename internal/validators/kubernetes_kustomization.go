package validators

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
)

type KubernetesKustomizationValidator struct {
	repoPath string
}

func NewKubernetesKustomizationValidator(repoPath string) *KubernetesKustomizationValidator {
	return &KubernetesKustomizationValidator{
		repoPath: repoPath,
	}
}

func (v *KubernetesKustomizationValidator) Name() string {
	return "Kubernetes Kustomization Validator"
}

func (v *KubernetesKustomizationValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all kustomization.yaml files
	kustomizationFiles, err := v.findKustomizationFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find kustomization files: %w", err)
	}

	for _, kustomizationFile := range kustomizationFiles {
		// Validate resource references
		if err := v.validateResourceReferences(kustomizationFile); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid resource references: %s", err.Error()),
				File:     kustomizationFile,
			})
		}

		// Validate patch references
		if err := v.validatePatchReferences(kustomizationFile); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid patch references: %s", err.Error()),
				File:     kustomizationFile,
			})
		}
	}

	return results, nil
}

func (v *KubernetesKustomizationValidator) findKustomizationFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "kustomization.yaml" || info.Name() == "kustomization.yml" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (v *KubernetesKustomizationValidator) validateResourceReferences(kustomizationFile string) error {
	file, err := os.Open(kustomizationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var kustomization map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&kustomization); err != nil {
		return err
	}

	resources, ok := kustomization["resources"].([]interface{})
	if !ok {
		return nil // No resources to validate
	}

	baseDir := filepath.Dir(kustomizationFile)
	seenResources := make(map[string]bool)

	for _, resource := range resources {
		resourcePath, ok := resource.(string)
		if !ok {
			continue
		}

		// Check for duplicate resource references
		if seenResources[resourcePath] {
			return fmt.Errorf("duplicate resource reference: '%s'", resourcePath)
		}
		seenResources[resourcePath] = true

		// Normalize and resolve path
		fullPath, shouldProcess := ResolvePath(baseDir, resourcePath)
		if !shouldProcess {
			// Skip remote resources for now
			continue
		}

		// Check if file/directory exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("resource '%s' does not exist", resourcePath)
		}
	}

	return nil
}

func (v *KubernetesKustomizationValidator) validatePatchReferences(kustomizationFile string) error {
	file, err := os.Open(kustomizationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var kustomization map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&kustomization); err != nil {
		return err
	}

	// Check patches
	if patches, ok := kustomization["patches"].([]interface{}); ok {
		baseDir := filepath.Dir(kustomizationFile)
		seenPatches := make(map[string]bool)

		for _, patch := range patches {
			if patchMap, ok := patch.(map[string]interface{}); ok {
				if path, ok := patchMap["path"].(string); ok {
					// Check for duplicate patch references
					if seenPatches[path] {
						return fmt.Errorf("duplicate patch reference: '%s'", path)
					}
					seenPatches[path] = true

					fullPath, shouldProcess := ResolvePath(baseDir, path)
					if shouldProcess {
						if _, err := os.Stat(fullPath); os.IsNotExist(err) {
							return fmt.Errorf("patch file '%s' does not exist", path)
						}
					}
				}
			}
		}
	}

	// Check patchesStrategicMerge
	if patches, ok := kustomization["patchesStrategicMerge"].([]interface{}); ok {
		baseDir := filepath.Dir(kustomizationFile)
		seenStrategicPatches := make(map[string]bool)

		for _, patch := range patches {
			if patchPath, ok := patch.(string); ok {
				// Check for duplicate strategic merge patch references
				if seenStrategicPatches[patchPath] {
					return fmt.Errorf("duplicate strategic merge patch reference: '%s'", patchPath)
				}
				seenStrategicPatches[patchPath] = true

				fullPath, shouldProcess := ResolvePath(baseDir, patchPath)
				if shouldProcess {
					if _, err := os.Stat(fullPath); os.IsNotExist(err) {
						return fmt.Errorf("strategic merge patch '%s' does not exist", patchPath)
					}
				}
			}
		}
	}

	return nil
}
