package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for gitops-validator
type Config struct {
	GitOpsValidator GitOpsValidatorConfig `yaml:"gitops-validator"`
}

// GitOpsValidatorConfig contains all configuration options
type GitOpsValidatorConfig struct {
	// Basic settings
	Path    string `yaml:"path"`
	Verbose bool   `yaml:"verbose"`

	// Entry points configuration
	EntryPoints EntryPointsConfig `yaml:"entry-points"`

	// Validation rules
	Rules RulesConfig `yaml:"rules"`

	// Deprecated APIs configuration
	DeprecatedAPIs DeprecatedAPIsConfig `yaml:"deprecated-apis"`

	// Chart configuration
	Chart ChartConfig `yaml:"chart"`

	// Ignore patterns for files/directories
	Ignore IgnoreConfig `yaml:"ignore"`

	// Exit code configuration
	ExitCodes ExitCodeConfig `yaml:"exit-codes"`
}

// EntryPointsConfig defines how to identify entry point resources
type EntryPointsConfig struct {
	Resources  []string `yaml:"resources"`  // Specific resource names
	Namespaces []string `yaml:"namespaces"` // Namespaces to consider
	Types      []string `yaml:"types"`      // Resource types
	Patterns   []string `yaml:"patterns"`   // Glob patterns
}

// RulesConfig defines which validation rules to run
type RulesConfig struct {
	FluxKustomization               RuleConfig `yaml:"flux-kustomization"`
	FluxPostBuildVariables          RuleConfig `yaml:"flux-postbuild-variables"`
	KubernetesKustomization         RuleConfig `yaml:"kubernetes-kustomization"`
	KustomizationVersionConsistency RuleConfig `yaml:"kustomization-version-consistency"`
	OrphanedResources               RuleConfig `yaml:"orphaned-resources"`
	DeprecatedAPIs                  RuleConfig `yaml:"deprecated-apis"`
	DoubleReferences                RuleConfig `yaml:"double-references"`
	CircularDependencies            RuleConfig `yaml:"circular-dependencies"`
}

// RuleConfig defines a single validation rule
type RuleConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Severity string `yaml:"severity"`
}

// DeprecatedAPIsConfig defines deprecated API configuration
type DeprecatedAPIsConfig struct {
	UseEmbedded bool                    `yaml:"use-embedded"`
	CustomAPIs  []DeprecatedAPIInfo     `yaml:"custom-apis"`
	Overrides   map[string]OverrideInfo `yaml:"overrides"`
	Disabled    []string                `yaml:"disabled"`
}

// DeprecatedAPIInfo represents a custom deprecated API
type DeprecatedAPIInfo struct {
	APIVersion       string `yaml:"api_version"`
	DeprecationInfo  string `yaml:"deprecation_info"`
	Severity         string `yaml:"severity"`
	OperatorCategory string `yaml:"operator_category"`
}

// OverrideInfo represents an override for an embedded deprecated API
type OverrideInfo struct {
	Severity string `yaml:"severity"`
}

// ChartConfig defines chart generation settings
type ChartConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Format          string `yaml:"format"`           // mermaid, tree, json
	Output          string `yaml:"output"`           // output file path
	IncludeOrphaned bool   `yaml:"include-orphaned"` // include orphaned resources
	IncludeMetadata bool   `yaml:"include-metadata"` // include resource metadata
}

// IgnoreConfig defines patterns to ignore during validation
type IgnoreConfig struct {
	Directories []string `yaml:"directories"` // Directory patterns to ignore
	Files       []string `yaml:"files"`       // File patterns to ignore
}

