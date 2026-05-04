# Workspace Diagnostics Reporting

| Field          | Value                                                                                                                                                       |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID             | WS-004                                                                                                                                                      |
| Status         | Draft                                                                                                                                                       |
| Realises       | [Repo Local Specs Host](../../../product/engine/ENG-009-repo-local-specs-host.md)                                                                           |
| Implemented By | [Human Diagnostics Output](../../use-cases/doctor/DOC-001-human-diagnostics.md), [Workspace Diagnostics](../../services/workspace/workspace-diagnostics.md) |

## Requirement

The diagnostics workflow shall report the effective repo-local specs root,
framework directory, and related artifact paths without misclassifying missing
graph cache files as a layout failure.

## Rationale

Maintainers need to distinguish a broken layout from an expected missing cache
or baseline artifact when iterating on local host configuration.

## Verification

- Run `./bin/specs doctor` from the repo root.
- Confirm the output reports the resolved specs root, framework dir, graph
  manifest, and baseline file.
- Confirm missing derived artifacts are shown as missing without hiding the
  resolved layout.
