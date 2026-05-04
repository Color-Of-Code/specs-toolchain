# Verify traceability graph

## Summary

Check that the canonical traceability graph is internally valid and points at
real artifacts:

- Every canonical node id resolves to a markdown artifact in the repo.
- Every relation family uses valid source and target kinds.
- Every baseline entry points at a configured repo mapping.

## Owner

**Reviewer** *(role)* — see [../roles.md](../roles.md). Typically the **Author**, **Analyst**, or **Architect** themselves run this check before submitting a CR for review:

- **Author** owns the product requirement ↔ model requirement links.
- **Analyst** owns the requirement ↔ use case (satisfy) links.
- **Architect** owns the requirement ↔ component (refine) links.

## Purpose

Guarantee that the canonical traceability graph is well-formed before review
or release automation consumes it.

## Entry point

`specs graph validate`

## Exit point

Zero exit on a valid graph. Non-zero with the first validation error
(missing artifact, invalid node id, invalid relation endpoint, unknown repo,
and similar graph-integrity failures) otherwise.

## Iteration

Fix the reported graph or artifact file(s), re-run `specs graph validate`. Pair with
[`specs visualize traceability`](visualize-traceability.md) when
diagnosing structural gaps rather than a specific validation failure.
