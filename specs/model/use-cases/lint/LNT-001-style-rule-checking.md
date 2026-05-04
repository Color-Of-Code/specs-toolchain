# Style Rule Checking

| Field        | Value                                                                                    |
| ------------ | ---------------------------------------------------------------------------------------- |
| ID           | LNT-001                                                                                  |
| Status       | Draft                                                                                    |
| Requirements | [Style Rule Enforcement](../../requirements/lint/LNT-001-style-rule-enforcement.md)      |

## Workflow

Load the active rule set (compiled-in defaults merged with the framework's
`style.yaml`), walk the targeted markdown files, evaluate each rule, and
report every violation with file path and line number.

## Engine Surface

- `specs lint --style` activates this check category.
- Rules in `framework/lint/style.yaml` override or augment compiled defaults.
- Each violation is printed as `<file>:<line>: <rule>: <message>`.
- Exit code is non-zero when at least one violation is found.

## Validation

Introduce a known style violation. Run `specs lint --style` and confirm the
violation is reported with the correct file path and line. Fix the violation
and confirm the command exits zero.
