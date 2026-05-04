---
id: CR-003
realises:
    - ../../../product/engine/ENG-005-change-request-workflow.md
implemented_by:
    - ../../use-cases/cr/CR-003-cr-drain.md
---

# Change Request Drain to Canonical Paths

## Requirement

`specs cr drain --id <NNN>` shall resolve the matching `CR-NNN-*` directory,
map each artifact to its canonical model destination based on its subtree
type, and move each file using a version-control-aware rename that preserves
history. The command shall prompt for
confirmation before each move unless `--yes` is given. `--dry-run` shall
print the planned move operations without executing them. The CR
directory shall be left in place after the drain.

## Rationale

Renaming files with history preservation during the promotion of a change
request to the canonical model prevents losing authorship and context.
Per-file confirmation prevents accidental overwrites
of existing canonical artifacts during an interactive drain.

## Verification

- Create a CR directory containing one requirement file.
- Run `specs cr drain --id 1 --dry-run` and confirm the planned move is
  printed with the correct canonical destination.
- Run without `--dry-run`, confirm the file is moved to the canonical model
  path with history preserved, and confirm the CR directory remains.
