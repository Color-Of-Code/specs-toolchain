---
id: WS-005
status: Draft
realises:
    - ../../../product/engine/ENG-008-environment-diagnostics.md
implemented_by:
    - ../../use-cases/doctor/DOC-002-json-diagnostics-integration.md
---

# Doctor Json Contract

## Requirement

The machine-readable doctor output shall expose the resolved repo-local specs
and framework paths, mode information, and compatibility fields needed by the
VS Code extension and local tooling.

## Rationale

The extension and other tooling cannot depend on human-oriented doctor text if
they need a stable contract for local host diagnostics.

## Verification

- Run `./bin/specs doctor --json`.
- Confirm the payload includes the resolved `specs_root`, `framework_dir`,
  `framework_mode`, and compatibility information.
- Confirm the reported framework dir matches the repo-local `framework/`
  directory.
