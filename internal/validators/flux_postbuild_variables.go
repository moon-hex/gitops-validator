package validators

import (
	"fmt"
	"regexp"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
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

// Validate implements the GraphValidator interface
func (v *FluxPostBuildVariablesValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Flux Kustomization resources from the graph
	fluxKustomizations := ctx.Graph.GetFluxKustomizations()

	for _, kustomization := range fluxKustomizations {
		// Extract postBuild substitute variable names from the parsed content
		variables := v.extractPostBuildVariables(kustomization)

		for _, variable := range variables {
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

type VariableInfo struct {
	Name string
	Line int
}

// extractPostBuildVariables extracts postBuild substitute variable names from a parsed Flux Kustomization
func (v *FluxPostBuildVariablesValidator) extractPostBuildVariables(kustomization *parser.ParsedResource) []VariableInfo {
	var variables []VariableInfo

	// Extract spec from the parsed content
	spec, exists := kustomization.Content["spec"]
	if !exists {
		return variables
	}

	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return variables
	}

	// Extract postBuild from spec
	postBuild, exists := specMap["postBuild"]
	if !exists {
		return variables
	}

	postBuildMap, ok := postBuild.(map[string]interface{})
	if !ok {
		return variables
	}

	// Extract substitute from postBuild
	substitute, exists := postBuildMap["substitute"]
	if !exists {
		return variables
	}

	substituteMap, ok := substitute.(map[string]interface{})
	if !ok {
		return variables
	}

	// Extract variable names from substitute map
	for varName := range substituteMap {
		variables = append(variables, VariableInfo{
			Name: varName,
			Line: 0, // Line number not available from parsed content
		})
	}

	return variables
}
