package validators

import (
	"fmt"
	"os"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
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

// Validate implements the GraphValidator interface
func (v *FluxKustomizationValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Flux Kustomization resources from the graph
	fluxKustomizations := ctx.Graph.GetFluxKustomizations()

	for _, kustomization := range fluxKustomizations {
		// Validate path references
		if err := v.validatePathReference(kustomization, ctx); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "flux-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid path reference: %s", err.Error()),
				File:     kustomization.File,
				Resource: kustomization.Name,
			})
		}

		// Validate source references
		if err := v.validateSourceReference(kustomization, ctx); err != nil {
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

func (v *FluxKustomizationValidator) validatePathReference(kustomization *parser.ParsedResource, ctx *context.ValidationContext) error {
	// Extract path from the parsed resource
	path, exists := kustomization.Content["spec"].(map[string]interface{})["path"]
	if !exists || path == nil {
		return fmt.Errorf("path is required")
	}

	pathStr, ok := path.(string)
	if !ok {
		return fmt.Errorf("path must be a string")
	}

	// Normalize and resolve path
	fullPath, shouldProcess := ResolvePath(ctx.RepoPath, pathStr)
	if !shouldProcess {
		return fmt.Errorf("unsupported path format: %s", pathStr)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("path '%s' does not exist", pathStr)
	}

	return nil
}

func (v *FluxKustomizationValidator) validateSourceReference(kustomization *parser.ParsedResource, ctx *context.ValidationContext) error {
	// Extract sourceRef from the parsed resource
	spec, exists := kustomization.Content["spec"]
	if !exists {
		return fmt.Errorf("spec is required")
	}

	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("spec must be an object")
	}

	sourceRef, exists := specMap["sourceRef"]
	if !exists || sourceRef == nil {
		return fmt.Errorf("sourceRef is required")
	}

	sourceRefMap, ok := sourceRef.(map[string]interface{})
	if !ok {
		return fmt.Errorf("sourceRef must be an object")
	}

	sourceName, exists := sourceRefMap["name"]
	if !exists || sourceName == nil {
		return fmt.Errorf("sourceRef.name is required")
	}

	sourceNameStr, ok := sourceName.(string)
	if !ok {
		return fmt.Errorf("sourceRef.name must be a string")
	}

	if sourceNameStr == "" {
		return fmt.Errorf("sourceRef.name cannot be empty")
	}

	// Check if the source exists in the graph
	sources := ctx.Graph.GetFluxSources()
	found := false
	for _, source := range sources {
		if source.Name == sourceNameStr {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("source '%s' not found in repository", sourceNameStr)
	}

	return nil
}
