# GitOps Validator

A comprehensive validation tool for GitOps repositories that checks for common issues in Flux and Kubernetes configurations.

## Features

- **Flux Kustomization Validation**: Validates Flux Kustomization resources for broken path and source references
- **Kubernetes Kustomization Validation**: Validates kustomization.yaml files for broken resource and patch references
- **Orphaned Resource Detection**: Identifies YAML files that are not referenced by any kustomization
- **Deprecated API Detection**: Warns about usage of deprecated Kubernetes API versions
- **GitHub Actions Integration**: Ready-to-use workflow for CI/CD pipelines

## Installation

### From Source

```bash
git clone https://github.com/your-org/gitops-validator.git
cd gitops-validator
go build -o gitops-validator ./main.go
```

### Using Go Install

```bash
go install github.com/your-org/gitops-validator@latest
```

## Usage

### Basic Usage

```bash
# Validate current directory
./gitops-validator

# Validate specific directory
./gitops-validator --path /path/to/gitops-repo

# Verbose output
./gitops-validator --verbose

# Use custom config file
./gitops-validator --config my-config.yaml
```

### Configuration

Create a `.gitops-validator.yaml` file in your repository root:

```yaml
# Repository path to validate
path: "."

# Verbose output
verbose: false

# Validation rules
rules:
  flux-kustomization:
    enabled: true
    severity: "error"
  kubernetes-kustomization:
    enabled: true
    severity: "error"
  orphaned-resources:
    enabled: true
    severity: "warning"
  deprecated-apis:
    enabled: true
    severity: "warning"

# Ignore patterns
ignore:
  - "**/.git/**"
  - "**/node_modules/**"
  - "**/.github/**"

# Custom deprecated APIs
custom-deprecated-apis:
  "mycompany.com/v1alpha1": "Deprecated in v1.0, will be removed in v2.0"
```

## GitHub Actions Integration

Add this workflow to your `.github/workflows/` directory:

```yaml
name: GitOps Validation

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        
    - name: Install dependencies
      run: go mod download
      
    - name: Build validator
      run: go build -o gitops-validator ./main.go
      
    - name: Run validation
      run: ./gitops-validator --path . --verbose
```

## Validation Rules

### Flux Kustomization Validation

Validates Flux Kustomization resources for:
- Missing or invalid `path` references
- Missing or invalid `sourceRef.name` references
- Broken file system paths

### Kubernetes Kustomization Validation

Validates kustomization.yaml files for:
- Broken `resources` references
- Broken `patches` references
- Broken `patchesStrategicMerge` references

### Orphaned Resource Detection

Identifies YAML files that:
- Are not referenced by any kustomization
- Are not entry points (kustomization files or Flux Kustomization resources)

### Deprecated API Detection

Warns about usage of deprecated Kubernetes API versions:
- `extensions/v1beta1` (removed in v1.22)
- `apps/v1beta1` (removed in v1.16)
- `apps/v1beta2` (removed in v1.16)
- And many more...

## Output Format

The validator provides clear, actionable output:

```
📋 Validation Results (3 issues found):

❌ [ERROR] Invalid path reference: path 'apps/backend' does not exist (File: flux/kustomizations/backend.yaml:15) (Resource: backend)
⚠️ [WARNING] File 'unused-config.yaml' is not referenced by any kustomization and is not an entry point (File: config/unused-config.yaml)
⚠️ [WARNING] Using deprecated API version 'extensions/v1beta1' for resource 'Deployment' 'my-app' - Deprecated in v1.16, removed in v1.22 (File: apps/my-app.yaml:3)
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.