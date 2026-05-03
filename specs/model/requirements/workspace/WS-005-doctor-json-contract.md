# Doctor Json Contract

| Field          | Value                                                                                                                                             |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID             | WS-005                                                                                                                                            |
| Status         | Draft                                                                                                                                             |
| Realises       | [Repo Local Specs Host](../../../product/engine/ENG-009-repo-local-specs-host.md)                                                                 |
| Implemented By | [JSON Diagnostics Integration](../../features/doctor/DOC-002-json-diagnostics-integration.md), [Doctor Json](../../apis/workspace/doctor-json.md) |

## Requirement

The machine-readable doctor output shall expose the resolved repo-local specs
and framework paths, mode information, and compatibility fields needed by the
VS Code extension and local tooling.

## Rationale

The extension and other tooling cannot depend on human-oriented doctor text if
they need a stable contract for local host diagnostics.

## Verification

- Run `./bin/specs doctor --json`.
- Confirm the payload includes the resolved `specs_root`, `framework_dir`,
  `framework_mode`, and compatibility information.
- Confirm the reported framework dir matches the repo-local `framework/`
  directory.
