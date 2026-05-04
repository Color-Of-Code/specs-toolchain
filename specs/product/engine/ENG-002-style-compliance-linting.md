# Style Compliance Linting

| Field       | Value                                                                                     |
| ----------- | ----------------------------------------------------------------------------------------- |
| ID          | ENG-002                                                                                   |
| Status      | Draft                                                                                     |
| Stakeholder | Spec author, project lead                                                                 |
| Source      | [Overview](../../../docs/overview.md)                                                     |
| Realised By | [Style Rule Enforcement](../../model/requirements/lint/LNT-001-style-rule-enforcement.md) |

## Summary

Teams need the engine to check that spec files conform to a configurable
house style guide so that structural and stylistic problems are caught
automatically before review rather than through manual inspection.

## User Value

- Authors get immediate, precise feedback on broken link targets, missing
  required sections, and style violations without waiting for a reviewer.
- Project leads can enforce consistent artifact structure across the model
  by configuring rules once in `style.yaml`.
- CI gates block merging spec content that violates the agreed-upon style
  without requiring reviewers to memorise the full rule set.

## Acceptance Signal

`specs lint --style` reports each violation with file path and line number.
Rules can be tuned in the framework's `style.yaml`. `specs lint --links`
reports broken cross-references. Running `specs lint` without flags runs all
checks. Every check mode exits non-zero when violations are found.
