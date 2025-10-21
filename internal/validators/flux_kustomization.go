package validators

import (
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/checks"
	"github.com/moon-hex/gitops-validator/internal/validators/common"
)

type FluxKustomizationValidator struct {
	*common.BaseValidator
}

func NewFluxKustomizationValidator(repoPath string) *FluxKustomizationValidator {
	return &FluxKustomizationValidator{
		BaseValidator: common.NewBaseValidator("Flux Kustomization Validator", repoPath),
	}
}

// Validate implements the GraphValidator interface
func (v *FluxKustomizationValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all Flux Kustomization resources from the graph
	fluxKustomizations := ctx.Graph.GetFluxKustomizations()

	for _, kustomization := range fluxKustomizations {
		// Run path validation checks
		pathResults := checks.FluxKustomizationPathCheck(kustomization, ctx)
		results = append(results, pathResults...)

		// Run source validation checks
		sourceResults := checks.FluxKustomizationSourceCheck(kustomization, ctx)
		results = append(results, sourceResults...)
	}

	return results, nil
}
