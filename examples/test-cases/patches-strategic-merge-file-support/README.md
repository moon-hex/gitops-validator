# patchesStrategicMerge Test Cases

This directory contains test cases for `patchesStrategicMerge` validation.

## Supported Format

Only string format is supported:

```yaml
patchesStrategicMerge:
  - patch1.yaml
  - ./patches/patch2.yaml
```

## Test Cases

- `valid-example.yaml` - Valid kustomization with multiple patches
- `missing-file.yaml` - Invalid kustomization with missing patch file (should fail validation)

## Expected Behavior

The validator should:
1. ✅ Accept string format patches
2. ✅ Validate file existence
3. ✅ Allow multiple patches targeting the same resource (no duplicate detection)
4. ✅ Apply path resolution logic
5. ❌ Fail validation for missing files
