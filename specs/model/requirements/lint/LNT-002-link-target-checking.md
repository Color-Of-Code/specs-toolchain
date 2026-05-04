---
id: LNT-002
status: Draft
realises: []
implemented_by:
    - ../../use-cases/lint/LNT-002-cross-reference-validation.md
---

# Link Target Checking

## Requirement

`specs lint --links` shall detect markdown link targets that do not resolve to
an existing file in the host, report each broken link with its source file path
and line number, and exit non-zero when at least one broken link is found.

## Rationale

Broken cross-references are a common source of silent traceability rot. An
automated check gives authors early feedback before a reviewer has to trace
links manually.

## Verification

- Add a markdown link pointing to a non-existent file.
- Run `specs lint --links` and confirm the broken link is reported with the
  correct source location.
- Fix or remove the broken link, rerun, and confirm the command exits zero.
