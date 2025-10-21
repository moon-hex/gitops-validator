package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	// new: optional output format ("", "markdown", "json")
	outputFormat string
	// Phase III: parallel validation
	parallel bool
	// Phase III: validation pipelines
	pipeline    *validators.ValidationPipeline
	usePipeline bool
	// Phase III: result aggregation
	aggregationOptions *types.AggregationOptions
	useAggregation     bool
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
		repoPath:           repoPath,
		verbose:            verbose,
		yamlPath:           yamlPath,
		config:             cfg,
		parser:             parser.NewResourceParser(repoPath, cfg),
		results:            make([]types.ValidationResult, 0),
		outputFormat:       "",
		parallel:           false, // Default to sequential for backward compatibility
		pipeline:           nil,   // Pipeline disabled by default
		usePipeline:        false,
		aggregationOptions: nil, // Aggregation disabled by default
		useAggregation:     false,
	}
}

// NewValidatorWithParallel creates a validator with parallel execution enabled
func NewValidatorWithParallel(repoPath string, verbose bool, yamlPath string, parallel bool) *Validator {
	v := NewValidator(repoPath, verbose, yamlPath)
	v.parallel = parallel
	return v
}

// SetParallel enables or disables parallel validation
func (v *Validator) SetParallel(parallel bool) {
	v.parallel = parallel
}

// SetPipeline sets the validation pipeline
func (v *Validator) SetPipeline(pipeline *validators.ValidationPipeline) {
	v.pipeline = pipeline
	v.usePipeline = pipeline != nil
}

// SetPipelineByName sets a predefined pipeline by name
func (v *Validator) SetPipelineByName(pipelineName string) error {
	switch pipelineName {
	case "default":
		v.SetPipeline(validators.GetDefaultPipeline())
	case "fast":
		v.SetPipeline(validators.GetFastPipeline())
	case "comprehensive":
		v.SetPipeline(validators.GetComprehensivePipeline())
	default:
		return fmt.Errorf("unknown pipeline: %s", pipelineName)
	}
	return nil
}

// SetAggregationOptions sets the result aggregation options
func (v *Validator) SetAggregationOptions(options *types.AggregationOptions) {
	v.aggregationOptions = options
	v.useAggregation = options != nil
}

