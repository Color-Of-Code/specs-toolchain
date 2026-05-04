---
id: EXT-002
status: Draft
stakeholder: VS Code user, spec author
source: "[Commands](../../../docs/commands.md), [Extension README](../../../extension/README.md)"
realised_by:
    - ../../model/requirements/extension/VEXT-002-palette-command-dispatch.md
---

# Palette Command Access

## Summary

VS Code users need the most common specs operations to be reachable from the
Command Palette without opening a terminal, so that everyday authoring
actions stay within the editor and do not interrupt the editing flow.

## User Value

- Authors can run lint, format checks, and change-request status without
  leaving the editor or remembering the exact CLI syntax.
- Teams new to the toolchain can discover available commands through the
  familiar Command Palette search without consulting the documentation.
- Output from terminal-based commands is shown in an integrated terminal
  panel, keeping context visible alongside the edited files.

## Acceptance Signal

`Specs: Lint`, `Specs: Doctor`, `Specs: CR Status`, `Specs: Framework
Update`, and the lint variant commands (`--links`, `--style`, `--baselines`)
are all available in the Command Palette. Each command runs in the
integrated terminal rooted at the detected specs root. Commands that require
user input (scaffold, CR new, CR drain) prompt with an input box or quick
pick rather than requiring flags.
