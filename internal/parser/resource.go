package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParsedResource represents a parsed Kubernetes resource
type ParsedResource struct {
	File         string                 // Source file path
	Line         int                    // Line number in file
	APIVersion   string                 // apiVersion
	Kind         string                 // kind
	Name         string                 // metadata.name
	Namespace    string                 // metadata.namespace
	Content      map[string]interface{} // Full resource content
	Dependencies []ResourceReference    // What this resource references
	ReferencedBy []ResourceReference    // What references this resource
}

// ResourceReference represents a reference from one resource to another
type ResourceReference struct {
	Type          string // "kustomization", "helmrelease", "flux-source", etc.
	Name          string // Resource name
	File          string // Source file
	Line          int    // Line number
	ReferenceType string // "path", "sourceRef", "chart", etc.
	Path          string // The actual path/reference value
	IsRelative    bool   // Whether the path is relative to the file or repo root
}

// ResourceType represents the type of a resource
type ResourceType string

const (
	ResourceTypeFluxKustomization       ResourceType = "flux-kustomization"
	ResourceTypeKubernetesKustomization ResourceType = "kubernetes-kustomization"
	ResourceTypeHelmRelease             ResourceType = "helm-release"
	ResourceTypeFluxSource              ResourceType = "flux-source"
	ResourceTypeFluxImage               ResourceType = "flux-image"
	ResourceTypeFluxNotification        ResourceType = "flux-notification"
	ResourceTypeKubernetesResource      ResourceType = "kubernetes-resource"
)

// ReferenceType represents the type of a reference
type ReferenceType string

const (
	ReferenceTypePath      ReferenceType = "path"
	ReferenceTypeSourceRef ReferenceType = "sourceRef"
	ReferenceTypeChart     ReferenceType = "chart"
	ReferenceTypeImage     ReferenceType = "image"
	ReferenceTypeResource  ReferenceType = "resource"
)

// GetResourceKey returns a unique key for the resource
func (r *ParsedResource) GetResourceKey() string {
	if r.Namespace != "" {
		return fmt.Sprintf("%s/%s", r.Namespace, r.Name)
	}
	return r.Name
}

// ClassifyResource determines the type of a resource
func ClassifyResource(resource *ParsedResource) ResourceType {
	switch {
	case resource.Kind == "Kustomization" && strings.HasPrefix(resource.APIVersion, "kustomize.toolkit.fluxcd.io/"):
		return ResourceTypeFluxKustomization
	case resource.Kind == "HelmRelease" && strings.HasPrefix(resource.APIVersion, "helm.toolkit.fluxcd.io/"):
		return ResourceTypeHelmRelease
	case resource.Kind == "GitRepository" && strings.HasPrefix(resource.APIVersion, "source.toolkit.fluxcd.io/"):
		return ResourceTypeFluxSource
	case resource.Kind == "HelmRepository" && strings.HasPrefix(resource.APIVersion, "source.toolkit.fluxcd.io/"):
		return ResourceTypeFluxSource
	case resource.Kind == "ImageRepository" && strings.HasPrefix(resource.APIVersion, "image.toolkit.fluxcd.io/"):
		return ResourceTypeFluxImage
	case resource.Kind == "ImagePolicy" && strings.HasPrefix(resource.APIVersion, "image.toolkit.fluxcd.io/"):
		return ResourceTypeFluxImage
	case resource.Kind == "ImageUpdateAutomation" && strings.HasPrefix(resource.APIVersion, "image.toolkit.fluxcd.io/"):
		return ResourceTypeFluxImage
	case resource.Kind == "Alert" && strings.HasPrefix(resource.APIVersion, "notification.toolkit.fluxcd.io/"):
		return ResourceTypeFluxNotification
	case resource.Kind == "Provider" && strings.HasPrefix(resource.APIVersion, "notification.toolkit.fluxcd.io/"):
		return ResourceTypeFluxNotification
	case resource.Kind == "Receiver" && strings.HasPrefix(resource.APIVersion, "notification.toolkit.fluxcd.io/"):
		return ResourceTypeFluxNotification
	default:
		return ResourceTypeKubernetesResource
	}
}

// IsKustomizationFile checks if a file is a kustomization.yaml file
func IsKustomizationFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	return fileName == "kustomization.yaml" || fileName == "kustomization.yml"
}

// ExtractReferences extracts all references from a parsed resource
func ExtractReferences(resource *ParsedResource, repoPath string) []ResourceReference {
	var references []ResourceReference

	switch ClassifyResource(resource) {
	case ResourceTypeFluxKustomization:
		references = append(references, extractFluxKustomizationReferences(resource, repoPath)...)
	case ResourceTypeKubernetesKustomization:
		references = append(references, extractKubernetesKustomizationReferences(resource, repoPath)...)
	case ResourceTypeHelmRelease:
		references = append(references, extractHelmReleaseReferences(resource, repoPath)...)
	}

	return references
}

