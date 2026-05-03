# Style Rule Enforcement

| Field          | Value                                                                                   |
| -------------- | --------------------------------------------------------------------------------------- |
| ID             | LNT-001                                                                                 |
| Status         | Draft                                                                                   |
| Realises       | [Style Compliance Linting](../../../product/engine/ENG-002-style-compliance-linting.md) |
| Implemented By | —                                                                                       |

## Requirement

`specs lint --style` shall check every markdown file against the rule set
compiled from the active framework's `style.yaml`, report each violation with
file path and line number, and exit non-zero when at least one violation is
found.

## Rationale

Precise file-and-line reporting lets authors fix violations without guessing
where problems are. Combining compiled-in defaults with a configurable
`style.yaml` means teams can tighten or relax rules without patching the
engine.

## Verification

- Introduce a deliberate style violation in a markdown file.
- Run `specs lint --style` and confirm a violation is reported with the
  correct file path and approximate line number.
- Remove the violation, rerun, and confirm the command exits zero.
