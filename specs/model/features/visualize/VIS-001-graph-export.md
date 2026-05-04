# Graph Export

| Field        | Value   |
| ------------ | ------- |
| ID           | VIS-001 |
| Status       | Draft   |
| Requirements | —       |

## Workflow

Render the canonical requirement-to-implementation traceability graph into a
requested output format and write it to stdout or a file.

## Engine Surface

- `specs visualize traceability --format <mermaid|json>` renders the
  graph in the selected format.
- `--out <path>` writes the output to a file instead of stdout.
- The JSON format is the canonical payload consumed by the interactive UI
  and the VS Code extension panel.
- Exit code is non-zero when the canonical graph is invalid.

## Validation

Run `specs visualize traceability --format json` and confirm the output is
valid JSON containing node and edge data. Run with `--format mermaid` and
confirm Mermaid flowchart syntax is produced.
