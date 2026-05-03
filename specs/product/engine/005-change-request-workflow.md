# Change Request Workflow

| Field       | Value                                                                    |
| ----------- | ------------------------------------------------------------------------ |
| Status      | Draft                                                                    |
| Stakeholder | Spec author, project lead                                                |
| Source      | [Overview](../../../docs/overview.md), [Actors](../../../docs/actors.md) |
| Realised By | —                                                                        |

## Summary

Teams need an isolated, numbered draft area for proposing, reviewing, and
integrating sets of related spec changes so that work-in-progress content
is kept separate from the canonical model until it has been reviewed and
approved.

## User Value

- Authors can iterate on a set of related requirements, features, and
  components inside a change-request folder without affecting the shared
  model visible to other team members.
- Project leads can review a self-contained change request in a single git
  diff before it is promoted to the canonical model.
- Once a change request is approved, a single drain command moves each file
  to its canonical home without requiring manual `git mv` operations.

## Acceptance Signal

`specs cr new` creates a numbered change-request directory from the framework
template. `specs cr status` lists all change requests with per-area file
counts. `specs cr drain --id <NNN>` interactively moves each CR-local file
to its canonical model path using `git mv`. `--dry-run` is available on every
write command. Change-request directories follow the `CR-NNN-slug` naming
convention.
