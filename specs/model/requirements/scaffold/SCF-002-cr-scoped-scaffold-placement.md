# Change-Request Scoped Scaffold Placement

| Field          | Value                                                                                                         |
| -------------- | ------------------------------------------------------------------------------------------------------------- |
| ID             | SCF-002                                                                                                       |
| Status         | Draft                                                                                                         |
| Realises       | [Artifact Scaffolding From Templates](../../../product/engine/ENG-004-artifact-scaffolding-from-templates.md) |
| Implemented By | [CR-Scoped Placement](../../features/scaffold/SCF-002-cr-scoped-placement.md)                                 |

## Requirement

When `--cr <NNN>` is supplied, `specs scaffold` shall locate the
`CR-NNN-*` directory under `change-requests/`, normalise the id to three
digits, place the scaffolded file under that directory's `<kind>s/` subtree,
and return an error when no matching CR directory exists.

## Rationale

Placing work-in-progress artifacts directly into the correct CR subtree
ensures draft content never lands in the canonical model tree by accident,
removing the need for manual moves before a change request is drained.

## Verification

- Create a `CR-001-smoke` directory under `change-requests/`.
- Run `specs scaffold requirement --cr 1 core/test --dry-run` and confirm the
  printed path is inside `change-requests/CR-001-smoke/requirements/`.
- Run without `--dry-run` and confirm the file is created at that path.
- Run with a non-existent CR id and confirm an error is returned.
