# Test Cases

This directory contains test cases for the gitops-validator, including intentionally invalid configurations to verify that the validators correctly detect errors.

**⚠️ Important:** This directory is excluded from validation in CI/CD workflows to prevent false failures. The examples here are meant for manual testing and documentation purposes.

## Directory Structure

- `flux-postbuild-test.yaml` - Examples of valid and invalid Flux postBuild variable names
- `kustomization-version-consistency/` - Examples of version consistency checks
- `patches-strategic-merge-file-support/` - Examples of patchesStrategicMerge with file object format support

## Usage

To test these examples manually:

```bash
# Test a specific directory (will show errors as expected)
./gitops-validator --path examples/test-cases/kustomization-version-consistency

# To run validation on the entire repository (excludes test-cases by default)
./gitops-validator --path .
```

## Purpose

These test cases serve as:
1. **Documentation** - Examples showing what validators catch
2. **Manual Testing** - Verify validators work correctly
3. **Development** - Test cases for adding new validation rules

The test cases are excluded from automated validation via the ignore patterns in `.gitops-validator.yaml`:

```yaml
ignore:
  directories:
    - "examples/test-cases/**"
```

