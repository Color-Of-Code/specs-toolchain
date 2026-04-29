# Development


## CLI

```bash
go test ./...
go build ./...
go install ./cmd/specs
```

## VS Code extension

```bash
pnpm --filter ./extension install
pnpm --filter ./extension run compile
```

See [extension/README.md](../extension/README.md) for the extension-specific settings, packaging notes, and platform matrix.

## Markdown lint & format

Repo-level docs are checked with [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2) and formatted with [Prettier](https://prettier.io/). Configs: [`.markdownlint-cli2.jsonc`](../.markdownlint-cli2.jsonc), [`.prettierrc.json`](../.prettierrc.json), [`.prettierignore`](../.prettierignore).

```bash
pnpm install         # once, at the repo root
pnpm run md:check    # prettier --check + markdownlint (what CI runs)
pnpm run md:format   # prettier --write
pnpm run md:fix      # markdownlint --fix
```

The `markdown` job in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs the same checks on every push and pull request.

## Build (extension & CLI)

To build both the VS Code extension and the CLI binary in one step:

```bash
pnpm run build
```

- The CLI binary is built to `specs-cli/specs`.
- The extension is compiled in `specs-cli/extension`.

You can also build them individually:

```bash
pnpm run build:cli         # builds Go CLI
pnpm run build:extension   # builds VS Code extension
```

## Markdown lint & format (docs only)

All markdown lint and formatting checks only apply to `docs/*.md`.

```bash
pnpm run md:check   # prettier --check + markdownlint (docs only)
pnpm run md:format  # prettier --write (docs only)
pnpm run md:fix     # markdownlint --fix (docs only)
```

The `markdown` job in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs the same checks on every push and pull request.

## Releases

Cross-platform release builds are produced by GoReleaser on git tags (`v*.*.*`). See [`.goreleaser.yaml`](../.goreleaser.yaml). Per-platform `.vsix` artifacts are attached to GitHub releases via [`scripts/build-extension.sh`](../scripts/build-extension.sh).

## Status

Phase 1 — lint, layout auto-detection, `init` / `bootstrap` / `tools update`, **managed mode** (cache + auto-fetch).
**Phase 2** — `scaffold`, `cr {new,status,drain}`, `baseline {check,update}`, `link check`, `vscode init` shipped.
**Phase 3** — `visualize traceability` (DOT and Mermaid), `templates_schema` enforcement, `--layout submodule` shipped.
**VS Code extension** under `extension/` (in progress; see [extension/README.md](../extension/README.md)).

## pnpm

This repo uses [pnpm](https://pnpm.io/) for all dev tooling. If you have npm artifacts from a previous version, remove them:

```bash
rm -rf node_modules package-lock.json
pnpm install
```

All scripts work the same way as with npm, just use `pnpm run ...` instead.
