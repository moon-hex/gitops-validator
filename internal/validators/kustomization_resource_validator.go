package validators

import (
	"fmt"

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

func (v *KustomizationResourceValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all kustomization files
	kustomizationFiles, err := v.parser.FindKustomizationFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find kustomization files: %w", err)
	}

	// Create validation rule set
	ruleSet := NewValidationRuleSet()
	ruleSet.AddRule(&ResourceReferenceRule{})

	// Validate each kustomization file
	for _, kustomizationFile := range kustomizationFiles {
		kustomization, err := v.parser.ParseKustomizationFile(kustomizationFile)
		if err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Failed to parse kustomization file: %s", err.Error()),
				File:     kustomizationFile,
			})
			continue
		}

		// Run validation rules
		ruleResults := ruleSet.Validate(kustomization)
		results = append(results, ruleResults...)
	}

	return results, nil
}
