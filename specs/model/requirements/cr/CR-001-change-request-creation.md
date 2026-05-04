---
id: CR-001
status: Draft
realises:
    - ../../../product/engine/ENG-005-change-request-workflow.md
implemented_by:
    - ../../use-cases/cr/CR-001-cr-creation.md
---

# Change Request Creation

## Requirement

`specs cr new --id <NNN> --slug <slug>` shall copy the framework's
`templates/change-request/` tree into a new `CR-NNN-<slug>/` directory under
`change-requests/`, substitute `CR-XXX` tokens with the normalised id, and
rewrite the `_index.md` H1 to include the id and title. `--title` shall set
the human-readable heading. `--force` shall allow recreating an existing CR
directory. `--dry-run` shall print the planned tree without writing.
`--json` shall emit `{path, id, slug, title}` on success.

## Rationale

A numbered, template-derived directory provides a consistent, isolated draft
area for every change request. Token substitution prevents manually-copied
templates from retaining placeholder values.

## Verification

- Run `specs cr new --id 1 --slug smoke-test --dry-run` and confirm the output
  shows the correct target path and title.
- Run without `--dry-run` and confirm the directory and `_index.md` are
  created with the correct content.
- Run `specs cr new --id 1 --slug smoke-test --json` and confirm a valid JSON
  object is emitted.
