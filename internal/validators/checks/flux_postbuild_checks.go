package checks

import (
	"fmt"
	"regexp"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// FluxPostBuildVariableCheck validates Flux postBuild substitute variable naming
func FluxPostBuildVariableCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Extract postBuild substitute variable names
	variables := extractPostBuildVariables(kustomization)

	for _, variable := range variables {
		if !isValidFluxVariableName(variable.Name) {
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

	return results
}

// PostBuildVariable represents a postBuild substitute variable
type PostBuildVariable struct {
	Name string
	Line int
}

// extractPostBuildVariables extracts postBuild substitute variable names from a parsed resource
func extractPostBuildVariables(resource *parser.ParsedResource) []PostBuildVariable {
	var variables []PostBuildVariable

	// Navigate to postBuild.substitute
	if postBuild, exists := resource.Content["postBuild"]; exists {
		if postBuildMap, ok := postBuild.(map[string]interface{}); ok {
			if substitute, exists := postBuildMap["substitute"]; exists {
				if substituteMap, ok := substitute.(map[string]interface{}); ok {
					// Extract variable names from the substitute map
					for key := range substituteMap {
						variables = append(variables, PostBuildVariable{
							Name: key,
							Line: resource.Line, // We don't have exact line numbers for nested values
						})
					}
				}
			}
		}
	}

	return variables
}

// isValidFluxVariableName checks if a variable name follows Flux naming conventions
func isValidFluxVariableName(name string) bool {
	// Flux variable names must start with underscore or letter, followed by letters, digits, or underscores
	// No dashes allowed
	fluxVariableNamePattern := regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
	return fluxVariableNamePattern.MatchString(name)
}
