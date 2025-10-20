package validators

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/types"
)

// ValidationRule represents a single validation rule
type ValidationRule interface {
	Name() string
	Validate(kustomization *KustomizationFile) []types.ValidationResult
}

// ResourceReferenceRule validates that referenced resource files exist
type ResourceReferenceRule struct{}

func (r *ResourceReferenceRule) Name() string {
	return "Resource Reference Rule"
}

func (r *ResourceReferenceRule) Validate(kustomization *KustomizationFile) []types.ValidationResult {
	var results []types.ValidationResult
	seenResources := make(map[string]bool)

	for _, resourcePath := range kustomization.GetResources() {
		// Check for duplicate resource references
		if seenResources[resourcePath] {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("duplicate resource reference: '%s'", resourcePath),
				File:     kustomization.Path,
			})
			continue
		}
		seenResources[resourcePath] = true

		// Check if file/directory exists
		if err := kustomization.ValidateFileExists(resourcePath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid resource references: %s", err.Error()),
				File:     kustomization.Path,
			})
		}
	}

	return results
}

// PatchReferenceRule validates that referenced patch files exist
type PatchReferenceRule struct{}

func (r *PatchReferenceRule) Name() string {
	return "Patch Reference Rule"
}

func (r *PatchReferenceRule) Validate(kustomization *KustomizationFile) []types.ValidationResult {
	var results []types.ValidationResult
	seenPatches := make(map[string]bool)

	for _, patchPath := range kustomization.GetPatches() {
		// Check for duplicate patch references
		if seenPatches[patchPath] {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("duplicate patch reference: '%s'", patchPath),
				File:     kustomization.Path,
			})
			continue
		}
		seenPatches[patchPath] = true

		// Check if file exists
		if err := kustomization.ValidateFileExists(patchPath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid patch references: %s", err.Error()),
				File:     kustomization.Path,
			})
		}
	}

	return results
}

// StrategicMergePatchReferenceRule validates that referenced strategic merge patch files exist
type StrategicMergePatchReferenceRule struct{}

func (r *StrategicMergePatchReferenceRule) Name() string {
	return "Strategic Merge Patch Reference Rule"
}

func (r *StrategicMergePatchReferenceRule) Validate(kustomization *KustomizationFile) []types.ValidationResult {
	var results []types.ValidationResult

	for _, patchPath := range kustomization.GetStrategicMergePatches() {
		// Check if file exists
		if err := kustomization.ValidateFileExists(patchPath); err != nil {
			results = append(results, types.ValidationResult{
				Type:     "kubernetes-kustomization",
				Severity: "error",
				Message:  fmt.Sprintf("Invalid patch references: %s", err.Error()),
				File:     kustomization.Path,
			})
		}
	}

	return results
}

// ValidationRuleSet manages a collection of validation rules
type ValidationRuleSet struct {
	rules []ValidationRule
}

// NewValidationRuleSet creates a new ValidationRuleSet
func NewValidationRuleSet() *ValidationRuleSet {
	return &ValidationRuleSet{
		rules: make([]ValidationRule, 0),
	}
}

// AddRule adds a validation rule to the set
func (s *ValidationRuleSet) AddRule(rule ValidationRule) {
	s.rules = append(s.rules, rule)
}

// Validate runs all rules against a kustomization file
func (s *ValidationRuleSet) Validate(kustomization *KustomizationFile) []types.ValidationResult {
	var results []types.ValidationResult

	for _, rule := range s.rules {
		ruleResults := rule.Validate(kustomization)
		results = append(results, ruleResults...)
	}

	return results
}
