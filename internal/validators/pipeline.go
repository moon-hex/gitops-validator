package validators

import (
	"fmt"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/context"
	"github.com/moon-hex/gitops-validator/internal/types"
)

// ValidationPipeline represents a configurable validation execution pipeline
type ValidationPipeline struct {
	Name        string
	Description string
	Stages      []PipelineStage
	Parallel    bool
}

// PipelineStage represents a stage in the validation pipeline
type PipelineStage struct {
	Name        string
	Description string
	Validators  []string // Validator names to run in this stage
	Parallel    bool     // Whether to run validators in this stage in parallel
	Required    bool     // Whether this stage must succeed for the pipeline to continue
	Condition   string   // Optional condition for running this stage
}

// PipelineExecutor executes validation pipelines
type PipelineExecutor struct {
	validators map[string]GraphValidator
	verbose    bool
}

// NewPipelineExecutor creates a new pipeline executor
func NewPipelineExecutor(validators map[string]GraphValidator, verbose bool) *PipelineExecutor {
	return &PipelineExecutor{
		validators: validators,
		verbose:    verbose,
	}
}

// ExecutePipeline executes a validation pipeline
func (pe *PipelineExecutor) ExecutePipeline(pipeline *ValidationPipeline, ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var allResults []types.ValidationResult

	if pe.verbose {
		fmt.Printf("Executing pipeline: %s\n", pipeline.Name)
		if pipeline.Description != "" {
			fmt.Printf("Description: %s\n", pipeline.Description)
		}
	}

	for stageIndex, stage := range pipeline.Stages {
		if pe.verbose {
			fmt.Printf("Executing stage %d: %s\n", stageIndex+1, stage.Name)
		}

		stageResults, err := pe.executeStage(&stage, ctx)
		if err != nil {
			if stage.Required {
				return allResults, fmt.Errorf("required stage '%s' failed: %w", stage.Name, err)
			}

			// Add stage failure as a validation result
			allResults = append(allResults, types.ValidationResult{
				Type:     "pipeline-stage-error",
				Severity: "error",
				Message:  fmt.Sprintf("Stage '%s' failed: %s", stage.Name, err.Error()),
			})

			if pe.verbose {
				fmt.Printf("Stage '%s' failed (non-required): %v\n", stage.Name, err)
			}
		} else {
			allResults = append(allResults, stageResults...)

			if pe.verbose {
				fmt.Printf("Stage '%s' completed with %d results\n", stage.Name, len(stageResults))
			}
		}
	}

	return allResults, nil
}

// executeStage executes a single pipeline stage
func (pe *PipelineExecutor) executeStage(stage *PipelineStage, ctx *context.ValidationContext) ([]types.ValidationResult, error) {
	var stageResults []types.ValidationResult

	// Check if stage should be executed based on condition
	if stage.Condition != "" {
		if !pe.evaluateCondition(stage.Condition, ctx) {
			if pe.verbose {
				fmt.Printf("Skipping stage '%s' due to condition: %s\n", stage.Name, stage.Condition)
			}
			return stageResults, nil
		}
	}

	// Collect validators for this stage
	var stageValidators []GraphValidator
	for _, validatorName := range stage.Validators {
		if validator, exists := pe.validators[validatorName]; exists {
			stageValidators = append(stageValidators, validator)
		} else {
			return stageResults, fmt.Errorf("validator '%s' not found", validatorName)
		}
	}

	if len(stageValidators) == 0 {
		return stageResults, fmt.Errorf("no validators found for stage '%s'", stage.Name)
	}

	// Execute validators in this stage
	if stage.Parallel {
		stageResults = pe.executeValidatorsParallel(stageValidators, ctx)
	} else {
		stageResults = pe.executeValidatorsSequential(stageValidators, ctx)
	}

	return stageResults, nil
}

// executeValidatorsSequential runs validators sequentially
func (pe *PipelineExecutor) executeValidatorsSequential(validators []GraphValidator, ctx *context.ValidationContext) []types.ValidationResult {
	var results []types.ValidationResult

	for _, validator := range validators {
		if pe.verbose {
			fmt.Printf("  Running validator: %s\n", validator.Name())
		}

		validatorResults, err := validator.Validate(ctx)
		if err != nil {
			results = append(results, types.ValidationResult{
				Type:     "validator-error",
				Severity: "error",
				Message:  fmt.Sprintf("Validator %s failed: %s", validator.Name(), err.Error()),
			})
			continue
		}

		results = append(results, validatorResults...)
	}

	return results
}

