# Status Bar Indicator

| Field        | Value                                                                                                     |
| ------------ | --------------------------------------------------------------------------------------------------------- |
| ID           | VEXT-004                                                                                                  |
| Status       | Draft                                                                                                     |
| Requirements | [Status Bar Workspace Indicator](../../requirements/extension/VEXT-004-status-bar-workspace-indicator.md) |

## Workflow

Display a persistent status bar item that reflects the engine-resolved state
of the open workspace. Update it whenever `.specs.yaml` changes or the window
gains focus, using `FileSystemWatcher` and `onDidChangeWindowState` — no
polling.

## VS Code Surface

- The item shows the detected specs root path when a host is found.
- When the engine is missing or configuration is broken, the item shows a
  warning indicator.
- Clicking the item opens the doctor output or the init wizard as appropriate.
- The item is hidden when no workspace folder is open.

## Validation

Open a workspace with a valid `.specs.yaml`. Confirm the status bar item
appears. Temporarily remove `.specs.yaml` and confirm the item updates to a
warning state. Restore the file and confirm the item recovers without
restarting the extension.
