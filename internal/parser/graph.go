package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResourceGraph represents the dependency graph of all resources
type ResourceGraph struct {
	Resources    map[string]*ParsedResource         // Key: "namespace/name" or "name"
	Files        map[string][]*ParsedResource       // Key: file path
	ByKind       map[string][]*ParsedResource       // Key: kind
	ByAPIVersion map[string][]*ParsedResource       // Key: apiVersion
	ByType       map[ResourceType][]*ParsedResource // Key: resource type
	// Phase III: Fast lookup index
	Index *ResourceIndex
}

// NewResourceGraph creates a new ResourceGraph
func NewResourceGraph() *ResourceGraph {
	return &ResourceGraph{
		Resources:    make(map[string]*ParsedResource),
		Files:        make(map[string][]*ParsedResource),
		ByKind:       make(map[string][]*ParsedResource),
		ByAPIVersion: make(map[string][]*ParsedResource),
		ByType:       make(map[ResourceType][]*ParsedResource),
		Index:        NewResourceIndex(),
	}
}

// AddResource adds a resource to the graph
func (g *ResourceGraph) AddResource(resource *ParsedResource) {
	key := resource.GetResourceKey()
	g.Resources[key] = resource

	// Add to file index
	g.Files[resource.File] = append(g.Files[resource.File], resource)

	// Add to kind index
	g.ByKind[resource.Kind] = append(g.ByKind[resource.Kind], resource)

	// Add to API version index
	g.ByAPIVersion[resource.APIVersion] = append(g.ByAPIVersion[resource.APIVersion], resource)

	// Add to type index
	resourceType := ClassifyResource(resource)
	g.ByType[resourceType] = append(g.ByType[resourceType], resource)
}

// BuildDependencyGraph extracts references and builds the dependency graph
func (g *ResourceGraph) BuildDependencyGraph(repoPath string) error {
	for _, resource := range g.Resources {
		// Extract references from the resource
		references := ExtractReferences(resource, repoPath)
		resource.Dependencies = references

		// For each reference, find the target resource and add reverse reference
		for _, ref := range references {
			targetResource := g.FindTargetResource(ref, resource, repoPath)
			if targetResource != nil {
				targetResource.ReferencedBy = append(targetResource.ReferencedBy, ResourceReference{
					Type:          ref.Type,
					Name:          resource.Name,
					File:          resource.File,
					Line:          resource.Line,
					ReferenceType: ref.ReferenceType,
					Path:          ref.Path,
					IsRelative:    ref.IsRelative,
				})
			}
		}
	}

	return nil
}

// FindTargetResource finds the target resource for a reference
func (g *ResourceGraph) FindTargetResource(ref ResourceReference, sourceResource *ParsedResource, repoPath string) *ParsedResource {
	switch ref.ReferenceType {
	case string(ReferenceTypePath):
		return g.findResourceByPath(ref.Path, ref.IsRelative, sourceResource.File, repoPath)
	case string(ReferenceTypeSourceRef):
		return g.findResourceByName(ref.Path)
	case string(ReferenceTypeChart):
		// For Helm charts, we might not have the chart as a resource
		// This could be extended to check HelmRepository resources
		return nil
	default:
		return nil
	}
}

// findResourceByPath finds a resource by its file path
func (g *ResourceGraph) findResourceByPath(path string, isRelative bool, sourceFile string, repoPath string) *ParsedResource {
	var fullPath string

	if isRelative {
		// Path is relative to the source file
		fullPath = filepath.Join(filepath.Dir(sourceFile), path)
	} else {
		// Path is relative to repo root
		fullPath = filepath.Join(repoPath, path)
	}

	// Look for resources in this file
	if resources, exists := g.Files[fullPath]; exists {
		// If it's a kustomization file, return the first resource (should be the kustomization)
		if IsKustomizationFile(fullPath) && len(resources) > 0 {
			return resources[0]
		}
		// For other files, return the first resource
		if len(resources) > 0 {
			return resources[0]
		}
	}

	return nil
}

