package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/config"
	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/parser"
	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators"
)

type Validator struct {
	repoPath string
	verbose  bool
	yamlPath string
	config   *config.Config
	parser   *parser.ResourceParser
	graph    *parser.ResourceGraph
	results  []types.ValidationResult
}

func NewValidator(repoPath string, verbose bool, yamlPath string) *Validator {
	// Load configuration from file
	cfg := config.DefaultConfig()

	// Try to load from .gitops-validator.yaml
	if _, err := os.Stat(".gitops-validator.yaml"); err == nil {
		if loadedConfig, err := config.LoadConfig(".gitops-validator.yaml"); err == nil {
			cfg = loadedConfig
		}
	}

	return &Validator{
		repoPath: repoPath,
		verbose:  verbose,
		yamlPath: yamlPath,
		config:   cfg,
		parser:   parser.NewResourceParser(repoPath, cfg),
		results:  make([]types.ValidationResult, 0),
	}
}

func (v *Validator) Validate() error {
	if v.verbose {
		fmt.Printf("Starting validation of repository: %s\n", v.repoPath)
	}

	// Check if repository path exists
	if _, err := os.Stat(v.repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path does not exist: %s", v.repoPath)
	}

	// Parse all resources into the graph
	if v.verbose {
		fmt.Printf("Parsing resources...\n")
	}

	graph, err := v.parser.ParseAllResources()
	if err != nil {
		return fmt.Errorf("failed to parse resources: %w", err)
	}
	v.graph = graph

	if v.verbose {
		fmt.Printf("Found %d resources in %d files\n", len(graph.Resources), len(graph.Files))
	}

	// Create validation context (for future use)
	_ = context.NewValidationContext(graph, v.config, v.repoPath, v.verbose)

	// For now, use the old validators until we refactor them
	// TODO: Refactor validators to use the new GraphValidator interface

	// Initialize validators
	var deprecatedAPIValidator validators.ValidatorInterface
	if v.yamlPath != "" {
		deprecatedAPIValidator = validators.NewDeprecatedAPIValidatorWithYAML(v.repoPath, v.yamlPath)
	} else {
		deprecatedAPIValidator = validators.NewDeprecatedAPIValidator(v.repoPath)
	}

	validatorList := []validators.ValidatorInterface{
		validators.NewFluxKustomizationValidator(v.repoPath),
		validators.NewKubernetesKustomizationValidator(v.repoPath),
		validators.NewOrphanedResourceValidator(v.repoPath),
		deprecatedAPIValidator,
	}

	// Run all validators
	for _, validator := range validatorList {
		if v.verbose {
			fmt.Printf("Running validator: %s\n", validator.Name())
		}

		results, err := validator.Validate()
		if err != nil {
			return fmt.Errorf("validator %s failed: %w", validator.Name(), err)
		}

		v.results = append(v.results, results...)
	}

	// Print results
	v.printResults()

	// Check if there are any errors
	hasErrors := false
	for _, result := range v.results {
		if result.Severity == "error" {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed with errors")
	}

	return nil
}

// GenerateChart generates a dependency chart in the specified format
func (v *Validator) GenerateChart(format string, outputFile string) error {
	if v.verbose {
		fmt.Printf("Generating dependency chart...\n")
	}

	// Parse all resources into the graph
	graph, err := v.parser.ParseAllResources()
	if err != nil {
		return fmt.Errorf("failed to parse resources: %w", err)
	}

	if v.verbose {
		fmt.Printf("Found %d resources in %d files\n", len(graph.Resources), len(graph.Files))
	}

	// Create validation context
	ctx := context.NewValidationContext(graph, v.config, v.repoPath, v.verbose)

	// Generate the chart
	chart, err := ctx.GenerateDependencyChart(format)
	if err != nil {
		return fmt.Errorf("failed to generate chart: %w", err)
	}

	// Output the chart
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(chart), 0644)
		if err != nil {
			return fmt.Errorf("failed to write chart to file %s: %w", outputFile, err)
		}
		if v.verbose {
			fmt.Printf("Chart written to: %s\n", outputFile)
		}
	} else {
		fmt.Println(chart)
	}

	return nil
}

// GenerateChartForEntryPoint generates a dependency chart for a specific entry point
func (v *Validator) GenerateChartForEntryPoint(format string, outputFile string, entryPointName string) error {
	if v.verbose {
		fmt.Printf("Generating dependency chart for entry point: %s\n", entryPointName)
	}

	// Parse all resources into the graph
	graph, err := v.parser.ParseAllResources()
	if err != nil {
		return fmt.Errorf("failed to parse resources: %w", err)
	}

	if v.verbose {
		fmt.Printf("Found %d resources in %d files\n", len(graph.Resources), len(graph.Files))
	}

	// Create validation context
	ctx := context.NewValidationContext(graph, v.config, v.repoPath, v.verbose)

	// Find the specific entry point
	entryPoints := ctx.FindEntryPoints()
	var targetEntryPoint *parser.ParsedResource
	for _, ep := range entryPoints {
		if ep.Name == entryPointName {
			targetEntryPoint = ep
			break
		}
	}

	if targetEntryPoint == nil {
		return fmt.Errorf("entry point '%s' not found. Available entry points: %v",
			entryPointName, getEntryPointNames(entryPoints))
	}

	// Generate the chart for this entry point
	chart, err := ctx.GenerateDependencyChartForEntryPoint(targetEntryPoint, format)
	if err != nil {
		return fmt.Errorf("failed to generate chart: %w", err)
	}

	// Output the chart
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(chart), 0644)
		if err != nil {
			return fmt.Errorf("failed to write chart to file %s: %w", outputFile, err)
		}
		if v.verbose {
			fmt.Printf("Chart written to: %s\n", outputFile)
		}
	} else {
		fmt.Println(chart)
	}

	return nil
}

// getEntryPointNames returns a slice of entry point names
func getEntryPointNames(entryPoints []*parser.ParsedResource) []string {
	names := make([]string, len(entryPoints))
	for i, ep := range entryPoints {
		names[i] = ep.Name
	}
	return names
}

func (v *Validator) printResults() {
	if len(v.results) == 0 {
		fmt.Println("✅ All validations passed!")
		return
	}

	fmt.Printf("\n📋 Validation Results (%d issues found):\n\n", len(v.results))

	for _, result := range v.results {
		icon := getSeverityIcon(result.Severity)
		fmt.Printf("%s [%s] %s", icon, strings.ToUpper(result.Severity), result.Message)

		if result.File != "" {
			fmt.Printf(" (File: %s", result.File)
			if result.Line > 0 {
				fmt.Printf(":%d", result.Line)
			}
			fmt.Printf(")")
		}

		if result.Resource != "" {
			fmt.Printf(" (Resource: %s)", result.Resource)
		}

		fmt.Println()
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "error":
		return "❌"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "📝"
	}
}

func (v *Validator) findYAMLFiles() ([]string, error) {
	var yamlFiles []string

	err := filepath.Walk(v.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and .git
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules") {
			return filepath.SkipDir
		}

		// Check for YAML files
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml")) {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	return yamlFiles, err
}
