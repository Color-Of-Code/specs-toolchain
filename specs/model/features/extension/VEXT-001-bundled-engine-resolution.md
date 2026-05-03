# Bundled Engine Resolution

| Field        | Value    |
| ------------ | -------- |
| ID           | VEXT-001 |
| Status       | Draft    |
| Requirements | —        |

## Workflow

Resolve the correct `specs` binary from one of three sources in priority
order: an explicit `specs.enginePath` setting, the `specs` binary found on
`PATH` when `specs.useGlobalBinary` is true, or the platform-specific binary
bundled inside the extension.

## Engine Surface

- The engine itself is not involved; resolution happens in `extension/src/engine.ts`.

## VS Code Surface

- `specs.enginePath` (string setting) — absolute path to an explicit binary.
- `specs.useGlobalBinary` (boolean setting) — prefer `specs` on `PATH`.
- Fallback: bundled binary in the extension's `bin/` directory.
- The resolved path is logged and used for every command invocation.

## Validation

Install the extension without any settings. Confirm commands invoke the
bundled binary. Set `specs.useGlobalBinary = true` and confirm the PATH
binary is preferred. Set `specs.enginePath` and confirm it takes precedence
over both.
