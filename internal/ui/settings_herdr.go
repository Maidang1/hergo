package ui

import (
	"errors"
	"strings"

	"github.com/Maidang1/hergo/internal/app"
	"github.com/Maidang1/hergo/internal/design"
	"github.com/Maidang1/hergo/internal/herdr"
	"github.com/Maidang1/hergo/internal/icons"
	"github.com/Maidang1/hergo/internal/session"
	native "github.com/egoist/mygo/ui"
)

type manifestSettings struct {
	loaded, busy bool
	entries      []herdr.AgentManifest
	message      string
	pluginPath   string
	handoff      string
	updating     bool
	update       string
}

var (
	startupModes      = []string{"auto", "herdr", "plain"}
	startupModeLabels = []string{"Automatic", "Herdr", "Plain terminal"}
)

func (w *Window) loadManifests(reload bool) {
	s := &w.manifests
	if s.busy || w.extra.Manifests == nil {
		return
	}
	s.loaded, s.busy, s.message = true, true, ""
	w.extra.Manifests(reload, func(entries []herdr.AgentManifest, err error) {
		s.busy = false
		if err != nil {
			s.message = err.Error()
			return
		}
		s.entries = entries
		if reload {
			s.message = "Agent manifests reloaded."
		}
	})
}

// settingsHerdr is the page about Herdr itself: whether it is installed,
// how the app opens, and the server-wide tools.
func (w *Window) settingsHerdr(c *native.Context, p design.Palette) {
	install := herdr.Installation{}
	if w.extra.Installation != nil {
		install = w.extra.Installation()
	}
	w.settingsSection(c, "settings-herdr-install", "Herdr", "Hergo uses the Herdr installed on this machine. It starts Herdr when the default session is not already running, and leaves it running when this window closes.", p, func(g *settingsGroup) {
		status, detail := "Not installed", "Install Herdr to manage workspaces, worktrees and agents."
		if install.Installed() {
			status, detail = fallback(install.Version, "Installed"), install.Path
		} else if install.Path != "" {
			status, detail = "Not responding", install.Error
		}
		g.row(status, detail, func() {
			if !install.Installed() && native.Button(c, "Installation Guide").Clicked() {
				c.OpenURL(herdr.InstallURL)
			}
		})
		if install.Installed() {
			w.herdrCompatibilityRow(c, g, install)
		}
		g.row("Startup", "Automatic opens Herdr, and offers to install it when it is missing. Plain terminal stays available. Applies the next time Hergo opens.", func() {
			index := indexOf(startupModes, w.startupMode, 0)
			if native.Segmented(c.Key("settings-startup-mode"), &index, startupModeLabels...).Shrink(0).Changed() {
				w.startupMode = startupModes[max(0, min(len(startupModes)-1, index))]
				if w.extra.StartupMode != nil {
					w.extra.StartupMode(w.startupMode)
				}
			}
		})
		g.row("Plain terminal", "A separate window with local shells, without Herdr.", func() {
			if native.Button(c, "Open").Disabled(w.extra.PlainTerminal == nil).Clicked() {
				w.manifests.message = ""
				if err := w.extra.PlainTerminal(); err != nil {
					w.manifests.message = err.Error()
				}
			}
		})
	})
	w.settingsSection(c, "settings-herdr-reminders", "Reminders", "", p, func(g *settingsGroup) {
		if g.toggle("share-reminders", "Show reminders in Herdr", "Output and agent reminders also appear as toasts in Herdr's terminal clients.", &w.shareReminders).Changed() && w.extra.ShareReminders != nil {
			w.extra.ShareReminders(w.shareReminders)
		}
	})
	names := []string(nil)
	if w.extra.LayoutNames != nil {
		names = w.extra.LayoutNames()
	}
	w.settingsSection(c, "settings-herdr-layouts", "Saved layouts", "Save a tab's layout from the Tab menu; open it again as a new tab.", p, func(g *settingsGroup) {
		if len(names) == 0 {
			g.row("No saved layouts", "", func() {})
		}
		for _, name := range names {
			n := name
			g.row(n, "", func() {
				if native.Button(c.Key("layout-open:"+n), "Open").Disabled(!w.Can("apply_layout")).Clicked() {
					w.applyLayout(n)
				}
				if native.Button(c.Key("layout-delete:"+n), "Delete").Disabled(w.extra.DeleteLayout == nil).Clicked() {
					w.extra.DeleteLayout(n)
				}
			})
		}
	})
	s := &w.manifests
	if !s.loaded && install.Installed() {
		w.loadManifests(false)
	}
	w.settingsSection(c, "settings-herdr-manifests", "Agent manifests", "How Herdr recognizes each agent and where the rules come from.", p, func(g *settingsGroup) {
		g.row("Manifests", "Reload after editing a local manifest.", func() {
			if native.Button(c, "Reload").Disabled(s.busy || w.extra.Manifests == nil).Clicked() {
				w.loadManifests(true)
			}
		})
		for _, m := range s.entries {
			detail := m.SourceKind + " · " + m.Source
			if m.ActiveVersion != nil {
				detail = "v" + *m.ActiveVersion + " · " + detail
			}
			if m.Warning != nil && *m.Warning != "" {
				detail += "\n" + *m.Warning
			}
			if m.UpdateError != nil && *m.UpdateError != "" {
				detail += "\n" + *m.UpdateError
			}
			g.row(m.Agent, detail, func() {})
		}
	})
	w.settingsSection(c, "settings-herdr-handoff", "Upgrade without restarting", "After installing a new Herdr, hand the running sessions to it. Terminals and agents keep running.", p, func(g *settingsGroup) {
		g.row("Live handoff", fallback(s.handoff, install.Path), func() {
			if native.Button(c, "Hand Off…").Disabled(w.extra.LiveHandoff == nil || !install.Installed()).Clicked() {
				w.ask(shortcutPrompt{title: "Hand sessions to the installed Herdr?", message: "The running server passes its sessions to " + install.Path + ". Hergo reconnects when it is done.", action: "Hand Off", confirm: true}, func(string) {
					s.handoff = "Handing off…"
					w.extra.LiveHandoff(func(err error) {
						if err != nil {
							s.handoff = "Handoff failed: " + err.Error()
							return
						}
						s.handoff = "Handed off to " + install.Path
						if w.retry != nil {
							w.retry()
						}
					})
				})
			}
		})
	})
	if s.busy {
		native.Text(c, "Reading Herdr…").TextColor(p.Secondary)
	}
	if s.message != "" {
		native.Text(c, s.message).FontSize(12).TextColor(p.Secondary)
	}
}

