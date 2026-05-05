# Glossary

Core vocabulary used throughout the toolchain docs. These definitions describe
the artifact kinds, paths, and framework terms that appear most often.

## Change request

A numbered work area under `change-requests/` where product requirements and
model artifacts are drafted before being drained into their canonical homes.

## Component

An implementation unit under `model/components/`, typically pinned to an
upstream repository.

## Use case

A model artifact under `model/use-cases/` describing an end-to-end interaction
scenario that satisfies one or more requirements.

## Framework dir

The directory holding the materialised framework content: `templates/`,
`process/`, `skills/`, `agents/`, and lint configuration. This is the
directory named by `framework_dir`.

## Framework source

The origin from which framework content is obtained. A framework source is
either a remote git URL, a local directory path, or an empty seeded skeleton.

## Host repo

The git repository that contains the specs root. It can be the specs root
itself or a larger repository that contains the specs root as a subdirectory or
submodule.

## Model artifact

Any artifact stored under `model/`: requirement, use case, or component.

## Model requirement

See [Requirement](glossary.md#requirement-model-requirement-technical-requirement).

## Product requirement

A stakeholder-facing artifact under `product/` describing what was asked for in
the stakeholder's vocabulary. Product requirements are realised by one or more
requirements.

## Requirement (model requirement; technical requirement)

A model artifact under `model/requirements/` that re-formulates one or more
product requirements into a precise, testable statement.

This is the artifact sometimes called a model requirement or technical
requirement. Across the docs, `requirement` is the usual short form unless the
product requirement versus requirement distinction needs to be explicit.

## Specs root

The directory that contains `.specs.yaml`, `model/`, `product/`, and
`change-requests/`. This is the root operated on by `specs`.

## Technical requirement

See [Requirement](glossary.md#requirement-model-requirement-technical-requirement).
