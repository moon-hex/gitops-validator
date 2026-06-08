# Release Notes

## Version 1.7.0 (upcoming) - Bug Fixes: External SourceRef Name Collision & Missing BaseDir in Kustomization Validators

### Bug Fixes

#### Bug 1 — External `sourceRef` lookup hits wrong resource due to name collision (Flux Kustomization Validator)

**Root cause:** `isExternalSourceRef` called `ctx.Graph.GetResource(sourceRefName)`, which looks up resources by the map key `"<namespace>/<name>"` or `"<name>"`. A namespaced `GitRepository` (e.g. `namespace: papla, name: papla`) has key `"papla/papla"`, but a `Namespace` resource also named `papla` has key `"papla"`. `GetResource("papla")` returned the `Namespace` instead of the `GitRepository`. The `Namespace` has no `spec.url`, so `isExternalSourceRef` returned `false` and path validation ran against paths that only exist in the remote repo.

**Fix (`internal/validators/checks/flux_kustomization_checks.go`):** Replaced `GetResource(sourceRefName)` with a new `findSourceByKindAndName` helper that searches only among resources of the expected kind (`GitRepository` / `OCIRepository`), eliminating any collision with same-named cluster-scoped resources.

**Impact:** Eliminates false-positive errors when a Flux Kustomization's external `GitRepository` shares its name with a `Namespace` or other cluster-scoped resource in the same repo.

#### Bug 2 — Missing `BaseDir` in kustomization validators causes all relative path checks to fail (Kustomization Validators)

**Root cause:** `KustomizationResourceValidator`, `KustomizationPatchValidator`, and `KustomizationStrategicMergeValidator` each constructed a `KustomizationFile` struct without setting `BaseDir`. `ValidateFileExists` resolves relative paths against `BaseDir`; when it is empty the path resolves against the process working directory instead of the kustomization file's own directory, causing every relative resource/patch reference to appear missing.

**Fix (`internal/validators/kustomization_resource_validator.go`, `kustomization_patch_validator.go`, `kustomization_strategic_merge_validator.go`):** Added `BaseDir: filepath.Dir(kustomization.File)` to the `KustomizationFile` struct literal in each validator's `Validate` method.

**Impact:** Eliminates false-positive errors for all relative path references in `resources:`, `patches:`, and `patchesStrategicMerge:` sections of Kubernetes kustomization files.

### Upgrade
Binary and bundle available on Releases. No configuration changes required — this is a pure bug-fix release.

---

## Version 1.6.0 (upcoming) - Bug Fixes: Unnamed Kustomization Files & External SourceRef Paths

### Bug Fixes

#### Bug 1 — Parser drops unnamed Kubernetes kustomization.yaml files (Orphaned Resource Validator)

**Root cause:** `parser.go` `parseResourceNode` required `metadata.name` to accept a resource (`apiVersion == "" || kind == "" || name == ""`). Kubernetes `kustomization.yaml` files (`apiVersion: kustomize.config.k8s.io/…`, `kind: Kustomization`) never carry `metadata.name` — the file path is their identity. Every Kubernetes kustomization.yaml was silently dropped from the resource graph, breaking dependency traversal and causing all files reachable only through those kustomization.yaml files to be reported as orphaned.

**Telltale symptom:** Index line `0 Kubernetes Kustomizations` despite many `kustomization.yaml` files in the repo.

**Fix (`internal/parser/parser.go`):** The `name == ""` guard is removed from the skip condition. When `apiVersion` and `kind` are present but `name` is empty, the file path is used as a synthetic name. The resource is correctly registered in `g.Files` and reachable by `findResourceByPath`.

**Impact:** Eliminates false-positive orphaned-resource warnings in repositories that use shared `common-resources/` patterns or unnamed intermediate kustomization.yaml files.

#### Bug 2 — Flux `spec.path` validated against local filesystem for external `sourceRef` (Flux Kustomization Validator)

**Root cause:** `FluxKustomizationPathCheck` always resolved `spec.path` against the local repo root. In Flux, `spec.path` is relative to the **source** repository named in `sourceRef`. When `sourceRef` points to an external `GitRepository` with a remote URL, the path does not exist locally and the check always fails.

**Fix (`internal/validators/checks/flux_kustomization_checks.go`):** Added `isExternalSourceRef` helper. Before the filesystem check, the validator now:
1. Reads `spec.sourceRef.kind` — proceeds only for `GitRepository` / `OCIRepository`
2. Looks up the referenced source resource in the graph
3. If the source has a remote `spec.url` (`https://`, `http://`, `ssh://`, `git@`, `git://`), or if the source cannot be found locally, path validation is skipped

**Impact:** Eliminates false-positive errors for Flux Kustomizations that deploy from external application repositories.

### Upgrade
Binary and bundle available on Releases. No configuration changes required — this is a pure bug-fix release.

---

## Version 1.5.1 (2026-05-26) - Bug Fix: False Positives in Resource Resolution

### Bug Fixes
- **Fixed false positives in Kustomization resource path resolution**: The resource graph now correctly resolves directory references by looking for `kustomization.yaml` / `kustomization.yml` inside the target directory. Previously, directory-style resource entries (e.g. `resources: [./some-dir]`) were not matched and caused spurious validation errors.
- **Fixed `ReferenceTypeResource` handling**: Kustomization `resources:` entries are now treated as relative paths (relative to the kustomization file), matching actual kustomize behaviour. Previously they fell through to the default unresolved case.
- **Fixed `ResourceTypeKubernetesKustomization` classification**: Resources with `apiVersion: kustomize.config.k8s.io/...` and `kind: Kustomization` are now correctly classified as `ResourceTypeKubernetesKustomization` instead of being left unclassified.

### Upgrade
Binary and bundle available on Releases. No configuration changes required — this is a pure bug-fix release reducing false-positive validation errors.

---

## Version 1.5.0 (2025-10-21) - Phase III: Clean Validator Structure

### Major Refactoring
- **Clean separation of validation checks from common code**: Reorganized `internal/validators/` into a clear three-layer structure:
  - `common/` — shared base validator and reusable check utilities
  - `checks/` — individual, single-purpose validation check functions (one file per validator domain)
  - Top-level validators now act as thin orchestrators, delegating to focused check functions

### New Infrastructure
- **Resource Indexing** (`internal/parser/index.go`): Fast lookup structures for large repositories
- **Validation Pipelines** (`internal/validators/pipeline.go`): Configurable, ordered validation execution
- **Result Aggregation** (`internal/types/aggregation.go`): Advanced filtering and grouping of validation results
- **Parallel Validation**: Validators can now run concurrently for better performance on large repos

### Developer Experience
- Individual checks can be unit-tested in isolation without a full validation context
- New checks can be added without modifying existing orchestrator code
- Each file has a single, clear responsibility

### Documentation
- Added `docs/VALIDATOR_STRUCTURE.md` with architecture overview and guide for adding new checks

### Breaking Changes
- None — all existing validation behaviour and CLI flags are preserved

### Upgrade
Binary and bundle available on Releases. No configuration changes required.

---

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
