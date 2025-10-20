package validators

import (
	"fmt"

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

func (v *KustomizationStrategicMergeValidator) Validate() ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Find all kustomization files
	kustomizationFiles, err := v.parser.FindKustomizationFiles()
	if err != nil {
		return results, fmt.Errorf("failed to find kustomization files: %w", err)
	}

	// Create validation rule set
	ruleSet := NewValidationRuleSet()
	ruleSet.AddRule(&StrategicMergePatchReferenceRule{})

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
