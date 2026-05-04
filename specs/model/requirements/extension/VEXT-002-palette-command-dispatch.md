---
id: VEXT-002
realises:
    - ../../../product/extension/EXT-002-palette-command-access.md
implemented_by:
    - ../../use-cases/extension/VEXT-002-palette-command-dispatch.md
---

# Palette Command Dispatch

## Requirement

`Specs: Lint`, `Specs: Lint Links`, `Specs: Lint Style`, `Specs: Lint
Baselines`, `Specs: Doctor`, `Specs: CR Status`, and `Specs: Framework
Update` shall be registered in the VS Code Command Palette. Each command
shall resolve the workspace folder and specs root before invoking the engine
binary in an integrated terminal panel named "Specs". A warning shall be
shown when no workspace folder is open.

## Rationale

Routing all commands through a named integrated terminal keeps engine output
visible alongside edited files and provides a consistent execution context
for all palette operations.

## Verification

- Open a workspace containing `.specs.yaml`.
- Run `Specs: Doctor` from the palette and confirm it executes in the
  integrated terminal at the specs root.
- Run `Specs: Lint Style` and confirm only the style check runs (no link or
  baseline output).
- Close all workspace folders and confirm a warning is shown instead of
  invoking the engine.
