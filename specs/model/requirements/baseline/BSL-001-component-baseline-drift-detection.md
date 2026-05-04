---
id: BSL-001
status: Draft
realises:
    - ../../../product/engine/ENG-006-component-baseline-tracking.md
implemented_by:
    - ../../use-cases/baseline/BSL-001-baseline-drift-detection.md
---

# Component Baseline Drift Detection

## Requirement

`specs lint --baselines` shall compare the recorded baseline SHA in each
component file against the current HEAD commit of its upstream repository,
report each drifted component with the recorded SHA and the current upstream
SHA, and exit non-zero when at least one component has drifted.

## Rationale

Automatic drift detection surfaces stale baselines before they mislead
reviewers about the maturity of an implementation. Reporting both the
recorded and upstream SHAs gives the author the exact information needed to
decide whether to refresh the baseline.

## Verification

- Record a baseline SHA that is one commit behind the upstream.
- Run `specs lint --baselines` and confirm the drifted component is reported
  with both SHAs.
- Update the recorded SHA to match the upstream and confirm the command exits
  zero.
