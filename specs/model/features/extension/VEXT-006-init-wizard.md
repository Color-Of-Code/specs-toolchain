# Init Wizard

| Field        | Value                                                                                       |
| ------------ | ------------------------------------------------------------------------------------------- |
| ID           | VEXT-006                                                                                    |
| Status       | Draft                                                                                       |
| Requirements | [Guided Host Init Wizard](../../requirements/extension/VEXT-006-guided-host-init-wizard.md) |

## Workflow

Guide the user through a multi-step input sequence to collect the framework
name, model layout preference, and optional VS Code tasks generation, then
invoke `specs init` with the appropriate flags in the integrated terminal.

## VS Code Surface

- `Specs: Init Host` (`specs.bootstrap`) opens the guided wizard.
- Each step uses a `showInputBox` or `showQuickPick` prompt with validation.
- If a `.specs.yaml` already exists, the wizard warns the user before
  proceeding.
- The wizard can be cancelled at any step without modifying the workspace.
- On completion the status bar indicator updates to reflect the new host.

## Validation

Open a workspace without `.specs.yaml`. Run `Specs: Init Host`, complete all
prompts, and confirm `specs init` is invoked in the integrated terminal with
the chosen options. Open a workspace that already has `.specs.yaml` and
confirm the wizard displays a warning before proceeding.
