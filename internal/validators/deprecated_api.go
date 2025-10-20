package validators

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
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
		// Check if the API version is deprecated
		deprecatedInfo := v.checkDeprecatedAPI(resource.APIVersion, ctx.Config)
		if deprecatedInfo != nil {
			results = append(results, types.ValidationResult{
				Type:     "deprecated-api",
				Severity: deprecatedInfo.Severity,
				Message:  fmt.Sprintf("'%s' API for '%s' '%s' - %s", resource.APIVersion, resource.Kind, resource.Name, deprecatedInfo.DeprecationInfo),
				File:     resource.File,
				Line:     resource.Line,
				Resource: fmt.Sprintf("%s/%s", resource.APIVersion, resource.Kind),
			})
		}
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
