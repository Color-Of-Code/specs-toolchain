# Update the framework content layer

## Summary

Refresh framework content (templates, process, skills, agents,
lint config) to a newer ref of the configured framework source.

## Owner

**Project owner** *(role)* — see [../ownership.md](../ownership.md). Picks up template fixes, new lint rules, or process updates published by a framework maintainer.

## Purpose

Pick up template fixes, new lint rules, or process updates published by
the framework maintainer without re-running the full setup.

## Entry point

`specs framework update [--to <ref>]`

Or VS Code palette: **Specs: Framework: Update**.

Pre-conditions: `framework_dir` points at a writable git checkout or
submodule; network access to the framework remote.

## Exit point

The framework checkout is fetched and optionally checked out to `--to <ref>`.
For submodule-based setups, the host repo records the updated submodule commit.

## Iteration

After updating, run [`specs doctor`](diagnose-environment.md) and
[`specs lint`](lint-and-format.md) to surface any new style rules or
template-schema changes; address findings before merging the bump.
