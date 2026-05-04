# Repo Local Specs Host

| Field        | Value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID           | ENG-009                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Status       | Draft                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Stakeholder  | specs-toolchain maintainer                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Source       | [Development guide](../../../docs/development.md), [Configuration reference](../../../docs/configuration.md)                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Realised By  | [Repo Local Specs Host](../../model/requirements/workspace/WS-001-repo-local-specs-host.md), [Config Relative Framework Directory](../../model/requirements/workspace/WS-002-config-relative-framework-directory.md), [Repo Local Engine Integration](../../model/requirements/workspace/WS-003-repo-local-engine-integration.md)                                                                                                                                                                                                             |

## Summary

The specs-toolchain repository shall contain its own specs host under `specs/`
and its own local framework under `framework/` so maintainers can develop and
verify host behavior without cloning or wiring a second repository.

## User Value

- Maintainers can run `specs doctor`, `specs scaffold`, and graph commands
	against a real host layout inside this repo.
- Local changes to framework templates, path resolution, and development
	scripts stay reviewable in one workspace.

## Acceptance Signal

`./bin/specs doctor` resolves `specs_root` to `./specs`, `framework_dir` to
`./framework`, and scaffold commands create artifacts inside the repo-local
host tree.
