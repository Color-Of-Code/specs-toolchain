# Traceability Graph Panel

| Field       | Value                                                                                   |
| ----------- | --------------------------------------------------------------------------------------- |
| Status      | Draft                                                                                   |
| Stakeholder | Product manager, architect                                                              |
| Source      | [Overview](../../../docs/overview.md), [Extension README](../../../extension/README.md) |
| Realised By | —                                                                                       |

## Summary

Product managers and architects need an interactive traceability graph
rendered inside VS Code so they can explore the full requirement-to-
implementation chain without switching to a browser or external tool.

## User Value

- Product managers can browse the product requirement → requirement →
  feature → component hierarchy visually in the same window where they
  author spec files.
- Architects can select a node in the graph to inspect its metadata and
  outbound relations without opening individual markdown files.
- Node positions can be saved so the layout is preserved across sessions
  and shared through the repository.

## Acceptance Signal

The `Specs: Visualize (Mermaid)` and `Specs: Visualize (DOT)` palette
commands generate output files in the model directory and open them in the
editor. The extension's WebView panel renders the Cytoscape-based graph
from the engine's JSON output. Clicking a node shows its relations. Saved
layouts persist across editor restarts via the canonical graph YAML.