// herdrCompatibilityRow names the Herdr this Hergo needs and offers to update
// an older one. The running server counts too: a new executable changes
// nothing until the sessions move to it.
func (w *Window) herdrCompatibilityRow(c *native.Context, g *settingsGroup, install herdr.Installation) {
	s := &w.manifests
	server := ""
	if w.extra.ServerVersion != nil {
		server = w.extra.ServerVersion()
	}
	detail := "Hergo works with Herdr " + herdr.CompatibleVersion + " or later."
	if install.Outdated() {
		detail = "Hergo needs Herdr " + herdr.CompatibleVersion + " or later for image paste, commands and notifications."
	} else if herdr.VersionOutdated(server) {
		detail = "The running Herdr server is still " + server + ". Hand off below, or restart Herdr, to use the installed version."
	}
	g.row("Compatible version", fallback(s.update, detail), func() {
		if install.Outdated() && native.Button(c, "Update Herdr…").Label("Update Herdr").Disabled(s.updating || w.extra.UpdateHerdr == nil).Clicked() {
			w.ask(shortcutPrompt{title: "Update Herdr?", message: "Hergo installs the latest Herdr for this system and moves running sessions to it where Herdr supports that.", action: "Update", confirm: true}, func(string) { w.updateHerdr() })
		}
	})
}

func (w *Window) updateHerdr() {
	s := &w.manifests
	s.updating, s.update = true, "Updating Herdr…"
	w.extra.UpdateHerdr(func(err error) {
		s.updating, s.update = false, ""
		if err != nil {
			s.update = "Update failed: " + err.Error()
			return
		}
		if w.retry != nil {
			w.retry()
		}
	})
}

