package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/config"
	"gopkg.in/yaml.v3"
)

// ResourceParser parses YAML files and extracts Kubernetes resources
type ResourceParser struct {
	repoPath string
	config   *config.Config
}

// NewResourceParser creates a new ResourceParser
func NewResourceParser(repoPath string, config *config.Config) *ResourceParser {
	return &ResourceParser{
		repoPath: repoPath,
		config:   config,
	}
}

// ParseAllResources parses all YAML files in the repository and returns a ResourceGraph
func (p *ResourceParser) ParseAllResources() (*ResourceGraph, error) {
	graph := NewResourceGraph()

	err := filepath.Walk(p.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if path should be ignored
		relPath, err := filepath.Rel(p.repoPath, path)
		if err != nil {
			return err
		}

		if p.config.ShouldIgnorePath(relPath) {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".yaml") && !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		resources, err := p.ParseFile(path)
		if err != nil {
			// Log error but continue parsing other files
			fmt.Printf("Warning: Failed to parse file %s: %v\n", path, err)
			return nil
		}

		for _, resource := range resources {
			graph.AddResource(resource)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk repository: %w", err)
	}

	// Extract references and build the dependency graph
	if err := graph.BuildDependencyGraph(p.repoPath); err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	return graph, nil
}

// ParseFile parses a single YAML file and extracts all resources (handles --- delimited resources)
func (p *ResourceParser) ParseFile(filePath string) ([]*ParsedResource, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var resources []*ParsedResource
	decoder := yaml.NewDecoder(file)

	for {
		var doc yaml.Node
		err := decoder.Decode(&doc)
		if err != nil {
			break // End of file or error
		}

		if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
			resource := p.parseResourceNode(doc.Content[0], filePath)
			if resource != nil {
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// parseResourceNode parses a single YAML document node into a ParsedResource
func (p *ResourceParser) parseResourceNode(node *yaml.Node, filePath string) *ParsedResource {
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var apiVersion, kind, name, namespace string
	var line int
	content := make(map[string]interface{})

	// Extract basic fields and build content map
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		if key.Value == "apiVersion" {
			apiVersion = value.Value
			line = value.Line
		} else if key.Value == "kind" {
			kind = value.Value
		} else if key.Value == "metadata" {
			if value.Kind == yaml.MappingNode {
				for j := 0; j < len(value.Content); j += 2 {
					if value.Content[j].Value == "name" {
						name = value.Content[j+1].Value
					} else if value.Content[j].Value == "namespace" {
						namespace = value.Content[j+1].Value
					}
				}
			}
		}

		// Build content map for further processing
		content[key.Value] = p.nodeToInterface(value)
	}

	// Skip if not a valid Kubernetes resource
	if apiVersion == "" || kind == "" || name == "" {
		return nil
	}

	resource := &ParsedResource{
		File:       filePath,
		Line:       line,
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
		Content:    content,
	}

	return resource
}

// nodeToInterface converts a YAML node to a Go interface{}
func (p *ResourceParser) nodeToInterface(node *yaml.Node) interface{} {
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Value
	case yaml.SequenceNode:
		var result []interface{}
		for _, item := range node.Content {
			result = append(result, p.nodeToInterface(item))
		}
		return result
	case yaml.MappingNode:
		result := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			result[key.Value] = p.nodeToInterface(value)
		}
		return result
	default:
		return nil
	}
}
