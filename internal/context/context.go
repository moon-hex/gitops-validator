package context

import (
	"fmt"

	"github.com/moon-hex/gitops-validator/internal/chart"
	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/parser"
)

// ValidationContext provides context for validators
type ValidationContext struct {
	Graph    *parser.ResourceGraph
	Config   *config.Config
	RepoPath string
	Verbose  bool
}

// NewValidationContext creates a new ValidationContext
func NewValidationContext(graph *parser.ResourceGraph, cfg *config.Config, repoPath string, verbose bool) *ValidationContext {
	return &ValidationContext{
		Graph:    graph,
		Config:   cfg,
		RepoPath: repoPath,
		Verbose:  verbose,
	}
}

// FindEntryPoints finds all entry point resources based on configuration
func (ctx *ValidationContext) FindEntryPoints() []*parser.ParsedResource {
	var entryPoints []*parser.ParsedResource

	// Add explicitly configured resources
	for _, resourceName := range ctx.Config.GetEntryPointResources() {
		if resource := ctx.Graph.GetResource(resourceName); resource != nil {
			entryPoints = append(entryPoints, resource)
		}
	}

	// Add resources matching patterns
	for _, pattern := range ctx.Config.GetEntryPointPatterns() {
		matches := ctx.Graph.GetResourcesMatchingPattern(pattern)
		entryPoints = append(entryPoints, matches...)
	}

	// Add resources of specified types
	for _, resourceType := range ctx.Config.GetEntryPointTypes() {
		switch resourceType {
		case "flux-kustomization":
			entryPoints = append(entryPoints, ctx.Graph.GetFluxKustomizations()...)
		case "helm-release":
			entryPoints = append(entryPoints, ctx.Graph.GetHelmReleases()...)
		case "git-repository":
			entryPoints = append(entryPoints, ctx.Graph.GetFluxSources()...)
		case "kubernetes-kustomization":
			entryPoints = append(entryPoints, ctx.Graph.GetKubernetesKustomizations()...)
		}
	}

	// Add resources in specified namespaces
	for _, namespace := range ctx.Config.GetEntryPointNamespaces() {
		entryPoints = append(entryPoints, ctx.Graph.GetResourcesByNamespace(namespace)...)
	}

	// Auto-detect common Flux entry points if no explicit entry points found
	if len(entryPoints) == 0 {
		entryPoints = ctx.detectEntryPoints()
	}

	return entryPoints
}

// detectEntryPoints automatically detects common Flux entry points
func (ctx *ValidationContext) detectEntryPoints() []*parser.ParsedResource {
	var entryPoints []*parser.ParsedResource

	// Flux Kustomizations are always entry points
	entryPoints = append(entryPoints, ctx.Graph.GetFluxKustomizations()...)

	// HelmReleases are entry points
	entryPoints = append(entryPoints, ctx.Graph.GetHelmReleases()...)

	// Resources in flux-system namespace
	entryPoints = append(entryPoints, ctx.Graph.GetResourcesByNamespace("flux-system")...)

	// Resources in common GitOps directories
	commonDirs := []string{"apps", "infrastructure", "clusters"}
	for _, dir := range commonDirs {
		entryPoints = append(entryPoints, ctx.Graph.GetResourcesInDirectory(dir)...)
	}

	return entryPoints
}

// FindOrphanedResources finds resources that are not referenced by any entry point
func (ctx *ValidationContext) FindOrphanedResources(entryPoints []*parser.ParsedResource) []*parser.ParsedResource {
	visited := make(map[string]bool)

	// Start traversal from all entry points
	for _, entryPoint := range entryPoints {
		ctx.traverseFromResource(entryPoint, visited)
	}

	// Find unvisited resources
	var orphaned []*parser.ParsedResource
	for _, resource := range ctx.Graph.Resources {
		if !visited[resource.GetResourceKey()] {
			orphaned = append(orphaned, resource)
		}
	}

	return orphaned
}

// traverseFromResource performs a depth-first traversal from a resource
func (ctx *ValidationContext) traverseFromResource(resource *parser.ParsedResource, visited map[string]bool) {
	key := resource.GetResourceKey()
	if visited[key] {
		return // Already visited
	}

	visited[key] = true

	// Traverse dependencies
	for _, dep := range resource.Dependencies {
		if dep.ReferenceType == string(parser.ReferenceTypePath) || dep.ReferenceType == string(parser.ReferenceTypeResource) {
			// Find the target resource
			targetResource := ctx.Graph.FindTargetResource(dep, resource, ctx.RepoPath)
			if targetResource != nil {
				ctx.traverseFromResource(targetResource, visited)
			}
		}
	}
}

// FindDoubleReferencedResources finds resources that are referenced by multiple sources
func (ctx *ValidationContext) FindDoubleReferencedResources() []DoubleReference {
	var doubleRefs []DoubleReference

	for _, resource := range ctx.Graph.Resources {
		if len(resource.ReferencedBy) > 1 {
			doubleRefs = append(doubleRefs, DoubleReference{
				Resource:    resource,
				Referencers: resource.ReferencedBy,
			})
		}
	}

	return doubleRefs
}

// DoubleReference represents a resource that is referenced by multiple sources
type DoubleReference struct {
	Resource    *parser.ParsedResource
	Referencers []parser.ResourceReference
}

// GenerateDependencyChart generates a dependency chart in the specified format
func (ctx *ValidationContext) GenerateDependencyChart(format string) (string, error) {
	entryPoints := ctx.FindEntryPoints()
	orphaned := ctx.FindOrphanedResources(entryPoints)

	generator := chart.NewChartGenerator(ctx.Graph)

	switch format {
	case "mermaid":
		return generator.GenerateMermaidChart(entryPoints, orphaned), nil
	case "tree":
		return generator.GenerateTreeChart(entryPoints, orphaned), nil
	case "json":
		return generator.GenerateJSONChart(entryPoints, orphaned), nil
	default:
		return "", fmt.Errorf("unsupported chart format: %s", format)
	}
}

// GenerateDependencyChartForEntryPoint generates a dependency chart for a specific entry point
func (ctx *ValidationContext) GenerateDependencyChartForEntryPoint(entryPoint *parser.ParsedResource, format string) (string, error) {
	orphaned := ctx.FindOrphanedResources([]*parser.ParsedResource{entryPoint})

	generator := chart.NewChartGenerator(ctx.Graph)

	switch format {
	case "mermaid":
		return generator.GenerateMermaidChartForEntryPoint(entryPoint, orphaned), nil
	case "tree":
		return generator.GenerateTreeChart([]*parser.ParsedResource{entryPoint}, orphaned), nil
	case "json":
		return generator.GenerateJSONChart([]*parser.ParsedResource{entryPoint}, orphaned), nil
	default:
		return "", fmt.Errorf("unsupported chart format: %s", format)
	}
}
