# Overview

A single-page summary of what the **specs engine** delivers. Everything below
is a feature of the `specs` binary; the VS Code extension is a thin wrapper
and is not covered here.

## What the engine is for

A small CLI that lets a team author a structured specification model in
markdown — product requirements, model requirements, use cases, components —
with guaranteed formatting, traceability, and baseline integrity.

## The model

Four artifact kinds, scaffolded from templates:

| Kind                | Owned by    | Purpose                                                                      |
| ------------------- | ----------- | ---------------------------------------------------------------------------- |
| Product requirement | Stakeholder | What was asked for, in the stakeholder's vocabulary. Lives under `product/`. |
| Requirement         | Author      | A single, testable re-formulation of one or more product requirements.       |
| Use case            | Analyst     | A scenario that satisfies one or more requirements.                          |
| Component           | Architect   | A unit of implementation pinned to an upstream repo.                         |

Product requirements live under `product/`; the rest live under `model/`.
All artifacts are written in markdown.

## The authoring chain

```text
Stakeholder ──► Author ──► Analyst ──► Architect
 product       requirements   use cases   components
 requirement
```

Stakeholders capture demands as **product requirements** inside change
requests; authors re-formulate them as model **requirements**; analysts and
architects refine those into use cases, components. See
[actors.md](actors.md) for details. Setup, review, and framework distribution
work is described as *operational roles* in [roles.md](roles.md).

## The change-request lifecycle

Work happens inside numbered change-request folders so it can be reviewed
in isolation before joining the canonical model.

```text
specs cr new      ──►  draft inside change-requests/NNN-slug/
specs scaffold    ──►  add requirements / use cases / components / ...
specs format      ──►  normalise markdown
specs lint        ──►  style + links + baselines
specs cr drain    ──►  git mv into canonical model paths
```

## What the engine guarantees

- **Consistent markdown** — `specs format` rewrites files in place,
  `specs format --check` is a CI gate.
- **Style compliance** — `specs lint --style` enforces the configured
  rules (defaults compiled in, overridable via `style.yaml`).
- **Traceability integrity** — `specs graph validate` verifies that the
  canonical traceability graph is well-formed, points at real markdown
  artifacts, and uses valid baseline repo mappings.
- **Component baselines** — `specs lint --baselines` detects drift between
  recorded SHAs and the real upstream commits;
  `specs baseline update` refreshes canonical baseline entries and regenerates
  component baseline fields.
- **Visual traceability** — `specs visualize traceability` renders the
  graph as Mermaid or JSON.
- **Diagnostics** — `specs doctor` prints engine version, resolved paths,
  framework mode, and version drift.

## What the engine does *not* do

- Render or publish documents — output is markdown, period.
- Manage tickets, sprints, or workflow approvals.
- Lock authors out of files — coordination is via change requests and
  normal git review.

## Where things live

| Path                        | Contents                                                                                                                   |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `.specs.yaml`               | Per-host configuration; resolves paths and framework.                                                                      |
| `product/`                  | Stakeholder-facing product requirements.                                                                                   |
| `model/`                    | Canonical model artifacts (requirements, use cases, components).                                                           |
| `change-requests/NNN-slug/` | Work in progress, drained into `product/` and `model/` on merge.                                                           |
| `.specs-framework`          | Framework content (templates, lint config). Either fetched into the user cache (managed) or supplied as a local directory. |

## Getting started

1. Install the binary: see [install.md](install.md).
2. Initialise a host: `specs init`.
3. Verify the setup: `specs doctor`.
4. Open a change request and start authoring: see [actors.md](actors.md)
   and [use-cases/](use-cases/README.md).

## Reference

- [Concepts](concepts.md) — paths, framework sources, modes.
- [Glossary](glossary.md) — core vocabulary and the product requirement vs.
  requirement distinction.
- [Commands](commands.md) — every `specs` subcommand.
- [Configuration](configuration.md) — `.specs.yaml` and registry schemas.
- [Use cases](use-cases/README.md) — task-oriented workflows.
