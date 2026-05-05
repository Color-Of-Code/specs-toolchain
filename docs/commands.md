# Command reference

Every command below is reachable as `specs <command>` on the terminal. Most are also exposed in the VS Code palette as **Specs: ...**; admin-only commands (`init`, `format`, `graph validate`, `graph import-markdown`, `graph generate-markdown`, `graph rebuild-cache`, `vscode init`, `framework seed`) are terminal-only. All write commands accept `--dry-run` where applicable.

## Core commands

- `specs version` (or `--version`) — print the installed binary version.
- `specs doctor` — diagnose environment, layout, and version drift.
- `specs init [<path>] [--framework <path-or-url>] [--with-model] [--with-vscode] [--force] [--dry-run]`
  Create or configure a host. `<path>` defaults to the current directory and is created if missing. `--framework` accepts either a local path (stored as `framework_dir`) or a remote git URL (materialised as a submodule at `specs/.framework`). When `--framework` is omitted, init defaults to `./framework`.
- `specs lint [--all] [--links] [--style]` — run lint checks. With no flag, all checks run.
- `specs format [--check] [--at <path>] [files...]` — format markdown files in place; `--check` exits non-zero if any file would change.
- `specs framework update [--to <ref>]` — update the framework content layer at the resolved `framework_dir`.
- `specs scaffold <kind> [--cr <NNN>] [--title <t>] [--force] [--dry-run] <path>` — instantiate a template (`product-requirement`, `requirement`, `use-case`, or `component`). Without `--cr`, `product-requirement` lands directly under `product/<path>.md`; the model kinds land under `model/<kind>s/<path>.md`. With `--cr`, every kind goes into the matching `change-requests/CR-NNN-*/<kind>s/` subtree.
- `specs cr new --id <NNN> --slug <slug> [--title <t>] [--force] [--dry-run]` — create a new change request from the template tree.
- `specs cr status` — list change requests with file counts per area.
- `specs cr drain --id <NNN> [--yes] [--dry-run]` — interactively `git mv` CR-local files to canonical model homes.
- `specs graph validate [--manifest <path>] [--json]` — validate the canonical traceability graph files, referenced markdown artifacts, and configured repo mappings.
- `specs graph import-markdown [--manifest <path>] [--force] [--dry-run] [--json]` — import the current markdown relationship fields into canonical graph YAML.
- `specs graph generate-markdown [--manifest <path>] [--dry-run] [--json]` — project canonical graph relations back into markdown field tables.
- `specs graph rebuild-cache [--manifest <path>] [--cache <path>] [--dry-run] [--json]` — rebuild the derived SQLite cache from canonical graph YAML.
- `specs graph save-relations [--manifest <path>] [--in <path>|-] [--json]` — update canonical relation entries from a JSON payload.
- `specs visualize traceability [--format mermaid|json] [--out <path>] [--serve] [--listen <addr>]` — render the canonical requirement ↔ implementer graph or host the local Cytoscape UI with layout switching and relation editing.
- `specs vscode init [--force]` — write `.vscode/tasks.json` with every Specs task.

## Framework commands

These commands support framework maintainers and project owners.

| Command                            | Purpose                                                   |
| ---------------------------------- | --------------------------------------------------------- |
| `specs framework seed --out <dir>` | create an empty framework skeleton in the given directory |
| `specs framework update [--to <ref>]` | update the configured framework checkout/submodule      |

### `specs framework seed`

Pre-seeds an empty directory with the minimal structure expected by the toolchain:

```text
<dir>/
├── templates/
├── process/
├── skills/
└── agents/
```

The command fails if the target directory already exists and is non-empty. After seeding, the caller is responsible for:

1. Running `git init` in the output directory.
2. Pushing it to a git remote for team use.
3. Referencing it from hosts via `specs init --framework <path-or-url>`.

This is an **advanced** operation intended for organisations that need a bespoke framework rather than forking an existing one.
