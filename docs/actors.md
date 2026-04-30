# Actors

The toolchain is built around a short, linear authoring chain. Four actors
cover everything the engine is concerned with; one person may hold several
of these roles in the same project.

Setup, review, and framework-distribution work happens **outside** this
chain and is described as *operational roles* in [roles.md](roles.md).

```text
Stakeholder ──► Author ──► Analyst ──► Architect
 product       requirements   features    components / services / APIs
 requirement
```

Each actor refines the work of the previous one. Their output lives in
canonical locations — product requirements under `product/`, the rest under
`model/` — drafted inside change requests.

## Stakeholder

Produces **product requirements** (PRs): what is being asked for, in the
stakeholder's own vocabulary. PRs live under `product/` once a CR is drained.

- Opens a change request: `specs cr new --id <NNN> --slug <slug>`.
- Scaffolds a product requirement:
  `specs scaffold product-requirement --cr <NNN> <path>`.
- Writes prose describing the demand inside `## Description`; lists the
  realising model requirements under `## Realised By` once the Author has
  written them.

## Author

Re-formulates product requirements as properly formed **model requirements**
(MRs): single, testable statements with stable IDs that the rest of the
model can hang implementation off. The Author owns wording, IDs, and scope
of each MR.

- Scaffolds a requirement: `specs scaffold requirement --cr <NNN> <path>`.
- Lists the originating PRs under `## Realises` and links each MR back from
  the PR's `## Realised By` section.
- Runs `specs format` and `specs lint` while iterating.

## Analyst

Derives **features** from approved requirements. The analyst groups related
requirements into features and ensures every requirement is implemented by
at least one feature.

- Scaffolds a feature: `specs scaffold feature --cr <NNN> <path>`.
- Lists implementing features under each requirement and the requirements
  under each feature.
- Runs `specs link check` to confirm the bidirectional links are symmetric.

## Architect

Decomposes features into **components, services, and APIs**. The architect
also keeps component baselines aligned with their upstream repositories.

- Scaffolds artifacts:
  `specs scaffold component|service|api --cr <NNN> <path>`.
- Maintains the components table: `specs baseline update`.
- Verifies traceability: `specs link check`,
  `specs visualize traceability`.

## How they work together

1. **Stakeholder** opens a CR and drafts product requirements (PRs) in it.
2. **Author** re-formulates those PRs as model requirements (MRs) inside
   the same CR, populating `## Realises` and `## Realised By` to keep the
   PR ↔ MR traceability symmetric.
3. **Analyst** adds features that implement those MRs.
4. **Architect** adds components, services, and APIs implementing the
   features.
5. Anyone runs `specs format`, `specs lint`, and `specs link check` to
   validate the CR.
6. The CR is drained (`specs cr drain --id <NNN>`) and merged: PR files
   move into `product/`, model files into their canonical homes under
   `model/`.

Setup tasks (installing the engine, initialising a host repo, managing
the framework registry) are one-time prerequisites and are not part of
the authoring chain. See [overview.md](overview.md) and
[install.md](install.md).
