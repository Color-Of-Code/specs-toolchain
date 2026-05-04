---
id: EXT-004
status: Draft
stakeholder: Spec author
source: "[Extension README](../../../extension/README.md)"
realised_by:
    - ../../model/requirements/extension/VEXT-004-status-bar-workspace-indicator.md
---

# Workspace Status Visibility

## Summary

Spec authors need a persistent, at-a-glance indicator in the VS Code status
bar that reflects the health of the current specs workspace so that they
are always aware of whether the open project is a recognised specs host and
whether the engine is available.

## User Value

- Authors working across multiple workspaces know immediately whether the
  currently active window contains a specs host, without running a command.
- Problems such as a missing or misconfigured engine are surfaced in the
  status bar as soon as the window opens, rather than only when the first
  command is attempted.
- The indicator updates automatically when relevant files change, so it
  stays accurate without requiring manual refresh.

## Acceptance Signal

A status bar item is visible whenever a workspace folder is open. It
reflects the engine-resolved state of the workspace (e.g. specs root
detected, engine missing, version information). The item is updated by
`FileSystemWatcher` events on `.specs.yaml` and on window focus changes,
not by polling. Clicking the item provides a relevant action (e.g. opens
the doctor output or the init wizard).
