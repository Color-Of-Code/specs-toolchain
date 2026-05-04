# Baseline Drift Detection

| Field        | Value                                                                                                           |
| ------------ | --------------------------------------------------------------------------------------------------------------- |
| ID           | BSL-001                                                                                                         |
| Status       | Draft                                                                                                           |
| Requirements | [Component Baseline Drift Detection](../../requirements/baseline/BSL-001-component-baseline-drift-detection.md) |

## Workflow

For each component that records a baseline SHA, fetch the current HEAD commit
of the upstream repository and compare it to the recorded SHA. Report any
component where the two differ.

## Engine Surface

- `specs lint --baselines` activates this check category.
- Each drifted component is reported with the recorded SHA and the current
  upstream SHA.
- Exit code is non-zero when at least one component has drifted.

## Validation

Record a baseline SHA that is one commit behind the upstream. Run `specs lint
--baselines` and confirm the drifted component is reported. Update the SHA
and confirm the command exits zero.
