# Set up a host

## Summary

Create or configure a host repository for use with the specs toolchain in
a single command: `specs init`. It works whether the target directory is
empty, brand-new, or already contains `model/` and `change-requests/`
content.

## Owner

**Project owner** *(role)* — see [../ownership.md](../ownership.md). Stands up the host repo and chooses managed vs. local mode. Not part of the authoring chain.

## Purpose

Get from "I have a folder" to "I can run `specs lint` and start
authoring" with no manual config or copy-pasting from a sibling repo.
Existing model content is left untouched.

## Entry point

```text
specs init [<path>]
           [--framework <path-or-url>]
           [--with-model] [--with-vscode]
           [--force] [--dry-run]
```

`<path>` defaults to the current directory and is created if missing.
`--framework` accepts a local path or a remote git URL. When omitted,
`specs init` defaults to `./framework`.

## Exit point

A directory containing `.specs.yaml`, a framework directory (or `specs/.framework`
submodule when `--framework` is a URL), and (when requested) `model/`, `change-requests/`, and
`.vscode/tasks.json`. The directory is committable as-is. `--force`
overwrites an existing `.specs.yaml`; otherwise the command refuses.

## Workflow

1. Choose a framework source: local path or remote git URL.
2. Run `specs init --framework <path-or-url>` (or just `specs init`) with
   `--dry-run` first to preview the file plan.
3. Re-run without `--dry-run` to materialise the host.
4. Run [`specs doctor`](diagnose-environment.md) to confirm paths,
   versions, and framework resolution.

### Iteration

Re-run with `--force` to rewrite `.specs.yaml`, or edit the file
manually for individual key changes (framework source, repo mappings,
lint config path). Switching framework source is just a matter of
updating `framework_dir` (or re-running `specs init --force --framework <path-or-url>`).
