# Doctor Json Contract

| Field          | Value                                                                                                                              |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Status         | Draft                                                                                                                              |
| Realises       | [Repo Local Specs Host](../../../product/toolchain/repo-local-specs-host.md)                                                       |
| Implemented By | [Doctor Json Integration](../../features/workspace/doctor-json-integration.md), [Doctor Json](../../apis/workspace/doctor-json.md) |

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
