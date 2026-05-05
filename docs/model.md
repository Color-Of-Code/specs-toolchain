# Model

The specs toolchain manages a small, stable artifact model. This page is the
short conceptual map for the artifacts, their homes, and how work moves between
draft and canonical state.

## Core paths

| Path                        | Meaning                                                                              |
| --------------------------- | ------------------------------------------------------------------------------------ |
| specs root                  | the directory containing `.specs.yaml`, `product/`, `model/`, and `change-requests/` |
| host repo                   | the git repository that contains the specs root                                      |
| `change-requests/NNN-slug/` | the draft work area before artifacts are drained into canonical paths                |

Framework-specific paths and modes are covered in [framework.md](framework.md).

## Artifact kinds

| Kind                | Canonical home        | Purpose                                                               | Typical owner |
| ------------------- | --------------------- | --------------------------------------------------------------------- | ------------- |
| Product requirement | `product/`            | what the stakeholder asked for, in the stakeholder's vocabulary       | Stakeholder   |
| Requirement         | `model/requirements/` | a precise, testable reformulation of one or more product requirements | Author        |
| Use case            | `model/use-cases/`    | a scenario that satisfies one or more requirements                    | Analyst       |
| Component           | `model/components/`   | an implementation unit pinned to an upstream repo                     | Architect     |

Exact term definitions live in [glossary.md](glossary.md).

## How work moves

1. Draft product requirements and model artifacts inside a numbered change request.
2. Iterate with `specs scaffold`, `specs format`, `specs lint`, and `specs graph validate`.
3. Drain approved files into `product/` and `model/`.

The end-to-end workflow is documented in [use-cases/author-change-request.md](use-cases/author-change-request.md).

## Traceability at a glance

- Product requirements motivate one or more requirements.
- Requirements are implemented through use cases and components.
- Markdown fields such as `## Realised By`, `## Realises`, `## Requirements`, and `## Implemented By` are projected into the canonical traceability graph.
- `specs graph validate` checks that graph before review, and `specs visualize traceability` renders it.

For the standard SysML relation names and arrow direction, see [relations.md](relations.md).
