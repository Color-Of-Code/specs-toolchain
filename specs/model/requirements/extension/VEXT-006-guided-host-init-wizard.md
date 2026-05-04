# Guided Host Init Wizard

| Field          | Value                                                                                          |
| -------------- | ---------------------------------------------------------------------------------------------- |
| ID             | VEXT-006                                                                                       |
| Status         | Draft                                                                                          |
| Realises       | [Guided Host Initialisation](../../../product/extension/EXT-006-guided-host-initialisation.md) |
| Implemented By | [Init Wizard](../../use-cases/extension/VEXT-006-init-wizard.md)                               |

## Requirement

The `Specs: Init Host` palette command (`specs.bootstrap`) shall open a
multi-step wizard that collects the framework name or URL, model layout
preference, and optional VS Code tasks generation using validated step-by-step
prompts. When `.specs.yaml` already
exists in the workspace the wizard shall warn the user before proceeding.
The wizard shall be cancellable at any step without modifying the workspace.
On completion the wizard shall invoke `specs init` in the integrated
terminal and the status bar indicator shall reflect the new host.

## Rationale

A guided wizard removes the need for first-time users to read the command
reference before initialising a host. Cancellability and the pre-existing
host warning prevent accidental reinitialisation of a configured workspace.

## Verification

- Open a workspace without `.specs.yaml`, run `Specs: Init Host`, complete
  all prompts, and confirm `specs init` is invoked with the chosen options.
- Cancel the wizard mid-way and confirm no files are modified.
- Open a workspace that already has `.specs.yaml` and confirm the wizard
  displays a warning before proceeding.
