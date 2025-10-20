package cli

import (
	"fmt"
	"os"

	"github.com/moon-hex/gitops-validator/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile      string
	repoPath        string
	verbose         bool
	yamlPath        string
	chartFormat     string
	chartOutput     string
	chartEntryPoint string
)

var (
	version = "1.3.0"
	commit  = "main"
	date    = "2025-01-20"
)

var rootCmd = &cobra.Command{
	Use:   "gitops-validator",
	Short: "Validate GitOps repositories for Flux and Kubernetes",
	Long: `A comprehensive validation tool for GitOps repositories that checks for:
- Flux Kustomization link integrity
- Flux postBuild variable naming (no dashes allowed)
- Kubernetes Kustomization link integrity
- Kustomization apiVersion consistency (no v1/v1beta1 mismatches)
- Orphaned resources not referenced by any Kustomization
- Deprecated Kubernetes API versions
- Dependency chart generation with Mermaid diagrams
- Configurable error handling and exit codes for CI/CD integration
- And more...

This tool helps maintain the health and integrity of your GitOps repositories
by identifying common issues before they cause problems in production.

Exit Codes:
- 0: Validation passed (or configured to not fail on found issues)
- 1: Validation failed with errors (default behavior)
- 2: Validation failed with warnings (when --fail-on-warnings is used)
- 3: Validation failed with info messages (when --fail-on-info is used)

Examples:
  gitops-validator --path . --verbose                    # Default: fail on errors only
  gitops-validator --path . --no-fail-on-errors          # Don't fail on errors
  gitops-validator --path . --fail-on-warnings           # Also fail on warnings
  gitops-validator --path . --chart mermaid              # Generate dependency chart
  gitops-validator --path . --chart mermaid --chart-output deps.md  # Save chart to file
  gitops-validator --path . --output-format markdown     # GitHub-friendly table output
  gitops-validator --path . --output-format json         # JSON for machine consumption

Version: ` + version + `
Commit: ` + commit + `
Built: ` + date,
	RunE: runValidation,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is data/gitops-validator.yaml)")
	rootCmd.PersistentFlags().StringVarP(&repoPath, "path", "p", "", "path to GitOps repository (default: current directory)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&yamlPath, "yaml-path", "", "path to deprecated APIs YAML file (default is data/deprecated-apis.yaml)")
	rootCmd.PersistentFlags().StringVar(&chartFormat, "chart", "", "generate dependency chart (mermaid, tree, json)")
	rootCmd.PersistentFlags().StringVar(&chartOutput, "chart-output", "", "output file for dependency chart (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&chartEntryPoint, "chart-entrypoint", "", "generate chart for specific entry point only")

	// Exit code configuration flags
	rootCmd.PersistentFlags().Bool("fail-on-errors", true, "exit with code 1 on errors (default: true)")
	rootCmd.PersistentFlags().Bool("no-fail-on-errors", false, "don't exit with code 1 on errors (useful for testing)")
	rootCmd.PersistentFlags().Bool("fail-on-warnings", false, "exit with code 2 on warnings (default: false)")
	rootCmd.PersistentFlags().Bool("no-fail-on-warnings", false, "don't exit with code 2 on warnings")
	rootCmd.PersistentFlags().Bool("fail-on-info", false, "exit with code 3 on info messages (default: false)")
	rootCmd.PersistentFlags().Bool("no-fail-on-info", false, "don't exit with code 3 on info messages")

	// Output formatting for CI (markdown/json)
	rootCmd.PersistentFlags().String("output-format", "", "output format for results: markdown, json, or default")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gitops-validator version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	})

	viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("yaml-path", rootCmd.PersistentFlags().Lookup("yaml-path"))
	viper.BindPFlag("chart", rootCmd.PersistentFlags().Lookup("chart"))
	viper.BindPFlag("chart-output", rootCmd.PersistentFlags().Lookup("chart-output"))
	viper.BindPFlag("chart-entrypoint", rootCmd.PersistentFlags().Lookup("chart-entrypoint"))
	viper.BindPFlag("fail-on-errors", rootCmd.PersistentFlags().Lookup("fail-on-errors"))
	viper.BindPFlag("no-fail-on-errors", rootCmd.PersistentFlags().Lookup("no-fail-on-errors"))
	viper.BindPFlag("fail-on-warnings", rootCmd.PersistentFlags().Lookup("fail-on-warnings"))
	viper.BindPFlag("no-fail-on-warnings", rootCmd.PersistentFlags().Lookup("no-fail-on-warnings"))
	viper.BindPFlag("fail-on-info", rootCmd.PersistentFlags().Lookup("fail-on-info"))
	viper.BindPFlag("no-fail-on-info", rootCmd.PersistentFlags().Lookup("no-fail-on-info"))
	viper.BindPFlag("output-format", rootCmd.PersistentFlags().Lookup("output-format"))
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(".gitops-validator")
		viper.SetConfigType("yaml")
	}

	// viper.AutomaticEnv() // Disabled to prevent PATH environment variable conflict

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}
}

