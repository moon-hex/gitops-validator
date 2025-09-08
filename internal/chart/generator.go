package chart

import (
	"fmt"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/parser"
)

// ChartGenerator generates dependency charts from resource graphs
type ChartGenerator struct {
	graph *parser.ResourceGraph
}

// NewChartGenerator creates a new ChartGenerator
func NewChartGenerator(graph *parser.ResourceGraph) *ChartGenerator {
	return &ChartGenerator{
		graph: graph,
	}
}

// GenerateMermaidChart generates a Mermaid diagram of the dependency graph
func (g *ChartGenerator) GenerateMermaidChart(entryPoints []*parser.ParsedResource, orphaned []*parser.ParsedResource) string {
	return g.generateMermaidChartInternal(entryPoints, orphaned, nil)
}

// GenerateMermaidChartForEntryPoint generates a Mermaid diagram for a specific entry point
func (g *ChartGenerator) GenerateMermaidChartForEntryPoint(entryPoint *parser.ParsedResource, orphaned []*parser.ParsedResource) string {
	return g.generateMermaidChartInternal([]*parser.ParsedResource{entryPoint}, orphaned, entryPoint)
}

// generateMermaidChartInternal is the internal implementation for Mermaid chart generation
func (g *ChartGenerator) generateMermaidChartInternal(entryPoints []*parser.ParsedResource, orphaned []*parser.ParsedResource, filterEntryPoint *parser.ParsedResource) string {
	var lines []string

	lines = append(lines, "graph TD")

	// Track visited nodes to avoid duplicates
	visited := make(map[string]bool)
	nodeCounter := 0
	nodeMap := make(map[string]string) // resource key -> node ID

	// Generate nodes and edges for entry points and their dependencies
	for _, entryPoint := range entryPoints {
		g.generateNodeAndEdges(entryPoint, &lines, visited, &nodeCounter, nodeMap)
	}

	// Add orphaned resources
	if len(orphaned) > 0 {
		lines = append(lines, "")
		lines = append(lines, "    %% Orphaned Resources")
		for _, resource := range orphaned {
			nodeID := g.getOrCreateNodeID(resource, &nodeCounter, nodeMap)
			visited[resource.GetResourceKey()] = true

			icon := g.getResourceIcon(resource)
			label := fmt.Sprintf("%s<br/>%s", resource.Name, icon)
			lines = append(lines, fmt.Sprintf("    %s[\"%s\"]", nodeID, label))
		}
	}

	// Add styling with better contrast
	lines = append(lines, "")
	lines = append(lines, "    %% Styling")
	lines = append(lines, "    classDef valid fill:#2E8B57,stroke:#1F5F3F,stroke-width:3px,color:#FFFFFF")
	lines = append(lines, "    classDef orphaned fill:#DC143C,stroke:#8B0000,stroke-width:3px,color:#FFFFFF")
	lines = append(lines, "    classDef error fill:#B22222,stroke:#8B0000,stroke-width:3px,color:#FFFFFF")
	lines = append(lines, "    classDef warning fill:#FF8C00,stroke:#CC7000,stroke-width:3px,color:#FFFFFF")

	// Apply styles
	lines = append(lines, "")
	lines = append(lines, "    %% Apply styles")
	for resourceKey, nodeID := range nodeMap {
		if resource, exists := g.graph.Resources[resourceKey]; exists {
			resourceType := parser.ClassifyResource(resource)
			switch resourceType {
			case parser.ResourceTypeFluxKustomization:
				lines = append(lines, fmt.Sprintf("    class %s valid", nodeID))
			case parser.ResourceTypeKubernetesKustomization:
				lines = append(lines, fmt.Sprintf("    class %s valid", nodeID))
			case parser.ResourceTypeHelmRelease:
				lines = append(lines, fmt.Sprintf("    class %s valid", nodeID))
			default:
				lines = append(lines, fmt.Sprintf("    class %s valid", nodeID))
			}
		}
	}

	// Style orphaned resources
	for _, resource := range orphaned {
		if nodeID, exists := nodeMap[resource.GetResourceKey()]; exists {
			lines = append(lines, fmt.Sprintf("    class %s orphaned", nodeID))
		}
	}

	return strings.Join(lines, "\n")
}

// generateNodeAndEdges recursively generates nodes and edges for a resource and its dependencies
func (g *ChartGenerator) generateNodeAndEdges(resource *parser.ParsedResource, lines *[]string, visited map[string]bool, nodeCounter *int, nodeMap map[string]string) {
	resourceKey := resource.GetResourceKey()
	if visited[resourceKey] {
		return
	}

	visited[resourceKey] = true

	// Create node for this resource
	nodeID := g.getOrCreateNodeID(resource, nodeCounter, nodeMap)
	icon := g.getResourceIcon(resource)
	label := fmt.Sprintf("%s<br/>%s", resource.Name, icon)
	*lines = append(*lines, fmt.Sprintf("    %s[\"%s\"]", nodeID, label))

	// Generate edges to dependencies
	for _, dep := range resource.Dependencies {
		if dep.ReferenceType == string(parser.ReferenceTypePath) || dep.ReferenceType == string(parser.ReferenceTypeResource) {
			// Find the target resource
			targetResource := g.graph.FindTargetResource(dep, resource, "")
			if targetResource != nil {
				targetNodeID := g.getOrCreateNodeID(targetResource, nodeCounter, nodeMap)
				edgeLabel := g.getEdgeLabel(dep)
				*lines = append(*lines, fmt.Sprintf("    %s -->|%s| %s", nodeID, edgeLabel, targetNodeID))

				// Recursively process the target resource
				g.generateNodeAndEdges(targetResource, lines, visited, nodeCounter, nodeMap)
			}
		}
	}
}

