# Framework

The framework layer provides the generic content a host project consumes:
templates, process docs, skills, agents, and lint configuration.

## Framework sources

A framework source is where that content comes from when `specs init` resolves a
host.

| Source kind | Example                                                               | Use it when                                                             |
| ----------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Remote URL  | `framework_url: https://github.com/Color-Of-Code/specs-framework.git` | you want the default managed workflow                                   |
| Local path  | `framework_dir: ../specs-framework`                                   | you want a checkout, submodule, or vendored snapshot under your control |
| Empty seed  | `specs framework seed --out /path/to/my-framework`                    | you are creating a brand-new framework from scratch                     |

Seeded frameworks are not managed after creation; the caller owns `git init`,
publishing, and subsequent maintenance.

## Framework registry

The framework registry maps short names to framework sources so users do not
have to remember raw URLs or paths.

On Linux the registry lives at `~/.config/specs/frameworks.yaml`; other
platforms use their standard application-data locations.

```yaml
frameworks:
  default:
    url: https://github.com/Color-Of-Code/specs-framework.git
    ref: v1.0.0
  acme:
    url: https://git.example.com/acme/specs-framework.git
    ref: main
  local-dev:
    path: ~/src/specs-framework
```

Relevant commands:

```bash
specs framework list
specs framework add <name> --url <URL> [--ref <ref>]
specs framework add <name> --path <dir>
specs framework remove <name>
specs framework seed --out <dir>
```

## Consumption modes

Once a framework source is resolved, a host uses it in one of two ways.

| Mode    | Where it lives                                      | Who updates it                                                   | Typical use                          |
| ------- | --------------------------------------------------- | ---------------------------------------------------------------- | ------------------------------------ |
| managed | user cache, shared across hosts on the same machine | the engine via `specs framework update`                          | ordinary specs authoring             |
| local   | a directory you point `framework_dir` at            | the user via `git pull`, submodule update, or vendoring workflow | framework editing or air-gapped work |

`specs doctor` prints the resolved framework dir and mode so you can confirm
what was detected.

## Quick decision

| You are...                        | Use this                        |
| --------------------------------- | ------------------------------- |
| writing specs in a host project   | managed                         |
| editing templates or process docs | local with an editable checkout |
| starting a brand-new framework    | seed, then switch to local      |
| working air-gapped                | local with a vendored snapshot  |

See [configuration.md](configuration.md) for the exact `.specs.yaml` and
registry schema details.
