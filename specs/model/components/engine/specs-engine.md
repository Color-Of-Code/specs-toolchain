---
requirements:
    - ../../requirements/workspace/WS-002-config-relative-framework-directory.md
    - ../../requirements/workspace/WS-003-repo-local-engine-integration.md
---

# Specs Engine

## Responsibilities

Own command behavior for config loading, diagnostics, scaffolding, graph
validation, and framework resolution.

## Key Paths

- `engine/internal/config/config.go`
- `engine/cmd/specs/doctor.go`
- `engine/cmd/specs/scaffold.go`
- `engine/cmd/specs/core_integration_test.go`

## Failure Modes

Relative paths can be anchored to the wrong root, local frameworks can be read
as managed frameworks, or scaffold can look in the wrong template directory.
