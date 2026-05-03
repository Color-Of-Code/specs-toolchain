# Change Request Creation

| Field        | Value   |
| ------------ | ------- |
| ID           | CR-001  |
| Status       | Draft   |
| Requirements | —       |

## Workflow

Copy the framework's `templates/change-request/` tree into a new numbered
directory under `change-requests/`, substituting `CR-XXX` tokens with the
normalised id and rewriting the `_index.md` H1 to include the id and title.

## Engine Surface

- `specs cr new --id <NNN> --slug <slug>` creates `CR-NNN-<slug>/`.
- `--title` sets the human-readable heading in `_index.md`.
- `--force` allows recreating an existing CR directory.
- `--dry-run` prints the planned tree without writing.
- `--json` emits `{path, id, slug, title}` on success.

## Validation

Run `specs cr new --id 1 --slug smoke-test --dry-run` and confirm the printed
output shows the correct target path and title. Run without `--dry-run` and
confirm the directory and `_index.md` are created with the correct content.
