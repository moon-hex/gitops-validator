# Release Notes

## Version 1.4.0 (2025-01-20) - Phase II: Graph-Based Validator Architecture

### Major Architecture Refactoring
- **Graph-Based Validator Architecture**: Complete refactoring of all validators to use the new `GraphValidator` interface
  - All validators now use the parsed resource graph instead of individual file parsing
  - Eliminated duplicate file walking and parsing across validators
  - Improved performance by parsing resources once and reusing across all validators
  - Enhanced maintainability with consistent validator interface

### Validator Refactoring
- **FluxKustomizationValidator**: Now uses parsed resource graph for Flux Kustomization validation
- **KubernetesKustomizationValidator**: Refactored to use graph-based approach with specialized validators
- **KustomizationVersionConsistencyValidator**: Simplified to use parsed resources instead of file walking
- **OrphanedResourceValidator**: Leverages ValidationContext's orphaned resource detection
- **DeprecatedAPIValidator**: Streamlined to use parsed resources from graph
- **FluxPostBuildVariablesValidator**: Refactored to extract variables from parsed content

### Technical Improvements
- **ValidationContext**: Enhanced context provides centralized access to parsed resources and graph operations
- **Resource Graph**: All validators now work with the unified resource graph
- **Eliminated Code Duplication**: Removed redundant file parsing and walking logic across validators
- **Improved Performance**: Single-pass parsing with multi-validator consumption
- **Better Error Handling**: Consistent error handling across all graph-based validators

### Developer Experience
- **Unified Interface**: All validators now implement the same `GraphValidator` interface
- **Easier Testing**: Validators can be tested with mock validation contexts
- **Simplified Maintenance**: Single source of truth for resource parsing
- **Enhanced Extensibility**: New validators can easily leverage the resource graph

### Breaking Changes
- **Validator Interface**: Legacy `ValidatorInterface` is deprecated in favor of `GraphValidator`
- **Constructor Changes**: Some validator constructors have been simplified (removed config parameters where no longer needed)

## Version 1.3.0 (2025-01-20)

### Refactoring & Architecture Improvements
- **Modular Validator Architecture**: Refactored Kubernetes Kustomization validator into specialized, composable validators
  - Split monolithic validator into focused components: Resource, Patch, and Strategic Merge validators
  - Created reusable validation rules that can be easily composed and tested
  - Extracted shared utilities for kustomization parsing and file operations
  - Improved maintainability and testability of validation logic

### Technical Improvements
- **Shared Utilities**: Created `KustomizationParser` and `KustomizationFile` utilities to eliminate code duplication
- **Validation Rules**: Implemented composable validation rules with `ValidationRuleSet` for better organization
- **Specialized Validators**: 
  - `KustomizationResourceValidator`: Handles resource reference validation
  - `KustomizationPatchValidator`: Handles patch reference validation  
  - `KustomizationStrategicMergeValidator`: Handles strategic merge patch validation
- **Backward Compatibility**: Maintained existing API and functionality while improving internal structure

### Developer Experience
- **Better Testability**: Individual components can now be tested in isolation
- **Easier Extension**: Adding new validation rules is now straightforward
- **Cleaner Code**: Single responsibility principle applied throughout validation logic

## Version 1.2.0 (2025-10-08)

### New Features
- **Kustomization Version Consistency Validation**: Added validator to ensure consistent `kustomize.config.k8s.io` apiVersions across dependency trees
  - Prevents mixing v1 with v1beta1 in the same dependency chain
  - Catches version mismatches that can cause kustomize build failures
  - Validates all resource references that point to other kustomizations
  - Helps maintain compatibility across your entire GitOps repository structure

### Configuration
- Added `kustomization-version-consistency` rule to configuration (enabled by default with error severity)
- Updated example configurations to include the new validation rule

### Documentation
- Added detailed Kustomization Version Consistency section to README
- Included explanation of why version consistency matters
- Added test examples demonstrating valid and invalid version combinations

### Upgrade
Binary and bundle available on Releases. The new validator is enabled by default and will catch apiVersion mismatches in your kustomization dependency trees.

---

## Version 1.1.5 (2025-10-08)