// executeValidatorsParallel runs validators in parallel
func (pe *PipelineExecutor) executeValidatorsParallel(validators []GraphValidator, ctx *context.ValidationContext) []types.ValidationResult {
	// This would use the same parallel execution logic as the main validator
	// For now, we'll use sequential execution
	// In a full implementation, this would use goroutines and channels
	return pe.executeValidatorsSequential(validators, ctx)
}

// evaluateCondition evaluates a condition string
func (pe *PipelineExecutor) evaluateCondition(condition string, ctx *context.ValidationContext) bool {
	// Simple condition evaluation
	// In a full implementation, this would support more complex conditions

	// Check for resource count conditions
	if strings.HasPrefix(condition, "resource_count >") {
		threshold := strings.TrimSpace(strings.TrimPrefix(condition, "resource_count >"))
		return len(ctx.Graph.Resources) > pe.parseInt(threshold)
	}

	if strings.HasPrefix(condition, "resource_count <") {
		threshold := strings.TrimSpace(strings.TrimPrefix(condition, "resource_count <"))
		return len(ctx.Graph.Resources) < pe.parseInt(threshold)
	}

	// Check for file count conditions
	if strings.HasPrefix(condition, "file_count >") {
		threshold := strings.TrimSpace(strings.TrimPrefix(condition, "file_count >"))
		return len(ctx.Graph.Files) > pe.parseInt(threshold)
	}

	// Default to true if condition is not recognized
	return true
}

// parseInt parses an integer from string, returns 0 on error
func (pe *PipelineExecutor) parseInt(s string) int {
	// Simple integer parsing
	// In a full implementation, this would use strconv.Atoi
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

// Predefined Pipelines

// GetDefaultPipeline returns the default validation pipeline
func GetDefaultPipeline() *ValidationPipeline {
	return &ValidationPipeline{
		Name:        "default",
		Description: "Default validation pipeline with all validators",
		Stages: []PipelineStage{
			{
				Name:        "basic-validation",
				Description: "Basic resource validation",
				Validators:  []string{"flux-kustomization", "kubernetes-kustomization", "deprecated-api"},
				Parallel:    true,
				Required:    true,
			},
			{
				Name:        "advanced-validation",
				Description: "Advanced validation and consistency checks",
				Validators:  []string{"kustomization-version-consistency", "flux-postbuild-variables"},
				Parallel:    true,
				Required:    false,
			},
			{
				Name:        "cleanup-validation",
				Description: "Cleanup and orphaned resource detection",
				Validators:  []string{"orphaned-resource"},
				Parallel:    false,
				Required:    false,
				Condition:   "resource_count > 10", // Only run for larger repositories
			},
		},
		Parallel: true,
	}
}

// GetFastPipeline returns a fast validation pipeline for CI/CD
func GetFastPipeline() *ValidationPipeline {
	return &ValidationPipeline{
		Name:        "fast",
		Description: "Fast validation pipeline for CI/CD",
		Stages: []PipelineStage{
			{
				Name:        "critical-validation",
				Description: "Critical validations only",
				Validators:  []string{"flux-kustomization", "kubernetes-kustomization"},
				Parallel:    true,
				Required:    true,
			},
		},
		Parallel: true,
	}
}

// GetComprehensivePipeline returns a comprehensive validation pipeline
func GetComprehensivePipeline() *ValidationPipeline {
	return &ValidationPipeline{
		Name:        "comprehensive",
		Description: "Comprehensive validation pipeline with all checks",
		Stages: []PipelineStage{
			{
				Name:        "syntax-validation",
				Description: "Syntax and basic structure validation",
				Validators:  []string{"flux-kustomization", "kubernetes-kustomization", "deprecated-api"},
				Parallel:    true,
				Required:    true,
			},
			{
				Name:        "consistency-validation",
				Description: "Consistency and version validation",
				Validators:  []string{"kustomization-version-consistency", "flux-postbuild-variables"},
				Parallel:    true,
				Required:    true,
			},
			{
				Name:        "cleanup-validation",
				Description: "Cleanup and optimization validation",
				Validators:  []string{"orphaned-resource"},
				Parallel:    false,
				Required:    false,
			},
		},
		Parallel: true,
	}
}
