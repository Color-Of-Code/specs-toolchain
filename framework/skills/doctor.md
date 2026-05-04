---
id: specs_doctor
name: Diagnose specs environment
description: Run the specs doctor command to check engine version, binary availability, and workspace health.
tags:
  - specs
inputSchema:
  type: object
  properties: {}
  additionalProperties: false
engineArgs:
  default: [doctor]
---

Use this tool to diagnose the specs engine environment for the current workspace.
It verifies the binary is available, reports its version, and checks that the
workspace configuration is well-formed.
