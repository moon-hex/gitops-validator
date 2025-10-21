package checks

import (
	"fmt"
	"regexp"

	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// DeprecatedAPICheck validates usage of deprecated Kubernetes API versions
func DeprecatedAPICheck(resource *parser.ParsedResource, config *config.Config) []types.ValidationResult {
	var results []types.ValidationResult

	// Check if the API version is deprecated
	deprecatedInfo := checkDeprecatedAPI(resource.APIVersion, config)
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

	return results
}

// checkDeprecatedAPI checks if an API version is deprecated
func checkDeprecatedAPI(apiVersion string, config *config.Config) *DeprecationInfo {
	// Check custom deprecated APIs from config
	for _, customAPI := range config.GitOpsValidator.DeprecatedAPIs.CustomAPIs {
		if matchesAPIVersion(apiVersion, customAPI.APIVersion) {
			return &DeprecationInfo{
				Severity:        customAPI.Severity,
				DeprecationInfo: customAPI.DeprecationInfo,
			}
		}
	}

	// Check built-in deprecated APIs
	return checkBuiltinDeprecatedAPI(apiVersion)
}

// DeprecationInfo represents information about a deprecated API
type DeprecationInfo struct {
	Severity        string
	DeprecationInfo string
}

// matchesAPIVersion checks if an API version matches a pattern
func matchesAPIVersion(apiVersion, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, apiVersion)
	return matched
}

// checkBuiltinDeprecatedAPI checks against built-in deprecated API patterns
func checkBuiltinDeprecatedAPI(apiVersion string) *DeprecationInfo {
	// Built-in deprecated API patterns
	deprecatedPatterns := map[string]DeprecationInfo{
		`^v1beta1/.*`: {
			Severity:        "warning",
			DeprecationInfo: "v1beta1 APIs are deprecated and will be removed in future Kubernetes versions",
		},
		`^v1alpha1/.*`: {
			Severity:        "warning",
			DeprecationInfo: "v1alpha1 APIs are experimental and may be removed without notice",
		},
		`^extensions/v1beta1/.*`: {
			Severity:        "error",
			DeprecationInfo: "extensions/v1beta1 APIs are deprecated and removed in Kubernetes 1.22+",
		},
		`^apps/v1beta1/.*`: {
			Severity:        "warning",
			DeprecationInfo: "apps/v1beta1 APIs are deprecated, use apps/v1 instead",
		},
		`^apps/v1beta2/.*`: {
			Severity:        "warning",
			DeprecationInfo: "apps/v1beta2 APIs are deprecated, use apps/v1 instead",
		},
	}

	for pattern, info := range deprecatedPatterns {
		if matchesAPIVersion(apiVersion, pattern) {
			return &info
		}
	}

	return nil
}
