# Guided Scaffold Wizard

| Field       | Value                                                                    |
| ----------- | ------------------------------------------------------------------------ |
| Status      | Draft                                                                    |
| Stakeholder | Spec author, stakeholder                                                 |
| Source      | [Commands](../../../docs/commands.md), [Actors](../../../docs/actors.md) |
| Realised By | —                                                                        |

## Summary

Authors and stakeholders need a step-by-step guided prompt inside VS Code
for creating new spec artifacts so they can scaffold requirements, features,
components, APIs, and services without memorising path conventions or flag
syntax.

## User Value

- Authors who rarely scaffold new artifacts do not need to recall the exact
  `specs scaffold` flag combinations — the palette command prompts for each
  required input.
- Input is validated inline before the command runs, so typos in paths or
  missing required fields are caught before the engine is invoked.
- Stakeholders can scaffold product requirements inside a change request by
  answering a small number of prompts, without learning the CR directory
  structure.

## Acceptance Signal

`Specs: Scaffold Requirement`, `Specs: Scaffold Feature`, `Specs: Scaffold
Component`, `Specs: Scaffold API`, and `Specs: Scaffold Service` are
available in the Command Palette. Each command shows a validated input box
for the relative path. The constructed `specs scaffold` invocation is run
in the integrated terminal and the newly created file is opened in the
editor on success.