// SetAggregationPreset sets a predefined aggregation preset
func (v *Validator) SetAggregationPreset(preset string) {
	switch preset {
	case "errors-only":
		v.SetAggregationOptions(&types.AggregationOptions{
			ShowOnlyErrors: true,
			SortBy:         "severity",
			SortOrder:      "desc",
			IncludeStats:   true,
		})
	case "warnings-only":
		v.SetAggregationOptions(&types.AggregationOptions{
			ShowOnlyWarnings: true,
			SortBy:           "type",
			SortOrder:        "asc",
			IncludeStats:     true,
		})
	case "summary":
		v.SetAggregationOptions(&types.AggregationOptions{
			IncludeStats: true,
			Limit:        50, // Show top 50 results
			SortBy:       "severity",
			SortOrder:    "desc",
		})
	case "grouped":
		v.SetAggregationOptions(&types.AggregationOptions{
			GroupBy:      "type",
			IncludeStats: true,
			SortBy:       "type",
			SortOrder:    "asc",
		})
	default:
		// No aggregation
		v.useAggregation = false
		v.aggregationOptions = nil
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

	// Build fast lookup index for large repositories (Phase III)
	if v.verbose {
		fmt.Printf("Building resource index...\n")
	}
	if err := graph.BuildIndex(); err != nil {
		return 1, fmt.Errorf("failed to build resource index: %w", err)
	}

	if v.verbose {
		stats := graph.Index.GetIndexStats()
		fmt.Printf("Index built: %d resources, %d Flux Kustomizations, %d Kubernetes Kustomizations\n",
			stats["total_resources"], stats["flux_kustomizations"], stats["kubernetes_kustomizations"])
	}

	// Create validation context
	validationContext := context.NewValidationContext(graph, v.config, v.repoPath, v.verbose)

	// Run validation using pipeline or traditional approach
	if v.usePipeline {
		v.runValidationWithPipeline(validationContext)
	} else {
		// Initialize graph-based validators
		validatorList := []validators.GraphValidator{
			validators.NewFluxKustomizationValidator(v.repoPath),
			validators.NewKubernetesKustomizationValidator(v.repoPath),
			validators.NewKustomizationVersionConsistencyValidator(v.repoPath),
			validators.NewOrphanedResourceValidator(v.repoPath),
			validators.NewDeprecatedAPIValidator(v.repoPath),
			validators.NewFluxPostBuildVariablesValidator(v.repoPath),
		}

		// Run all validators with context (parallel or sequential)
		if v.parallel {
			v.runValidatorsParallel(validatorList, validationContext)
		} else {
			v.runValidatorsSequential(validatorList, validationContext)
		}
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

// runValidatorsSequential runs validators sequentially (legacy behavior)
func (v *Validator) runValidatorsSequential(validatorList []validators.GraphValidator, validationContext *context.ValidationContext) {
	for _, validator := range validatorList {
		if v.verbose {
			fmt.Printf("Running validator: %s\n", validator.Name())
		}

		results, err := validator.Validate(validationContext)
		if err != nil {
			// Add error as validation result instead of failing completely
			v.results = append(v.results, types.ValidationResult{
				Type:     "validator-error",
				Severity: "error",
				Message:  fmt.Sprintf("Validator %s failed: %s", validator.Name(), err.Error()),
			})
			continue
		}

		v.results = append(v.results, results...)
	}
}

// runValidatorsParallel runs validators in parallel for better performance
func (v *Validator) runValidatorsParallel(validatorList []validators.GraphValidator, validationContext *context.ValidationContext) {
	if v.verbose {
		fmt.Printf("Running %d validators in parallel...\n", len(validatorList))
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a channel to collect results
	resultChan := make(chan []types.ValidationResult, len(validatorList))
	errorChan := make(chan error, len(validatorList))

	// Start all validators in parallel
	for _, validator := range validatorList {
		wg.Add(1)
		go func(validator validators.GraphValidator) {
			defer wg.Done()

			if v.verbose {
				mu.Lock()
				fmt.Printf("Starting validator: %s\n", validator.Name())
				mu.Unlock()
			}

			results, err := validator.Validate(validationContext)
			if err != nil {
				errorChan <- fmt.Errorf("validator %s failed: %w", validator.Name(), err)
				return
			}

			resultChan <- results
		}(validator)
	}

	// Wait for all validators to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	for {
		select {
		case results, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				v.results = append(v.results, results...)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				// Add error as validation result instead of failing completely
				v.results = append(v.results, types.ValidationResult{
					Type:     "validator-error",
					Severity: "error",
					Message:  err.Error(),
				})
			}
		}

		// Exit when both channels are closed
		if resultChan == nil && errorChan == nil {
			break
		}
	}

	if v.verbose {
		fmt.Printf("Parallel validation completed. Found %d total results.\n", len(v.results))
	}
}

// runValidationWithPipeline runs validation using a pipeline
func (v *Validator) runValidationWithPipeline(validationContext *context.ValidationContext) {
	if v.verbose {
		fmt.Printf("Running validation with pipeline: %s\n", v.pipeline.Name)
	}

	// Create validator registry
	validatorRegistry := map[string]validators.GraphValidator{
		"flux-kustomization":                validators.NewFluxKustomizationValidator(v.repoPath),
		"kubernetes-kustomization":          validators.NewKubernetesKustomizationValidator(v.repoPath),
		"kustomization-version-consistency": validators.NewKustomizationVersionConsistencyValidator(v.repoPath),
		"orphaned-resource":                 validators.NewOrphanedResourceValidator(v.repoPath),
		"deprecated-api":                    validators.NewDeprecatedAPIValidator(v.repoPath),
		"flux-postbuild-variables":          validators.NewFluxPostBuildVariablesValidator(v.repoPath),
	}

	// Create pipeline executor
	executor := validators.NewPipelineExecutor(validatorRegistry, v.verbose)

	// Execute pipeline
	results, err := executor.ExecutePipeline(v.pipeline, validationContext)
	if err != nil {
		v.results = append(v.results, types.ValidationResult{
			Type:     "pipeline-error",
			Severity: "error",
			Message:  fmt.Sprintf("Pipeline execution failed: %s", err.Error()),
		})
	} else {
		v.results = append(v.results, results...)
	}
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

	// Apply result aggregation if enabled
	var resultsToPrint []types.ValidationResult
	if v.useAggregation && v.aggregationOptions != nil {
		aggregator := types.NewResultAggregator(v.results)
		aggregated := aggregator.Aggregate(*v.aggregationOptions)
		resultsToPrint = aggregated.Results

		// Print summary if requested
		if v.aggregationOptions.IncludeStats {
			fmt.Println(aggregated.GetSummary())
			fmt.Println()
		}
	} else {
		resultsToPrint = v.results
	}

	// Default human-readable output
	if v.outputFormat == "" {
		fmt.Printf("\n📋 Validation Results (%d issues found):\n\n", len(resultsToPrint))
		for _, result := range resultsToPrint {
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
		return
	}

	// Markdown table output
	if v.outputFormat == "markdown" || v.outputFormat == "md" {
		fmt.Println("## GitOps Validator Results")
		fmt.Println()
		fmt.Printf("%d issues found\n\n", len(resultsToPrint))
		fmt.Println("| Severity | Type | Message | File | Line | Resource |")
		fmt.Println("|---|---|---|---|---:|---|")
		for _, r := range resultsToPrint {
			msg := strings.ReplaceAll(r.Message, "|", "\\|")
			fmt.Printf("| %s | %s | %s | %s | %d | %s |\n",
				strings.ToUpper(r.Severity), r.Type, msg, r.File, r.Line, r.Resource)
		}
		return
	}

	// JSON output
	if v.outputFormat == "json" {
		b, err := json.MarshalIndent(resultsToPrint, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting JSON output: %v\n", err)
			return
		}
		fmt.Println(string(b))
		return
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

// SetOutputFormat configures how results are printed: "markdown", "json" or default human output
func (v *Validator) SetOutputFormat(format string) {
	f := strings.ToLower(strings.TrimSpace(format))
	switch f {
	case "markdown", "md", "json":
		v.outputFormat = f
	default:
		v.outputFormat = ""
	}
}
