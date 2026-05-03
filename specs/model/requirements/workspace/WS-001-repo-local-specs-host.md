# Repo Local Specs Host

| Field          | Value                                                                                                                                                                                                                                                                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID             | WS-001                                                                                                                                                                                                                                                                                                                                    |
| Status         | Draft                                                                                                                                                                                                                                                                                                                                     |
| Realises       | [Repo Local Specs Host](../../../product/engine/ENG-009-repo-local-specs-host.md)                                                                                                                                                                                                                                                         |
| Implemented By | [Repo Local Host Layout](../../features/workspace/repo-local-host-layout.md)                                                                                                                                                                                                                                                              |

## Requirement

The repository shall provide a coherent repo-local specs host layout that can
be exercised from the repo root without introducing path collisions between the
content tree, framework directory, build outputs, and diagnostics workflows.

## Rationale

Without a single host-level requirement, the detailed technical requirements for
path resolution, engine integration, and diagnostics drift into isolated fixes
instead of describing one maintainable local development workflow.

## Verification

- Run `make build-engine` and `./bin/specs doctor` from the repo root.
- Run `./bin/specs scaffold product-requirement --dry-run toolchain/check`.
- Confirm the resolved specs and framework paths stay inside this repository.
