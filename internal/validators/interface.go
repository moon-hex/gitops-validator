package validators

import "gitops-validator/internal/types"

// ValidatorInterface defines the contract for all validators
type ValidatorInterface interface {
	Name() string
	Validate() ([]types.ValidationResult, error)
}
