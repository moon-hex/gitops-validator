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

	// Try to load from data/gitops-validator.yaml first, then .gitops-validator.yaml for backward compatibility
	if _, err := os.Stat("data/gitops-validator.yaml"); err == nil {
		if loadedConfig, err := config.LoadConfig("data/gitops-validator.yaml"); err == nil {
			cfg = loadedConfig
		}
	} else if _, err := os.Stat(".gitops-validator.yaml"); err == nil {
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

// NewValidatorWithExitCodes creates a validator with custom exit code configuration
func NewValidatorWithExitCodes(repoPath string, verbose bool, yamlPath string, failOnErrors, failOnWarnings, failOnInfo bool) *Validator {
	v := NewValidator(repoPath, verbose, yamlPath)

	// Override exit code configuration
	v.config.GitOpsValidator.ExitCodes.FailOnErrors = failOnErrors
	v.config.GitOpsValidator.ExitCodes.FailOnWarnings = failOnWarnings
	v.config.GitOpsValidator.ExitCodes.FailOnInfo = failOnInfo

	return v
}

func (v *Validator) Validate() (int, error) {
	if v.verbose {
		fmt.Printf("Starting validation of repository: %s\n", v.repoPath)
	}

	// Check if repository path exists
	if _, err := os.Stat(v.repoPath); os.IsNotExist(err) {
		return 1, fmt.Errorf("repository path does not exist: %s", v.repoPath)
	}

	// Parse all resources into the graph
	if v.verbose {
		fmt.Printf("Parsing resources...\n")
	}

	graph, err := v.parser.ParseAllResources()
	if err != nil {
		return 1, fmt.Errorf("failed to parse resources: %w", err)
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
		validators.NewOrphanedResourceValidatorWithConfig(v.repoPath, v.config),
		deprecatedAPIValidator,
	}

	// Run all validators
	for _, validator := range validatorList {
		if v.verbose {
			fmt.Printf("Running validator: %s\n", validator.Name())
		}

		results, err := validator.Validate()
		if err != nil {
			return 1, fmt.Errorf("validator %s failed: %w", validator.Name(), err)
		}

		v.results = append(v.results, results...)
	}

	// Print results
	v.printResults()

	// Check validation results based on configured exit codes
	hasErrors := false
	hasWarnings := false
	hasInfo := false

	for _, result := range v.results {
		switch result.Severity {
		case "error":
			hasErrors = true
		case "warning":
			hasWarnings = true
		case "info":
			hasInfo = true
		}
	}

	// Return appropriate exit code based on configuration
	if hasErrors && v.config.GitOpsValidator.ExitCodes.FailOnErrors {
		return 1, nil // Exit code 1 for errors, no error returned
	}
	if hasWarnings && v.config.GitOpsValidator.ExitCodes.FailOnWarnings {
		return 2, nil // Exit code 2 for warnings, no error returned
	}
	if hasInfo && v.config.GitOpsValidator.ExitCodes.FailOnInfo {
		return 3, nil // Exit code 3 for info, no error returned
	}

	return 0, nil // Exit code 0 for success, no error returned
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
