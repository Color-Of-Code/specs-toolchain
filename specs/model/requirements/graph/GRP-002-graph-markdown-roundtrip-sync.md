# Graph Markdown Round-Trip Sync

| Field          | Value                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------- |
| ID             | GRP-002                                                                                             |
| Status         | Draft                                                                                               |
| Realises       | [Traceability Graph Validation](../../../product/engine/ENG-003-traceability-graph-validation.md)   |
| Implemented By | [Graph Markdown Roundtrip](../../features/graph/GRP-002-graph-markdown-roundtrip.md)                |

## Requirement

`specs graph import-markdown` shall read the `Realises`, `Implemented By`, and
baseline table fields from markdown artifact files and write the extracted
relations into the canonical traceability YAML. `specs graph
generate-markdown` shall project the canonical YAML back into the
corresponding markdown fields. Both commands shall accept `--dry-run` and
`--json`. `specs graph rebuild-cache` shall regenerate the derived cache from
the canonical YAML after either sync direction.

## Rationale

Keeping markdown field values and canonical YAML in sync through explicit
import and generate commands prevents divergence caused by manual edits to
either representation. The `rebuild-cache` step ensures that downstream
queries always reflect the latest canonical state.

## Verification

- Edit a relationship field in a markdown file.
- Run `specs graph import-markdown` and confirm the canonical YAML is updated.
- Run `specs graph generate-markdown` and confirm the markdown field is
  regenerated to match the canonical YAML.
- Run `specs graph rebuild-cache` and confirm the cache is refreshed without
  error.
