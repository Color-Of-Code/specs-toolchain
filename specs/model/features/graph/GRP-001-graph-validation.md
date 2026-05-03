# Graph Validation

| Field        | Value   |
| ------------ | ------- |
| ID           | GRP-001 |
| Status       | Draft   |
| Requirements | —       |

## Workflow

Load the canonical traceability graph YAML files, verify that every node ID
resolves to an existing markdown artifact, check that all relation targets
are reachable, and validate baseline repository mappings.

## Engine Surface

- `specs graph validate` runs the full validation suite.
- Each invalid reference is printed with the graph file path and node ID.
- `--json` emits a machine-readable result payload for CI pipelines.
- Exit code is non-zero when the graph is invalid.

## Validation

Introduce a dangling node reference in a traceability YAML file. Run
`specs graph validate` and confirm the broken reference is reported. Restore
the file and confirm the command exits zero.
