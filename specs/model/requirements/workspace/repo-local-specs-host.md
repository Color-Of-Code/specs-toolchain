# Repo Local Specs Host

| Field          | Value                                                                                                                                                                                                                                                                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Status         | Draft                                                                                                                                                                                                                                                                                                                                     |
| Realises       | [Repo Local Specs Host](../../../product/toolchain/repo-local-specs-host.md)                                                                                                                                                                                                                                                              |
| Implemented By | [Repo Local Host Layout](../../features/workspace/repo-local-host-layout.md), [Specs Engine](../../components/engine/specs-engine.md), [VSCode Extension](../../components/extension/vscode-extension.md), [Workspace Diagnostics](../../services/workspace/workspace-diagnostics.md), [Doctor Json](../../apis/workspace/doctor-json.md) |

## Requirement

The repository shall resolve a repo-local specs host and repo-local framework
consistently across engine commands, extension development scripts, and
traceability tooling so local development uses the same paths in every entry
point.

## Rationale

Without a canonical local host layout, the repository either collides with the
engine build artifact at the root `specs` path or silently points diagnostics
and scaffolding at the wrong framework content.

## Verification

- Run `make build-engine` and `./bin/specs doctor` from the repo root.
- Run `./bin/specs scaffold product-requirement --dry-run toolchain/check`.
- Confirm the resolved specs and framework paths stay inside this repository.
