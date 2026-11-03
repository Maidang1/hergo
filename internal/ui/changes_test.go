package ui

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Maidang1/herdr-ui/internal/gitdiff"
	"github.com/Maidang1/herdr-ui/internal/session"
	"github.com/Maidang1/herdr-ui/internal/sidebarinfo"
	native "github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

func changesWindow(t *testing.T) *Window {
	t.Helper()
	return openChangesWindow(t, true)
}

func changesWindowWithoutGit(t *testing.T) *Window {
	t.Helper()
	return openChangesWindow(t, false)
}

func openChangesWindow(t *testing.T, git bool) *Window {
	t.Helper()
	w := referenceWindow(t)
	s, err := session.DecodeSnapshot([]byte(fmt.Sprintf(`{"type":"session_snapshot","snapshot":{"version":"test","protocol":22,"workspaces":[{"workspace_id":"a","label":"API","active_tab_id":"ta"},{"workspace_id":"b","label":"Web","active_tab_id":"tb"}],"tabs":[{"workspace_id":"a","tab_id":"ta","number":1,"label":"Agent"},{"workspace_id":"b","tab_id":"tb","number":1,"label":"Build"}],"panes":[{"pane_id":"pa","terminal_id":"term-a","workspace_id":"a","tab_id":"ta","cwd":%q},{"pane_id":"pb","terminal_id":"term-b","workspace_id":"b","tab_id":"tb","cwd":%q}],"layouts":[],"agents":[]}}`, filepath.Join(t.TempDir(), "api"), filepath.Join(t.TempDir(), "web"))))
	if err != nil {
		t.Fatal(err)
	}
	published := session.PublishedSnapshot{TargetID: 1, Epoch: 1, State: session.Ready, Snapshot: s}
	w.app.Update(published)
	if !git {
		return w
	}
	entries := map[string]sidebarinfo.Entry{}
	for _, ws := range sidebarinfo.BuildSource(published).Workspaces {
		root := ws.Path
		if root == "" {
			root = string(filepath.Separator)
		}
		entries[ws.ID] = sidebarinfo.Entry{Fingerprint: sidebarinfo.Fingerprint(ws), Info: sidebarinfo.Info{RepoRoot: root, GitKnown: true, Branch: "main"}}
	}
	w.app.UpdateSidebar(sidebarinfo.Publication{TargetID: 1, Entries: entries})
	return w
}

func changesFixture(w *Window) {
	w.hooks.Changes = func(dir string, done func(gitdiff.Changes, error)) func() {
		done(gitdiff.Changes{Root: dir, Base: strings.Repeat("a", 40), Files: []gitdiff.File{
			{Path: "internal/server.go", Status: "Modified", Added: 2, Removed: 1},
			{Path: "README.md", Status: "Untracked", Untracked: true},
		}}, nil)
		return func() {}
	}
	w.hooks.DiffFile = func(req gitdiff.Request, done func(gitdiff.Text, error)) func() {
		done(gitdiff.Text{Old: "package server\n\nfunc port() int { return 80 }\n", New: "package server\n\nfunc port() int {\n\treturn 8080\n}\n"}, nil)
		return func() {}
	}
}

