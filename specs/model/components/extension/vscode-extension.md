---
requirements:
    - ../../requirements/workspace/WS-003-repo-local-engine-integration.md
---

# Vscode Extension

## Responsibilities

Wrap the engine for day-to-day VS Code use and keep local development aligned
with the repo-local engine binary and framework layout.

## Key Paths

- `extension/src/engine.ts`
- `extension/src/commands.ts`
- `extension/package.json`
- `docs/development.md`

## Failure Modes

The extension can resolve the wrong binary, drift from the repo-local build
path, or surface stale assumptions about how a host repository is laid out.
