package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/checks"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

// HTTPRoutePolicyValidator checks that every HTTPRoute and VirtualService
// in the repository has a SecurityPolicy defined in the same namespace.
type HTTPRoutePolicyValidator struct {
	*common.BaseValidator
}

func NewHTTPRoutePolicyValidator(repoPath string) *HTTPRoutePolicyValidator {
	return &HTTPRoutePolicyValidator{
		BaseValidator: common.NewBaseValidator("HTTP Route Policy Validator", repoPath),
	}
}

// Validate implements the GraphValidator interface
func (v *HTTPRoutePolicyValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	results := checks.HTTPRoutePolicyCheck(ctx)
	return results, nil
}
