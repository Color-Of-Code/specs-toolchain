---
id: SCF-001
status: Draft
requirements:
    - ../../requirements/scaffold/SCF-001-template-based-artifact-instantiation.md
---

# Template Instantiation

## Workflow

Copy the framework template matching the requested artifact kind to the
computed destination path, replace the H1 with a title derived from the
slug (or the `--title` override), and create any missing intermediate
directories.

## Engine Surface

- `specs scaffold <kind> <path>` resolves the template from the active
  framework's `templates/` directory.
- The destination is `model/<kind>s/<path>.md` by default.
- `--title` overrides the H1; without it the title is derived from the slug
  by stripping leading `NNN-` digits and converting separators to spaces.
- `--force` allows overwriting an existing file.
- `--dry-run` prints the destination path without writing.

## Validation

Run `specs scaffold requirement core/001-test-req --dry-run` and confirm the
printed destination path is correct. Run without `--dry-run` and confirm the
file exists with the correct H1 and section structure.
