# Status Bar Workspace Indicator

| Field          | Value                                                                                            |
| -------------- | ------------------------------------------------------------------------------------------------ |
| ID             | VEXT-004                                                                                         |
| Status         | Draft                                                                                            |
| Realises       | [Workspace Status Visibility](../../../product/extension/EXT-004-workspace-status-visibility.md) |
| Implemented By | [Status Bar Indicator](../../features/extension/VEXT-004-status-bar-indicator.md)                |

## Requirement

The extension shall display a persistent status bar item that reflects the
engine-resolved state of the open workspace. The item shall show the detected
specs root path when a host is found, and a warning indicator when the engine
is missing or configuration is broken. The item shall be updated by
file change events on `.specs.yaml` and by window focus change events —
not by polling. The item shall be hidden when no workspace folder is
open. Clicking the item shall open the doctor output or the init wizard as
appropriate.

## Rationale

An event-driven indicator avoids the performance overhead of polling while
still providing up-to-date workspace health information. Exposing a
click-through action turns the indicator into a navigation shortcut rather
than a passive display.

## Verification

- Open a workspace with a valid `.specs.yaml` and confirm the status bar item
  appears with the specs root path.
- Temporarily remove `.specs.yaml` and confirm the item updates to a warning
  state.
- Restore the file and confirm the item recovers without restarting the
  extension.
- Close all workspace folders and confirm the item is hidden.
