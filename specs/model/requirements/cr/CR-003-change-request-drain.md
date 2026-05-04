# Change Request Drain to Canonical Paths

| Field          | Value                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------- |
| ID             | CR-003                                                                                    |
| Status         | Draft                                                                                     |
| Realises       | [Change Request Workflow](../../../product/engine/ENG-005-change-request-workflow.md)     |
| Implemented By | [Change Request Drain](../../features/cr/CR-003-cr-drain.md)                              |

## Requirement

`specs cr drain --id <NNN>` shall resolve the matching `CR-NNN-*` directory,
map each artifact to its canonical model destination based on its subtree
type, and move each file using `git mv`. The command shall prompt for
confirmation before each move unless `--yes` is given. `--dry-run` shall
print the planned `git mv` operations without executing them. The CR
directory shall be left in place after the drain.

## Rationale

Using `git mv` preserves history during the promotion of a change request
to the canonical model. Per-file confirmation prevents accidental overwrites
of existing canonical artifacts during an interactive drain.

## Verification

- Create a CR directory containing one requirement file.
- Run `specs cr drain --id 1 --dry-run` and confirm the planned move is
  printed with the correct canonical destination.
- Run without `--dry-run`, confirm the file is moved to the canonical model
  path via `git mv`, and confirm the CR directory remains.
