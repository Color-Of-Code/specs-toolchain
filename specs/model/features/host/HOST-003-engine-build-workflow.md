# Engine Build Workflow

| Field        | Value                                                                                                 |
| ------------ | ----------------------------------------------------------------------------------------------------- |
| ID           | HOST-003                                                                                              |
| Status       | Draft                                                                                                 |
| Requirements | [Repo Local Engine Integration](../../requirements/workspace/WS-003-repo-local-engine-integration.md) |

## Workflow

Build and use a repo-local engine binary during development without colliding
with the `specs/` content tree, then point VS Code integration at that same
binary path.

## Engine Surface

- `make build-engine` writes the engine to `bin/specs`.
- Repo-local authoring commands run from the repo root against `specs/`.
- Build outputs remain separate from the content tree and local framework.

## VS Code Surface

- `specs.enginePath` can target the repo-local `bin/specs` binary.
- Extension development scripts use the same local binary path assumption.

## Validation

Run `make build-engine`, confirm `bin/specs` exists, and exercise both a
terminal command and an extension development flow against that binary.
