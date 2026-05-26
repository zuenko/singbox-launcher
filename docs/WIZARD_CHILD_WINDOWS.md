# Wizard child windows: contract and overlay

The wizard can have **child windows** open on top of the main window (Config Wizard). To keep UX consistent and avoid multiple copies of the same window, all child windows follow one contract and share a single overlay.

## Child window types

1. **Rule dialogs** — Add/Edit Rule (from Rules tab). Multiple rule dialogs are not allowed at once; opening a new one closes any existing.
2. **View** — “Servers” window for one source (Sources tab, View button). Only one View window at a time.
3. **Outbound Edit/Add** — window to add or edit an outbound (Outbounds tab). Only one Edit window at a time.

## Contract

### When opening a child window

- **Register** the window with the presenter:
  - Rule dialogs: add to `presenter.OpenRuleDialogs()` map (key = rule index or -1 for add).
  - View: `presenter.SetViewWindow(win)`.
  - Outbound Edit: `presenter.SetOutboundEditWindow(win)` (via `OutboundEditPresenter`).
- Call **`presenter.UpdateChildOverlay()`** so the overlay is shown.

### When closing a child window

- **Unregister**: remove from the map (rule), or call `ClearViewWindow()` / `ClearOutboundEditWindow()`.
- Call **`presenter.UpdateChildOverlay()`** so the overlay is hidden when no child is left.

### Single-instance (View, Outbound Edit)

- Before creating a new window, check if one is already open (`OpenViewWindow()` / `OpenOutboundEditWindow()`).
- If yes, call **`RequestFocus()`** on that window and do not create a new one.

### Click on wizard while a child is open

- The main wizard content is covered by **ChildWindowsOverlay** (transparent, tappable).
- On tap, **FocusOpenChildWindows** (UIService) is invoked; it focuses one open child in this order: View → Outbound Edit → any rule dialog.
- So the user can bring the child window to front by clicking on the wizard.

## Implementation details

- **Overlay**: `GUIState.ChildWindowsOverlay` (created in `wizard.go`, `components.NewClickRedirect`). Show/hide via `UpdateChildOverlay()`.
- **Focus callback**: `UIService.FocusOpenChildWindows` is set in `wizard.go` and used in `click_redirect.go` when the user taps the overlay.
- Rule dialogs use `SetCloseIntercept` to unregister and call `UpdateChildOverlay` on close. View and Edit use `SetOnClosed`.

See: `ui/wizard/presentation/presenter.go` (SetViewWindow, ClearViewWindow, SetOutboundEditWindow, ClearOutboundEditWindow, UpdateChildOverlay), `ui/wizard/wizard.go` (overlay creation and FocusOpenChildWindows).

---

## Main-window overlay (separate, opt-in)

Historically the **main launcher window** had its own `components.ClickRedirect` overlay (created by `ui.InitWizardOverlay` in `ui/wizard_overlay.go`) that intercepted every click on the main window while the wizard was open and refocused the wizard. Effect: main window became read-only — Update / Restart / Servers tab / Start / Stop all silently swallowed clicks.

Since v0.9.8 this is **gated by a build-time constant**:

```go
// ui/wizard_overlay.go
const wizardOverlayEnabled = false  // default: main window stays usable
```

- `false` (current default) — `InitWizardOverlay` early-returns; `app.content` is just `app.tabs` (no overlay layer); the user can drive Update, Restart, start/stop sing-box, switch tabs etc. while the configurator is open as a parallel sibling window.
- `true` — restores the legacy behavior: main-window overlay layered on top of tabs, every click forwards focus to the wizard.

Flip the constant + rebuild to switch. Nothing else needs to change — the wizard's *internal* `ChildWindowsOverlay` (above) is independent and works the same in both modes.

**Why opt-in instead of removed:** the focus-the-wizard semantic is useful for some users (the launcher feels like it has a modal dialog open). Keeping the implementation around makes it a one-line flip if we ever decide to bring it back as a Settings opt-in.
