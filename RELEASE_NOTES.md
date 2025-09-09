# Release Notes

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
