# `.specs.yaml` reference

Lives next to the specs root.

## Framework directory

Point `framework_dir` at the framework directory on disk:

```yaml
framework_dir: .framework # or ../framework, or any absolute path
min_specs_version: 0.1.0
repos: ...
```

The directory can be a regular git checkout, a submodule, or a vendored snapshot — the toolchain treats them identically and never modifies the directory itself.

If `framework_dir` is omitted, the engine auto-detects `.framework` under the specs root (or host root when specs lives in a subdirectory).

## Optional knobs

| Key                   | Purpose                                                 |
| --------------------- | ------------------------------------------------------- |
| `change_requests_dir` | Override the default `change-requests/` path.           |
| `model_dir`           | Override the default `model/` path.                     |
| `product_dir`         | Override the default `product/` path.                   |
| `style_config`        | Path to a custom `style.yaml` for markdown style rules. |
| `templates_schema`    | Integer schema version expected by the engine.          |

Defaults are sensible; only set these when overriding.

### `style.yaml` schema

The `style.yaml` file controls markdown style checking with abstract, tool-independent rule names. If no `style_config` is set, the toolchain looks for `<framework_dir>/lint/style.yaml`, falling back to compiled-in defaults.

```yaml
rules:
  line_length: false # false (disabled) or integer max
  inline_html: false # allow inline HTML
  first_heading_h1: false # require first line to be h1
  heading_style: atx # "atx" (# Heading) or "setext"
  blank_lines_around_headings: true
  blank_lines_around_fences: true
  list_marker: dash # "dash" (-) or "asterisk" (*)
  no_trailing_whitespace: true
  no_consecutive_blank_lines: true
  fenced_code_language: true # require language on fenced code blocks
```

Only include keys you want to override — omitted keys use the compiled-in defaults.

