# Guided Scaffold Wizard

| Field          | Value                                                                                  |
| -------------- | -------------------------------------------------------------------------------------- |
| ID             | VEXT-005                                                                               |
| Status         | Draft                                                                                  |
| Realises       | [Guided Scaffold Wizard](../../../product/extension/EXT-005-guided-scaffold-wizard.md) |
| Implemented By | [Scaffold Wizard](../../use-cases/extension/VEXT-005-scaffold-wizard.md)               |

## Requirement

`Specs: Scaffold Requirement`, `Specs: Scaffold Feature`, `Specs: Scaffold
Component`, `Specs: Scaffold API`, and `Specs: Scaffold Service` shall be
registered in the Command Palette. Each command shall show a validated input
box for the artifact path; an empty value shall be rejected before the engine
is invoked. The constructed `specs scaffold` invocation shall run in the
integrated terminal and the newly created file shall be opened in the editor
on success.

## Rationale

Inline path validation catches typos before the engine is invoked, preventing
incomplete artifacts from being created. Opening the file automatically after
scaffolding reduces the steps an author must take before starting to edit.

## Verification

- Run `Specs: Scaffold Requirement` from the palette, enter a valid path, and
  confirm the engine is invoked in the integrated terminal and the new file
  opens in the editor.
- Enter an empty path and confirm the input is rejected without running the
  engine.