// ExitCodeConfig defines when the tool should exit with non-zero codes
type ExitCodeConfig struct {
	FailOnErrors   bool `yaml:"fail-on-errors"`   // Exit with code 1 on errors (default: true)
	FailOnWarnings bool `yaml:"fail-on-warnings"` // Exit with code 2 on warnings (default: false)
	FailOnInfo     bool `yaml:"fail-on-info"`     // Exit with code 3 on info messages (default: false)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		GitOpsValidator: GitOpsValidatorConfig{
			Path:    ".",
			Verbose: false,
			EntryPoints: EntryPointsConfig{
				Namespaces: []string{"flux-system"},
				Types:      []string{"flux-kustomization", "helm-release", "git-repository"},
				Patterns:   []string{"clusters/*", "apps/*", "infrastructure/*"},
			},
			Rules: RulesConfig{
				FluxKustomization:               RuleConfig{Enabled: true, Severity: "error"},
				FluxPostBuildVariables:          RuleConfig{Enabled: true, Severity: "error"},
				KubernetesKustomization:         RuleConfig{Enabled: true, Severity: "error"},
				KustomizationVersionConsistency: RuleConfig{Enabled: true, Severity: "error"},
				OrphanedResources:               RuleConfig{Enabled: true, Severity: "warning"},
				DeprecatedAPIs:                  RuleConfig{Enabled: true, Severity: "warning"},
				DoubleReferences:                RuleConfig{Enabled: true, Severity: "warning"},
				CircularDependencies:            RuleConfig{Enabled: true, Severity: "error"},
			},
			DeprecatedAPIs: DeprecatedAPIsConfig{
				UseEmbedded: true,
				CustomAPIs:  []DeprecatedAPIInfo{},
				Overrides:   make(map[string]OverrideInfo),
				Disabled:    []string{},
			},
			Chart: ChartConfig{
				Enabled:         false,
				Format:          "mermaid",
				Output:          "",
				IncludeOrphaned: true,
				IncludeMetadata: true,
			},
			Ignore: IgnoreConfig{
				Directories: []string{
					".git/**",
					".github/**",
					".gitlab-ci/**",
					".circleci/**",
					".azure-pipelines/**",
					"node_modules/**",
					"vendor/**",
					"tmp/**",
					"temp/**",
					"build/**",
					"dist/**",
					"bin/**",
					"examples/test-cases/**",
				},
				Files: []string{
					"*.log",
					"*.tmp",
					"*.temp",
					".DS_Store",
					"Thumbs.db",
				},
			},
			ExitCodes: ExitCodeConfig{
				FailOnErrors:   true,  // Default: fail on errors
				FailOnWarnings: false, // Default: don't fail on warnings
				FailOnInfo:     false, // Default: don't fail on info
			},
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Merge with defaults for any missing fields
	defaultConfig := DefaultConfig()

	// Merge ignore patterns
	if len(config.GitOpsValidator.Ignore.Directories) == 0 {
		config.GitOpsValidator.Ignore.Directories = defaultConfig.GitOpsValidator.Ignore.Directories
	}
	if len(config.GitOpsValidator.Ignore.Files) == 0 {
		config.GitOpsValidator.Ignore.Files = defaultConfig.GitOpsValidator.Ignore.Files
	}

	return &config, nil
}

// ShouldIgnorePath checks if a path should be ignored based on ignore patterns
func (c *Config) ShouldIgnorePath(path string) bool {
	// Normalize path separators to forward slashes for consistent matching
	normalizedPath := filepath.ToSlash(path)

	// Check directory patterns
	for _, pattern := range c.GitOpsValidator.Ignore.Directories {
		// Normalize pattern separators too
		normalizedPattern := filepath.ToSlash(pattern)

		if matched, _ := filepath.Match(normalizedPattern, normalizedPath); matched {
			return true
		}
		// Also check if the path is within an ignored directory
		if strings.Contains(normalizedPattern, "**") {
			dirPattern := strings.TrimSuffix(normalizedPattern, "/**")
			if strings.HasPrefix(normalizedPath, dirPattern+"/") {
				return true
			}
		}
	}

	// Check file patterns
	for _, pattern := range c.GitOpsValidator.Ignore.Files {
		// Normalize pattern separators
		normalizedPattern := filepath.ToSlash(pattern)

		// Try matching against the full path first
		if matched, _ := filepath.Match(normalizedPattern, normalizedPath); matched {
			return true
		}

		// Also try matching against just the filename for simple patterns
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate entry point patterns
	for _, pattern := range c.GitOpsValidator.EntryPoints.Patterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid entry point pattern: %s", pattern)
		}
	}

	// Validate deprecated API versions
	for _, api := range c.GitOpsValidator.DeprecatedAPIs.CustomAPIs {
		if api.APIVersion == "" {
			return fmt.Errorf("deprecated API version cannot be empty")
		}
		if api.Severity != "error" && api.Severity != "warning" && api.Severity != "info" {
			return fmt.Errorf("invalid severity '%s' for API '%s', must be error, warning, or info", api.Severity, api.APIVersion)
		}
	}

	// Validate rule severities
	rules := []RuleConfig{
		c.GitOpsValidator.Rules.FluxKustomization,
		c.GitOpsValidator.Rules.FluxPostBuildVariables,
		c.GitOpsValidator.Rules.KubernetesKustomization,
		c.GitOpsValidator.Rules.KustomizationVersionConsistency,
		c.GitOpsValidator.Rules.OrphanedResources,
		c.GitOpsValidator.Rules.DeprecatedAPIs,
		c.GitOpsValidator.Rules.DoubleReferences,
		c.GitOpsValidator.Rules.CircularDependencies,
	}

	for _, rule := range rules {
		if rule.Enabled && rule.Severity != "error" && rule.Severity != "warning" && rule.Severity != "info" {
			return fmt.Errorf("invalid rule severity '%s', must be error, warning, or info", rule.Severity)
		}
	}

	return nil
}

