package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// KustomizationResourceValidator validates resource references in kustomization files
type KustomizationResourceValidator struct {
	parser *KustomizationParser
}

// NewKustomizationResourceValidator creates a new KustomizationResourceValidator
func NewKustomizationResourceValidator(repoPath string) *KustomizationResourceValidator {
	return &KustomizationResourceValidator{
		parser: NewKustomizationParser(repoPath),
	}
}

func (v *KustomizationResourceValidator) Name() string {
	return "Kustomization Resource Validator"
}

// Validate implements the GraphValidator interface
func (v *KustomizationResourceValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Kubernetes Kustomization resources from the graph
	kustomizations := ctx.Graph.GetKubernetesKustomizations()

	// Create validation rule set
	ruleSet := NewValidationRuleSet()
	ruleSet.AddRule(&ResourceReferenceRule{})

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
