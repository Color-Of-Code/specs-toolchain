---
id: CR-002
status: Draft
realises:
    - ../../../product/engine/ENG-005-change-request-workflow.md
implemented_by:
    - ../../use-cases/cr/CR-002-cr-status.md
---

# Change Request Status Reporting

## Requirement

`specs cr status` shall scan the `change-requests/` directory, collect one
record per `CR-NNN-*` entry with the id, slug, per-area markdown file counts,
and whether `_index.md` is present, and print a tabular summary. `--json`
shall emit a JSON array of the same records. An empty `change-requests/`
directory shall produce an empty table without an error.

## Rationale

A tabular status summary lets project leads assess the volume and structure
of open change requests at a glance without navigating the file tree. The
JSON output enables programmatic consumption by the extension or CI scripts.

## Verification

- Create two CR directories with varying content.
- Run `specs cr status` and confirm both appear with correct per-area file
  counts and `_index.md` presence.
- Run `specs cr status --json` and confirm the payload is valid JSON containing
  both records.
- Remove all CR directories and confirm `specs cr status` exits zero with no
  output.