// pluginLinkRows add local plugin development to the plugins page.
func (w *Window) pluginLinkRows(c *native.Context, g *settingsGroup) {
	s := &w.manifests
	g.row("Local plugin", "Link a plugin directory for development.", func() {
		native.TextInput(c.Key("plugin-link-path"), &s.pluginPath).Label("Plugin directory").Placeholder("~/code/my-plugin").Width(200)
		if native.Button(c, "Link").Disabled(w.extra.PluginLink == nil || strings.TrimSpace(s.pluginPath) == "").Clicked() {
			w.pluginChange(func(done func(error)) { w.extra.PluginLink(s.pluginPath, done) })
		}
	})
}

// herdrMissing reports whether the session ended because Herdr is not
// installed, rather than because it could not be reached.
func herdrMissing(v app.View) bool {
	return v.State == session.Incompatible && v.Error == herdr.ErrNotInstalled.Error()
}

// herdrMissingView is the window's welcome when Herdr is not installed:
// install it, point at one that is already installed, or use a plain terminal.
func (w *Window) herdrMissingView(c *native.Context, p design.Palette) {
	native.Column(c.Key("herdr-missing")).Grow(1).FillWidth().Center().Padding(32).Children(func() {
		native.Column(c).Width(420).Gap(16).AlignItems(native.Stretch).Children(func() {
			native.Row(c).Gap(10).AlignItems(native.Center).Children(func() {
				native.Icon(c, icons.Terminal).Size(22, 22).TextColor(p.Secondary)
				native.Text(c, "Herdr is not installed").FontSize(18).Bold()
			})
			native.Text(c, "Herdr runs your workspaces and keeps them when this window closes. Hergo can install it, or use a Herdr you already have.").FontSize(13).TextColor(p.Secondary).FillWidth()
			native.Column(c).Gap(8).AlignItems(native.Stretch).Children(func() {
				busy := w.setupBusy
				if native.PrimaryButton(c, "Install Herdr").FillWidth().Disabled(busy || w.extra.InstallHerdr == nil).Clicked() {
					w.beginHerdrSetup("Installing Herdr…", w.extra.InstallHerdr)
				}
				if native.Button(c, "Choose Herdr…").FillWidth().Disabled(busy || w.extra.ChooseHerdr == nil).Clicked() {
					w.beginHerdrSetup("Looking for Herdr…", w.extra.ChooseHerdr)
				}
				if native.Button(c, "Installation Guide").FillWidth().Clicked() {
					c.OpenURL(herdr.InstallURL)
				}
				if native.Button(c, "Check Again").FillWidth().Disabled(busy).Clicked() && w.retry != nil {
					w.retry()
				}
				if native.Button(c, "Open Plain Terminal").FillWidth().Disabled(w.extra.PlainTerminal == nil).Clicked() {
					w.shortcutError(w.extra.PlainTerminal())
				}
			})
			if w.setupLabel != "" {
				native.Text(c, w.setupLabel).FontSize(12).TextColor(p.Secondary).FillWidth()
			}
			if w.setupError != "" {
				native.Text(c, w.setupError).FontSize(12).TextColor(p.Danger).FillWidth()
			}
			always := w.startupMode == "plain"
			if native.Checkbox(c.Key("herdr-missing-always-plain"), &always, "Open as a plain terminal next time").Changed() {
				w.startupMode = "auto"
				if always {
					w.startupMode = "plain"
				}
				if w.extra.StartupMode != nil {
					w.extra.StartupMode(w.startupMode)
				}
			}
		})
	})
}

// beginHerdrSetup marks the setup screen busy and runs install or choose.
// done runs on the UI thread. A nil error connects; a canceled file dialog
// leaves the screen as it was.
func (w *Window) beginHerdrSetup(label string, run func(func(error))) {
	if run == nil || w.setupBusy {
		return
	}
	w.setupBusy = true
	w.setupError = ""
	w.setupLabel = label
	run(func(err error) {
		w.setupBusy = false
		w.setupLabel = ""
		if errors.Is(err, herdr.ErrSetupCanceled) {
			return
		}
		if err != nil {
			w.setupError = err.Error()
			return
		}
		w.setupError = ""
		if w.retry != nil {
			w.retry()
		}
	})
}
