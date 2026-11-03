package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Maidang1/herdr-ui/internal/design"
	"github.com/Maidang1/herdr-ui/internal/gitdiff"
	"github.com/Maidang1/herdr-ui/internal/icons"
	"github.com/Maidang1/herdr-ui/internal/session"
	native "github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

type changesTarget struct {
	target, epoch   uint64
	workspace, path string
}

type changesWorkspace struct {
	target    uint64
	workspace string
}

type changesViewState struct {
	open                       bool
	selected, layout, revision string
}

type changedFilePreview struct {
	index       int
	busy        bool
	cancel      func()
	err, notice string
	diff        *editor.DiffState
}

type changesUI struct {
	open, busy    bool
	target        changesTarget
	generation    uint64
	cancel        func()
	list          gitdiff.Changes
	selected, err string
	diff          *editor.DiffState
	layout        string
	rows, diffs   native.ListState
	revs          native.ListState
	jumping       bool
	previews      map[string]*changedFilePreview
	queue         []gitdiff.File
	loading       int
	focus         bool
	revision      string
	commits       []gitdiff.Commit
	commitsErr    string
	commitsBusy   bool
	commitsGen    uint64
	commitsCancel func()
	workspaces    map[changesWorkspace]changesViewState
}

func (w *Window) workspaceRepo(id string) bool {
	if id == "" {
		return false
	}
	for _, row := range w.currentView().Rows {
		if row.ID == id {
			return row.RepoRoot != ""
		}
	}
	return false
}

func (w *Window) changesTarget() changesTarget {
	if w.app == nil {
		return changesTarget{}
	}
	v := w.currentView()
	if v.State != session.Ready || v.Stale || v.WorkspaceID == "" || !filepath.IsAbs(v.Path) {
		return changesTarget{}
	}
	return changesTarget{v.TargetID, v.Epoch, v.WorkspaceID, v.Path}
}

func (w *Window) toggleChanges() {
	w.syncChanges()
	if w.changes.open {
		w.closeChanges()
		return
	}
	w.openChanges()
}

func (w *Window) openChanges() {
	w.CloseCommandSurface()
	w.details.open = false
	w.details.generation++
	w.findClose()
	w.BlurInput()
	w.changes.open, w.changes.focus = true, true
	w.refreshChanges()
	w.refreshCommits()
}

func (w *Window) cancelChanges() {
	ch := &w.changes
	ch.generation++
	if ch.cancel != nil {
		ch.cancel()
		ch.cancel = nil
	}
	for _, preview := range ch.previews {
		if preview.cancel != nil {
			preview.cancel()
			preview.cancel = nil
		}
		preview.busy = false
	}
	ch.busy, ch.loading, ch.queue = false, 0, nil
}

func (w *Window) cancelCommits() {
	ch := &w.changes
	ch.commitsGen++
	if ch.commitsCancel != nil {
		ch.commitsCancel()
		ch.commitsCancel = nil
	}
	ch.commitsBusy = false
}

func (w *Window) closeChanges() {
	w.syncChanges()
	w.cancelChanges()
	w.cancelCommits()
	w.changes.open = false
	w.BlurInput()
}

// Keep visibility and selection per workspace, but reload its files on return.
// Never retain repository texts or in-flight work from the previous target.
func (w *Window) syncChanges() {
	ch := &w.changes
	target := w.changesTarget()
	if target == ch.target {
		return
	}
	w.cancelChanges()
	w.cancelCommits()
	if ch.workspaces == nil {
		ch.workspaces = make(map[changesWorkspace]changesViewState)
	}
	if ch.target.workspace != "" {
		ch.workspaces[changesWorkspace{ch.target.target, ch.target.workspace}] = changesViewState{ch.open, ch.selected, ch.layout, ch.revision}
	}
	state := ch.workspaces[changesWorkspace{target.target, target.workspace}]
	if state.open && !w.workspaceRepo(target.workspace) {
		state.open = false
	}
	*ch = changesUI{
		open: state.open, selected: state.selected, layout: state.layout, revision: state.revision,
		target: target, generation: ch.generation, commitsGen: ch.commitsGen,
		workspaces: ch.workspaces,
	}
	w.BlurInput()
	if ch.open {
		w.openChanges()
	}
}

func (w *Window) changesButton(c *native.Context, key, text, label string, selected bool, p design.Palette) native.Element {
	button := native.ButtonBase(c.Key(key)).Height(26).Shrink(0).Radius(13).Padding(0, 10).Label(label).Selected(selected)
	w.controlState(button, p)
	color := p.Secondary
	if selected {
		button.Background(p.Selected)
		color = p.Text
	}
	button.Children(func() { native.Text(c, text).FontSize(13).TextColor(color) })
	return button
}

