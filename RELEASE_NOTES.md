# Release Notes

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
