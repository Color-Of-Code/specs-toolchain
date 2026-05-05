---
id: LNT-003
requirements:
    - ../../requirements/lint/LNT-003-lint-check-flags.md
---

# Lint Flag Composition

## Workflow

Dispatch lint invocations to exactly the check categories selected by the
active flags. When no flag is given, activate all categories. Each category
runs independently and contributes its own exit status.

## Engine Surface

- `--style` and `--links` each enable their category.
- No flag is equivalent to `--style --links`.
- Each category prints its own header and result block.
- The overall exit code is non-zero if any enabled category found issues.

## Validation

Run `specs lint --style` and confirm only the style block appears. Run
`specs lint` with no flags and confirm both blocks appear. Introduce
violations in different categories and confirm mixed results are handled
correctly.
