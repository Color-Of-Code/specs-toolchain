---
id: WS-004
realises:
    - ../../../product/engine/ENG-008-environment-diagnostics.md
implemented_by:
    - ../../use-cases/doctor/DOC-001-human-diagnostics.md
---

# Workspace Diagnostics Reporting

## Requirement

The diagnostics workflow shall report the effective repo-local specs root,
framework directory, and related artifact paths without misclassifying missing
graph cache files as a layout failure.

## Rationale

Maintainers need to distinguish a broken layout from an expected missing cache
or derived artifact when iterating on local host configuration.

## Verification

- Run `./bin/specs doctor` from the repo root.
- Confirm the output reports the resolved specs root, framework dir, graph
  manifest, and graph cache file.
- Confirm missing derived artifacts are shown as missing without hiding the
  resolved layout.
