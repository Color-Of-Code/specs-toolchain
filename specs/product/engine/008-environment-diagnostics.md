# Environment Diagnostics

| Field       | Value                                                                      |
| ----------- | -------------------------------------------------------------------------- |
| Status      | Draft                                                                      |
| Stakeholder | Spec author, maintainer                                                    |
| Source      | [Install](../../../docs/install.md), [Commands](../../../docs/commands.md) |
| Realised By | —                                                                          |

## Summary

Maintainers and authors need a single command that verifies and reports the
complete engine environment — binary version, resolved paths, framework mode,
and version drift — so that configuration problems can be diagnosed and
communicated without guesswork.

## User Value

- Maintainers can confirm that the engine is resolving the specs root and
  framework directory to the expected locations after a fresh install or
  path change.
- Authors can share the `specs doctor` output as a reproducible environment
  snapshot when reporting problems.
- Version drift between the engine binary and the framework content is
  surfaced explicitly so teams know when to update.

## Acceptance Signal

`specs doctor` prints the binary version, resolved `specs_root`,
`framework_dir`, framework mode, and any detected version drift in a
human-readable table. It exits non-zero when critical configuration is
missing or invalid. A `--json` flag emits the same information as a
machine-readable payload for use in automated environment checks.
