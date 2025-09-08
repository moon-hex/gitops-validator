# GitOps Validator Exit Codes

This document explains the exit code behavior of the GitOps Validator tool and how to configure it for different CI/CD scenarios.

## Exit Code Reference

| Exit Code | Meaning | When It Occurs |
|-----------|---------|----------------|
| **0** | Success | Validation passed OR configured to not fail on found issues |
| **1** | Errors Found | Critical issues detected (default behavior) |
| **2** | Warnings Found | Non-critical issues detected (when `--fail-on-warnings` is used) |
| **3** | Info Found | Informational messages detected (when `--fail-on-info` is used) |

## CLI Flags

### Basic Usage
```bash
# Default behavior (fail on errors only)
gitops-validator --path .

# Don't fail on errors (useful for testing)
gitops-validator --path . --no-fail-on-errors

# Fail on warnings too (strict mode)
gitops-validator --path . --fail-on-warnings

# Fail on all severity levels (very strict)
gitops-validator --path . --fail-on-errors --fail-on-warnings --fail-on-info
```

### Flag Reference
- `--fail-on-errors`: Exit with code 1 on errors (default: true)
- `--no-fail-on-errors`: Don't exit with code 1 on errors (useful for testing)
- `--fail-on-warnings`: Exit with code 2 on warnings (default: false)
- `--no-fail-on-warnings`: Don't exit with code 2 on warnings
- `--fail-on-info`: Exit with code 3 on info messages (default: false)
- `--no-fail-on-info`: Don't exit with code 3 on info messages

## Configuration File

Configure exit codes in `.gitops-validator.yaml`:

```yaml
exit-codes:
  fail-on-errors: true      # Exit with code 1 on errors
  fail-on-warnings: false  # Exit with code 2 on warnings
  fail-on-info: false      # Exit with code 3 on info messages
```

## GitHub Actions Examples

### Basic Validation (Fail on Errors Only)
```yaml
- name: Validate GitOps
  run: gitops-validator --path . --verbose
  # Will fail the job if errors are found (exit code 1)
```

### Testing Mode (Don't Fail on Errors)
```yaml
- name: Test GitOps Validation
  run: gitops-validator --path . --verbose --no-fail-on-errors
  # Will not fail the job even if errors are found (exit code 0)
```

### Strict Validation (Fail on Warnings Too)
```yaml
- name: Strict GitOps Validation
  run: gitops-validator --path . --verbose --fail-on-warnings
  # Will fail the job on errors (exit code 1) or warnings (exit code 2)
```

### Advanced Error Handling
```yaml
- name: Validate GitOps Repository
  id: validation
  run: gitops-validator --path . --verbose
  continue-on-error: true  # Capture result without failing immediately

- name: Generate Chart (Only if validation passes)
  if: steps.validation.outcome == 'success'
  run: gitops-validator --path . --chart mermaid --chart-output chart.md

- name: Fail Job on Validation Errors
  if: steps.validation.outcome == 'failure'
  run: |
    echo "❌ GitOps validation failed. Please fix the issues before merging."
    exit 1
```

## Common Use Cases

### Development Environment
```bash
# Don't fail on errors during development
gitops-validator --path . --no-fail-on-errors
```

### CI/CD Pipeline (Default)
```bash
# Fail on errors in CI/CD
gitops-validator --path . --verbose
```

### Quality Gate (Strict)
```bash
# Fail on warnings too for high-quality standards
gitops-validator --path . --fail-on-warnings
```

### Testing/Validation Only
```bash
# Just check for issues without failing
gitops-validator --path . --no-fail-on-errors --no-fail-on-warnings
```

## Troubleshooting

### Exit Code 1 (Errors)
- Check for broken Flux Kustomization references
- Verify Kubernetes Kustomization resource paths
- Fix any missing files or invalid configurations

### Exit Code 2 (Warnings)
- Review orphaned resources
- Check for deprecated API usage
- Consider addressing warnings for better GitOps health

### Exit Code 3 (Info)
- Review informational messages
- Consider enabling info-level validation for comprehensive checks