# Component Baseline Refresh

| Field          | Value                                                                                         |
| -------------- | --------------------------------------------------------------------------------------------- |
| ID             | BSL-002                                                                                       |
| Status         | Draft                                                                                         |
| Realises       | [Component Baseline Tracking](../../../product/engine/ENG-006-component-baseline-tracking.md) |
| Implemented By | [Baseline Refresh](../../use-cases/baseline/BSL-002-baseline-refresh.md)                      |

## Requirement

`specs baseline update` shall fetch the current HEAD commit SHA from the
upstream repository of each component with a stale or missing baseline and
rewrite the component's baseline field in its markdown file. `--only <substr>`
shall limit the update to components whose name matches the given substring.
`--dry-run` shall print the planned updates without modifying any files.

## Rationale

A single command that refreshes all stale baselines eliminates the need for
per-component manual git queries and prevents human error when copying SHAs.
The `--only` filter supports incremental updates during active development.

## Verification

- Run `specs baseline update --dry-run` and confirm the planned SHA updates
  are printed without modifying any files.
- Run without `--dry-run` and confirm the markdown baseline fields are updated.
- Follow with `specs lint --baselines` and confirm it exits zero.
