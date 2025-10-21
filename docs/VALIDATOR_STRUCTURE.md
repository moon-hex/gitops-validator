# Validator Structure Overview

This document explains the clean separation between individual validation checks and common code/helpers in the gitops-validator project.

## 📁 Directory Structure

```
internal/validators/
├── common/                    # Common utilities and base classes
│   ├── base_validator.go     # Base validator with common functionality
│   └── checks.go             # Reusable validation check functions
├── checks/                   # Individual validation checks (focused)
│   ├── flux_kustomization_checks.go
│   ├── kustomization_checks.go
│   ├── deprecated_api_checks.go
│   ├── orphaned_resource_checks.go
│   ├── flux_postbuild_checks.go
│   └── version_consistency_checks.go
├── flux_kustomization.go     # Main validator (orchestrator)
├── kubernetes_kustomization.go
├── deprecated_api.go
├── orphaned_resources.go
├── flux_postbuild_variables.go
├── kustomization_version_consistency.go
├── interface.go              # GraphValidator interface
├── pipeline.go               # Validation pipelines
└── path_utils.go             # Path utility functions
```

## 🎯 Clear Separation of Concerns

### 1. **Individual Validation Checks** (`internal/validators/checks/`)

Each check file contains **focused, single-purpose validation logic**:

- **`flux_kustomization_checks.go`**: Path and source reference validation for Flux Kustomizations
- **`kustomization_checks.go`**: Resource, patch, and strategic merge validation for Kubernetes Kustomizations  
- **`deprecated_api_checks.go`**: API version deprecation checking
- **`orphaned_resource_checks.go`**: Orphaned file detection
- **`flux_postbuild_checks.go`**: PostBuild variable naming validation
- **`version_consistency_checks.go`**: API version consistency across dependencies

**Characteristics:**
- ✅ **Single responsibility**: Each check does one specific thing
- ✅ **Pure functions**: No side effects, easy to test
- ✅ **Reusable**: Can be used by multiple validators
- ✅ **Focused**: Clear input/output, no complex state

### 2. **Common Code & Helpers** (`internal/validators/common/`)

Shared utilities and base functionality:

- **`base_validator.go`**: Common validator functionality, result creation helpers
- **`checks.go`**: Reusable validation functions (file existence, path validation, etc.)

**Characteristics:**
- ✅ **DRY principle**: No code duplication
- ✅ **Consistent**: Standardized result creation and error handling
- ✅ **Testable**: Isolated functions easy to unit test
- ✅ **Maintainable**: Changes in one place affect all validators

### 3. **Main Validators** (`internal/validators/*.go`)

Orchestrators that coordinate multiple checks:

- **`flux_kustomization.go`**: Coordinates Flux Kustomization checks
- **`orphaned_resources.go`**: Coordinates orphaned resource detection
- **`deprecated_api.go`**: Coordinates deprecated API checking

**Characteristics:**
- ✅ **Orchestration**: Coordinates multiple checks
- ✅ **Graph integration**: Works with the resource graph
- ✅ **Result aggregation**: Combines results from multiple checks
- ✅ **Clean interface**: Simple, focused public API

## 🔄 Data Flow

```
1. Main Validator receives ValidationContext
2. Main Validator calls specific Check functions
3. Check functions use Common utilities
4. Check functions return ValidationResults
5. Main Validator aggregates and returns results
```

## 📝 Example: Flux Kustomization Validation

### Before (Mixed Concerns):
```go
// All validation logic mixed with helper code in one file
func (v *FluxKustomizationValidator) Validate(ctx *context.ValidationContext) {
    // Path validation logic mixed with file I/O
    // Source validation logic mixed with graph traversal
    // Result creation mixed with validation logic
}
```

### After (Clean Separation):

**Main Validator** (`flux_kustomization.go`):
```go
func (v *FluxKustomizationValidator) Validate(ctx *context.ValidationContext) {
    for _, kustomization := range fluxKustomizations {
        // Delegate to focused checks
        pathResults := checks.FluxKustomizationPathCheck(kustomization, ctx)
        sourceResults := checks.FluxKustomizationSourceCheck(kustomization, ctx)
        results = append(results, pathResults..., sourceResults...)
    }
}
```

**Individual Check** (`checks/flux_kustomization_checks.go`):
```go
func FluxKustomizationPathCheck(kustomization *parser.ParsedResource, ctx *context.ValidationContext) []types.ValidationResult {
    // Focused path validation logic only
    path, err := common.ExtractStringFromContent(kustomization.Content, "spec", "path")
    if err != nil {
        return []types.ValidationResult{...}
    }
    return common.PathValidationCheck(baseDir, path)
}
```

**Common Utility** (`common/checks.go`):
```go
func PathValidationCheck(baseDir, path string) error {
    // Reusable path validation logic
    return FileExistenceCheck(baseDir, path)
}
```

## 🎯 Benefits of This Structure

### 1. **Maintainability**
- ✅ Easy to find and modify specific validation logic
- ✅ Changes to common utilities automatically benefit all validators
- ✅ Clear separation makes debugging easier

### 2. **Testability**
- ✅ Individual checks can be unit tested in isolation
- ✅ Common utilities can be tested independently
- ✅ Main validators can be tested with mocked checks

### 3. **Reusability**
- ✅ Common utilities can be used across multiple validators
- ✅ Individual checks can be composed in different ways
- ✅ New validators can easily reuse existing checks

### 4. **Readability**
- ✅ Each file has a clear, single purpose
- ✅ Code is self-documenting through clear structure
- ✅ Easy to understand data flow and dependencies

### 5. **Extensibility**
- ✅ New checks can be added without modifying existing code
- ✅ New validators can be created by composing existing checks
- ✅ Common utilities can be extended for new use cases

## 🚀 Adding New Validation Checks

1. **Create focused check function** in `internal/validators/checks/`
2. **Use common utilities** from `internal/validators/common/`
3. **Integrate into main validator** by calling the check function
4. **Test independently** - each check can be unit tested

This structure ensures that individual validation logic is clean, focused, and easily maintainable while maximizing code reuse and consistency.
