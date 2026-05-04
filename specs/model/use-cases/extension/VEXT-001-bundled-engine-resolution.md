---
id: VEXT-001
status: Draft
requirements:
    - ../../requirements/extension/VEXT-001-bundled-engine-resolution-priority.md
---

# Bundled Engine Resolution

## Workflow

Resolve the correct `specs` binary from four sources in priority order: an
explicit `specs.enginePath` setting, the `specs` binary found on `PATH` when
`specs.useGlobalBinary` is true, a workspace-local `bin/specs` build when one
is present, or the platform-specific binary bundled inside the extension.

## Engine Surface

- The engine itself is not involved; resolution happens in `extension/src/engine.ts`.

## VS Code Surface

- `specs.enginePath` (string setting) — absolute path to an explicit binary.
- `specs.useGlobalBinary` (boolean setting) — prefer `specs` on `PATH`.
- Next fallback: workspace-local binary in the workspace `bin/` directory.
- Final fallback: bundled binary in the extension's `bin/` directory.
- The resolved path is logged and used for every command invocation.

## Validation

Install the extension without any settings. Confirm commands invoke the
bundled binary when the workspace does not contain `bin/specs`. Open a
workspace that does contain `bin/specs` and confirm commands prefer that
local binary. Set `specs.useGlobalBinary = true` and confirm the PATH binary
is preferred. Set `specs.enginePath` and confirm it takes precedence over
all other sources.
