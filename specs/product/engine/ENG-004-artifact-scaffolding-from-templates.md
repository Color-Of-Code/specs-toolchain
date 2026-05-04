# Artifact Scaffolding From Templates

| Field       | Value                                                                                                                                                                                                                                              |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ID          | ENG-004                                                                                                                                                                                                                                            |
| Status      | Draft                                                                                                                                                                                                                                              |
| Stakeholder | Spec author, stakeholder                                                                                                                                                                                                                           |
| Source      | [Commands](../../../docs/commands.md), [Actors](../../../docs/actors.md)                                                                                                                                                                           |
| Realised By | [Template-Based Artifact Instantiation](../../model/requirements/scaffold/SCF-001-template-based-artifact-instantiation.md), [Change-Request Scoped Scaffold Placement](../../model/requirements/scaffold/SCF-002-cr-scoped-scaffold-placement.md) |

## Summary

Authors and stakeholders need to instantiate new spec artifacts from
framework-provided templates without copying files manually or remembering
the canonical directory layout, so that new artifacts start with the correct
structure and land in the right location.

## User Value

- Authors can scaffold a requirement with a single command and get a
  fully-structured file with the correct heading, field table, and section
  stubs in the right model subdirectory.
- Stakeholders can scaffold product requirements inside a change request
  without knowing where the canonical paths are.
- Framework templates are used consistently across the team, preventing
  structural drift between artifacts created by different people.

## Acceptance Signal

`specs scaffold <kind> <path>` creates a file populated from the matching
framework template with its H1 derived from the path slug. The `--title`
flag overrides the derived heading. The `--cr <NNN>` flag redirects the
output into the matching change-request subtree. `--dry-run` prints the
target path without writing. `--force` allows overwriting an existing file.
