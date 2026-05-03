# Repo Local Engine Integration

| Field          | Value                                                                                                                                                                                                     |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID             | WS-003                                                                                                                                                                                                    |
| Status         | Draft                                                                                                                                                                                                     |
| Realises       | [Repo Local Specs Host](../../../product/engine/ENG-009-repo-local-specs-host.md)                                                                                                                         |
| Implemented By | [Repo Local Host Layout](../../features/workspace/repo-local-host-layout.md), [Specs Engine](../../components/engine/specs-engine.md), [Vscode Extension](../../components/extension/vscode-extension.md) |

## Requirement

The repository shall build the local engine into a non-conflicting repo path
and the extension shall resolve that local binary consistently during
development workflows.

## Rationale

The repo-local host layout only stays usable if the engine build artifact does
not collide with the `specs/` content tree and extension development uses the
same local engine path assumptions as terminal workflows.

## Verification

- Run `make build-engine` and confirm the binary is written to `bin/specs`.
- Run the extension development workflow and confirm it resolves the same local
  engine path.
- Confirm the top-level `specs/` directory remains a content tree, not a build
  artifact.
