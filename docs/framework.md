# Framework

The framework layer provides the generic content a host project consumes:
templates, process docs, skills, agents, and lint configuration.

## Framework sources

A framework source is where that content comes from when `specs init` resolves a
host.

| Source kind | Example                                                               | Use it when                                                             |
| ----------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Remote URL  | `specs init --framework https://github.com/Color-Of-Code/specs-framework.git` | you want init to materialise `specs/.framework` as a submodule |
| Local path  | `framework_dir: ../specs-framework`                                   | you want a checkout, submodule, or vendored snapshot under your control |
| Empty seed  | `specs framework seed --out /path/to/my-framework`                    | you are creating a brand-new framework from scratch                     |

Seeded frameworks are not managed after creation; the caller owns `git init`,
publishing, and subsequent maintenance.

## Consumption model

Framework content is always consumed from a local directory (`framework_dir`).
That directory may be a regular checkout, a submodule, or a vendored snapshot.
`specs doctor` prints the resolved framework dir and mode so you can confirm
what was detected.

## Quick decision

| You are...                        | Use this                        |
| --------------------------------- | ------------------------------- |
| writing specs in a host project   | `specs/.framework` submodule or a local checkout |
| editing templates or process docs | local with an editable checkout |
| starting a brand-new framework    | seed, then switch to local      |
| working air-gapped                | local with a vendored snapshot  |

See [configuration.md](configuration.md) for the exact `.specs.yaml` details.
