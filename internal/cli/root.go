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
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "gitops-validator",
	Short: "Validate GitOps repositories for Flux and Kubernetes",
	Long: `A comprehensive validation tool for GitOps repositories that checks for:
- Flux Kustomization link integrity
- Kubernetes Kustomization link integrity  
- Orphaned resources not referenced by any Kustomization
- Deprecated Kubernetes API versions
- Dependency chart generation with Mermaid diagrams
- And more...

This tool helps maintain the health and integrity of your GitOps repositories
by identifying common issues before they cause problems in production.

Version: ` + version + `
Commit: ` + commit + `
Built: ` + date,
	RunE: runValidation,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is .gitops-validator.yaml)")
	rootCmd.PersistentFlags().StringVarP(&repoPath, "path", "p", "", "path to GitOps repository (default: current directory)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&yamlPath, "yaml-path", "", "path to deprecated APIs YAML file (default is data/deprecated-apis.yaml)")
	rootCmd.PersistentFlags().StringVar(&chartFormat, "chart", "", "generate dependency chart (mermaid, tree, json)")
	rootCmd.PersistentFlags().StringVar(&chartOutput, "chart-output", "", "output file for dependency chart (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&chartEntryPoint, "chart-entrypoint", "", "generate chart for specific entry point only")

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

	v := validator.NewValidator(path, verbose, yamlPath)

	// If chart generation is requested, handle it separately
	if chartFormat != "" {
		if chartEntryPoint != "" {
			return v.GenerateChartForEntryPoint(chartFormat, chartOutput, chartEntryPoint)
		}
		return v.GenerateChart(chartFormat, chartOutput)
	}

	return v.Validate()
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
