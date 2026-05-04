---
id: VEXT-002
requirements:
    - ../../requirements/extension/VEXT-002-palette-command-dispatch.md
---

# Palette Command Dispatch

## Workflow

Register each user-facing specs command in the VS Code command palette, resolve
the specs root of the open workspace, and invoke the engine in an integrated
terminal rooted at that directory.

## VS Code Surface

- `Specs: Lint`, `Specs: Lint Links`, `Specs: Lint Style`, `Specs: Lint
  Baselines`, `Specs: Doctor`, `Specs: CR Status`, and `Specs: Framework
  Update` are all registered palette commands.
- Each command resolves the workspace folder and specs root before invoking
  the engine.
- Output appears in an integrated terminal panel named "Specs".
- Commands show a warning if no workspace folder is open.

## Validation

Open a workspace containing `.specs.yaml`. Run `Specs: Doctor` from the
palette and confirm it executes in the integrated terminal at the specs root.
Run `Specs: Lint Style` and confirm only the style check runs.
