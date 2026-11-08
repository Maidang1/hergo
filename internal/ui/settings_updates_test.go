package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/Maidang1/hergo/internal/herdr"
	native "github.com/egoist/mygo/ui"
)

func TestUpdateSettingsSharePreferencesAndManualCheck(t *testing.T) {
	w := settingsWindow(t)
	state := UpdateSettings{Version: "0.0.7", Enabled: true, AutomaticChecks: true}
	checks := 0
	w.SetActionHooks(ActionHooks{
		UpdateSettings:           func() UpdateSettings { return state },
		CheckForUpdates:          func() { checks++ },
		AutomaticUpdateChecks:    func(on bool) { state.AutomaticChecks = on },
		AutomaticUpdateDownloads: func(on bool) { state.AutomaticDownloads = on },
	})
	tester := native.NewTester(w.SettingsView, 720, 1000)
	for _, label := range []string{"Updates", "Automatically check for updates", "Automatically download and install", "Check for Updates"} {
		if err := tester.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if state.AutomaticChecks || !state.AutomaticDownloads || checks != 1 {
		t.Fatalf("update hooks: state=%+v checks=%d", state, checks)
	}
	// Changes made by the update window appear on the next settings frame.
	state.LastCheck = time.Date(2026, 10, 27, 12, 34, 0, 0, time.Local)
	tester.Frame()
	if _, ok := tester.Find("2026-10-27 12:34"); !ok {
		t.Fatal("last check did not refresh")
	}
	if err := tester.Click("Automatically check for updates"); err != nil {
		t.Fatal(err)
	}
	if !state.AutomaticChecks {
		t.Fatal("automatic checks did not re-enable")
	}
}

func TestUpdateSettingsDevelopmentBuildCanCheckManually(t *testing.T) {
	w := settingsWindow(t)
	checked := false
	w.SetActionHooks(ActionHooks{CheckForUpdates: func() { checked = true }})
	tester := native.NewTester(w.SettingsView, 720, 1000)
	if err := tester.Click("Updates"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tester.Find("Automatic updates unavailable"); !ok {
		t.Fatal("missing unavailable explanation")
	}
	if _, ok := tester.Find("Automatically check for updates"); ok {
		t.Fatal("disabled updater offers automatic checks")
	}
	if err := tester.Click("Check for Updates"); err != nil {
		t.Fatal(err)
	}
	if !checked {
		t.Fatal("manual check unavailable for disabled build")
	}
}

func TestHerdrCompatibilityOffersUpdateOnlyWhenOutdated(t *testing.T) {
	w := settingsWindow(t)
	install := herdr.Installation{Path: "/bin/herdr", Version: "herdr 0.8.0"}
	server := "0.8.0"
	updates := 0
	var finish func(error)
	w.SetExtraHooks(ExtraHooks{
		Installation:  func() herdr.Installation { return install },
		ServerVersion: func() string { return server },
		UpdateHerdr:   func(done func(error)) { updates++; finish = done },
	})
	tester := native.NewTester(w.SettingsView, 720, 1000)
	if err := tester.Click("Herdr"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tester.Find("Hergo needs Herdr " + herdr.CompatibleVersion + " or later for image paste, commands and notifications."); !ok {
		t.Fatal("outdated Herdr did not name the compatible version")
	}
	if err := tester.Click("Update Herdr"); err != nil {
		t.Fatal(err)
	}
	if updates != 0 {
		t.Fatal("update ran before it was confirmed")
	}
	w.controller.shortcut.prompt.submit("")
	if updates != 1 {
		t.Fatalf("confirmed update ran %d times", updates)
	}
	tester.Frame()
	if _, ok := tester.Find("Updating Herdr…"); !ok {
		t.Fatal("running update is not shown")
	}
	finish(errors.New("download failed"))
	tester.Frame()
	if _, ok := tester.Find("Update failed: download failed"); !ok {
		t.Fatal("update failure is not shown")
	}
	// The executable is new but the sessions have not moved to it.
	w.manifests.update = ""
	install.Version = "herdr 0.9.5"
	tester.Frame()
	if _, ok := tester.Find("The running Herdr server is still 0.8.0. Hand off below, or restart Herdr, to use the installed version."); !ok {
		t.Fatal("old running server is not shown")
	}
	if _, ok := tester.Find("Update Herdr"); ok {
		t.Fatal("compatible executable still offers an update")
	}
	server = "0.9.5"
	tester.Frame()
	if _, ok := tester.Find("Hergo works with Herdr " + herdr.CompatibleVersion + " or later."); !ok {
		t.Fatal("compatible state is not shown")
	}
}
