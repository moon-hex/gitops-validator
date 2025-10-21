package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResourceIndex provides fast lookup structures for large repositories
type ResourceIndex struct {
	// By API version and kind
	byAPIVersionKind map[string]map[string][]*ParsedResource

	// By file path
	byFilePath map[string]*ParsedResource

	// By resource name
	byResourceName map[string][]*ParsedResource

	// By namespace
	byNamespace map[string][]*ParsedResource

	// By directory
	byDirectory map[string][]*ParsedResource

	// By resource type (Flux vs Kubernetes)
	fluxKustomizations       []*ParsedResource
	kubernetesKustomizations []*ParsedResource
	helmReleases             []*ParsedResource
	otherResources           []*ParsedResource

	// Dependency graph for fast traversal
	dependencyGraph     map[string][]string
	reverseDependencies map[string][]string
}

// NewResourceIndex creates a new resource index
func NewResourceIndex() *ResourceIndex {
	return &ResourceIndex{
		byAPIVersionKind:         make(map[string]map[string][]*ParsedResource),
		byFilePath:               make(map[string]*ParsedResource),
		byResourceName:           make(map[string][]*ParsedResource),
		byNamespace:              make(map[string][]*ParsedResource),
		byDirectory:              make(map[string][]*ParsedResource),
		fluxKustomizations:       make([]*ParsedResource, 0),
		kubernetesKustomizations: make([]*ParsedResource, 0),
		helmReleases:             make([]*ParsedResource, 0),
		otherResources:           make([]*ParsedResource, 0),
		dependencyGraph:          make(map[string][]string),
		reverseDependencies:      make(map[string][]string),
	}
}

// BuildIndex builds the index from a list of parsed resources
func (ri *ResourceIndex) BuildIndex(resources []*ParsedResource) error {
	// Clear existing index
	ri.clear()

	for _, resource := range resources {
		if err := ri.addResource(resource); err != nil {
			return fmt.Errorf("failed to index resource %s: %w", resource.File, err)
		}
	}

	// Build dependency graph
	ri.buildDependencyGraph(resources)

	return nil
}

// addResource adds a single resource to the index
func (ri *ResourceIndex) addResource(resource *ParsedResource) error {
	// Index by file path
	ri.byFilePath[resource.File] = resource

	// Index by API version and kind
	if ri.byAPIVersionKind[resource.APIVersion] == nil {
		ri.byAPIVersionKind[resource.APIVersion] = make(map[string][]*ParsedResource)
	}
	ri.byAPIVersionKind[resource.APIVersion][resource.Kind] = append(
		ri.byAPIVersionKind[resource.APIVersion][resource.Kind], resource)

	// Index by resource name
	fullName := resource.Name
	if resource.Namespace != "" {
		fullName = fmt.Sprintf("%s/%s", resource.Namespace, resource.Name)
	}
	ri.byResourceName[fullName] = append(ri.byResourceName[fullName], resource)

	// Index by namespace
	if resource.Namespace != "" {
		ri.byNamespace[resource.Namespace] = append(ri.byNamespace[resource.Namespace], resource)
	}

	// Index by directory
	dir := filepath.Dir(resource.File)
	ri.byDirectory[dir] = append(ri.byDirectory[dir], resource)

	// Categorize by resource type
	ri.categorizeResource(resource)

	return nil
}

// categorizeResource categorizes a resource by type
func (ri *ResourceIndex) categorizeResource(resource *ParsedResource) {
	switch {
	case ri.isFluxKustomization(resource):
		ri.fluxKustomizations = append(ri.fluxKustomizations, resource)
	case ri.isKubernetesKustomization(resource):
		ri.kubernetesKustomizations = append(ri.kubernetesKustomizations, resource)
	case ri.isHelmRelease(resource):
		ri.helmReleases = append(ri.helmReleases, resource)
	default:
		ri.otherResources = append(ri.otherResources, resource)
	}
}

// buildDependencyGraph builds the dependency graph for fast traversal
func (ri *ResourceIndex) buildDependencyGraph(resources []*ParsedResource) {
	for _, resource := range resources {
		// Convert ResourceReference to string paths
		var depPaths []string
		for _, dep := range resource.Dependencies {
			depPaths = append(depPaths, dep.Path)
		}
		ri.dependencyGraph[resource.File] = depPaths

		// Build reverse dependencies
		for _, dep := range resource.Dependencies {
			ri.reverseDependencies[dep.Path] = append(ri.reverseDependencies[dep.Path], resource.File)
		}
	}
}

// Query methods for fast lookups

// GetByAPIVersionKind returns resources matching the given API version and kind
func (ri *ResourceIndex) GetByAPIVersionKind(apiVersion, kind string) []*ParsedResource {
	if ri.byAPIVersionKind[apiVersion] != nil {
		return ri.byAPIVersionKind[apiVersion][kind]
	}
	return nil
}