func (w *Window) gitChangesButton(c *native.Context, label string, selected bool, p design.Palette) native.Element {
	button := native.ButtonBase(c.Key("show-changes")).Size(26, 26).Shrink(0).Radius(13).Center().Label(label).Selected(selected)
	w.controlState(button, p)
	color := p.Secondary
	if selected && !button.Hovered() && !button.Pressed() {
		button.Background(p.Selected)
		color = p.Text
	}
	button.Children(func() { native.Icon(c, icons.GitBranch).Size(15, 15).TextColor(color) })
	return button
}

func (w *Window) refreshChanges() {
	ch := &w.changes
	if !ch.open || ch.target.path == "" || ch.revision == "" && w.hooks.Changes == nil {
		return
	}
	w.cancelChanges()
	ch.busy, ch.err = true, ""
	ch.previews = make(map[string]*changedFilePreview)
	ch.diff = nil
	ch.list = gitdiff.Changes{}
	ch.diffs = native.ListState{}
	gen, target := ch.generation, ch.target
	done := func(list gitdiff.Changes, err error) {
		if !ch.open || ch.generation != gen || w.changesTarget() != target {
			return
		}
		ch.busy, ch.cancel = false, nil
		if err != nil {
			ch.err = err.Error()
			ch.list, ch.diff, ch.selected = gitdiff.Changes{}, nil, ""
			return
		}
		ch.list = list
		selected := ch.selected
		ch.diff, ch.selected = nil, ""
		for _, f := range list.Files {
			if f.Path == selected {
				w.selectChangedFile(f)
				return
			}
		}
		if len(list.Files) > 0 {
			w.selectChangedFile(list.Files[0])
		}
	}
	if ch.revision != "" {
		if w.hooks.CommitChanges == nil {
			ch.busy = false
			ch.err = "Commit history is unavailable."
			return
		}
		ch.cancel = w.hooks.CommitChanges(target.path, ch.revision, done)
		return
	}
	ch.cancel = w.hooks.Changes(target.path, done)
}

func (w *Window) refreshCommits() {
	ch := &w.changes
	if !ch.open || ch.target.path == "" || w.hooks.Commits == nil {
		return
	}
	w.cancelCommits()
	ch.commitsBusy, ch.commitsErr = true, ""
	gen, target := ch.commitsGen, ch.target
	ch.commitsCancel = w.hooks.Commits(target.path, func(list []gitdiff.Commit, err error) {
		if !ch.open || ch.commitsGen != gen || w.changesTarget() != target {
			return
		}
		ch.commitsBusy, ch.commitsCancel = false, nil
		if err != nil {
			ch.commits, ch.commitsErr = nil, err.Error()
			return
		}
		ch.commits = list
	})
}

func (w *Window) selectRevision(hash string) {
	ch := &w.changes
	if hash == ch.revision {
		return
	}
	ch.revision = hash
	ch.rows = native.ListState{}
	w.refreshChanges()
}

func (w *Window) selectChangedFile(f gitdiff.File) {
	ch := &w.changes
	for i, file := range ch.list.Files {
		if file.Path == f.Path {
			ch.selected = f.Path
			ch.diff = w.loadChangedFile(f, i).diff
			ch.diffs.ScrollTo(i, native.Start)
			ch.jumping = true
			return
		}
	}
}

func (w *Window) loadChangedFile(f gitdiff.File, index int) *changedFilePreview {
	ch := &w.changes
	if preview := ch.previews[f.Path]; preview != nil {
		return preview
	}
	preview := &changedFilePreview{index: index, busy: true}
	ch.previews[f.Path] = preview
	ch.queue = append(ch.queue, f)
	w.loadChangedFiles()
	return preview
}

