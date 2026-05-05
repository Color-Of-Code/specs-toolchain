---
id: DOC-001
requirements:
    - ../../requirements/workspace/WS-004-workspace-diagnostics-reporting.md
---

# Human Diagnostics Output

## Workflow

Diagnose the repo-local host layout from the repo root and separate resolved
path information from missing derived artifacts such as caches.

## Engine Surface

- `specs doctor` reports specs root, framework dir, manifest, graph manifest,
  and graph cache paths.
- Missing graph cache or similar derived files remain warnings, not layout
  failures.

## VS Code Surface

- Support and maintenance workflows can rely on the same diagnostics behavior
  when surfaced through editor commands.
- Repo-local troubleshooting uses the same underlying workspace report as the
  terminal flow.

## Validation

Run `./bin/specs doctor` and confirm resolved paths are reported even when the
graph cache is still missing.
