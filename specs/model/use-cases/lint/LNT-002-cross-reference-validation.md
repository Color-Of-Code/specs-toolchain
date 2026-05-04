# Cross-Reference Validation

| Field        | Value                                                                                  |
| ------------ | -------------------------------------------------------------------------------------- |
| ID           | LNT-002                                                                                |
| Status       | Draft                                                                                  |
| Requirements | [Link Target Checking](../../requirements/lint/LNT-002-link-target-checking.md)        |

## Workflow

Collect every markdown link target in the scanned files, resolve each target
relative to the source file, and report any path that does not exist on disk.

## Engine Surface

- `specs lint --links` activates this check category.
- Both relative paths and anchor-only fragments are evaluated.
- Each broken link is printed as `<file>:<line>: broken link: <target>`.
- Exit code is non-zero when at least one broken link is found.

## Validation

Add a link to a non-existent file. Run `specs lint --links` and confirm it
is reported with the correct source location. Remove the broken link and
confirm the command exits zero.
