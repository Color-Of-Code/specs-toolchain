# Doctor Json Integration

| Field        | Value                                                                        |
| ------------ | ---------------------------------------------------------------------------- |
| Status       | Draft                                                                        |
| Requirements | [Doctor Json Contract](../../requirements/workspace/WS-005-doctor-json-contract.md) |

## Workflow

Expose machine-readable workspace diagnostics so the VS Code extension and
other tooling can consume resolved repo-local path and compatibility data.

## Engine Surface

- `specs doctor --json` emits the stable diagnostics payload.
- The JSON output reflects the same resolved paths as human doctor output.

## VS Code Surface

- The extension can consume repo-local engine diagnostics without parsing
  human prose.
- Local tooling can verify `specs_root`, `framework_dir`, and compatibility
  fields directly.

## Validation

Run `./bin/specs doctor --json` and confirm the payload exposes resolved local
path fields and compatibility metadata.
