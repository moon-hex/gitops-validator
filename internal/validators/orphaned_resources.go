package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/checks"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

type OrphanedResourceValidator struct {
	*common.BaseValidator
}

func NewOrphanedResourceValidator(repoPath string) *OrphanedResourceValidator {
	return &OrphanedResourceValidator{
		BaseValidator: common.NewBaseValidator("Orphaned Resource Validator", repoPath),
	}
}

// Validate implements the GraphValidator interface
func (v *OrphanedResourceValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	// Use the focused orphaned resource check
	results := checks.OrphanedResourceCheck(ctx)
	return results, nil
}
