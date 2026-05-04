---
id: WS-002
status: Draft
realises:
    - ../../../product/engine/ENG-009-repo-local-specs-host.md
implemented_by:
    - ../../components/engine/specs-engine.md
    - ../../use-cases/host/HOST-002-config-relative-framework.md
---

# Config Relative Framework Directory

## Requirement

The engine shall resolve an explicit `framework_dir` relative to the directory
containing `.specs.yaml` so a host can use `framework_dir: ./framework` while
also setting `specs_root: ./specs`.

## Rationale

Anchoring `framework_dir` to `specs_root` makes repo-local framework layouts
surprising and forces host repositories to use path traversal in config for a
framework directory that sits next to `.specs.yaml`.

## Verification

- Set `specs_root: ./specs` and `framework_dir: ./framework` in `.specs.yaml`.
- Run `./bin/specs doctor` from the repo root.
- Confirm the reported framework dir resolves to the repo-local `framework/`
  directory.
