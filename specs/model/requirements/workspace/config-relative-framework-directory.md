# Config Relative Framework Directory

| Field          | Value                                                                                                                                 |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Status         | Draft                                                                                                                                 |
| Realises       | [Repo Local Specs Host](../../../product/toolchain/repo-local-specs-host.md)                                                          |
| Implemented By | [Repo Local Host Layout](../../features/workspace/repo-local-host-layout.md), [Specs Engine](../../components/engine/specs-engine.md) |

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