// Only visible files and a few neighbours request previews. Bound parallel Git
// reads even when a fast scroll visits many files before their reads finish.
func (w *Window) loadChangedFiles() {
	ch := &w.changes
	for ch.open && ch.loading < 4 && len(ch.queue) > 0 && w.hooks.DiffFile != nil {
		f := ch.queue[0]
		ch.queue = ch.queue[1:]
		preview := ch.previews[f.Path]
		ch.loading++
		gen, target := ch.generation, ch.target
		commit := ch.list.Commit
		if commit == "" {
			commit = ch.revision
		}
		req := gitdiff.Request{Root: ch.list.Root, Base: ch.list.Base, Commit: commit, File: f}
		cancel := w.hooks.DiffFile(req, func(text gitdiff.Text, err error) {
			if !ch.open || ch.generation != gen || ch.target != target || w.changesTarget() != target || ch.previews[f.Path] != preview {
				return
			}
			preview.busy, preview.cancel = false, nil
			ch.loading--
			if err != nil {
				preview.err = err.Error()
			} else {
				preview.setText(f.Path, text)
			}
			if ch.selected == f.Path {
				ch.diff = preview.diff
			}
			w.loadChangedFiles()
		})
		if preview.busy {
			preview.cancel = cancel
		} else if cancel != nil {
			cancel()
		}
	}
}

func (preview *changedFilePreview) setText(path string, text gitdiff.Text) {
	preview.notice = text.Notice
	if text.Patch != "" {
		files, err := editor.ParsePatch(text.Patch)
		if err != nil || len(files) == 0 {
			preview.err = "The patch cannot be previewed."
			return
		}
		if len(files) != 1 {
			preview.notice = "This large rename cannot be previewed. Use git diff to inspect it."
			return
		}
		preview.diff = &editor.DiffState{}
		preview.diff.SetPatch(files[0])
	} else if text.Notice == "" {
		lang := editor.LanguageFor(filepath.Base(path))
		old, now := editor.NewBuffer(text.Old), editor.NewBuffer(text.New)
		old.SetLanguage(lang)
		now.SetLanguage(lang)
		preview.diff = &editor.DiffState{Old: old, New: now}
	}
}

func (w *Window) changesView(c *native.Context, p design.Palette) {
	ch := &w.changes
	panel := native.Column(c.Key("changes-panel")).Grow(1).MinHeight(0).FillWidth().AlignItems(native.Stretch).Focusable().Label("Changes")
	if ch.focus {
		panel.Focus()
		ch.focus = false
	}
	panel.Children(func() {
		native.Row(c.Key("changes-header")).Height(40).Shrink(0).Padding(0, 12).Gap(10).AlignItems(native.Center).Children(func() {
			native.Text(c, "Changes").FontWeight(600)
			native.Text(c, ch.summary()).FontSize(12).TextColor(p.Secondary).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
			if w.changesButton(c, "diff-unified", "Unified", "Unified diff", ch.layout != "split", p).Clicked() {
				ch.layout = "unified"
			}
			if w.changesButton(c, "diff-split", "Split", "Side-by-side diff", ch.layout == "split", p).Clicked() {
				ch.layout = "split"
			}
			if w.changesButton(c, "refresh-changes", "Refresh", "Refresh changes", false, p).Disabled(ch.busy || ch.target.path == "").Clicked() {
				w.refreshChanges()
				w.refreshCommits()
			}
			if w.iconButton(c, "close-changes", icons.X, "Close changes", p).Clicked() {
				w.closeChanges()
			}
		})
		if ch.target.path == "" {
			w.changesMessage(c, p, "Select a connected workspace with a local directory.")
			return
		}
		native.Box(c).Height(1).Shrink(0).FillWidth().Background(p.Border)
		native.Row(c.Key("changes-body")).Grow(1).MinHeight(0).FillWidth().AlignItems(native.Stretch).Children(func() {
			width := max(0, w.viewportWidth-w.shape.Gap*2)
			if w.sidebarVisible {
				width -= w.sidebarWidth + w.shape.Gap
			}
			listWidth := min(float32(280), width*.36)
			native.Column(c.Key("changes-nav")).Width(listWidth).Shrink(0).MinHeight(0).Children(func() {
				w.revisionList(c, p)
				native.Box(c).Height(1).Shrink(0).FillWidth().Background(p.Border)
				if len(ch.list.Files) == 0 {
					return
				}
				ch.rows.Key = func(i int) any { return ch.list.Files[i].Path }
				native.List(c.Key("changed-files"), &ch.rows, len(ch.list.Files), func(i int) {
					f := ch.list.Files[i]
					row := native.ButtonBase(c.Key(f.Path)).FillWidth().Padding(9, 10).Radius(6).Label("Changed file " + f.Path).Selected(ch.selected == f.Path)
					if ch.selected == f.Path {
						row.Background(p.Selected)
					} else {
						w.controlState(row, p)
					}
					row.Children(func() {
						native.Column(c).FillWidth().Gap(4).Children(func() {
							native.Text(c, f.Path).FontSize(12).TextColor(p.Text).SingleLine().Ellipsis("…").FillWidth()
							native.Row(c).Gap(6).Children(func() {
								native.Text(c, f.Status).FontSize(11).TextColor(p.Secondary)
								if f.Added > 0 {
									native.Text(c, fmt.Sprintf("+%d", f.Added)).FontSize(11).TextColor(p.Success)
								}
								if f.Removed > 0 {
									native.Text(c, fmt.Sprintf("−%d", f.Removed)).FontSize(11).TextColor(p.Danger)
								}
							})
						})
					})
					if row.Clicked() && !ch.busy {
						w.selectChangedFile(f)
					}
				}).Grow(1).FillWidth().MinHeight(0).Padding(6)
			})
			native.Box(c).Width(1).Shrink(0).Background(p.Border)
			if ch.err != "" {
				w.changesMessage(c, p, ch.err)
			} else if ch.busy && len(ch.list.Files) == 0 {
				native.Text(c, "Reading changes…").TextColor(p.Secondary).Padding(8, 12)
			} else if len(ch.list.Files) == 0 {
				message := "No uncommitted changes."
				if ch.revision != "" {
					message = "This commit does not change any files."
				}
				w.changesMessage(c, p, message)
			} else {
				w.continuousChangesView(c, p)
			}
		})
	})
	if panel.Shortcut(0, native.KeyEscape) && !w.modalOpen() {
		w.closeChanges()
	}
}

