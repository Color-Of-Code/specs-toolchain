# Use cases

Task-oriented workflows exposed by the specs engine. Each page describes one
use case in terms of **Summary**, **Owner**, **Purpose**, **Entry point**, and
**Exit point**.

Use [../overview.md](../overview.md) when you need a fast router through the
docs. Use [../ownership.md](../ownership.md) when you need the short answer to
who normally owns a task.

## Quick paths

- First-time setup: [Set up a host](setup-host.md) → [Diagnose the environment](diagnose-environment.md)
- Daily authoring: [Author a change request](author-change-request.md) → [Scaffold a model artifact](scaffold-model-artifact.md) → [Lint and format specifications](lint-and-format.md) → [Verify traceability links](verify-traceability.md)
- Framework administration: [Seed a brand-new framework](seed-new-framework.md) → [Update the framework content layer](update-framework.md) → [Configure VS Code integration](configure-vscode.md)

## Authoring (day-to-day)

| Use case                                                         | Owner                                                                 |
| ---------------------------------------------------------------- | --------------------------------------------------------------------- |
| [Author a change request](author-change-request.md)              | Stakeholder → Author *(actors)*                                       |
| [Scaffold a model artifact](scaffold-model-artifact.md)          | Author / Analyst / Architect *(actors)*                               |
| [Lint and format specifications](lint-and-format.md)             | Any authoring actor                                                   |
| [Verify traceability links](verify-traceability.md)              | Reviewer *(role)* — typically Author / Analyst / Architect themselves |
| [Visualize the traceability graph](visualize-traceability.md)    | Any authoring actor; consumed by Reviewers and Stakeholders           |

## Setup and maintenance (one-off or occasional)

| Use case                                                      | Owner                              |
| ------------------------------------------------------------- | ---------------------------------- |
| [Set up a host](setup-host.md)                                | Project owner *(role)*             |
| [Diagnose the environment](diagnose-environment.md)           | Any user *(role)*                  |
| [Update the framework content layer](update-framework.md)     | Project owner *(role)*             |
| [Configure VS Code integration](configure-vscode.md)          | Project owner *(role)* or Any user |
| [Seed a brand-new framework](seed-new-framework.md)           | Framework maintainer *(role)*      |