// GetEntryPointTypes returns the resource types that should be considered entry points
func (c *Config) GetEntryPointTypes() []string {
	return c.GitOpsValidator.EntryPoints.Types
}

// GetEntryPointNamespaces returns the namespaces that should be considered entry points
func (c *Config) GetEntryPointNamespaces() []string {
	return c.GitOpsValidator.EntryPoints.Namespaces
}

// GetEntryPointPatterns returns the patterns that should be considered entry points
func (c *Config) GetEntryPointPatterns() []string {
	return c.GitOpsValidator.EntryPoints.Patterns
}

// GetEntryPointResources returns the specific resources that should be considered entry points
func (c *Config) GetEntryPointResources() []string {
	return c.GitOpsValidator.EntryPoints.Resources
}

// IsRuleEnabled checks if a specific rule is enabled
func (c *Config) IsRuleEnabled(ruleName string) bool {
	switch ruleName {
	case "flux-kustomization":
		return c.GitOpsValidator.Rules.FluxKustomization.Enabled
	case "flux-postbuild-variables":
		return c.GitOpsValidator.Rules.FluxPostBuildVariables.Enabled
	case "kubernetes-kustomization":
		return c.GitOpsValidator.Rules.KubernetesKustomization.Enabled
	case "kustomization-version-consistency":
		return c.GitOpsValidator.Rules.KustomizationVersionConsistency.Enabled
	case "orphaned-resources":
		return c.GitOpsValidator.Rules.OrphanedResources.Enabled
	case "deprecated-apis":
		return c.GitOpsValidator.Rules.DeprecatedAPIs.Enabled
	case "double-references":
		return c.GitOpsValidator.Rules.DoubleReferences.Enabled
	case "circular-dependencies":
		return c.GitOpsValidator.Rules.CircularDependencies.Enabled
	default:
		return false
	}
}

// GetRuleSeverity returns the severity for a specific rule
func (c *Config) GetRuleSeverity(ruleName string) string {
	switch ruleName {
	case "flux-kustomization":
		return c.GitOpsValidator.Rules.FluxKustomization.Severity
	case "flux-postbuild-variables":
		return c.GitOpsValidator.Rules.FluxPostBuildVariables.Severity
	case "kubernetes-kustomization":
		return c.GitOpsValidator.Rules.KubernetesKustomization.Severity
	case "kustomization-version-consistency":
		return c.GitOpsValidator.Rules.KustomizationVersionConsistency.Severity
	case "orphaned-resources":
		return c.GitOpsValidator.Rules.OrphanedResources.Severity
	case "deprecated-apis":
		return c.GitOpsValidator.Rules.DeprecatedAPIs.Severity
	case "double-references":
		return c.GitOpsValidator.Rules.DoubleReferences.Severity
	case "circular-dependencies":
		return c.GitOpsValidator.Rules.CircularDependencies.Severity
	default:
		return "warning"
	}
}
