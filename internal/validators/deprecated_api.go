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
	repoPath string
}

// DeprecatedAPIs contains a list of deprecated Kubernetes API versions
var DeprecatedAPIs = map[string]string{
	"extensions/v1beta1":                    "Deprecated in v1.16, removed in v1.22",
	"apps/v1beta1":                          "Deprecated in v1.9, removed in v1.16",
	"apps/v1beta2":                          "Deprecated in v1.9, removed in v1.16",
	"policy/v1beta1":                        "Deprecated in v1.21, removed in v1.25",
	"rbac.authorization.k8s.io/v1beta1":     "Deprecated in v1.17, removed in v1.22",
	"rbac.authorization.k8s.io/v1alpha1":    "Deprecated in v1.17, removed in v1.22",
	"storage.k8s.io/v1beta1":                "Deprecated in v1.19, removed in v1.22",
	"admissionregistration.k8s.io/v1beta1":  "Deprecated in v1.19, removed in v1.22",
	"networking.k8s.io/v1beta1":             "Deprecated in v1.19, removed in v1.22",
	"scheduling.k8s.io/v1beta1":             "Deprecated in v1.19, removed in v1.22",
	"coordination.k8s.io/v1beta1":           "Deprecated in v1.19, removed in v1.22",
	"node.k8s.io/v1beta1":                   "Deprecated in v1.19, removed in v1.22",
	"discovery.k8s.io/v1beta1":              "Deprecated in v1.19, removed in v1.22",
	"flowcontrol.apiserver.k8s.io/v1beta1":  "Deprecated in v1.20, removed in v1.25",
	"flowcontrol.apiserver.k8s.io/v1alpha1": "Deprecated in v1.20, removed in v1.25",
}

func NewDeprecatedAPIValidator(repoPath string) *DeprecatedAPIValidator {
	return &DeprecatedAPIValidator{
		repoPath: repoPath,
	}
}

func (v *DeprecatedAPIValidator) Name() string {
	return "Deprecated API Validator"
}

func (v *DeprecatedAPIValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

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
	if deprecationInfo, isDeprecated := DeprecatedAPIs[apiVersion]; isDeprecated {
		severity := "warning"
		if strings.Contains(deprecationInfo, "removed") {
			severity = "error"
		}

		message := fmt.Sprintf("Using deprecated API version '%s' for resource '%s'", apiVersion, kind)
		if name != "" {
			message += fmt.Sprintf(" '%s'", name)
		}
		message += fmt.Sprintf(" - %s", deprecationInfo)

		results = append(results, types.ValidationResult{
			Type:     "deprecated-api",
			Severity: severity,
			Message:  message,
			File:     filePath,
			Line:     line,
			Resource: fmt.Sprintf("%s/%s", apiVersion, kind),
		})
	}

	return results
}
