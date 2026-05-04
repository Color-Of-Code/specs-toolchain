---
id: FMT-002
status: Draft
realises:
    - ../../../product/engine/ENG-001-consistent-markdown-formatting.md
implemented_by:
    - ../../use-cases/format/FMT-002-ci-gate.md
---

# Format Check Mode

## Requirement

`specs format --check` shall exit non-zero and print the path of every file
that would change, without modifying any file, so that CI pipelines can gate
on formatting compliance.

## Rationale

A non-destructive check mode is the standard pattern for formatter CI gates.
It allows automated checks to fail fast and report which files need attention
without mutating the working tree.

## Verification

- Run `specs format --check` on a directory containing at least one
  non-canonical file.
- Confirm the command exits non-zero and prints the non-canonical file path.
- Confirm the file's content is unchanged after the command runs.
- Run `specs format --check` on a directory where all files are already
  canonical and confirm the command exits zero.
