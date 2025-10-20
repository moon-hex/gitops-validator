package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// KustomizationStrategicMergeValidator validates strategic merge patch references in kustomization files
type KustomizationStrategicMergeValidator struct {
	parser *KustomizationParser
}

// NewKustomizationStrategicMergeValidator creates a new KustomizationStrategicMergeValidator
func NewKustomizationStrategicMergeValidator(repoPath string) *KustomizationStrategicMergeValidator {
	return &KustomizationStrategicMergeValidator{
		parser: NewKustomizationParser(repoPath),
	}
}

func (v *KustomizationStrategicMergeValidator) Name() string {
	return "Kustomization Strategic Merge Validator"
}

// Validate implements the GraphValidator interface
func (v *KustomizationStrategicMergeValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Kubernetes Kustomization resources from the graph
	kustomizations := ctx.Graph.GetKubernetesKustomizations()

	// Create validation rule set
	ruleSet := NewValidationRuleSet()
	ruleSet.AddRule(&StrategicMergePatchReferenceRule{})

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
