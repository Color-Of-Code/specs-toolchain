# Scaffold a model artifact

## Summary

Instantiate a single template — product requirement, requirement, feature,
component, API, or service — at a target path with placeholders pre-filled.

## Owner

Depends on the artifact kind (see [../actors.md](../actors.md)):

- **Stakeholder** *(actor)* for `product-requirement`.
- **Author** *(actor)* for `requirement`.
- **Analyst** *(actor)* for `feature`.
- **Architect** *(actor)* for `component`, `service`, `api`.

## Purpose

Eliminate copy-paste of template files; ensure new artifacts start
from the framework's current template version with a consistent header
structure.

## Entry point

`specs scaffold <kind> [--cr <NNN>] [--title <t>] [--force] [--dry-run]
<path>`

`<kind>` ∈ `product-requirement | requirement | feature | component | api | service`.

Without `--cr`, `product-requirement` lands directly under `product/<path>.md`;
the model kinds land under `model/<kind>s/<path>.md`. With `--cr`, every kind
goes into the matching `change-requests/CR-NNN-*/<kind>s/` subtree.

Or VS Code palette: **Specs: Scaffold …**.

Pre-conditions: the host is initialised; the target path does not
exist unless `--force`; `--cr` references an existing CR folder.

## Exit point

A new markdown file at `<path>` (or under
`change-requests/NNN-slug/...` when `--cr` is set) ready to edit.

## Iteration

Re-run with `--force` to regenerate from the current template; or edit
in place. Subsequent template-schema bumps surface via
[`specs lint`](lint-and-format.md) — re-scaffold a fresh sibling and
diff if needed.
