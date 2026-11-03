# Changelog

## 0.0.20

- Follow herdr-app's startup. Hergo does not bundle Herdr. When Herdr is missing, the window can install it with the official installer or use an executable you choose. When Herdr is installed and the default session is not listening, Hergo starts `herdr server` and leaves it running after the window closes. An explicit socket or named session is left untouched. `-plain` and the Plain terminal startup choice still open a plain terminal.
- Replace the toolbar Changes label with a Git icon. Hide it, and the workspace menu item, when the workspace is not a Git repository. The Changes view lists the 100 most recent commits; choosing one shows that commit's changes against its first parent.
- Disable Git signature verification while loading commit history, so configured external verification programs do not run when opening Changes.

## 0.0.19

- Fix a Windows named-pipe listener shutdown race that could hang session cleanup and block release tests. Upgrade go-winio to v0.6.4 and run its listener-close regression tests in Windows CI.
- Include the Changes view and macOS Command-scroll fix from 0.0.18, whose release was blocked by the Windows test failure.

## 0.0.18

- Add a read-only Changes view for workspace changes relative to HEAD, including staged, unstaged, and untracked files. Open it from the toolbar, View menu, command palette, or workspace context menu.
- Browse continuous file diffs with unified and side-by-side layouts, syntax highlighting, search, manual refresh, and per-workspace view state. Load previews on demand and show notices or bounded patches for binary, non-UTF-8, and large files.
- On macOS, Command-scroll now scrolls terminal history without changing the font size. Keyboard font-size shortcuts continue to work.

Known limitation: in long file diffs, search can select a result outside the visible area without scrolling it into view. Physical Windows and Linux desktop validation remains outstanding.

## 0.0.17

- On macOS, `Cmd+Left`/`Cmd+Right` move to the start/end of the line, `Cmd+Backspace` deletes to its start, and `Option+Left`/`Option+Right` move by word, as in Ghostty. Applies to Herdr panes, plain terminals and command views; a Herdr binding on the same key wins.

## 0.0.16

- Fix macOS release tests by using a short temporary directory for Unix sockets.

## 0.0.15

- Detect whether Herdr is installed (PATH and the usual install directories). Without Herdr, the window explains how to install it and offers a plain terminal.
- Add a plain terminal mode: local shells in tabs, a quiet one-line strip, and a Herdr status chip that leads back to Herdr once it is installed. Settings → Herdr chooses Automatic, Herdr, or Plain terminal at startup; `-plain` and `-connect` override it once.
- Reorder workspaces and tabs by dragging them, or with Move Up/Down in the workspace menu (`workspace.move`, `tab.move`).
- Move a pane to a new tab, a new workspace, or another tab (`pane.move`).
- Send a prompt to an agent, get notified once when it stops, and start an agent in a pane (`agent.prompt`, `agent.wait`, `agent.start`).
- Filter the sidebar's agents by view (All, Needs attention, Working, This workspace) and share the view with Herdr's clients (`agent.view.set`).
- Save a tab's layout and open it again as a new tab (`layout.export`, `layout.apply`); saved layouts keep splits, ratios, names and directories, not environments.
- Optionally repeat Hergo's reminders in Herdr's terminal clients (`notification.show`).
- Settings → Herdr shows the installation, agent manifests with reload, saved layouts, and live handoff to a newly installed Herdr. The plugins page links and unlinks local plugins.
- Name the window after the selected workspace and tab.

## 0.0.14

- Add agent and process details, detection diagnostics, and independent agent naming.
- Add Herdr plugin management, session restore settings, saved SSH machines, and local session entry points.
- Add cancellable output matching reminders with substring and regular expression matching.
- Add official agent marks for Prime Agent, OMP, Droid, Qoder CLI, Hermes, Letta, and Mastra Code, with light and dark theme support.

Live SSH access, installed-plugin mutations, successful naming of a real coding agent, and Windows command surfaces remain unverified on physical systems. Physical Linux X11/Wayland desktop and IME validation remains outstanding, as does physical Windows desktop and IME validation.

## 0.0.13

- Follow Herdr theme configuration automatically, including light/dark switching and custom color overrides. Apply configuration edits to windows and terminals while the app is running.
- Show all themes as preview cards, with System, Hergo Light and Hergo Dark alongside Herdr built-in themes. Use the Follow Herdr switch to choose automatic or manual colors.
- Split Spaces and Agents into equal-height sidebar sections with fixed headers and independent scrolling, and hide their scrollbars.
- Improve default Herdr text contrast.

Physical Linux X11/Wayland desktop and IME validation remains outstanding, as does physical Windows desktop and IME validation.

## 0.0.12

- Rebrand the desktop application and platform packages as Hergo.
- Adopt a new ram app icon inspired by Herdr and mygo.
- Rewrite the README with installation, development, and release guidance, and adopt the MIT license.

Physical Linux X11/Wayland desktop and IME validation remains outstanding, as does physical Windows desktop and IME validation.

## 0.0.11

- Add native Linux desktop packages for amd64 and arm64, including Debian packages, portable archives, a user-local installer, and signed in-app updates.
- Test Linux daemon, PTY, native GTK window integration, and packaged terminal startup in the multi-platform Release workflow.
- Publish compatible architecture-specific macOS update manifests alongside the universal update manifest.

Physical Linux X11/Wayland desktop and IME validation remains outstanding, as does physical Windows desktop and IME validation.

## 0.0.10

- Publish macOS universal and Windows x64 packages together through one multi-platform Release workflow.
- Include the Windows installer, portable ZIP, SHA-256 checksums, and signed Windows update archives and manifest in GitHub Releases.
- Run both platforms' tests and packaging in parallel, and publish only after both succeed. Manual workflow runs build artifacts without publishing.

Physical Windows desktop and IME validation remains outstanding.

## 0.0.9

- Add agent integration settings to inspect, install, and remove integrations managed by Herdr.
- Show contextual agent notifications without numeric location labels, match terminal backgrounds to the window theme, and add a compact sidebar settings shortcut.
- Add native Windows named-pipe transport and terminal surface rendering with semantic keyboard, Unicode text, paste, mouse, and resize input.
- Improve native terminal split-pane geometry, wide grapheme rendering, and reconnect history handling, with automated Windows tests and packaging artifacts.

Windows packages remain available as workflow artifacts; this release publishes the universal macOS app and signed updates. Physical Windows desktop and IME validation remains outstanding.

## 0.0.8

- Draw agent marks in their brand colors in the sidebar, tabs, split cards, pane list, and command palette.
- Add marks for Pi, Devin, Kilo, Qwen, Antigravity, and Grok. GitHub Copilot uses its multicolor mark. Black ink turns white in a dark window.
- Keep the content panel's corner radius on terminal cards that meet those corners.
- Keep font changes in the settings preview until Apply. Keyboard zoom still applies immediately.
- Replace the sidebar connection chip with an Update to x.y.z button when a newer release is available.

## 0.0.7

- Add native automatic updates: check after launch and daily while the app runs.
- Add Check for Updates to the application menu and an Updates page in Settings.
- Support release notes, download progress, skipping a version, and optional automatic installation on the next launch.
- Verify update archives with Ed25519 signatures and publish signed universal macOS update archives and manifests.

Users of v0.0.6 and earlier must install this version manually once to enable future in-app updates.
