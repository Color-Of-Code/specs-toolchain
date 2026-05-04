---
id: SCF-001
status: Draft
realises:
    - ../../../product/engine/ENG-004-artifact-scaffolding-from-templates.md
implemented_by:
    - ../../use-cases/scaffold/SCF-001-template-instantiation.md
---

# Template-Based Artifact Instantiation

## Requirement

`specs scaffold <kind> <path>` shall copy the framework template for the
requested artifact kind to `model/<kind>s/<path>.md`, derive the H1 heading
from the path slug (stripping leading digit prefixes and converting separators
to spaces), create any missing intermediate directories, and exit non-zero
when the template is not found. `--title` shall override the derived H1.
`--force` shall allow overwriting an existing file. `--dry-run` shall print
the destination path without writing any file.

## Rationale

Deriving the heading from the slug eliminates a common source of copy-paste
errors and ensures new artifacts start with structurally correct content.
The dry-run mode lets authors verify the computed path before committing.

## Verification

- Run `specs scaffold requirement core/001-test-req --dry-run` and confirm the
  printed destination path matches `model/requirements/core/001-test-req.md`.
- Run without `--dry-run` and confirm the file exists with the correct H1 and
  section stubs.
- Run again without `--force` and confirm an error is returned instead of
  overwriting.
