---
id: specs_lint_workspace
name: Lint specs workspace
description: Run all lint checks on the specs model (links and style).
tags:
  - specs
inputSchema:
  type: object
  properties:
    check:
      type: string
      enum: [all, links, style]
      description: Which lint check to run. Defaults to all.
  additionalProperties: false
engineArgs:
  all: [lint]
  links: [lint, --links]
  style: [lint, --style]
---

Use this tool to lint the specs model in the current workspace.
Provide `check: "all"` (default) to run every check, or a specific value
to run only links or markdown style validation.
