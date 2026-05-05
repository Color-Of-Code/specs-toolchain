# Roles

Operational hats people put on **outside** the authoring chain — setup, review, and framework distribution. For artifact ownership inside the authoring chain (Stakeholder, Author, Analyst, Architect), see [actors.md](actors.md).

For the short ownership map, start with [ownership.md](ownership.md). This
page is the longer role-by-role detail.

One person typically wears several roles in the same repository.

| Role                 | Responsibility                                                                          | Typical commands                                                |
| -------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Any user             | Get the engine working locally                                                          | `specs doctor`                                                   |
| Project owner        | Stand up a host repository, choose framework source, and seed editor tasks              | `specs init`, `specs vscode init`, `specs framework update`     |
| Reviewer             | Confirm a proposed change request is structurally sound and traceable before it lands   | `specs graph validate`, `specs visualize traceability`          |
| Framework maintainer | Create, evolve, and distribute framework content for downstream host repos to consume   | `specs framework seed`; publishes a framework repo              |

## Notes

- **Any user** owns their own machine's installation. There is no separate platform-admin role.
- **Project owner** chooses a framework source at init time: local path or remote URL. URL mode materialises `specs/.framework` as a submodule.
- **Framework maintainer** is also a project owner and reviewer on the framework's own repository; the role here describes only the *downstream-facing* responsibility of publishing framework content.
