# Command reference

Every command below is reachable as `specs <command>` on the terminal. Most are also exposed in the VS Code palette as **Specs: …**; admin-only commands (`init`, `format`, `graph validate`, `graph import-markdown`, `graph generate-markdown`, `graph rebuild-cache`, `vscode init`, `framework list|add|remove|seed`) are terminal-only. All write commands accept `--dry-run` where applicable.

## Core commands

- `specs version` (or `--version`) — print the installed binary version.
- `specs doctor` — diagnose environment, layout, and version drift.
- `specs init [<path>] [--framework <name>[@ref]] [--with-model] [--with-vscode] [--force] [--dry-run]`
  Create or configure a host. `<path>` defaults to the current directory and is created if missing. `--framework` takes a name registered with `specs framework add` (for example `acme`, or `acme@v2.1` to override the registered ref). When `--framework` is omitted the registry's `default` entry is used; if no entries are registered, `specs init` fails. URL-based entries are fetched into the user cache (managed mode); path-based entries are recorded in `framework_dir` and left untouched, so the host can hold the framework as a plain folder, a git submodule, or a vendored snapshot — whichever fits.
- `specs lint [--all] [--links] [--style] [--baselines]` — run lint checks. With no flag, all checks run.
- `specs format [--check] [--at <path>] [files...]` — format markdown files in place; `--check` exits non-zero if any file would change.
- `specs framework update [--to <ref>]` — update the `.specs-framework` content layer.
- `specs scaffold <kind> [--cr <NNN>] [--title <t>] [--force] [--dry-run] <path>` — instantiate a template (`product-requirement`, `requirement`, `feature`, `component`, `api`, or `service`). Without `--cr`, `product-requirement` lands directly under `product/<path>.md`; the model kinds land under `model/<kind>s/<path>.md`. With `--cr`, every kind goes into the matching `change-requests/CR-NNN-*/<kind>s/` subtree.
- `specs cr new --id <NNN> --slug <slug> [--title <t>] [--force] [--dry-run]` — create a new change request from the template tree.
- `specs cr status` — list change requests with file counts per area.
- `specs cr drain --id <NNN> [--yes] [--dry-run]` — interactively `git mv` CR-local files to canonical model homes.
- `specs baseline update [--only <substr>] [--dry-run]` — refresh stale canonical baseline SHAs from `git log` and regenerate component baseline fields.
- `specs graph validate [--manifest <path>] [--json]` — validate the canonical traceability graph files, referenced markdown artifacts, and baseline repo mappings.
- `specs graph import-markdown [--manifest <path>] [--force] [--dry-run] [--json]` — import the current markdown relationship fields and baseline table into canonical graph YAML.
- `specs graph generate-markdown [--manifest <path>] [--dry-run] [--json]` — project canonical graph relations back into markdown field tables.
- `specs graph rebuild-cache [--manifest <path>] [--cache <path>] [--dry-run] [--json]` — rebuild the derived SQLite cache from canonical graph YAML.
- `specs visualize traceability [--format dot|mermaid|json] [--out <path>] [--serve] [--listen <addr>]` — render the canonical requirement ↔ implementer graph or host the local Cytoscape UI.
- `specs vscode init [--force]` — write `.vscode/tasks.json` with every Specs task.

## Framework management commands

These commands manage the framework registry and support creating new frameworks from scratch. They are **not** needed for day-to-day specs authoring — only for framework maintainers and administrators.

| Command                                                | Purpose                                                   |
| ------------------------------------------------------ | --------------------------------------------------------- |
| `specs framework list`                                 | show all registered framework entries                     |
| `specs framework add <name> --url <URL> [--ref <ref>]` | register a remote framework source by name                |
| `specs framework add <name> --path <dir>`              | register a local directory as a framework source          |
| `specs framework remove <name>`                        | unregister a framework entry                              |
| `specs framework seed --out <dir>`                     | create an empty framework skeleton in the given directory |

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
3. Registering it in the framework registry with `specs framework add <name> --url <git-url>` (or `--path <dir>`), so `specs init --framework <name>` can resolve it.

This is an **advanced** operation intended for organisations that need a bespoke framework rather than forking an existing one.
