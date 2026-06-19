package checks

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// HTTPRoutePolicyCheck reports any HTTPRoute or VirtualService that has no SecurityPolicy
// defined in the same namespace.
//
// Matching rule: a route is considered protected if at least one SecurityPolicy resource
// exists in the same namespace. Routes whose namespace cannot be determined (empty
// metadata.namespace) are skipped with an info-level notice rather than a false positive.
func HTTPRoutePolicyCheck(ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	// Index namespaces that have at least one SecurityPolicy
	protectedNamespaces := buildProtectedNamespaceIndex(ctx)

	// Check Gateway API HTTPRoutes
	for _, route := range ctx.Graph.GetHTTPRoutes() {
		results = append(results, checkRouteProtection(route, protectedNamespaces, "HTTPRoute")...)
	}

	// Check Istio VirtualServices
	for _, vs := range ctx.Graph.GetVirtualServices() {
		results = append(results, checkRouteProtection(vs, protectedNamespaces, "VirtualService")...)
	}

	return results
}

// buildProtectedNamespaceIndex returns the set of namespaces that contain at least one
// SecurityPolicy resource.
func buildProtectedNamespaceIndex(ctx *context.ValidationContext) map[string]bool {
	protected := make(map[string]bool)
	for _, policy := range ctx.Graph.GetSecurityPolicies() {
		if policy.Namespace != "" {
			protected[policy.Namespace] = true
		}
	}
	return protected
}

// checkRouteProtection returns a validation result if the route is not protected.
func checkRouteProtection(route *parser.ParsedResource, protectedNamespaces map[string]bool, kind string) []types.ValidationResult {
	var results []types.ValidationResult

	if route.Namespace == "" {
		// Namespace not present in the YAML — may be set via kustomization patch.
		// Emit info rather than a spurious warning.
		results = append(results, types.ValidationResult{
			Type:     "http-route-policy",
			Severity: "info",
			Message: fmt.Sprintf(
				"%s '%s' has no metadata.namespace — cannot verify SecurityPolicy coverage (namespace may be injected by kustomize)",
				kind, route.Name,
			),
			File:     route.File,
			Line:     route.Line,
			Resource: route.Name,
		})
		return results
	}

	if !protectedNamespaces[route.Namespace] {
		results = append(results, types.ValidationResult{
			Type:     "http-route-policy",
			Severity: "warning",
			Message: fmt.Sprintf(
				"%s '%s' in namespace '%s' is not protected by any SecurityPolicy",
				kind, route.Name, route.Namespace,
			),
			File:     route.File,
			Line:     route.Line,
			Resource: route.Name,
		})
	}

	return results
}
