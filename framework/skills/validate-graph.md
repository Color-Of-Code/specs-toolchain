---
id: specs_validate_graph
name: Validate traceability graph
description: Validate the traceability graph for integrity — checks that all referenced nodes exist and the YAML round-trips cleanly.
tags:
  - specs
  - traceability
inputSchema:
  type: object
  properties: {}
  additionalProperties: false
engineArgs:
  default: [graph, validate]
---

Use this tool to validate the traceability graph of the specs workspace.
It checks that all edges reference existing nodes and that the canonical
YAML files can be read back without data loss.
Returns a summary with the node count and any errors found.