// getOrCreateNodeID gets or creates a unique node ID for a resource
func (g *ChartGenerator) getOrCreateNodeID(resource *parser.ParsedResource, nodeCounter *int, nodeMap map[string]string) string {
	resourceKey := resource.GetResourceKey()
	if nodeID, exists := nodeMap[resourceKey]; exists {
		return nodeID
	}

	*nodeCounter++
	nodeID := fmt.Sprintf("N%d", *nodeCounter)
	nodeMap[resourceKey] = nodeID
	return nodeID
}

// getResourceIcon returns an appropriate icon for the resource type
func (g *ChartGenerator) getResourceIcon(resource *parser.ParsedResource) string {
	resourceType := parser.ClassifyResource(resource)
	switch resourceType {
	case parser.ResourceTypeFluxKustomization:
		return "📁 flux-kustomization"
	case parser.ResourceTypeKubernetesKustomization:
		return "📁 kustomization"
	case parser.ResourceTypeHelmRelease:
		return "🚀 helm-release"
	case parser.ResourceTypeFluxSource:
		return "📦 flux-source"
	case parser.ResourceTypeFluxImage:
		return "🖼️ flux-image"
	case parser.ResourceTypeFluxNotification:
		return "🔔 flux-notification"
	default:
		return "📄 kubernetes-resource"
	}
}

// getEdgeLabel returns a label for the edge based on the reference type
func (g *ChartGenerator) getEdgeLabel(ref parser.ResourceReference) string {
	switch ref.ReferenceType {
	case string(parser.ReferenceTypePath):
		return "path"
	case string(parser.ReferenceTypeSourceRef):
		return "sourceRef"
	case string(parser.ReferenceTypeChart):
		return "chart"
	case string(parser.ReferenceTypeResource):
		return "resource"
	default:
		return ref.ReferenceType
	}
}

// GenerateTreeChart generates a text-based tree chart
func (g *ChartGenerator) GenerateTreeChart(entryPoints []*parser.ParsedResource, orphaned []*parser.ParsedResource) string {
	var lines []string

	visited := make(map[string]bool)

	// Generate tree for each entry point
	for _, entryPoint := range entryPoints {
		g.generateTreeNode(entryPoint, "", &lines, visited, true)
	}

	// Add orphaned resources
	if len(orphaned) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Orphaned Resources:")
		for _, resource := range orphaned {
			icon := g.getResourceIcon(resource)
			lines = append(lines, fmt.Sprintf("└── %s %s", icon, resource.Name))
		}
	}

	return strings.Join(lines, "\n")
}

// generateTreeNode recursively generates tree nodes
func (g *ChartGenerator) generateTreeNode(resource *parser.ParsedResource, prefix string, lines *[]string, visited map[string]bool, isLast bool) {
	resourceKey := resource.GetResourceKey()
	if visited[resourceKey] {
		return
	}

	visited[resourceKey] = true

	icon := g.getResourceIcon(resource)
	nodePrefix := "└── "
	if !isLast {
		nodePrefix = "├── "
	}

	*lines = append(*lines, fmt.Sprintf("%s%s %s", prefix, nodePrefix, icon))

	// Add dependencies
	deps := resource.Dependencies
	for i, dep := range deps {
		if dep.ReferenceType == string(parser.ReferenceTypePath) || dep.ReferenceType == string(parser.ReferenceTypeResource) {
			targetResource := g.graph.FindTargetResource(dep, resource, "")
			if targetResource != nil {
				childPrefix := prefix
				if isLast {
					childPrefix += "    "
				} else {
					childPrefix += "│   "
				}

				isLastDep := i == len(deps)-1
				g.generateTreeNode(targetResource, childPrefix, lines, visited, isLastDep)
			}
		}
	}
}

// GenerateJSONChart generates a JSON representation of the dependency graph
func (g *ChartGenerator) GenerateJSONChart(entryPoints []*parser.ParsedResource, orphaned []*parser.ParsedResource) string {
	// This would generate a JSON structure for the graph
	// For now, return a simple representation
	return fmt.Sprintf(`{
  "entryPoints": %d,
  "totalResources": %d,
  "orphanedResources": %d,
  "note": "Full JSON chart generation not yet implemented"
}`, len(entryPoints), len(g.graph.Resources), len(orphaned))
}
