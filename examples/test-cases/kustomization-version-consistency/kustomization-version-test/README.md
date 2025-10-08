# Kustomization Version Consistency Test

This directory demonstrates the kustomization version consistency validation.

## Valid Structure (Passes Validation)

- `base-v1/` - Uses `kustomize.config.k8s.io/v1`
- `common-v1/` - Uses `kustomize.config.k8s.io/v1`
- ✅ Both use the same version, validation passes

## What Gets Caught (Invalid Scenarios)

The validator will catch these version mismatch errors:

### Scenario 1: Mixing v1 and v1beta1

```yaml
# base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../common  # ERROR: common uses v1 but this uses v1beta1
```

```yaml
# common/kustomization.yaml  
apiVersion: kustomize.config.k8s.io/v1
kind: Kustomization
resources:
  - config.yaml
```

**Error Message:**
```
❌ [ERROR] Kustomization apiVersion mismatch: 'base/kustomization.yaml' references '../common' (version: kustomize.config.k8s.io/v1) but uses version kustomize.config.k8s.io/v1beta1
```

### Why This Matters

When a kustomization references another kustomization (via a directory resource), they should use the same apiVersion to avoid compatibility issues with kustomize builds. Mixing versions can cause:
- Build failures
- Unexpected behavior
- Maintenance confusion

### Best Practice

Choose one version and stick with it across your entire repository:
- **Recommended**: `kustomize.config.k8s.io/v1` (stable, current)
- **Legacy**: `kustomize.config.k8s.io/v1beta1` (deprecated)

