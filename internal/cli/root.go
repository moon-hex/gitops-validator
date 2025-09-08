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

var rootCmd = &cobra.Command{
	Use:   "gitops-validator",
	Short: "Validate GitOps repositories for Flux and Kubernetes",
	Long: `A comprehensive validation tool for GitOps repositories that checks for:
- Flux Kustomization link integrity
- Kubernetes Kustomization link integrity  
- Orphaned resources not referenced by any Kustomization
- Deprecated Kubernetes API versions
- And more...`,
	RunE: runValidation,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is .gitops-validator.yaml)")
	rootCmd.PersistentFlags().StringVarP(&repoPath, "path", "p", ".", "path to GitOps repository")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&yamlPath, "yaml-path", "", "path to deprecated APIs YAML file (default is data/deprecated-apis.yaml)")
	rootCmd.PersistentFlags().StringVar(&chartFormat, "chart", "", "generate dependency chart (mermaid, tree, json)")
	rootCmd.PersistentFlags().StringVar(&chartOutput, "chart-output", "", "output file for dependency chart (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&chartEntryPoint, "chart-entrypoint", "", "generate chart for specific entry point only")

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

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}
}

func runValidation(cmd *cobra.Command, args []string) error {
	path := viper.GetString("path")
	verbose := viper.GetBool("verbose")
	yamlPath := viper.GetString("yaml-path")
	chartFormat := viper.GetString("chart")
	chartOutput := viper.GetString("chart-output")
	chartEntryPoint := viper.GetString("chart-entrypoint")

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

func Execute() error {
	return rootCmd.Execute()
}
