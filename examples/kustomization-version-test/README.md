# Kustomization Version Consistency Test

This directory demonstrates the kustomization version consistency validation.

## Valid Structure

- `base-v1/` - Uses `kustomize.config.k8s.io/v1`
- `common-v1/` - Uses `kustomize.config.k8s.io/v1`
- ✅ Both use the same version, validation passes

## Invalid Structure (Version Mismatch)

- `base-v1beta1-MISMATCH/` - Uses `kustomize.config.k8s.io/v1beta1`
- References `../common-v1` which uses `kustomize.config.k8s.io/v1`
- ❌ Version mismatch detected, validation fails

## Expected Validation Result

When running the validator on this directory:

```bash
❌ [ERROR] Kustomization apiVersion mismatch: 'base-v1beta1-MISMATCH/kustomization.yaml' references '../common-v1' (version: kustomize.config.k8s.io/v1) but uses version kustomize.config.k8s.io/v1beta1
```

This error prevents build issues where different kustomization versions are mixed in the same dependency tree.