### New Features
- **Flux PostBuild Variables Validation**: Added new validator for Flux Kustomization `postBuild.substitute` variable naming
  - Enforces Flux variable naming rules: must start with underscore or letter, followed by letters, digits, or underscores only
  - Catches common mistakes like using dashes (Kubernetes naming convention) in Flux variables
  - Pattern: `^[_a-zA-Z][_a-zA-Z0-9]*$`
  - Helps prevent runtime errors from invalid variable names

### Configuration
- Added `flux-postbuild-variables` rule to configuration (enabled by default with error severity)
- Updated example configurations to include the new validation rule

### Documentation
- Added detailed Flux PostBuild Variables validation section to README
- Included valid/invalid variable name examples
- Added test examples demonstrating correct and incorrect variable usage

### Upgrade
Binary and bundle available on Releases. The new validator is enabled by default and will catch invalid Flux variable names in your `postBuild.substitute` sections.

---

## Version 1.1.4 (2025-09-18)

- removed several Flux apis from warning

## Version 1.1.2 (2025-09-16)

### New
- **CI-friendly output**: Added `--output-format` flag supporting `markdown` and `json` for easy PR comments and automation.

### Docs
- README: usage examples for `--output-format` and updated GitHub Actions workflow to post results as a Markdown table.
- Examples: `examples/validate-gitops.yml` now demonstrates commenting results back to the PR.

### Config Updates
- ESO: mark `external-secrets.io/v1beta1` as removed in ESO `v0.17.0`.
- Added Istio legacy CRDs to deprecated list.
- Corrected Kubernetes deprecation windows (node/discovery/flowcontrol; RBAC v1alpha1) and removed incorrect Prometheus Operator deprecations.

### Upgrade
Binary and bundle available on Releases. For PR comments, run with `--output-format markdown` and pipe to your comment step.

## Version 1.0.8 (2025-09-09)

### Configuration Consolidation
- **Single config file**: Consolidated all configuration into one `data/gitops-validator.yaml` file
- **Integrated deprecated APIs**: Moved deprecated API definitions into main config file
- **Simplified bundling**: Release bundles now contain only one config file

### What's New
- Deprecated APIs are now loaded from the main config file instead of separate YAML
- Cleaner bundle structure with single source of truth
- Eliminated config file duplication and synchronization issues

### Removed
- Separate `data/deprecated-apis.yaml` file
- Duplicate config files from bundle directory

### Upgrade
Download the latest bundle from the [releases page](https://github.com/moon-hex/gitops-validator/releases) - no changes to usage required.

---

## Version 1.0.7 (2025-09-09)

### Configuration
- **Config file ignore**: Added `gitops-validator.yaml` and `deprecated-apis.yaml` to default ignore patterns
- **Cleaner validation**: Config files are no longer reported as orphaned resources

### What's New
- Config files are automatically ignored during validation
- Reduced noise in validation output for configuration files

### Upgrade
Download the latest bundle from the [releases page](https://github.com/moon-hex/gitops-validator/releases) - no changes to usage required.

---

## Version 1.0.6 (2025-09-09)

### Documentation
- **Flux kustomization paths**: Added clarification that Flux kustomization `path` fields must be relative to repository root
- **Path resolution guide**: Updated documentation to explain the difference between Flux and Kubernetes kustomization path resolution

### What's New
- Clear documentation on Flux kustomization path requirements
- Better understanding of path resolution contexts

### Upgrade
Download the latest bundle from the [releases page](https://github.com/moon-hex/gitops-validator/releases) - no changes to usage required.

---

## Version 1.0.5 (2025-09-09)

### Improvements
- **Consistent path handling**: All validators now use a unified path normalization system
- **Enhanced Flux support**: Flux kustomization validator now properly handles `./` prefixes in path references
- **Code refactoring**: Created reusable path utility functions for better maintainability

### What's New
- Unified path normalization across Kubernetes and Flux kustomization validators
- Better support for relative paths with `./` prefixes in all validators
- Improved code organization with shared utility functions

### Upgrade
Download the latest bundle from the [releases page](https://github.com/moon-hex/gitops-validator/releases) - no changes to usage required.

---

## Version 1.0.4 (2025-09-09)

### Bug Fixes
- **Fixed bundle structure**: Config files now extract to `data/` directory instead of `bundle/data/`
- **Enhanced path validation**: Added support for `./` prefixes in kustomization resource and patch references

### What's New
- Validator now correctly handles paths like `./namespace.yml` and `./patches/production-patch.yaml`
- Improved bundle extraction process
- Better error handling for missing files
