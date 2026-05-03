# Format CI Gate

| Field        | Value                                                                                   |
| ------------ | --------------------------------------------------------------------------------------- |
| ID           | FMT-002                                                                                 |
| Status       | Draft                                                                                   |
| Requirements | [Format Check Mode](../../requirements/format/FMT-002-format-check-mode.md)             |

## Workflow

Run the same normalisation logic as FMT-001 but without writing any output.
Exit non-zero and print each non-canonical path so CI pipelines can gate on
formatting compliance.

## Engine Surface

- `specs format --check` activates the read-only comparison mode.
- Exit code is 0 when all files are canonical, non-zero otherwise.
- Output lists only files that would change; no diff is printed.

## Validation

Run `specs format --check` against a non-canonical file. Confirm it exits
non-zero and prints the path without modifying the file. Run against a
canonical directory and confirm it exits zero.
