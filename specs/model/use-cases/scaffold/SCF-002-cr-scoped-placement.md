---
id: SCF-002
requirements:
    - ../../requirements/scaffold/SCF-002-cr-scoped-scaffold-placement.md
---

# CR-Scoped Placement

## Workflow

When `--cr <NNN>` is given, redirect the scaffold output to the matching
change-request subtree instead of the canonical model tree, so work-in-progress
artifacts land in the right draft area automatically.

## Engine Surface

- `specs scaffold <kind> --cr <NNN> <path>` looks up the `CR-NNN-*` directory
  under `change-requests/` and places the file under its `<kind>s/` subtree.
- The id is normalised to three digits (`4` → `004`).
- An error is returned when no matching CR directory is found.

## Validation

Create a change-request directory. Run `specs scaffold requirement --cr 1
core/test --dry-run` and confirm the printed path is inside the CR directory.
Run without `--dry-run` and confirm the file is created in the CR subtree.
