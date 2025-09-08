package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
)

type FluxKustomizationValidator struct {
	repoPath string
}

func NewFluxKustomizationValidator(repoPath string) *FluxKustomizationValidator {
	return &FluxKustomizationValidator{
		repoPath: repoPath,
	}
}

func (v *FluxKustomizationValidator) Name() string {
	return "Flux Kustomization Validator"
}

func (v *FluxKustomizationValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all Flux Kustomization resources
	kustomizations, err := v.findFluxKustomizations()
	if err != nil {
		return results, fmt.Errorf("failed to find Flux Kustomizations: %w", err)
	}

	for _, kustomization := range kustomizations {
		// Validate path references
		if err := v.validatePathReference(kustomization); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "flux-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid path reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}

		// Validate source references
		if err := v.validateSourceReference(kustomization); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "flux-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid source reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}
	}

	return results, nil
}

type FluxKustomization struct {
	File   string
	Name   string
	Path   string
	Source string
	Line   int
}

func (v *FluxKustomizationValidator) findFluxKustomizations() ([]FluxKustomization, error) {
	var kustomizations []FluxKustomization

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".yaml") && !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
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
				if v.isFluxKustomization(resource) {
					kustomization := v.extractKustomizationInfo(resource, path)
					if kustomization != nil {
						kustomizations = append(kustomizations, *kustomization)
					}
				}
			}
		}

		return nil
	})

	return kustomizations, err
}

func (v *FluxKustomizationValidator) isFluxKustomization(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}

	var kind, apiVersion string

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		if key.Value == "kind" && value.Value == "Kustomization" {
			kind = value.Value
		}

		if key.Value == "apiVersion" && strings.HasPrefix(value.Value, "kustomize.toolkit.fluxcd.io/") {
			apiVersion = value.Value
		}
	}

	return kind == "Kustomization" && apiVersion != ""
}

func (v *FluxKustomizationValidator) extractKustomizationInfo(node *yaml.Node, filePath string) *FluxKustomization {
	var name, path, source string

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		switch key.Value {
		case "metadata":
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					if value.Content[j].Value == "name" {
						name = value.Content[j+1].Value
					}
				}
			}
		case "spec":
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					switch value.Content[j].Value {
					case "path":
						path = value.Content[j+1].Value
					case "sourceRef":
						if value.Content[j+1].Kind == yaml.MappingNode {
							for k := 0; k < len(value.Content[j+1].Content); k += 2 {
								if value.Content[j+1].Content[k].Value == "name" {
									source = value.Content[j+1].Content[k+1].Value
								}
							}
						}
					}
				}
			}
		}
	}

	if name == "" {
		return nil
	}

	return &FluxKustomization{
		File:   filePath,
		Name:   name,
		Path:   path,
		Source: source,
		Line:   node.Line,
	}
}

func (v *FluxKustomizationValidator) validatePathReference(kustomization FluxKustomization) error {
	if kustomization.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Check if path exists relative to repository root
	path := filepath.Join(v.repoPath, kustomization.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path '%s' does not exist", kustomization.Path)
	}

	return nil
}

func (v *FluxKustomizationValidator) validateSourceReference(kustomization FluxKustomization) error {
	if kustomization.Source == "" {
		return fmt.Errorf("sourceRef.name is required")
	}

	// This would need to check against Flux Source resources
	// For now, we'll just validate that it's not empty
	return nil
}
