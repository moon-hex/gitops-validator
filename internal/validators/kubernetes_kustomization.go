package validators

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/types"
)

// KubernetesKustomizationValidator is now a composite validator that uses specialized validators
type KubernetesKustomizationValidator struct {
	resourceValidator       *KustomizationResourceValidator
	patchValidator          *KustomizationPatchValidator
	strategicMergeValidator *KustomizationStrategicMergeValidator
}

func NewKubernetesKustomizationValidator(repoPath string) *KubernetesKustomizationValidator {
	return &KubernetesKustomizationValidator{
		resourceValidator:       NewKustomizationResourceValidator(repoPath),
		patchValidator:          NewKustomizationPatchValidator(repoPath),
		strategicMergeValidator: NewKustomizationStrategicMergeValidator(repoPath),
	}
}

func (v *KubernetesKustomizationValidator) Name() string {
	return "Kubernetes Kustomization Validator"
}

func (v *KubernetesKustomizationValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Run all specialized validators
	validators := []struct {
		name     string
		validate func() ([]types.ValidationResult, error)
	}{
		{v.resourceValidator.Name(), v.resourceValidator.Validate},
		{v.patchValidator.Name(), v.patchValidator.Validate},
		{v.strategicMergeValidator.Name(), v.strategicMergeValidator.Validate},
	}

	for _, validator := range validators {
		validatorResults, err := validator.validate()
		if err != nil {
			// Add error as validation result instead of failing completely
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Validator %s failed: %s", validator.name, err.Error()),
			})
			continue
		}
		results = append(results, validatorResults...)
	}

	return results, nil
}
