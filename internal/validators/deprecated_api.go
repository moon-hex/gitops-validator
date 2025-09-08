package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
)

type DeprecatedAPIValidator struct {
	repoPath       string
	yamlPath       string
	deprecatedAPIs map[string]DeprecatedAPIInfo
}

type DeprecatedAPIInfo struct {
	DeprecationInfo  string
	Severity         string
	OperatorCategory string
}

// Default YAML path relative to the binary
const defaultYAMLPath = "data/deprecated-apis.yaml"

func NewDeprecatedAPIValidator(repoPath string) *DeprecatedAPIValidator {
	return &DeprecatedAPIValidator{
		repoPath:       repoPath,
		yamlPath:       defaultYAMLPath,
		deprecatedAPIs: make(map[string]DeprecatedAPIInfo),
	}
}

func NewDeprecatedAPIValidatorWithYAML(repoPath, yamlPath string) *DeprecatedAPIValidator {
	return &DeprecatedAPIValidator{
		repoPath:       repoPath,
		yamlPath:       yamlPath,
		deprecatedAPIs: make(map[string]DeprecatedAPIInfo),
	}
}

func (v *DeprecatedAPIValidator) Name() string {
	return "Deprecated API Validator"
}

func (v *DeprecatedAPIValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Load deprecated APIs from CSV
	if err := v.loadDeprecatedAPIs(); err != nil {
		return results, fmt.Errorf("failed to load deprecated APIs: %w", err)
	}

	// Find all YAML files
	yamlFiles, err := v.findYAMLFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find YAML files: %w", err)
	}

	for _, yamlFile := range yamlFiles {
		fileResults, err := v.validateFile(yamlFile)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		results = append(results, fileResults...)
	}

	return results, nil
}

func (v *DeprecatedAPIValidator) loadDeprecatedAPIs() error {
	// Try to find YAML file relative to the binary first
	yamlPath := v.yamlPath
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		// Try relative to current working directory
		yamlPath = filepath.Join(".", v.yamlPath)
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			return fmt.Errorf("YAML file not found at %s or %s", v.yamlPath, yamlPath)
		}
	}

	file, err := os.Open(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to open YAML file %s: %w", yamlPath, err)
	}
	defer file.Close()

	var config struct {
		DeprecatedAPIs map[string][]struct {
			APIVersion       string `yaml:"api_version"`
			DeprecationInfo  string `yaml:"deprecation_info"`
			Severity         string `yaml:"severity"`
			OperatorCategory string `yaml:"operator_category"`
		} `yaml:"deprecated_apis"`
	}

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("failed to decode YAML file: %w", err)
	}

	// Flatten the hierarchical structure into a flat map
	for _, category := range config.DeprecatedAPIs {
		for _, api := range category {
			v.deprecatedAPIs[api.APIVersion] = DeprecatedAPIInfo{
				DeprecationInfo:  api.DeprecationInfo,
				Severity:         api.Severity,
				OperatorCategory: api.OperatorCategory,
			}
		}
	}

	return nil
}

func (v *DeprecatedAPIValidator) findYAMLFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (v *DeprecatedAPIValidator) validateFile(filePath string) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	file, err := os.Open(filePath)
	if err != nil {
		return results, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var doc yaml.Node

	for {
		err := decoder.Decode(&doc)
		if err != nil {
			break
		}

		if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
			resource := doc.Content[0]
			if resource.Kind == yaml.MappingNode {
				docResults := v.validateResource(resource, filePath)
				results = append(results, docResults...)
			}
		}
	}

	return results, nil
}

func (v *DeprecatedAPIValidator) validateResource(resource *yaml.Node, filePath string) []types.ValidationResult {
	var results []types.ValidationResult

	var apiVersion, kind, name string
	var line int

	for i := 0; i < len(resource.Content); i += 2 {
		key := resource.Content[i]
		value := resource.Content[i+1]

		switch key.Value {
		case "apiVersion":
			apiVersion = value.Value
			line = value.Line
		case "kind":
			kind = value.Value
		case "metadata":
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					if value.Content[j].Value == "name" {
						name = value.Content[j+1].Value
					}
				}
			}
		}
	}

	// Check if API version is deprecated
	if apiInfo, isDeprecated := v.deprecatedAPIs[apiVersion]; isDeprecated {
		message := fmt.Sprintf("Using deprecated API version '%s' for resource '%s'", apiVersion, kind)
		if name != "" {
			message += fmt.Sprintf(" '%s'", name)
		}
		message += fmt.Sprintf(" - %s", apiInfo.DeprecationInfo)

		// Add operator category context
		if apiInfo.OperatorCategory != "" {
			message += fmt.Sprintf(" (%s)", apiInfo.OperatorCategory)
		}

		results = append(results, types.ValidationResult{
			Type:     "deprecated-api",
			Severity: apiInfo.Severity,
			Message:  message,
			File:     filePath,
			Line:     line,
			Resource: fmt.Sprintf("%s/%s", apiVersion, kind),
		})
	}

	return results
}