func (ch *changesUI) summary() string {
	if ch.revision != "" {
		subject := "Commit"
		for _, commit := range ch.commits {
			if commit.Hash == ch.revision && commit.Subject != "" {
				subject = commit.Subject
				break
			}
		}
		short := ch.revision
		if len(short) > 7 {
			short = short[:7]
		}
		if ch.list.Root == "" {
			return short + " · " + subject
		}
		return fmt.Sprintf("%s · %s · %d files", short, subject, len(ch.list.Files))
	}
	if ch.list.Root != "" {
		return fmt.Sprintf("%d files · %s", len(ch.list.Files), ch.list.Root)
	}
	return "Uncommitted · compared with HEAD"
}

func (w *Window) revisionList(c *native.Context, p design.Palette) {
	ch := &w.changes
	count := 1 + len(ch.commits)
	ch.revs.Key = func(i int) any {
		if i == 0 {
			return "uncommitted"
		}
		return ch.commits[i-1].Hash
	}
	native.List(c.Key("change-revisions"), &ch.revs, count, func(i int) {
		if i == 0 {
			w.revisionButton(c, p, "uncommitted", "", "Uncommitted changes", "Uncommitted", "Working tree", ch.revision == "")
			return
		}
		commit := ch.commits[i-1]
		title := commit.Subject
		if title == "" {
			title = "(no message)"
		}
		meta := shortHash(commit.Hash)
		if stamp := commitStamp(commit.Time); stamp != "" {
			meta += " · " + stamp
		}
		w.revisionButton(c, p, commit.Hash, commit.Hash, "Commit "+commit.Hash, title, meta, ch.revision == commit.Hash)
	}).Height(min(float32(count)*48+8, 240)).Shrink(0).FillWidth().MinWidth(0).Padding(4, 6, 2, 6)
	if ch.commitsBusy && len(ch.commits) == 0 {
		native.Text(c, "Reading history…").FontSize(11).TextColor(p.Secondary).Padding(2, 10, 6, 10)
	} else if ch.commitsErr != "" {
		native.Text(c, ch.commitsErr).FontSize(11).TextColor(p.Secondary).FillWidth().Padding(2, 10, 6, 10)
	} else if len(ch.commits) == gitdiff.MaxCommits {
		native.Text(c, "Showing 100 recent commits").FontSize(11).TextColor(p.Secondary).Padding(2, 10, 6, 10)
	}
}

func (w *Window) revisionButton(c *native.Context, p design.Palette, key, hash, label, title, meta string, selected bool) {
	row := native.ButtonBase(c.Key("rev:"+key)).FillWidth().Padding(8, 10).Radius(6).Label(label).Selected(selected)
	if selected {
		row.Background(p.Selected)
	} else {
		w.controlState(row, p)
	}
	row.Children(func() {
		native.Column(c).FillWidth().Gap(2).Children(func() {
			native.Text(c, title).FontSize(12).TextColor(p.Text).SingleLine().Ellipsis("…").FillWidth()
			native.Text(c, meta).FontSize(11).TextColor(p.Secondary).SingleLine().Ellipsis("…").FillWidth()
		})
	})
	if row.Clicked() {
		w.selectRevision(hash)
	}
}

