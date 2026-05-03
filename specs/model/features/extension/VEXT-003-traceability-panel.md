# Traceability Graph Panel

| Field        | Value    |
| ------------ | -------- |
| ID           | VEXT-003 |
| Status       | Draft    |
| Requirements | —        |

## Workflow

Invoke `specs visualize traceability --format json`, load the JSON payload
into a Cytoscape WebView panel inside VS Code, and allow the user to browse
nodes, inspect relations, and save node layout positions back to the canonical
graph.

## VS Code Surface

- `Specs: Visualize (Mermaid)` and `Specs: Visualize (DOT)` generate output
  files and open them in the editor.
- The WebView panel renders the Cytoscape graph from the JSON payload.
- Clicking a node shows its ID, title, and outbound relations in an inspector
  pane.
- Layout changes are posted back to `specs graph save-layout`.

## Validation

Run `Specs: Visualize (DOT)` and confirm a `.dot` file is created and opened.
Open the WebView panel and confirm nodes from the canonical traceability graph
are rendered. Select a node and confirm its metadata appears in the inspector.
