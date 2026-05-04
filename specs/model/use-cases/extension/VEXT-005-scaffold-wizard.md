---
id: VEXT-005
status: Draft
requirements:
    - ../../requirements/extension/VEXT-005-guided-scaffold-wizard.md
---

# Scaffold Wizard

## Workflow

Present a validated input box for the artifact path, construct the
`specs scaffold` invocation with the correct kind and flags, run it in the
integrated terminal, and open the newly created file in the editor on success.

## VS Code Surface

- `Specs: Scaffold Requirement`, `Feature`, `Component`, `API`, and `Service`
  are registered palette commands.
- Each command shows a kind-specific placeholder in the input box (e.g.
  `core/012-some-requirement` for requirements).
- The path input is validated inline; an empty value is rejected before
  invoking the engine.
- The created file is opened automatically after a successful scaffold.

## Validation

Run `Specs: Scaffold Requirement` from the palette. Enter a valid path and
confirm the engine is invoked in the integrated terminal and the new file
opens in the editor. Enter an empty path and confirm the input is rejected
without running the engine.
