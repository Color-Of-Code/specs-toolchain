# In-Place Rewriting

| Field        | Value                                                                                 |
| ------------ | ------------------------------------------------------------------------------------- |
| ID           | FMT-001                                                                               |
| Status       | Draft                                                                                 |
| Requirements | [Format In Place](../../requirements/format/FMT-001-format-in-place.md)               |

## Workflow

Parse every targeted markdown file, apply canonical whitespace and table
normalisation, write the result back in place, and print the path of each
file that changed.

## Engine Surface

- `specs format [--at <dir>] [files…]` selects which files to process.
- Table columns are padded to the widest cell in each column.
- Blank-line rules, list indentation, and heading spacing are normalised.
- Files already in canonical form are left untouched (idempotent).

## Validation

Run `specs format` on a directory containing a non-canonical file. Confirm
the file is rewritten and its path is printed. Run again and confirm no
output is produced.
