package common

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// BaseValidator provides common functionality for all validators
type BaseValidator struct {
	name     string
	repoPath string
}

// NewBaseValidator creates a new base validator
func NewBaseValidator(name, repoPath string) *BaseValidator {
	return &BaseValidator{
		name:     name,
		repoPath: repoPath,
	}
}

// Name returns the validator name
func (b *BaseValidator) Name() string {
	return b.name
}

// RepoPath returns the repository path
func (b *BaseValidator) RepoPath() string {
	return b.repoPath
}

// CreateResult creates a validation result with consistent formatting
func (b *BaseValidator) CreateResult(resultType, severity, message, file, resource string, line int) types.ValidationResult {
	return types.ValidationResult{
		Type:     resultType,
		Severity: severity,
		Message:  message,
		File:     file,
		Resource: resource,
		Line:     line,
	}
}

// CreateErrorResult creates an error result
func (b *BaseValidator) CreateErrorResult(resultType, message, file, resource string) types.ValidationResult {
	return b.CreateResult(resultType, "error", message, file, resource, 0)
}

// CreateWarningResult creates a warning result
func (b *BaseValidator) CreateWarningResult(resultType, message, file, resource string) types.ValidationResult {
	return b.CreateResult(resultType, "warning", message, file, resource, 0)
}

// CreateInfoResult creates an info result
func (b *BaseValidator) CreateInfoResult(resultType, message, file, resource string) types.ValidationResult {
	return b.CreateResult(resultType, "info", message, file, resource, 0)
}

// ValidateContext provides a common interface for validation context
type ValidateContext interface {
	*context.ValidationContext
}

// ValidatorFunc represents a validation function
type ValidatorFunc func(ctx *context.ValidationContext) ([]types.ValidationResult, error)

// ValidationCheck represents a single validation check
type ValidationCheck struct {
	Name        string
	Description string
	CheckFunc   ValidatorFunc
	Severity    string
}

// NewValidationCheck creates a new validation check
func NewValidationCheck(name, description string, checkFunc ValidatorFunc, severity string) *ValidationCheck {
	return &ValidationCheck{
		Name:        name,
		Description: description,
		CheckFunc:   checkFunc,
		Severity:    severity,
	}
}