// findResourceByName finds a resource by its name
func (g *ResourceGraph) findResourceByName(name string) *ParsedResource {
	// Try exact match first
	if resource, exists := g.Resources[name]; exists {
		return resource
	}

	// Try namespace/name format
	for key, resource := range g.Resources {
		if strings.HasSuffix(key, "/"+name) {
			return resource
		}
	}

	return nil
}

// Query Functions

// GetResource returns a resource by its key
func (g *ResourceGraph) GetResource(key string) *ParsedResource {
	return g.Resources[key]
}

// GetResourcesByKind returns all resources of a specific kind
func (g *ResourceGraph) GetResourcesByKind(kind string) []*ParsedResource {
	return g.ByKind[kind]
}

// GetResourcesByAPIVersion returns all resources of a specific API version
func (g *ResourceGraph) GetResourcesByAPIVersion(apiVersion string) []*ParsedResource {
	return g.ByAPIVersion[apiVersion]
}

// GetResourcesByType returns all resources of a specific type
func (g *ResourceGraph) GetResourcesByType(resourceType ResourceType) []*ParsedResource {
	return g.ByType[resourceType]
}

// GetResourcesByNamespace returns all resources in a specific namespace
func (g *ResourceGraph) GetResourcesByNamespace(namespace string) []*ParsedResource {
	var resources []*ParsedResource
	for _, resource := range g.Resources {
		if resource.Namespace == namespace {
			resources = append(resources, resource)
		}
	}
	return resources
}

// GetResourcesInDirectory returns all resources in a specific directory
func (g *ResourceGraph) GetResourcesInDirectory(dir string) []*ParsedResource {
	var resources []*ParsedResource
	for filePath, fileResources := range g.Files {
		if strings.Contains(filePath, "/"+dir+"/") || strings.HasPrefix(filePath, dir+"/") {
			resources = append(resources, fileResources...)
		}
	}
	return resources
}

// GetResourcesMatchingPattern returns all resources matching a glob pattern
func (g *ResourceGraph) GetResourcesMatchingPattern(pattern string) []*ParsedResource {
	var resources []*ParsedResource
	for filePath, fileResources := range g.Files {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			resources = append(resources, fileResources...)
		}
	}
	return resources
}

// Flux-specific query functions

// GetFluxKustomizations returns all Flux Kustomization resources
func (g *ResourceGraph) GetFluxKustomizations() []*ParsedResource {
	return g.ByType[ResourceTypeFluxKustomization]
}

// GetKubernetesKustomizations returns all Kubernetes kustomization.yaml files
func (g *ResourceGraph) GetKubernetesKustomizations() []*ParsedResource {
	var resources []*ParsedResource
	for filePath, fileResources := range g.Files {
		if IsKustomizationFile(filePath) {
			resources = append(resources, fileResources...)
		}
	}
	return resources
}

// GetHelmReleases returns all HelmRelease resources
func (g *ResourceGraph) GetHelmReleases() []*ParsedResource {
	return g.ByType[ResourceTypeHelmRelease]
}

// GetFluxSources returns all Flux Source resources
func (g *ResourceGraph) GetFluxSources() []*ParsedResource {
	return g.ByType[ResourceTypeFluxSource]
}

// Validation helper functions

// ValidatePathReference checks if a path reference exists
func (g *ResourceGraph) ValidatePathReference(path string, isRelative bool, sourceFile string, repoPath string) error {
	var fullPath string

	if isRelative {
		fullPath = filepath.Join(filepath.Dir(sourceFile), path)
	} else {
		fullPath = filepath.Join(repoPath, path)
	}

	// Check if file exists
	if _, exists := g.Files[fullPath]; !exists {
		return fmt.Errorf("path '%s' does not exist", path)
	}

	return nil
}

// BuildIndex builds the fast lookup index for the graph
func (g *ResourceGraph) BuildIndex() error {
	// Convert map to slice for indexing
	var resources []*ParsedResource
	for _, resource := range g.Resources {
		resources = append(resources, resource)
	}

	return g.Index.BuildIndex(resources)
}

// ValidateResourceReference checks if a resource reference exists
func (g *ResourceGraph) ValidateResourceReference(ref ResourceReference) error {
	targetResource := g.findResourceByName(ref.Path)
	if targetResource == nil {
		return fmt.Errorf("resource '%s' not found", ref.Path)
	}

	return nil
}
