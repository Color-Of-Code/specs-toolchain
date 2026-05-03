# Repo Local Host Layout

| Field        | Value                                                                          |
| ------------ | ------------------------------------------------------------------------------ |
| Status       | Draft                                                                          |
| Requirements | [Repo Local Specs Host](../../requirements/workspace/WS-001-repo-local-specs-host.md) |

## Workflow

Maintain a host-shaped content tree under `specs/`, keep the project-specific
framework under `templates/`, rebuild the engine into `bin/specs`, and run the
same diagnostic and scaffolding commands that external host repositories use.

## Engine Surface

- `.specs.yaml` resolves `specs_root: ./specs` and `framework_dir: ../templates`.
- `make build-engine` writes the local engine binary to `bin/specs`.
- `specs doctor`, `specs scaffold`, and `specs graph validate` consume the
	same repo-local host layout.

## VS Code Surface

- `specs.enginePath` can point at the repo-local binary while testing the
	extension.
- `pnpm run deploy-dev` and the extension's engine resolution must use the same
	repo-local binary path assumptions.

## Validation

Run `make build-engine`, `./bin/specs doctor`, and one
`./bin/specs scaffold ... --dry-run` command from the repo root.
