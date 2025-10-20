package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// KustomizationPatchValidator validates patch references in kustomization files
type KustomizationPatchValidator struct {
	parser *KustomizationParser
}

// NewKustomizationPatchValidator creates a new KustomizationPatchValidator
func NewKustomizationPatchValidator(repoPath string) *KustomizationPatchValidator {
	return &KustomizationPatchValidator{
		parser: NewKustomizationParser(repoPath),
	}
}

func (v *KustomizationPatchValidator) Name() string {
	return "Kustomization Patch Validator"
}

// Validate implements the GraphValidator interface
func (v *KustomizationPatchValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Kubernetes Kustomization resources from the graph
	kustomizations := ctx.Graph.GetKubernetesKustomizations()

	// Create validation rule set
	ruleSet := NewValidationRuleSet()
	ruleSet.AddRule(&PatchReferenceRule{})

	// Validate each kustomization
	for _, kustomization := range kustomizations {
		// Convert ParsedResource to KustomizationFile format for compatibility
		kustomizationFile := &KustomizationFile{
			Path:    kustomization.File,
			Content: kustomization.Content,
		}

		// Run validation rules
		ruleResults := ruleSet.Validate(kustomizationFile)
		results = append(results, ruleResults...)
	}

	return results, nil
}
