# Specs Engine

| Field        | Value                                                                                                                                                                                                      |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Status       | Draft                                                                                                                                                                                                      |
| Requirements | [Config Relative Framework Directory](../../requirements/workspace/config-relative-framework-directory.md), [Repo Local Engine Integration](../../requirements/workspace/repo-local-engine-integration.md) |
| Baseline     | —                                                                                                                                                                                                          |

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
