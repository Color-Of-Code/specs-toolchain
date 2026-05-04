# Config Relative Framework Resolution

| Field        | Value                                                                                                             |
| ------------ | ----------------------------------------------------------------------------------------------------------------- |
| ID           | HOST-002                                                                                                          |
| Status       | Draft                                                                                                             |
| Requirements | [Config Relative Framework Directory](../../requirements/workspace/WS-002-config-relative-framework-directory.md) |

## Workflow

Resolve a repo-local framework directory from the directory containing
`.specs.yaml` even when the specs content tree lives in a nested `specs/`
folder.

## Engine Surface

- `config.Load` reads `.specs.yaml` and resolves `framework_dir` against the
  config file location.
- `specs doctor` reports the resolved framework dir and manifest.
- `specs scaffold` loads templates from the same resolved framework path.

## VS Code Surface

- The extension inherits the same framework resolution behavior through the
  engine.
- Repo-local command invocations can keep `framework/` next to `.specs.yaml`
  without extra path traversal.

## Validation

Set `framework_dir: ./framework` in `.specs.yaml`, then run `./bin/specs
doctor` and one `./bin/specs scaffold ... --dry-run` command.
