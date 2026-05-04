---
id: FMT-001
status: Draft
realises:
    - ../../../product/engine/ENG-001-consistent-markdown-formatting.md
implemented_by:
    - ../../use-cases/format/FMT-001-in-place-rewriting.md
---

# Format In Place

## Requirement

The `specs format` command shall rewrite every markdown file it processes to
the canonical layout in place, applying deterministic whitespace and table
normalisation, and print the path of each file it modifies.

## Rationale

In-place rewriting with printed output lets authors and CI see exactly which
files changed without requiring a separate diff step. Determinism ensures that
running the command twice on the same input is a no-op.

## Verification

- Run `specs format` on a directory containing at least one non-canonical file.
- Confirm the file is rewritten and its path is printed.
- Run `specs format` again on the same directory and confirm no output is
  printed (idempotency).