func TestChangesViewSelectionLayoutRefreshAndReadOnly(t *testing.T) {
	w := changesWindow(t)
	changesFixture(w)
	calls := 0
	load := w.hooks.Changes
	w.hooks.Changes = func(dir string, done func(gitdiff.Changes, error)) func() {
		calls++
		return load(dir, done)
	}
	tester := native.NewTester(w.View, 1200, 800)
	if err := tester.Click("Show changes"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !w.changes.open || w.changes.diff == nil || w.changes.diff.Editable || len(w.changes.diff.Hunks()) == 0 {
		t.Fatal("missing read-only diff")
	}
	if err := tester.Click("Side-by-side diff"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if w.changes.diff.Layout != editor.DiffSplit {
		t.Fatal("layout did not change")
	}
	if err := tester.Click("Changed file README.md"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if w.changes.selected != "README.md" {
		t.Fatal("file selection did not change")
	}
	if err := tester.Click("Refresh changes"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if calls != 2 || w.changes.selected != "README.md" {
		t.Fatal("refresh lost file selection", calls)
	}
	rect, ok := tester.Find("Code diff README.md")
	if !ok || rect.H < 50 || rect.W < 100 {
		t.Fatal("diff has no usable bounds", rect)
	}
	before := w.changes.diff.New.String()
	tester.ClickAt(rect.X+rect.W*.7, rect.Y+20)
	tester.Type("should not edit")
	tester.Frame()
	if got := w.changes.diff.New.String(); got != before {
		t.Fatal("read-only preview accepted an edit")
	}
	w.Dispatch("find")
	tester.Frame()
	if _, ok := tester.Find("Find"); !ok || w.findOpen() {
		t.Fatal("Find did not open in the diff")
	}
	tester.Key(0, native.KeyEscape)
	tester.Frame()
	if !w.changes.open {
		t.Fatal("closing Find also closed Changes")
	}
	w.ClosePaneKey()
	tester.Frame()
	if w.changes.open || w.currentView().PaneID != "pa" {
		t.Fatal("close shortcut did not return to the terminal")
	}
	w.Dispatch("changes")
	if calls != 3 {
		t.Fatal("reopening did not refresh changes")
	}
}

func TestChangesVisibilityBelongsToWorkspace(t *testing.T) {
	w := changesWindow(t)
	changesFixture(w)
	w.Dispatch("changes")
	w.changes.layout = "split"
	w.selectChangedFile(w.changes.list.Files[1])
	tester := native.NewTester(w.View, 1200, 800)

	w.app.SelectWorkspace("b")
	tester.Frame()
	if w.changes.open {
		t.Fatal("Changes followed the selection into another workspace")
	}
	if _, ok := tester.Find("All file diffs"); ok {
		t.Fatal("another workspace still displays the diff")
	}
	w.Dispatch("changes")
	tester.Frame()
	if !w.changes.open || w.changes.layout == "split" {
		t.Fatal("another workspace did not open with its own view state")
	}
	w.app.SelectWorkspace("a")
	tester.Frame()
	if !w.changes.open || w.changes.layout != "split" {
		t.Fatal("opening another workspace lost the first workspace's view")
	}
	w.app.SelectWorkspace("b")
	tester.Frame()
	if !w.changes.open || w.changes.layout == "split" {
		t.Fatal("the second workspace did not remember its own open view")
	}
	w.closeChanges()

	w.app.SelectWorkspace("a")
	tester.Frame()
	if !w.changes.open || w.changes.layout != "split" || w.changes.selected != "README.md" {
		t.Fatal("returning to the workspace did not restore its diff")
	}
	if _, ok := tester.Find("All file diffs"); !ok {
		t.Fatal("restored workspace did not render its diff")
	}
	w.closeChanges()
	for _, id := range []string{"b", "a"} {
		w.app.SelectWorkspace(id)
		tester.Frame()
		if w.changes.open {
			t.Fatalf("closed Changes reopened in workspace %s", id)
		}
	}
	// Dispatching before a render must toggle the newly selected workspace.
	w.Dispatch("changes")
	w.app.SelectWorkspace("b")
	w.Dispatch("changes")
	if !w.changes.open || w.changes.target.workspace != "b" {
		t.Fatal("toggle acted on the previously rendered workspace")
	}
}

func TestChangesWorkspaceContextMenuOpensSelectedProject(t *testing.T) {
	w := changesWindow(t)
	changesFixture(w)
	w.Dispatch("changes")
	tester := native.NewTester(w.View, 1200, 800)
	if err := tester.RightClick("Web"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.ChooseMenuItem("Show Changes"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !w.changes.open || w.changes.target.workspace != "b" {
		t.Fatal("Show Changes did not open the selected project")
	}
	w.app.SelectWorkspace("a")
	tester.Frame()
	if !w.changes.open {
		t.Fatal("context menu closed the previous project's Changes")
	}
}

func TestChangesDiscardsLateWorkspacesAndFiles(t *testing.T) {
	w := changesWindow(t)
	var lists []func(gitdiff.Changes, error)
	var texts []func(gitdiff.Text, error)
	cancelled := 0
	w.hooks.Changes = func(_ string, done func(gitdiff.Changes, error)) func() {
		lists = append(lists, done)
		return func() { cancelled++ }
	}
	w.hooks.DiffFile = func(_ gitdiff.Request, done func(gitdiff.Text, error)) func() {
		texts = append(texts, done)
		return func() { cancelled++ }
	}
	w.Dispatch("changes")
	w.app.SelectWorkspace("b")
	// The old result must be rejected even before the next render syncs.
	lists[0](gitdiff.Changes{Root: "wrong", Files: []gitdiff.File{{Path: "wrong"}}}, nil)
	if w.changes.list.Root != "" {
		t.Fatal("old workspace installed before the next frame")
	}
	w.syncChanges()
	if w.changes.open || len(lists) != 1 || cancelled == 0 {
		t.Fatal("workspace change did not hide Changes and cancel the request")
	}
	w.app.SelectWorkspace("a")
	w.syncChanges()
	if len(lists) != 2 || cancelled == 0 {
		t.Fatal("returning to the workspace did not reload")
	}
	lists[0](gitdiff.Changes{Root: "wrong"}, nil)
	if w.changes.list.Root == "wrong" {
		t.Fatal("restored workspace accepted its old request")
	}
	files := []gitdiff.File{{Path: "first.go"}, {Path: "second.go"}}
	lists[1](gitdiff.Changes{Root: w.changes.target.path, Files: files}, nil)
	w.selectChangedFile(files[1])
	texts[1](gitdiff.Text{New: "current\n"}, nil)
	texts[0](gitdiff.Text{New: "late\n"}, nil)
	if w.changes.diff.New.String() != "current\n" {
		t.Fatal("old file replaced current preview")
	}
	w.refreshChanges()
	w.closeChanges()
	lists[2](gitdiff.Changes{Root: "late"}, nil)
	if w.changes.open || w.changes.list.Root == "late" {
		t.Fatal("closed view accepted a late response")
	}
}

func TestChangesEmptyErrorAndEscape(t *testing.T) {
	w := changesWindow(t)
	changesFixture(w)
	w.hooks.Changes = func(_ string, done func(gitdiff.Changes, error)) func() {
		done(gitdiff.Changes{}, errors.New("Cannot read repository"))
		return func() {}
	}
	w.Dispatch("changes")
	tester := native.NewTester(w.View, 820, 560)
	if !tester.HasText("Cannot read repository") {
		t.Fatal("missing error", tester.Texts())
	}
	w.hooks.Changes = func(dir string, done func(gitdiff.Changes, error)) func() {
		done(gitdiff.Changes{Root: dir}, nil)
		return func() {}
	}
	if err := tester.Click("Refresh changes"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("No uncommitted changes.") {
		t.Fatal("missing empty state", tester.Texts())
	}
	tester.Key(0, native.KeyEscape)
	tester.Frame()
	if w.changes.open {
		t.Fatal("Escape did not close Changes")
	}
}

func manyChangesFixture(w *Window, count int) []gitdiff.File {
	files := make([]gitdiff.File, count)
	for i := range files {
		files[i] = gitdiff.File{Path: fmt.Sprintf("file-%02d.go", i), Status: "Untracked", Untracked: true}
	}
	w.hooks.Changes = func(dir string, done func(gitdiff.Changes, error)) func() {
		done(gitdiff.Changes{Root: dir, Files: files}, nil)
		return func() {}
	}
	return files
}

func TestChangesScrollsAcrossFiles(t *testing.T) {
	for _, layout := range []string{"unified", "split"} {
		t.Run(layout, func(t *testing.T) {
			w := changesWindow(t)
			files := manyChangesFixture(w, 40)
			reads := make(map[string]int)
			w.hooks.DiffFile = func(req gitdiff.Request, done func(gitdiff.Text, error)) func() {
				reads[req.File.Path]++
				done(gitdiff.Text{New: strings.Repeat("// "+req.File.Path+"\n", 36)}, nil)
				return func() {}
			}
			w.Dispatch("changes")
			w.changes.layout = layout
			tester := native.NewTester(w.View, 1200, 800)
			bounds, ok := tester.Find("All file diffs")
			if !ok || len(reads) >= len(files) {
				t.Fatal("diffs are not loaded by visible range", len(reads))
			}
			first, _ := w.changes.diffs.Visible()
			// Scroll over the code, not the file list or empty space.
			tester.Scroll(bounds.X+bounds.W*.75, bounds.Y+100, 0, 950)
			tester.Frame()
			next, last := w.changes.diffs.Visible()
			if next <= first || last < next {
				t.Fatalf("scrolling the code did not advance to another file: %d -> %d..%d", first, next, last)
			}
			if w.changes.selected != files[next].Path || w.changes.diff != w.changes.previews[files[next].Path].diff {
				t.Fatal("active file did not follow scrolling", w.changes.selected, next)
			}
			if reads[files[next].Path] == 0 {
				t.Fatal("scrolling did not load the next file")
			}
			w.changes.rows.ScrollTo(15, native.Start)
			tester.Frame()
			if err := tester.Click("Changed file " + files[15].Path); err != nil {
				t.Fatal(err)
			}
			tester.Frame()
			first, _ = w.changes.diffs.Visible()
			if first != 15 {
				t.Fatal("file list did not jump within the continuous diff", first)
			}
			for _, index := range []int{0, 10, 20, 30, 39} {
				w.selectChangedFile(files[index])
				tester.Frame()
			}
			if len(w.changes.previews) > 24 {
				t.Fatal("visited previews were retained without a bound", len(w.changes.previews))
			}
			w.refreshChanges()
			tester.Frame()
			if w.changes.selected != files[39].Path {
				t.Fatal("refresh lost the current file", w.changes.selected)
			}
		})
	}
}

func TestChangesContinuousReadsBoundedAndCancelled(t *testing.T) {
	w := changesWindow(t)
	manyChangesFixture(w, 100)
	var callbacks []func(gitdiff.Text, error)
	cancelled := 0
	w.hooks.DiffFile = func(req gitdiff.Request, done func(gitdiff.Text, error)) func() {
		callbacks = append(callbacks, done)
		return func() { cancelled++ }
	}
	w.Dispatch("changes")
	tester := native.NewTester(w.View, 1200, 800)
	if len(callbacks) != 4 || w.changes.loading != 4 || len(w.changes.queue) == 0 {
		t.Fatal("parallel preview reads are not bounded", len(callbacks), w.changes.loading)
	}
	callbacks[0](gitdiff.Text{New: "first\n"}, nil)
	tester.Frame()
	if len(callbacks) != 5 || w.changes.loading != 4 {
		t.Fatal("completing a preview did not start the next queued file", len(callbacks), w.changes.loading)
	}
	w.app.SelectWorkspace("b")
	tester.Frame()
	if cancelled != 4 || w.changes.open || len(w.changes.queue) != 0 {
		t.Fatal("workspace switch did not cancel all outstanding reads", cancelled)
	}
	for _, done := range callbacks[1:] {
		done(gitdiff.Text{New: "late\n"}, nil)
	}
	if w.changes.diff != nil || w.changes.loading != 0 {
		t.Fatal("another workspace accepted a stale preview")
	}
}

func TestChangesContinuousFileFailuresDoNotHideOtherDiffs(t *testing.T) {
	w := changesWindow(t)
	files := manyChangesFixture(w, 3)
	w.hooks.DiffFile = func(req gitdiff.Request, done func(gitdiff.Text, error)) func() {
		switch req.File.Path {
		case files[0].Path:
			done(gitdiff.Text{}, errors.New("Cannot read this file"))
		case files[1].Path:
			done(gitdiff.Text{Notice: "Binary file cannot be previewed."}, nil)
		default:
			done(gitdiff.Text{Patch: "diff --git a/file-02.go b/file-02.go\n--- a/file-02.go\n+++ b/file-02.go\n@@ -1 +1 @@\n-old\n+new\n"}, nil)
		}
		return func() {}
	}
	w.Dispatch("changes")
	tester := native.NewTester(w.View, 1200, 800)
	for _, text := range []string{"Cannot read this file", "Binary file cannot be previewed.", files[2].Path} {
		if !tester.HasText(text) {
			t.Fatal("missing per-file content", text)
		}
	}
	if d := w.changes.previews[files[2].Path].diff; d == nil || !d.FitContent || d.Editable {
		t.Fatal("valid patch after file failures was not shown as a read-only diff")
	}
}

func TestChangesHiddenOutsideGitRepository(t *testing.T) {
	w := changesWindowWithoutGit(t)
	changesFixture(w)
	tester := native.NewTester(w.View, 1200, 800)
	if _, ok := tester.Find("Show changes"); ok {
		t.Fatal("git control shown outside a repository")
	}
	w.Dispatch("changes")
	if w.changes.open {
		t.Fatal("opened changes outside a repository")
	}
	if err := tester.RightClick("Web"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.ChooseMenuItem("Show Changes"); err == nil {
		t.Fatal("workspace menu offered changes outside a repository")
	}
}

func TestCommitSelectionShowsThatCommit(t *testing.T) {
	w := changesWindow(t)
	changesFixture(w)
	hash := strings.Repeat("b", 40)
	parent := strings.Repeat("c", 40)
	other := strings.Repeat("d", 40)
	var seen []string
	w.hooks.Commits = func(dir string, done func([]gitdiff.Commit, error)) func() {
		done([]gitdiff.Commit{{Hash: hash, Subject: "Add server", Author: "Ada", Time: 1_700_000_000}}, nil)
		return func() {}
	}
	w.hooks.CommitChanges = func(dir, got string, done func(gitdiff.Changes, error)) func() {
		seen = append(seen, got)
		files := []gitdiff.File{{Path: "old.go", Status: "Modified", Added: 1, Removed: 1}}
		if got == other {
			files = []gitdiff.File{{Path: "other.go", Status: "Added", Added: 1}}
		}
		done(gitdiff.Changes{Root: dir, Base: parent, Commit: got, Files: files}, nil)
		return func() {}
	}
	w.hooks.DiffFile = func(req gitdiff.Request, done func(gitdiff.Text, error)) func() {
		if req.Commit == "" {
			done(gitdiff.Text{Old: "package server\n", New: "package server\n\nfunc port() int { return 8080 }\n"}, nil)
			return func() {}
		}
		if req.Commit != hash && req.Commit != other || req.Base != parent {
			t.Fatalf("diff request %+v", req)
		}
		done(gitdiff.Text{Old: "a\n", New: "b\n"}, nil)
		return func() {}
	}
	w.Dispatch("changes")
	tester := native.NewTester(w.View, 1200, 800)
	tester.Frame()
	if _, ok := tester.Find("Commit " + hash); !ok {
		t.Fatal("history did not list the commit")
	}
	if err := tester.Click("Commit " + hash); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if w.changes.revision != hash || w.changes.selected != "old.go" || len(seen) != 1 {
		t.Fatal("commit selection did not load that commit", w.changes.revision, w.changes.selected, seen)
	}
	if !tester.HasText("old.go") || !tester.HasText("Add server") {
		t.Fatal("commit diff was not shown", tester.Texts())
	}
	if err := tester.Click("Uncommitted changes"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if w.changes.revision != "" || w.changes.selected != "internal/server.go" {
		t.Fatal("uncommitted changes did not return", w.changes.revision, w.changes.selected)
	}
	w.changes.commits = append(w.changes.commits, gitdiff.Commit{Hash: other, Subject: "Other"})
	w.selectRevision(other)
	tester.Frame()
	w.app.SelectWorkspace("b")
	tester.Frame()
	if w.changes.revision != "" {
		t.Fatal("commit selection followed another workspace")
	}
	w.app.SelectWorkspace("a")
	tester.Frame()
	if w.changes.revision != other || w.changes.selected != "other.go" {
		t.Fatal("workspace did not restore its commit", w.changes.revision, w.changes.selected)
	}
}

func TestChangesRenderSizesAndThemes(t *testing.T) {
	for _, size := range [][2]int{{820, 560}, {1200, 800}, {1600, 1000}} {
		for _, dark := range []bool{false, true} {
			w := changesWindow(t)
			changesFixture(w)
			w.Dispatch("changes")
			tester := native.NewTester(w.View, size[0], size[1])
			tester.SetDark(dark)
			tester.SetTitleBar(native.TitleBar{Left: 80, Height: 44})
			tester.Frame()
			for _, label := range []string{"Refresh changes", "Close changes", "Unified diff", "Side-by-side diff", "All file diffs", "Code diff internal/server.go"} {
				r, ok := tester.Find(label)
				if !ok || r.W <= 0 || r.H <= 0 || r.X < 0 || r.X+r.W > float32(size[0]) || r.Y+r.H > float32(size[1]) {
					t.Fatalf("%v dark=%t %s has invalid bounds: %+v", size, dark, label, r)
				}
			}
			if dir := os.Getenv("HERDR_UI_RENDER_DIR"); dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				name := filepath.Join(dir, fmt.Sprintf("changes-%dx%d-dark-%t.png", size[0], size[1], dark))
				file, err := os.Create(name)
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, tester.Image())
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
