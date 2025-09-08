package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// GraphValidator defines the contract for graph-based validators
type GraphValidator interface {
	Name() string
	Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error)
}

// Legacy ValidatorInterface for backward compatibility
type ValidatorInterface interface {
	Name() string
	Validate() ([]types.ValidationResult, error)
}