// GetByFilePath returns the resource at the given file path
func (ri *ResourceIndex) GetByFilePath(filePath string) *ParsedResource {
	return ri.byFilePath[filePath]
}

// GetByResourceName returns resources with the given name
func (ri *ResourceIndex) GetByResourceName(name string) []*ParsedResource {
	return ri.byResourceName[name]
}

// GetByNamespace returns resources in the given namespace
func (ri *ResourceIndex) GetByNamespace(namespace string) []*ParsedResource {
	return ri.byNamespace[namespace]
}

// GetByDirectory returns resources in the given directory
func (ri *ResourceIndex) GetByDirectory(directory string) []*ParsedResource {
	return ri.byDirectory[directory]
}

// GetFluxKustomizations returns all Flux Kustomization resources
func (ri *ResourceIndex) GetFluxKustomizations() []*ParsedResource {
	return ri.fluxKustomizations
}

// GetKubernetesKustomizations returns all Kubernetes Kustomization resources
func (ri *ResourceIndex) GetKubernetesKustomizations() []*ParsedResource {
	return ri.kubernetesKustomizations
}

// GetHelmReleases returns all Helm Release resources
func (ri *ResourceIndex) GetHelmReleases() []*ParsedResource {
	return ri.helmReleases
}

// GetDependencies returns direct dependencies of a resource
func (ri *ResourceIndex) GetDependencies(filePath string) []string {
	return ri.dependencyGraph[filePath]
}

// GetReverseDependencies returns resources that depend on the given resource
func (ri *ResourceIndex) GetReverseDependencies(filePath string) []string {
	return ri.reverseDependencies[filePath]
}

// FindResourcesByPattern finds resources matching a pattern
func (ri *ResourceIndex) FindResourcesByPattern(pattern string) []*ParsedResource {
	var results []*ParsedResource

	for _, resource := range ri.byFilePath {
		if strings.Contains(resource.File, pattern) ||
			strings.Contains(resource.Name, pattern) ||
			strings.Contains(resource.Kind, pattern) {
			results = append(results, resource)
		}
	}

	return results
}

// GetIndexStats returns statistics about the index
func (ri *ResourceIndex) GetIndexStats() map[string]interface{} {
	return map[string]interface{}{
		"total_resources":           len(ri.byFilePath),
		"flux_kustomizations":       len(ri.fluxKustomizations),
		"kubernetes_kustomizations": len(ri.kubernetesKustomizations),
		"helm_releases":             len(ri.helmReleases),
		"other_resources":           len(ri.otherResources),
		"unique_api_versions":       len(ri.byAPIVersionKind),
		"unique_kinds":              ri.countUniqueKinds(),
		"unique_namespaces":         len(ri.byNamespace),
		"unique_directories":        len(ri.byDirectory),
		"dependency_relationships":  len(ri.dependencyGraph),
	}
}

// countUniqueKinds counts unique kinds across all API versions
func (ri *ResourceIndex) countUniqueKinds() int {
	uniqueKinds := make(map[string]bool)

	for _, kinds := range ri.byAPIVersionKind {
		for kind := range kinds {
			uniqueKinds[kind] = true
		}
	}

	return len(uniqueKinds)
}

// clear clears all index data
func (ri *ResourceIndex) clear() {
	ri.byAPIVersionKind = make(map[string]map[string][]*ParsedResource)
	ri.byFilePath = make(map[string]*ParsedResource)
	ri.byResourceName = make(map[string][]*ParsedResource)
	ri.byNamespace = make(map[string][]*ParsedResource)
	ri.byDirectory = make(map[string][]*ParsedResource)
	ri.fluxKustomizations = make([]*ParsedResource, 0)
	ri.kubernetesKustomizations = make([]*ParsedResource, 0)
	ri.helmReleases = make([]*ParsedResource, 0)
	ri.otherResources = make([]*ParsedResource, 0)
	ri.dependencyGraph = make(map[string][]string)
	ri.reverseDependencies = make(map[string][]string)
}

// Helper methods for resource type detection

func (ri *ResourceIndex) isFluxKustomization(resource *ParsedResource) bool {
	return resource.APIVersion == "kustomize.toolkit.fluxcd.io/v1" &&
		resource.Kind == "Kustomization"
}

func (ri *ResourceIndex) isKubernetesKustomization(resource *ParsedResource) bool {
	return resource.APIVersion == "kustomize.config.k8s.io/v1beta1" &&
		resource.Kind == "Kustomization"
}

func (ri *ResourceIndex) isHelmRelease(resource *ParsedResource) bool {
	return resource.APIVersion == "helm.toolkit.fluxcd.io/v2beta1" &&
		resource.Kind == "HelmRelease"
}
