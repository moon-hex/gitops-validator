package validators

import (
	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators/checks"
)

type DeprecatedAPIValidator struct {
	repoPath string
}

func NewDeprecatedAPIValidator(repoPath string) *DeprecatedAPIValidator {
	return &DeprecatedAPIValidator{
		repoPath: repoPath,
	}
}

func (v *DeprecatedAPIValidator) Name() string {
	return "Deprecated API Validator"
}

// Validate implements the GraphValidator interface
func (v *DeprecatedAPIValidator) Validate(ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var results []types.ValidationResult

	// Get all resources from the graph
	allResources := ctx.Graph.Resources

	for _, resource := range allResources {
		// Use the focused deprecated API check
		checkResults := checks.DeprecatedAPICheck(resource, ctx.Config)
		results = append(results, checkResults...)
	}

	return results, nil
}

// DeprecatedAPIInfo represents information about a deprecated API
type DeprecatedAPIInfo struct {
	DeprecationInfo  string
	Severity         string
	OperatorCategory string
}

// checkDeprecatedAPI checks if an API version is deprecated
func (v *DeprecatedAPIValidator) checkDeprecatedAPI(apiVersion string, config *config.Config) *DeprecatedAPIInfo {
	// Check against config's deprecated APIs
	for _, api := range config.GitOpsValidator.DeprecatedAPIs.CustomAPIs {
		if api.APIVersion == apiVersion {
			return &DeprecatedAPIInfo{
				DeprecationInfo:  api.DeprecationInfo,
				Severity:         api.Severity,
				OperatorCategory: api.OperatorCategory,
			}
		}
	}

	return nil
}
