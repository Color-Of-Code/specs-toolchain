# Component Baseline Tracking

| Field       | Value                                                                                                                                                                                                                        |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID          | ENG-006                                                                                                                                                                                                                      |
| Status      | Draft                                                                                                                                                                                                                        |
| Stakeholder | Architect                                                                                                                                                                                                                    |
| Source      | [Overview](../../../docs/overview.md), [Commands](../../../docs/commands.md)                                                                                                                                                 |
| Realised By | [Component Baseline Drift Detection](../../model/requirements/baseline/BSL-001-component-baseline-drift-detection.md), [Component Baseline Refresh](../../model/requirements/baseline/BSL-002-component-baseline-refresh.md) |

## Summary

Architects need the engine to detect when recorded component baseline SHAs
have drifted from the actual latest commits in upstream repositories, so
that stale baselines are surfaced before they cause traceability
inconsistencies or mislead reviewers about the maturity of an implementation.

## User Value

- Architects are alerted when a component's recorded baseline SHA no longer
  matches the real upstream commit, without having to inspect each upstream
  repository manually.
- Updating baselines for all stale components in one command eliminates
  tedious per-component git queries.
- CI can block merging when any component baseline is out of date, ensuring
  the canonical model always reflects the current state of upstream code.

## Acceptance Signal

`specs lint --baselines` reports each component whose recorded SHA has
drifted from the upstream commit. `specs baseline update` fetches the
current SHA from each upstream and rewrites the component baseline fields
in place. `--only <substr>` limits the update to components whose name
matches the substring. `--dry-run` prints the planned changes without
modifying files.
