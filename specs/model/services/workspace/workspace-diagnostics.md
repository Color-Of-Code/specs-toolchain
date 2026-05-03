# Workspace Diagnostics

| Field        | Value                                                                          |
| ------------ | ------------------------------------------------------------------------------ |
| Status       | Draft                                                                          |
| Requirements | [Repo Local Specs Host](../../requirements/workspace/repo-local-specs-host.md) |

## Responsibilities

Resolve the current workspace configuration and report the effective specs
paths, framework mode, and compatibility information for local development and
support workflows.

## Inputs and Outputs

- Inputs: current working directory, `.specs.yaml`, local framework manifest,
  filesystem state.
- Outputs: human-readable `specs doctor` output and stable JSON for tooling.

## Operational Notes

Relative paths are anchored to `specs_root`, local framework directories are
read as framework roots, and missing graph or baseline files are reported
without misclassifying the layout itself.
