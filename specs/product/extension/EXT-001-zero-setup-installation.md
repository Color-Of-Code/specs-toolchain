# Zero-Setup Installation

| Field       | Value                                                                                                                   |
| ----------- | ----------------------------------------------------------------------------------------------------------------------- |
| ID          | EXT-001                                                                                                                 |
| Status      | Draft                                                                                                                   |
| Stakeholder | VS Code user, spec author                                                                                               |
| Source      | [Install](../../../docs/install.md), [Extension README](../../../extension/README.md)                                   |
| Realised By | [Bundled Engine Resolution Priority](../../model/requirements/extension/VEXT-001-bundled-engine-resolution-priority.md) |

## Summary

VS Code users need the specs engine to be ready to use immediately after
installing the extension without performing a separate engine installation
step, so that the barrier to getting started is as low as possible.

## User Value

- Authors on a new machine can install one `.vsix` file and start using
  every specs command without installing Go, downloading binaries manually,
  or editing `PATH`.
- Teams can distribute the extension as the canonical onboarding step and
  be confident every member has a matching engine version.
- Users who already have a global `specs` binary can opt in to using it
  instead of the bundled one via a single setting, giving experienced users
  full control.

## Acceptance Signal

After installing the `.vsix`, running any palette command (e.g. `Specs:
Doctor`) in a workspace without a local `bin/specs` invokes the bundled
binary without any additional configuration. In workspaces that do contain a
local `bin/specs`, the extension prefers that local build so development
workflows match terminal usage. The `specs.enginePath` setting overrides the
other sources when set. The `specs.useGlobalBinary` setting makes the
extension prefer the `specs` binary found on `PATH`. The bundled binary is
platform-specific and ships for each supported platform as a separate `.vsix`.
