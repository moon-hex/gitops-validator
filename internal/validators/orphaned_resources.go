package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
)

type OrphanedResourceValidator struct {
	repoPath string
	config   *config.Config
}

func NewOrphanedResourceValidator(repoPath string) *OrphanedResourceValidator {
	return &OrphanedResourceValidator{
		repoPath: repoPath,
		config:   config.DefaultConfig(),
	}
}

func NewOrphanedResourceValidatorWithConfig(repoPath string, cfg *config.Config) *OrphanedResourceValidator {
	return &OrphanedResourceValidator{
		repoPath: repoPath,
		config:   cfg,
	}
}

func (v *OrphanedResourceValidator) Name() string {
	return "Orphaned Resource Validator"
}

func (v *OrphanedResourceValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all YAML files
	yamlFiles, err := v.findYAMLFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find YAML files: %w", err)
	}

	// Find all kustomization files
	kustomizationFiles, err := v.findKustomizationFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find kustomization files: %w", err)
	}

	// Build a map of referenced files
	referencedFiles := make(map[string]bool)

	// Add kustomization files themselves as referenced
	for _, kustomizationFile := range kustomizationFiles {
		referencedFiles[kustomizationFile] = true
	}

	// Add files referenced by kustomizations
	for _, kustomizationFile := range kustomizationFiles {
		referenced, err := v.getReferencedFiles(kustomizationFile)
		if err != nil {
			continue // Skip invalid kustomization files
		}

		for _, file := range referenced {
			referencedFiles[file] = true
		}
	}

	// Check for orphaned files
	for _, yamlFile := range yamlFiles {
		if !referencedFiles[yamlFile] && !v.isEntryPoint(yamlFile) {
			results = append(results, types.ValidationResult{
				Type:     "orphaned-resource",
				Severity: "warning",
				Message:  fmt.Sprintf("File '%s' is not referenced by any kustomization and is not an entry point", filepath.Base(yamlFile)),
				File:     yamlFile,
			})
		}
	}

	return results, nil
}

func (v *OrphanedResourceValidator) findYAMLFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if path should be ignored
		relPath, err := filepath.Rel(v.repoPath, path)
		if err != nil {
			return err
		}

		if v.config.ShouldIgnorePath(relPath) {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (v *OrphanedResourceValidator) findKustomizationFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if path should be ignored
		relPath, err := filepath.Rel(v.repoPath, path)
		if err != nil {
			return err
		}

		if v.config.ShouldIgnorePath(relPath) {
			return nil
		}

		if info.Name() == "kustomization.yaml" || info.Name() == "kustomization.yml" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (v *OrphanedResourceValidator) getReferencedFiles(kustomizationFile string) ([]string, error) {
	var referencedFiles []string

	file, err := os.Open(kustomizationFile)
	if err != nil {
		return referencedFiles, err
	}
	defer file.Close()

	var kustomization map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&kustomization); err != nil {
		return referencedFiles, err
	}

	baseDir := filepath.Dir(kustomizationFile)

	// Check resources
	if resources, ok := kustomization["resources"].([]interface{}); ok {
		for _, resource := range resources {
			if resourcePath, ok := resource.(string); ok {
				if !strings.HasPrefix(resourcePath, "http://") && !strings.HasPrefix(resourcePath, "https://") {
					var fullPath string
					if strings.HasPrefix(resourcePath, "/") {
						fullPath = resourcePath
					} else {
						// Strip ./ prefix if present
						cleanPath := strings.TrimPrefix(resourcePath, "./")
						fullPath = filepath.Join(baseDir, cleanPath)
					}
					referencedFiles = append(referencedFiles, fullPath)
				}
			}
		}
	}

	// Check patches
	if patches, ok := kustomization["patches"].([]interface{}); ok {
		for _, patch := range patches {
			if patchMap, ok := patch.(map[string]interface{}); ok {
				if path, ok := patchMap["path"].(string); ok {
					// Strip ./ prefix if present
					cleanPath := strings.TrimPrefix(path, "./")
					fullPath := filepath.Join(baseDir, cleanPath)
					referencedFiles = append(referencedFiles, fullPath)
				}
			}
		}
	}

	// Check patchesStrategicMerge
	if patches, ok := kustomization["patchesStrategicMerge"].([]interface{}); ok {
		for _, patch := range patches {
			if patchPath, ok := patch.(string); ok {
				// Strip ./ prefix if present
				cleanPath := strings.TrimPrefix(patchPath, "./")
				fullPath := filepath.Join(baseDir, cleanPath)
				referencedFiles = append(referencedFiles, fullPath)
			}
		}
	}

	return referencedFiles, nil
}

func (v *OrphanedResourceValidator) isEntryPoint(filePath string) bool {
	fileName := filepath.Base(filePath)

	// Check if it's a kustomization file
	if fileName == "kustomization.yaml" || fileName == "kustomization.yml" {
		return true
	}

	// Check if it's a Flux Kustomization resource
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	var doc yaml.Node
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&doc); err != nil {
		return false
	}

	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		resource := doc.Content[0]
		if resource.Kind == yaml.MappingNode {
			for i := 0; i < len(resource.Content); i += 2 {
				key := resource.Content[i]
				value := resource.Content[i+1]

				if key.Value == "kind" && value.Value == "Kustomization" {
					// Check if it's a Flux Kustomization
					for j := 0; j < len(resource.Content); j += 2 {
						if resource.Content[j].Value == "apiVersion" {
							apiVersion := resource.Content[j+1].Value
							if strings.HasPrefix(apiVersion, "kustomize.toolkit.fluxcd.io/") {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}
