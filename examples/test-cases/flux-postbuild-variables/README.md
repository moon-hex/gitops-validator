# Flux PostBuild Variables Test

This directory demonstrates Flux postBuild variable naming validation.

## Variable Naming Rules

Flux variable names must follow the pattern: `^[_a-zA-Z][_a-zA-Z0-9]*$`

- Must start with underscore (`_`) or letter (`a-z`, `A-Z`)  
- Can only contain letters, digits, and underscores
- **No dashes allowed** (common mistake from Kubernetes naming)

## Valid Examples ✅

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: my-app
  namespace: flux-system
spec:
  interval: 10m
  path: ./apps/production
  sourceRef:
    kind: GitRepository
    name: flux-system
  postBuild:
    substitute:
      CLUSTER_NAME: "production"      # ✅ Valid
      ENV: "prod"                     # ✅ Valid
      _private_var: "secret"          # ✅ Valid (starts with _)
      region123: "us-east-1"          # ✅ Valid (contains digits)
      API_ENDPOINT: "https://api.com" # ✅ Valid
```

## Invalid Examples ❌

The following will be caught by the validator:

```yaml
postBuild:
  substitute:
    cluster-name: "staging"        # ❌ ERROR: contains dash
    my-env-var: "staging"          # ❌ ERROR: contains dashes
    region.name: "us-west-1"       # ❌ ERROR: contains dot
    123var: "value"                # ❌ ERROR: starts with digit
    my@var: "value"                # ❌ ERROR: contains special char
```

**Error Message:**
```
❌ [ERROR] Invalid Flux variable name 'cluster-name': must start with underscore or letter, followed by letters, digits, or underscores only (no dashes allowed). Pattern: ^[_a-zA-Z][_a-zA-Z0-9]*$
```

## Why This Matters

Using invalid variable names will cause Flux to fail at runtime when trying to substitute variables. This validator catches these errors before deployment.

## Common Mistake

People often use Kubernetes resource naming conventions (which allow dashes) for Flux variables. Remember:
- **Kubernetes resources**: `cluster-name` ✅ (dashes allowed)
- **Flux variables**: `CLUSTER_NAME` ✅ (no dashes, use underscores instead)

