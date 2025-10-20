package validators

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// KustomizationFile represents a parsed kustomization file
type KustomizationFile struct {
	Path    string
	Content map[string]interface{}
	BaseDir string
}

// KustomizationParser handles parsing of kustomization files
type KustomizationParser struct {
	repoPath string
}

// NewKustomizationParser creates a new KustomizationParser
func NewKustomizationParser(repoPath string) *KustomizationParser {
	return &KustomizationParser{
		repoPath: repoPath,
	}
}

// FindKustomizationFiles finds all kustomization.yaml files in the repository
func (p *KustomizationParser) FindKustomizationFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(p.repoPath, func(path string, info os.FileInfo, err error) error {
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

// ParseKustomizationFile parses a kustomization file and returns its content
func (p *KustomizationParser) ParseKustomizationFile(filePath string) (*KustomizationFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open kustomization file %s: %w", filePath, err)
	}
	defer file.Close()

	var kustomization map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&kustomization); err != nil {
		return nil, fmt.Errorf("failed to parse kustomization file %s: %w", filePath, err)
	}

	return &KustomizationFile{
		Path:    filePath,
		Content: kustomization,
		BaseDir: filepath.Dir(filePath),
	}, nil
}

// GetResources returns the resources list from a kustomization file
func (k *KustomizationFile) GetResources() []string {
	var resources []string

	if resourcesList, ok := k.Content["resources"].([]interface{}); ok {
		for _, resource := range resourcesList {
			if resourcePath, ok := resource.(string); ok {
				resources = append(resources, resourcePath)
			}
		}
	}

	return resources
}

// GetPatches returns the patches list from a kustomization file
func (k *KustomizationFile) GetPatches() []string {
	var patches []string

	if patchesList, ok := k.Content["patches"].([]interface{}); ok {
		for _, patch := range patchesList {
			if patchMap, ok := patch.(map[string]interface{}); ok {
				if path, ok := patchMap["path"].(string); ok {
					patches = append(patches, path)
				}
			}
		}
	}

	return patches
}

// GetStrategicMergePatches returns the patchesStrategicMerge list from a kustomization file
func (k *KustomizationFile) GetStrategicMergePatches() []string {
	var patches []string

	if patchesList, ok := k.Content["patchesStrategicMerge"].([]interface{}); ok {
		for _, patch := range patchesList {
			if patchPath, ok := patch.(string); ok {
				patches = append(patches, patchPath)
			}
		}
	}

	return patches
}

// ValidateFileExists checks if a file exists relative to the kustomization base directory
func (k *KustomizationFile) ValidateFileExists(filePath string) error {
	fullPath, shouldProcess := ResolvePath(k.BaseDir, filePath)
	if shouldProcess {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist", filePath)
		}
	}
	return nil
}
