# Traceability Graph Panel

| Field          | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| ID             | VEXT-003                                                                                   |
| Status         | Draft                                                                                      |
| Realises       | [Traceability Graph Panel](../../../product/extension/EXT-003-traceability-graph-panel.md) |
| Implemented By | [Traceability Graph Panel](../../features/extension/VEXT-003-traceability-panel.md)        |

## Requirement

The extension shall invoke `specs visualize traceability --format json` and
load the JSON payload into a Cytoscape WebView panel. The panel shall render
the graph with layered, organic, and grid layout options. Selecting a node
shall display its ID, title, and outbound relations in an inspector pane. The
`Specs: Visualize (Mermaid)` palette command shall generate an output file and
open it in the editor. The panel shall also allow exporting the canonical JSON
payload.

## Rationale

An in-editor panel removes the context switch to a browser for the most
common graph review workflow. The inspector pane surfaces relation data
without requiring the author to open individual markdown files.

## Verification

- Run `Specs: Visualize (Mermaid)` and confirm a Mermaid file is created and
  opened in the editor.
- Open the WebView panel and confirm nodes from the canonical traceability
  graph are rendered.
- Select a node and confirm its metadata and outbound relations appear in the
  inspector.
- Switch between layouts and confirm the graph relayouts without reloading.
