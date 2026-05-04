---
id: ENG-001
stakeholder: Spec author
source: "[Overview](../../../docs/overview.md)"
realised_by:
    - ../../model/requirements/format/FMT-001-format-in-place.md
    - ../../model/requirements/format/FMT-002-format-check-mode.md
    - ../../model/requirements/lint/LNT-003-lint-check-flags.md
---

# Consistent Markdown Formatting

## Summary

Spec authors need every markdown file in the host to be formatted to a
uniform, deterministic style so that diffs in pull requests reflect only
intentional content changes and not incidental whitespace or layout noise.

## User Value

- Authors working in different editors no longer produce divergent whitespace
  that pollutes diffs and obscures real changes.
- CI can reject unformatted files before review, eliminating formatting
  back-and-forth from the review cycle.
- Running a single command both reformats files in place and reports which
  files are out of compliance without modifying them.

## Acceptance Signal

`specs format` rewrites files to the canonical layout in place. `specs format
--check` exits non-zero when at least one file would change and prints the
affected paths. Both modes are runnable in CI without additional configuration.
