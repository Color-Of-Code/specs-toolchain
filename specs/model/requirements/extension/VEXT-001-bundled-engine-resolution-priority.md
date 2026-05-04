# Bundled Engine Resolution Priority

| Field          | Value                                                                                       |
| -------------- | ------------------------------------------------------------------------------------------- |
| ID             | VEXT-001                                                                                    |
| Status         | Draft                                                                                       |
| Realises       | [Zero-Setup Installation](../../../product/extension/EXT-001-zero-setup-installation.md)    |
| Implemented By | [Bundled Engine Resolution](../../features/extension/VEXT-001-bundled-engine-resolution.md) |

## Requirement

The extension shall resolve the `specs` binary from four sources in
descending priority: an explicit `specs.enginePath` setting, the `specs`
binary found on `PATH` when `specs.useGlobalBinary` is true, a
workspace-local `bin/specs` build when present, and the platform-specific
binary bundled inside the extension. The resolved path shall be used for
every command invocation.

## Rationale

A fixed priority order guarantees deterministic binary selection across
diverse machine configurations. Bundling a platform-specific binary as the
final fallback achieves zero-setup for new users while giving experienced
users full control through the settings.

## Verification

- Install the extension without any settings and confirm commands invoke the
  bundled binary when the workspace does not contain `bin/specs`.
- Open a workspace with `bin/specs` and confirm the local binary is
  preferred.
- Set `specs.useGlobalBinary = true` and confirm the PATH binary takes
  precedence over the local build.
- Set `specs.enginePath` and confirm it takes precedence over all other
  sources.
