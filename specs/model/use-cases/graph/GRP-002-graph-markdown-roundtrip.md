---
id: GRP-002
status: Draft
requirements:
    - ../../requirements/graph/GRP-002-graph-markdown-roundtrip-sync.md
---

# Graph Markdown Roundtrip

## Workflow

Synchronise between canonical traceability YAML and the relationship fields
embedded in markdown artifact files. Import reads the markdown fields and
writes them into canonical YAML; generate projects the canonical YAML back
into the markdown fields.

## Engine Surface

- `specs graph import-markdown` reads `Realises`, `Implemented By`, and
  baseline table fields from markdown and writes canonical YAML entries.
- `specs graph generate-markdown` writes those fields back into each
  markdown file from the canonical YAML.
- Both commands accept `--dry-run` and `--json`.
- `specs graph rebuild-cache` regenerates the derived SQLite cache from the
  canonical YAML after either direction of sync.

## Validation

Edit a relationship field in a markdown file. Run `specs graph
import-markdown` and confirm the canonical YAML is updated. Run `specs
graph generate-markdown` and confirm the markdown field is regenerated
consistently.
