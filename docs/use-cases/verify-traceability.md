# Verify traceability links

## Summary

Check that every traceability edge in the spec corpus has a matching reverse
edge:

- A product requirement listed as `Realised By` a model requirement is
  reciprocally listed in that model requirement's `Realises` section.
- A model requirement listed as `Implemented By` a feature or component is
  reciprocally listed in that feature/component's `Requirements` section.

## Owner

**Reviewer** *(role)* — see [../roles.md](../roles.md). Typically the **Author**, **Analyst**, or **Architect** themselves run this check before submitting a CR for review:

- **Author** owns the product requirement ↔ model requirement links.
- **Analyst** owns the requirement ↔ feature links.
- **Architect** owns the feature ↔ component / service / API links.

## Purpose

Guarantee bidirectional traceability across the model. One-way links
silently rot; this check turns them into a CI failure.

## Entry point

`specs link check`

Or VS Code palette: **Specs: Check links**.

## Exit point

Zero exit on full symmetry. Non-zero with a list of asymmetric links
(missing back-reference, dangling target, wrong id) otherwise.

## Iteration

Fix the reported file(s), re-run `specs link check`. Pair with
[`specs visualize traceability`](visualize-traceability.md) when
diagnosing structural gaps rather than individual missing links.
