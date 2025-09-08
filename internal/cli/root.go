package cli

import (
	"fmt"
	"os"

	"github.com/moon-hex/gitops-validator/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	repoPath   string
	verbose    bool
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

	viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
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

	if verbose {
		fmt.Printf("Validating GitOps repository at: %s\n", path)
	}

	v := validator.NewValidator(path, verbose)
	return v.Validate()
}

func Execute() error {
	return rootCmd.Execute()
}
