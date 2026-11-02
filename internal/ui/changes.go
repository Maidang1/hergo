package ui

import (
	"fmt"
	"path/filepath"

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
	open             bool
	selected, layout string
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
	jumping       bool
	previews      map[string]*changedFilePreview
	queue         []gitdiff.File
	loading       int
	focus         bool
	workspaces    map[changesWorkspace]changesViewState
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

func (w *Window) closeChanges() {
	w.syncChanges()
	w.cancelChanges()
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
	if ch.workspaces == nil {
		ch.workspaces = make(map[changesWorkspace]changesViewState)
	}
	if ch.target.workspace != "" {
		ch.workspaces[changesWorkspace{ch.target.target, ch.target.workspace}] = changesViewState{ch.open, ch.selected, ch.layout}
	}
	state := ch.workspaces[changesWorkspace{target.target, target.workspace}]
	*ch = changesUI{
		open: state.open, selected: state.selected, layout: state.layout,
		target: target, generation: ch.generation,
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

func (w *Window) refreshChanges() {
	ch := &w.changes
	if !ch.open || ch.target.path == "" || w.hooks.Changes == nil {
		return
	}
	w.cancelChanges()
	ch.busy, ch.err = true, ""
	ch.previews = make(map[string]*changedFilePreview)
	ch.diff = nil
	ch.list = gitdiff.Changes{}
	ch.diffs = native.ListState{}
	gen, target := ch.generation, ch.target
	ch.cancel = w.hooks.Changes(target.path, func(list gitdiff.Changes, err error) {
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
	})
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
		req := gitdiff.Request{Root: ch.list.Root, Base: ch.list.Base, File: f}
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
			summary := "Uncommitted · compared with HEAD"
			if ch.list.Root != "" {
				summary = fmt.Sprintf("%d files · %s", len(ch.list.Files), ch.list.Root)
			}
			native.Text(c, summary).FontSize(12).TextColor(p.Secondary).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
			if w.changesButton(c, "diff-unified", "Unified", "Unified diff", ch.layout != "split", p).Clicked() {
				ch.layout = "unified"
			}
			if w.changesButton(c, "diff-split", "Split", "Side-by-side diff", ch.layout == "split", p).Clicked() {
				ch.layout = "split"
			}
			if w.changesButton(c, "refresh-changes", "Refresh", "Refresh changes", false, p).Disabled(ch.busy || ch.target.path == "").Clicked() {
				w.refreshChanges()
			}
			if w.iconButton(c, "close-changes", icons.X, "Close changes", p).Clicked() {
				w.closeChanges()
			}
		})
		if ch.busy {
			native.Text(c, "Reading changes…").TextColor(p.Secondary).Padding(8, 12)
		}
		if ch.err != "" {
			w.changesMessage(c, p, ch.err)
			return
		}
		if ch.target.path == "" {
			w.changesMessage(c, p, "Select a connected workspace with a local directory.")
			return
		}
		if len(ch.list.Files) == 0 {
			if !ch.busy {
				w.changesMessage(c, p, "No uncommitted changes.")
			}
			return
		}
		native.Box(c).Height(1).Shrink(0).FillWidth().Background(p.Border)
		native.Row(c.Key("changes-body")).Grow(1).MinHeight(0).FillWidth().AlignItems(native.Stretch).Children(func() {
			width := max(0, w.viewportWidth-w.shape.Gap*2)
			if w.sidebarVisible {
				width -= w.sidebarWidth + w.shape.Gap
			}
			listWidth := min(float32(250), width*.32)
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
			}).Width(listWidth).Shrink(0).MinHeight(0).Padding(6)
			native.Box(c).Width(1).Shrink(0).Background(p.Border)
			w.continuousChangesView(c, p)
		})
	})
	if panel.Shortcut(0, native.KeyEscape) && !w.modalOpen() {
		w.closeChanges()
	}
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
