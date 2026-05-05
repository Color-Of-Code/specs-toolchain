# Seed a brand-new framework

## Summary

Pre-create the minimal directory skeleton of a fresh framework
(`templates/`, `process/`, `skills/`, `agents/`) in an empty
directory, ready for the caller to author content and publish to a git
remote.

## Owner

**Framework maintainer** *(role)* — see [../ownership.md](../ownership.md). Authoring a bespoke framework that downstream host projects can consume. Not part of the authoring chain inside a host project.

## Purpose

Support organisations that need a bespoke framework rather than forking
the default one. This is an advanced, low-level operation: the seeded
tree is **not** managed by the toolchain after creation.

## Entry point

`specs framework seed --out <dir>`

Pre-conditions: `<dir>` does not exist or is empty.

## Exit point

The output directory contains the empty skeleton. The caller is
responsible for `git init`, pushing to a remote, and pointing hosts at
the framework via `specs init --framework <path-or-url>`.

## Workflow

1. Run `specs framework seed --out /path/to/my-framework`.
2. Author templates, process docs, skills, and agents.
3. `git init`, commit, push to a remote.
4. Use it: [`specs init --framework <remote-or-path>`](setup-host.md).

### Iteration

Once published, evolve the framework like any other repository.
Consumers pick up changes via
[`specs framework update`](update-framework.md).
