---
id: LNT-003
realises:
    - ../../../product/engine/ENG-001-consistent-markdown-formatting.md
implemented_by:
    - ../../use-cases/lint/LNT-003-flag-composition.md
---

# Lint Check Flags

## Requirement

`specs lint` shall expose `--style`, `--links`, and `--baselines` flags that
each enable exactly their corresponding check category. Invoking `specs lint`
without any flag shall run all check categories as if all three flags were
given.

## Rationale

Teams need to run individual check categories in separate CI steps (e.g. link
checking only during a fast pre-commit hook) without running the full suite.
The default-all behaviour keeps single-command usage simple.

## Verification

- Run `specs lint --style` and confirm only style output is produced.
- Run `specs lint --links` and confirm only link output is produced.
- Run `specs lint` with no flags and confirm all check categories run.
- Confirm each mode exits non-zero only when its own category finds issues.
