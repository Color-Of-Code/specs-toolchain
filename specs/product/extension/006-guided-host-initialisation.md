# Guided Host Initialisation

| Field       | Value                                                                      |
| ----------- | -------------------------------------------------------------------------- |
| Status      | Draft                                                                      |
| Stakeholder | VS Code user, project lead                                                 |
| Source      | [Install](../../../docs/install.md), [Commands](../../../docs/commands.md) |
| Realised By | —                                                                          |

## Summary

New teams and first-time users need a step-by-step wizard inside VS Code
to initialise a specs host in the current workspace so that the initial
setup is guided and correct without requiring familiarity with the CLI flags
or configuration file format.

## User Value

- Project leads setting up a new specs host for the first time complete the
  process through familiar VS Code input boxes rather than reading a
  command reference.
- The wizard validates each input (framework name, model layout choice)
  before invoking the engine, preventing common misconfiguration mistakes.
- Once the wizard completes, the workspace is immediately usable — the
  status bar updates, and palette commands resolve to the newly created
  specs root.

## Acceptance Signal

`Specs: Init Host` (palette command `specs.bootstrap`) opens a multi-step
wizard that collects the framework name (or URL), model layout preference,
and optional VS Code tasks generation. On completion it invokes `specs init`
with the appropriate flags in the integrated terminal. If a specs host
already exists in the workspace the wizard warns the user before
proceeding. The wizard can be cancelled at any step without modifying the
workspace.