// extractFluxKustomizationReferences extracts references from Flux Kustomization resources
func extractFluxKustomizationReferences(resource *ParsedResource, repoPath string) []ResourceReference {
	var references []ResourceReference

	// Extract path reference (relative to repo root)
	if spec, ok := resource.Content["spec"].(map[string]interface{}); ok {
		if path, ok := spec["path"].(string); ok {
			references = append(references, ResourceReference{
				Type:          "flux-kustomization-path",
				Name:          resource.Name,
				File:          resource.File,
				Line:          resource.Line,
				ReferenceType: string(ReferenceTypePath),
				Path:          path,
				IsRelative:    false, // Flux paths are relative to repo root
			})
		}

		// Extract sourceRef reference
		if sourceRef, ok := spec["sourceRef"].(map[string]interface{}); ok {
			if name, ok := sourceRef["name"].(string); ok {
				references = append(references, ResourceReference{
					Type:          "flux-source",
					Name:          name,
					File:          resource.File,
					Line:          resource.Line,
					ReferenceType: string(ReferenceTypeSourceRef),
					Path:          name,
					IsRelative:    false,
				})
			}
		}
	}

	return references
}

// extractKubernetesKustomizationReferences extracts references from kustomization.yaml files
func extractKubernetesKustomizationReferences(resource *ParsedResource, repoPath string) []ResourceReference {
	var references []ResourceReference

	// Extract resources references (relative to kustomization file)
	if resources, ok := resource.Content["resources"].([]interface{}); ok {
		for _, res := range resources {
			if resPath, ok := res.(string); ok {
				references = append(references, ResourceReference{
					Type:          "kustomization-resource",
					Name:          resource.Name,
					File:          resource.File,
					Line:          resource.Line,
					ReferenceType: string(ReferenceTypeResource),
					Path:          resPath,
					IsRelative:    true, // K8s kustomization paths are relative to the file
				})
			}
		}
	}

	// Extract patches references
	if patches, ok := resource.Content["patches"].([]interface{}); ok {
		for _, patch := range patches {
			if patchMap, ok := patch.(map[string]interface{}); ok {
				if path, ok := patchMap["path"].(string); ok {
					references = append(references, ResourceReference{
						Type:          "kustomization-patch",
						Name:          resource.Name,
						File:          resource.File,
						Line:          resource.Line,
						ReferenceType: string(ReferenceTypePath),
						Path:          path,
						IsRelative:    true, // K8s kustomization paths are relative to the file
					})
				}
			}
		}
	}

	// Extract patchesStrategicMerge references
	if patches, ok := resource.Content["patchesStrategicMerge"].([]interface{}); ok {
		for _, patch := range patches {
			if patchPath, ok := patch.(string); ok {
				references = append(references, ResourceReference{
					Type:          "kustomization-patch-strategic",
					Name:          resource.Name,
					File:          resource.File,
					Line:          resource.Line,
					ReferenceType: string(ReferenceTypePath),
					Path:          patchPath,
					IsRelative:    true, // K8s kustomization paths are relative to the file
				})
			}
		}
	}

	return references
}

// extractHelmReleaseReferences extracts references from HelmRelease resources
func extractHelmReleaseReferences(resource *ParsedResource, repoPath string) []ResourceReference {
	var references []ResourceReference

	if spec, ok := resource.Content["spec"].(map[string]interface{}); ok {
		// Extract chart reference
		if chart, ok := spec["chart"].(map[string]interface{}); ok {
			if spec, ok := chart["spec"].(map[string]interface{}); ok {
				if chart, ok := spec["chart"].(string); ok {
					references = append(references, ResourceReference{
						Type:          "helm-chart",
						Name:          resource.Name,
						File:          resource.File,
						Line:          resource.Line,
						ReferenceType: string(ReferenceTypeChart),
						Path:          chart,
						IsRelative:    false,
					})
				}

				// Extract sourceRef reference
				if sourceRef, ok := spec["sourceRef"].(map[string]interface{}); ok {
					if name, ok := sourceRef["name"].(string); ok {
						references = append(references, ResourceReference{
							Type:          "helm-source",
							Name:          name,
							File:          resource.File,
							Line:          resource.Line,
							ReferenceType: string(ReferenceTypeSourceRef),
							Path:          name,
							IsRelative:    false,
						})
					}
				}
			}
		}
	}

	return references
}
