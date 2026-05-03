# Maintain component baselines

## Summary

Verify or refresh the SHAs recorded for each tracked component in the
canonical baseline graph entries, so reviewers see exactly which upstream commit a
component spec was written against.

## Owner

**Architect** *(actor)* — see [../actors.md](../actors.md). Owns components and the canonical baseline entries that pin each of them to an upstream commit.

## Purpose

Detect when an upstream `repos:` entry has moved on without the spec
catching up, and provide a one-shot way to bring the canonical baseline data in sync.

## Entry point

- Verify: `specs lint --baselines`
- Update: `specs baseline update [--only <substr>] [--dry-run]`

Or VS Code palette: **Specs: Update baselines**.

Pre-conditions: `repos:` mappings in `.specs.yaml` resolve to local git
checkouts; the canonical graph manifest (default `model/traceability/graph.yaml`)
exists and contains baseline entries.

## Exit point

After **update**: the canonical baseline entries contain current `git log`
SHAs and the generated component `Baseline` fields are refreshed. After **lint --baselines**: zero exit when the recorded commits are in sync,
non-zero with a per-component diff otherwise.

## Workflow

1. Run `specs lint --baselines` to detect drift.
2. If drift is reported, run `specs baseline update --dry-run` to
   preview the canonical baseline updates.
3. Re-run without `--dry-run` (optionally narrowing with `--only`).
4. Review the resulting diff in `model/traceability/baselines.yaml` and any regenerated component markdown; commit alongside
   any related spec changes.

### Iteration

Repeat per CR or per release. The `--only <substr>` flag scopes
updates to a subset of components when only some are in scope.
