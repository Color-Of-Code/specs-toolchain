---
id: BSL-002
requirements:
    - ../../requirements/baseline/BSL-002-component-baseline-refresh.md
---

# Baseline Refresh

## Workflow

For each component with a stale or missing baseline SHA, fetch the current
HEAD from the upstream repository and rewrite the component's baseline field
in its markdown file.

## Engine Surface

- `specs baseline update` refreshes all stale component baseline SHAs.
- `--only <substr>` limits the update to components whose name matches the
  substring.
- `--dry-run` prints the planned updates without writing any files.
- After updating, `specs lint --baselines` should exit zero.

## Validation

Run `specs baseline update --dry-run` and confirm the planned SHA updates are
printed. Run without `--dry-run` and confirm the markdown baseline fields are
updated. Follow with `specs lint --baselines` and confirm it exits zero.
