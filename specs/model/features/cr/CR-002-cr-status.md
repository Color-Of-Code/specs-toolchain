# Change Request Status

| Field        | Value                                                                                              |
| ------------ | -------------------------------------------------------------------------------------------------- |
| ID           | CR-002                                                                                             |
| Status       | Draft                                                                                              |
| Requirements | [Change Request Status Reporting](../../requirements/cr/CR-002-change-request-status-reporting.md) |

## Workflow

Scan the `change-requests/` directory, collect one record per `CR-NNN-*`
entry, count markdown files in each artifact subtree, and report a summary
table or a JSON array.

## Engine Surface

- `specs cr status` prints a tabular summary: id, slug, per-area file counts,
  and whether `_index.md` is present.
- `--json` emits a JSON array of CR records with the same fields.
- An empty `change-requests/` directory produces an empty table (no error).

## Validation

Create two CR directories with varying content. Run `specs cr status` and
confirm both appear with correct file counts. Run `specs cr status --json`
and confirm the payload is valid JSON containing both records.
