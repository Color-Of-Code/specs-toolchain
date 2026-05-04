# Traceability Graph Export

| Field          | Value                                                                                         |
| -------------- | --------------------------------------------------------------------------------------------- |
| ID             | VIS-001                                                                                       |
| Status         | Draft                                                                                         |
| Realises       | [Traceability Visualization](../../../product/engine/ENG-007-traceability-visualization.md)   |
| Implemented By | [Graph Export](../../features/visualize/VIS-001-graph-export.md)                              |

## Requirement

`specs visualize traceability --format <mermaid|json>` shall render the
canonical requirement-to-implementation traceability graph to stdout in the
selected format. `--out <path>` shall redirect the output to a file. The JSON
format shall be the canonical payload consumed by the interactive UI and the
VS Code extension panel. The command shall exit non-zero when the canonical
graph is invalid.

## Rationale

Two output formats serve distinct audiences: Mermaid output can be embedded
directly in design documents or repository wikis, while JSON output feeds
downstream tooling including the interactive server and the extension panel.
A shared canonical payload format prevents divergence between the CLI and
extension views.

## Verification

- Run `specs visualize traceability --format json` and confirm the output is
  valid JSON containing node and edge data.
- Run `specs visualize traceability --format mermaid` and confirm Mermaid
  flowchart syntax is produced.
- Run with `--out <path>` and confirm the output is written to the file
  instead of stdout.
