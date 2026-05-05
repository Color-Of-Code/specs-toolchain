# Ownership

Use this page when you want the short answer to "who normally does this?".
One person often holds several of these hats in the same repository.

## Authoring chain

The authoring chain owns the specification artifacts themselves.

| Party       | Produces                                 | Typical commands                                                                           |
| ----------- | ---------------------------------------- | ------------------------------------------------------------------------------------------ |
| Stakeholder | product requirements in a change request | `specs cr new`, `specs scaffold product-requirement --cr <NNN> ...`                        |
| Author      | model requirements                       | `specs scaffold requirement --cr <NNN> ...`, `specs format`, `specs lint`                  |
| Analyst     | use cases                                | `specs scaffold use-case --cr <NNN> ...`, `specs graph validate`                           |
| Architect   | components                               | `specs scaffold component --cr <NNN> ...`, `specs graph validate`                          |

See [actors.md](actors.md) for the longer narrative version of the authoring
chain.

## Operational roles

Operational roles own setup, review, and framework distribution work around the
model.

| Role                 | Responsibility                                                      | Typical commands                                                                        |
| -------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Any user             | local installation and diagnostics                                  | `specs doctor`                                                                           |
| Project owner        | host setup, framework source selection, editor task setup           | `specs init`, `specs vscode init`, `specs framework update`                             |
| Reviewer             | traceability and structure checks before merge                      | `specs graph validate`, `specs visualize traceability`                                  |
| Framework maintainer | create and publish framework content for downstream hosts           | `specs framework seed`                                                                   |

See [roles.md](roles.md) for the longer role-by-role detail.

## Quick reading guide

- Use [use-cases/README.md](use-cases/README.md) when you already know the task.
- Use [model.md](model.md) when you need the artifact model and traceability context.
- Use [framework.md](framework.md) when you need framework source and directory-mode details.
