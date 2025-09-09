# Flux Kustomization Path Requirements

## Overview

Flux kustomization resources use **repository root** as the base directory for path resolution, which is different from Kubernetes kustomization files that use the kustomization file's directory as the base.

## Path Resolution

### Flux Kustomization
- **Base directory**: Repository root (where `gitops-validator` is run)
- **Path field**: Must be relative to repository root
- **Example**: `./clusters/production` resolves to `repo-root/clusters/production`

### Kubernetes Kustomization
- **Base directory**: Directory containing the kustomization file
- **Resource paths**: Relative to kustomization file location
- **Example**: `./namespace.yml` resolves to `kustomization-dir/namespace.yml`

## Examples

### ✅ Correct Flux Kustomization
```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: production
  namespace: flux-system
spec:
  path: ./clusters/production  # ✅ Relative to repo root
  sourceRef:
    kind: GitRepository
    name: gitops-repo
  interval: 1m
  prune: true
```

### ❌ Incorrect Flux Kustomization
```yaml
spec:
  path: ./examples/sample-gitops/clusters/production  # ❌ Wrong - includes repo path
```

### ✅ Correct Kubernetes Kustomization
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ./namespace.yml           # ✅ Relative to kustomization file
  - ../../apps/backend        # ✅ Relative to kustomization file
  - ./patches/production-patch.yaml  # ✅ Relative to kustomization file
```

## Validation

The GitOps validator correctly handles both path resolution contexts:

- **Flux kustomization validator**: Uses repository root as base directory
- **Kubernetes kustomization validator**: Uses kustomization file directory as base directory

## Common Issues

1. **Double path inclusion**: Including the repository path in Flux kustomization paths
2. **Wrong base directory**: Using Flux-style paths in Kubernetes kustomizations
3. **Missing `./` prefix**: Both validators handle `./` prefixes correctly

## Best Practices

1. **Flux kustomizations**: Use paths relative to repository root
2. **Kubernetes kustomizations**: Use paths relative to kustomization file
3. **Consistent naming**: Use `./` prefix for clarity (optional but recommended)
4. **Test validation**: Run `gitops-validator` to verify path references
