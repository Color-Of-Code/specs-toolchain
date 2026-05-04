---
id: CR-003
status: Draft
requirements:
    - ../../requirements/cr/CR-003-change-request-drain.md
---

# Change Request Drain

## Workflow

Move every file in a CR's artifact subtrees to its canonical model home
using `git mv`, prompting for confirmation before each file (or skipping
prompts when `--yes` is given).

## Engine Surface

- `specs cr drain --id <NNN>` resolves the matching `CR-NNN-*` directory.
- Each artifact is mapped to its canonical destination under `model/` or
  `product/` based on its subtree type.
- `--yes` suppresses per-file confirmation.
- `--dry-run` prints the planned `git mv` operations without executing them.
- The CR directory is left in place after drain; removal is the caller's
  responsibility.

## Validation

Create a CR directory with one requirement file. Run `specs cr drain --id 1
--dry-run` and confirm the planned move is printed. Run without `--dry-run`
and confirm the file is moved to the canonical model path via `git mv`.