func runValidation(cmd *cobra.Command, args []string) error {
	// Check if we should show help BEFORE doing any validation
	chartFormat := viper.GetString("chart")
	verbose := viper.GetBool("verbose")
	yamlPath := viper.GetString("yaml-path")
	chartOutput := viper.GetString("chart-output")
	chartEntryPoint := viper.GetString("chart-entrypoint")
	outputFormat := viper.GetString("output-format")

	// Check if path was explicitly set by user (not just default)
	pathExplicitlySet := cmd.Flags().Changed("path")

	// If no validation or chart generation is requested, show help
	if chartFormat == "" && !verbose && yamlPath == "" && chartOutput == "" && chartEntryPoint == "" && !pathExplicitlySet {
		return cmd.Help()
	}

	// Only proceed with validation if we have a valid request
	path := viper.GetString("path")
	if path == "" {
		path = "."
	}

	if verbose {
		fmt.Printf("Validating GitOps repository at: %s\n", path)
		if yamlPath != "" {
			fmt.Printf("Using deprecated APIs YAML: %s\n", yamlPath)
		}
		if chartFormat != "" {
			if chartEntryPoint != "" {
				fmt.Printf("Generating dependency chart for entry point '%s' in %s format\n", chartEntryPoint, chartFormat)
			} else {
				fmt.Printf("Generating dependency chart in %s format\n", chartFormat)
			}
		}
	}

	// Get exit code configuration from flags
	failOnErrors := viper.GetBool("fail-on-errors") && !viper.GetBool("no-fail-on-errors")
	failOnWarnings := viper.GetBool("fail-on-warnings") && !viper.GetBool("no-fail-on-warnings")
	failOnInfo := viper.GetBool("fail-on-info") && !viper.GetBool("no-fail-on-info")

	v := validator.NewValidatorWithExitCodes(path, verbose, yamlPath, failOnErrors, failOnWarnings, failOnInfo)
	if outputFormat != "" {
		v.SetOutputFormat(outputFormat)
	}

	// If chart generation is requested, handle it separately
	if chartFormat != "" {
		var err error
		if chartEntryPoint != "" {
			err = v.GenerateChartForEntryPoint(chartFormat, chartOutput, chartEntryPoint)
		} else {
			err = v.GenerateChart(chartFormat, chartOutput)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
		return nil // This line is unreachable but required by Go compiler
	}

	// Handle validation and exit with appropriate code
	exitCode, err := v.Validate()
	if err != nil {
		// For parsing errors, show the error and exit
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// Always exit with the validation result code (0 for success, 1/2/3 for different failure types)
	// This prevents Cobra from showing help text since we never return an error from RunE
	os.Exit(exitCode)
	return nil // This line is unreachable but required by Go compiler
}

// hasValidationFlags checks if any validation-related flags are set
func hasValidationFlags() bool {
	// Check if any flags were explicitly set by the user
	return viper.GetBool("verbose") ||
		viper.GetString("yaml-path") != "" ||
		viper.GetString("chart") != "" ||
		viper.GetString("chart-output") != "" ||
		viper.GetString("chart-entrypoint") != "" ||
		viper.GetString("config") != "" ||
		viper.IsSet("path") // Check if path was explicitly set
}

func Execute() error {
	return rootCmd.Execute()
}
