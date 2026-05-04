# Doctor Json

| Field        | Value                                                                                 |
| ------------ | ------------------------------------------------------------------------------------- |
| Status       | Draft                                                                                 |
| Requirements | [Repo Local Specs Host](../../requirements/workspace/WS-001-repo-local-specs-host.md) |

## Command or Contract

`specs doctor --json` emits the effective config path, resolved specs and host
roots, framework path and mode, key artifact paths, compatibility status, and
framework manifest details.

## Consumers

- The VS Code extension when it needs machine-readable environment details.
- Maintainers verifying repo-local development behavior from the terminal.

## Validation

Run `./bin/specs doctor --json` and confirm the payload resolves
`specs_root` to `specs/`, `framework_dir` to `templates/`, and
`framework_mode` to `local`.
