package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/types"

	"gopkg.in/yaml.v3"
)

type FluxPostBuildVariablesValidator struct {
	repoPath string
}

func NewFluxPostBuildVariablesValidator(repoPath string) *FluxPostBuildVariablesValidator {
	return &FluxPostBuildVariablesValidator{
		repoPath: repoPath,
	}
}

func (v *FluxPostBuildVariablesValidator) Name() string {
	return "Flux PostBuild Variables Validator"
}

// Flux variable naming pattern: must start with _ or letter, followed by letters, digits, or underscores
// Pattern: ^[_[:alpha:]][_[:alpha:][:digit:]]*$
var fluxVariableNamePattern = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

func (v *FluxPostBuildVariablesValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all Flux Kustomization resources
	kustomizations, err := v.findFluxKustomizations()
	if err != nil {
		return results, fmt.Errorf("failed to find Flux Kustomizations: %w", err)
	}

	for _, kustomization := range kustomizations {
		// Validate postBuild substitute variable names
		for _, variable := range kustomization.PostBuildVariables {
			if !fluxVariableNamePattern.MatchString(variable.Name) {
				results = append(results, types.ValidationResult{
					Type:     "flux-postbuild-variables",
					Severity: "error",
					Message: fmt.Sprintf("Invalid Flux variable name '%s': must start with underscore or letter, followed by letters, digits, or underscores only (no dashes allowed). Pattern: ^[_a-zA-Z][_a-zA-Z0-9]*$",
						variable.Name),
					File:     kustomization.File,
					Line:     variable.Line,
					Resource: kustomization.Name,
				})
			}
		}
	}

	return results, nil
}

type FluxKustomizationWithPostBuild struct {
	File               string
	Name               string
	PostBuildVariables []VariableInfo
}

type VariableInfo struct {
	Name string
	Line int
}

func (v *FluxPostBuildVariablesValidator) findFluxKustomizations() ([]FluxKustomizationWithPostBuild, error) {
	var kustomizations []FluxKustomizationWithPostBuild

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
					kustomization := v.extractPostBuildVariables(resource, path)
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

func (v *FluxPostBuildVariablesValidator) isFluxKustomization(node *yaml.Node) bool {
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

func (v *FluxPostBuildVariablesValidator) extractPostBuildVariables(node *yaml.Node, filePath string) *FluxKustomizationWithPostBuild {
	var name string
	var variables []VariableInfo

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
					if value.Content[j].Value == "postBuild" {
						variables = v.extractVariablesFromPostBuild(value.Content[j+1])
					}
				}
			}
		}
	}

	if name == "" {
		return nil
	}

	// Only return if there are variables to validate
	if len(variables) == 0 {
		return nil
	}

	return &FluxKustomizationWithPostBuild{
		File:               filePath,
		Name:               name,
		PostBuildVariables: variables,
	}
}

func (v *FluxPostBuildVariablesValidator) extractVariablesFromPostBuild(postBuildNode *yaml.Node) []VariableInfo {
	var variables []VariableInfo

	if postBuildNode.Kind != yaml.MappingNode {
		return variables
	}

	for i := 0; i < len(postBuildNode.Content); i += 2 {
		key := postBuildNode.Content[i]
		value := postBuildNode.Content[i+1]

		if key.Value == "substitute" {
			// substitute is a map of variable names to values
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					varName := value.Content[j].Value
					varLine := value.Content[j].Line
					variables = append(variables, VariableInfo{
						Name: varName,
						Line: varLine,
					})
				}
			}
		}
	}

	return variables
}