func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

func commitStamp(unix int64) string {
	if unix <= 0 {
		return ""
	}
	when := time.Unix(unix, 0)
	if when.Year() == time.Now().Year() {
		return when.Format("Jan 2")
	}
	return when.Format("Jan 2, 2006")
}

func (w *Window) continuousChangesView(c *native.Context, p design.Palette) {
	ch := &w.changes
	first, last := ch.diffs.Visible()
	if first <= last && first < len(ch.list.Files) {
		selected := ch.previews[ch.selected]
		if !ch.jumping && (selected == nil || selected.index < first || selected.index > last) {
			f := ch.list.Files[first]
			ch.selected = f.Path
			ch.diff = w.loadChangedFile(f, first).diff
			ch.rows.ScrollIntoView(first)
		}
		// A rapid jump should not wait behind files that left the viewport.
		pending := ch.queue[:0]
		for _, f := range ch.queue {
			preview := ch.previews[f.Path]
			if f.Path == ch.selected || preview.index >= first-2 && preview.index <= last+2 {
				pending = append(pending, f)
			} else {
				delete(ch.previews, f.Path)
			}
		}
		ch.queue = pending
	}
	ch.jumping = false
	// Keep a small cache of visited files, rather than retaining every buffer
	// in a large repository. In-flight and nearby previews remain intact.
	if len(ch.previews) > 24 {
		for path, preview := range ch.previews {
			if !preview.busy && path != ch.selected && (preview.index < first-4 || preview.index > last+4) {
				delete(ch.previews, path)
			}
		}
	}
	ch.diffs.Key = func(i int) any { return ch.list.Files[i].Path }
	native.List(c.Key("all-file-diffs"), &ch.diffs, len(ch.list.Files), func(i int) {
		f := ch.list.Files[i]
		preview := w.loadChangedFile(f, i)
		native.Column(c.Key("file-diff:" + f.Path)).FillWidth().MinWidth(0).Shrink(0).AlignItems(native.Stretch).Children(func() {
			native.Row(c.Key("diff-header")).Height(34).Shrink(0).Padding(0, 10).Gap(8).AlignItems(native.Center).Children(func() {
				native.Text(c, f.Path).FontSize(12).FontWeight(500).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
				native.Text(c, f.Status).FontSize(11).TextColor(p.Secondary)
			})
			w.changedFileView(c, p, f, preview)
			native.Box(c).Height(1).FillWidth().Background(p.Border)
		})
	}).Grow(1).MinWidth(0).MinHeight(0).Label("All file diffs")
}

func (w *Window) changedFileView(c *native.Context, p design.Palette, f gitdiff.File, preview *changedFilePreview) {
	ch := &w.changes
	if preview.busy {
		native.Text(c, "Reading file…").TextColor(p.Secondary).Padding(12).MinHeight(80)
	} else if preview.err != "" {
		native.Text(c, preview.err).TextColor(p.Secondary).FillWidth().Padding(12).MinHeight(80)
	} else {
		if preview.notice != "" {
			native.Text(c, preview.notice).FontSize(12).TextColor(p.Secondary).FillWidth().Padding(12).MinHeight(80)
		}
		if d := preview.diff; d != nil {
			d.Layout = editor.DiffUnified
			if ch.layout == "split" {
				d.Layout = editor.DiffSplit
			}
			d.Font = native.Font{Family: w.settings.snapshot.Font.Family, Size: max(11, w.settings.snapshot.Font.Size)}
			if d.Font.Family == "" {
				d.Font.Family = "monospace"
			}
			theme := editor.DefaultTheme(c.Theme())
			theme.Background, theme.Text = p.Content, p.Text
			theme.Added, theme.Deleted = p.Success, p.Danger
			d.Theme = theme
			d.Editable = false
			d.FitContent = true
			view := editor.DiffView(c.Key("diff"), d).Label("Code diff " + f.Path).MinWidth(0).Shrink(0).FillWidth()
			first, last := ch.diffs.Visible()
			if view.FocusWithin() && preview.index >= first && preview.index <= last {
				ch.selected, ch.diff = f.Path, d
			}
		}
	}
}

func (w *Window) changesMessage(c *native.Context, p design.Palette, message string) {
	native.Column(c).Grow(1).FillWidth().Center().Padding(24).Children(func() {
		native.Text(c, message).TextColor(p.Secondary).FillWidth().TextAlign(native.Center)
	})
}
