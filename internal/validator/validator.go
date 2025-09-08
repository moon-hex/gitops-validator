package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moon-hex/gitops-validator/internal/types"
	"github.com/moon-hex/gitops-validator/internal/validators"
)

type Validator struct {
	repoPath string
	verbose  bool
	results  []types.ValidationResult
}

func NewValidator(repoPath string, verbose bool) *Validator {
	return &Validator{
		repoPath: repoPath,
		verbose:  verbose,
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

	// Initialize validators
	validators := []validators.ValidatorInterface{
		validators.NewFluxKustomizationValidator(v.repoPath),
		validators.NewKubernetesKustomizationValidator(v.repoPath),
		validators.NewOrphanedResourceValidator(v.repoPath),
		validators.NewDeprecatedAPIValidator(v.repoPath),
	}

	// Run all validators
	for _, validator := range validators {
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
