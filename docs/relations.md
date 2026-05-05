# Relations in SysML

The source and target columns below follow the actual SysML dependency
direction: client/dependent element → supplier/independent element. Most
arrows therefore point opposite to the top-down requirement flow people often
expect.

| Source                                          | Target                                             | Standard name                       | Purpose                                                                                                                                          |
| ----------------------------------------------- | -------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Composite requirement                           | Child requirement                                  | Containment (namespace containment) | Builds a requirement hierarchy by decomposing a broader requirement into sub-requirements.                                                       |
| Derived requirement                             | Source requirement                                 | deriveReqt                          | Captures a requirement derived by analysis from another requirement, typically at a lower level or a more detailed abstraction.                  |
| Refining model element, or refining requirement | Refined requirement or other refined model element | refine                              | Clarifies or elaborates meaning/context. Unlike deriveReqt, it is not limited to requirement-to-requirement links.                               |
| Design or implementation element                | Requirement                                        | satisfy                             | States which design element is intended to satisfy a requirement. This is an allocation/assertion, not proof.                                    |
| Test case or other named element                | Requirement                                        | verify                              | States how a requirement is verified, for example by test, inspection, analysis, or demonstration.                                               |
| Copied requirement                              | Original/source requirement                        | copy                                | Reuses a requirement in another namespace/context while keeping the copied text read-only from the original.                                     |
| Any tracing model element, often a requirement  | Any traced model element                           | trace                               | Generic weak traceability link when no more specific relation fits. Often used to connect requirements to source documentation or related specs. |

Two useful clarifications:

1. deriveReqt is the correct SysML standard name for the derivation relation. If you say “derive” informally, that is understandable, but the standard keyword is deriveReqt.
2. Containment is a special case in this list: it is structural ownership of a sub-requirement by a composite requirement, not one of the stereotyped dependency names like deriveReqt or satisfy.

For the toolchain's artifact vocabulary, see [model.md](model.md) and
[glossary.md](glossary.md).
