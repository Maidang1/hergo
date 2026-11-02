package gitdiff

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func gitFixture(t *testing.T, dir, input string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"--no-optional-locks", "-C", dir}, args...)...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFixture(t *testing.T, root, name, text string) {
	t.Helper()
	full := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func repoFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	gitFixture(t, root, "", "init", "-b", "main")
	for name, text := range files {
		writeFixture(t, root, name, text)
	}
	gitFixture(t, root, "", "add", "--all")
	tree := gitFixture(t, root, "", "write-tree")
	// A synthetic object keeps fixtures independent of user identity, signing
	// configuration and hooks. No Git author configuration is changed.
	commit := "tree " + tree + "\nauthor Fixture <fixture@example.invalid> 1 +0000\ncommitter Fixture <fixture@example.invalid> 1 +0000\n\nfixture\n"
	oid := gitFixture(t, root, commit, "hash-object", "-t", "commit", "-w", "--stdin")
	gitFixture(t, root, "", "update-ref", "refs/heads/main", oid)
	return root
}

func listFixture(t *testing.T, root string) Changes {
	t.Helper()
	res, err := List(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func readFixture(t *testing.T, list Changes, name string) Text {
	t.Helper()
	for _, f := range list.Files {
		if f.Path == name {
			text, err := Read(context.Background(), Request{Root: list.Root, Base: list.Base, File: f})
			if err != nil {
				t.Fatal(err)
			}
			return text
		}
	}
	t.Fatalf("missing file %q in %+v", name, list.Files)
	return Text{}
}

func TestListCombinesStagedUnstagedAndUntrackedWithoutChangingIndex(t *testing.T) {
	root := repoFixture(t, map[string]string{"main.go": "old\n", ".gitignore": "ignored\n"})
	writeFixture(t, root, "main.go", "staged\n")
	gitFixture(t, root, "", "add", "--", "main.go")
	writeFixture(t, root, "main.go", "working\n")
	writeFixture(t, root, "new file.txt", "new\n")
	writeFixture(t, root, "ignored", "ignored\n")
	indexPath := filepath.Join(root, ".git", "index")
	before, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	// A terminal's inherited Git overrides must not redirect the viewer.
	t.Setenv("GIT_DIR", "/not/this/repo")
	t.Setenv("GIT_WORK_TREE", "/not/this/worktree")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "diff.external")
	t.Setenv("GIT_CONFIG_VALUE_0", "must-not-run")
	list := listFixture(t, root)
	if len(list.Files) != 2 || list.Files[0].Path != "main.go" || list.Files[0].Added != 1 || list.Files[0].Removed != 1 || !list.Files[1].Untracked {
		t.Fatal(list.Files)
	}
	if got := readFixture(t, list, "main.go"); got.Old != "old\n" || got.New != "working\n" {
		t.Fatal(got)
	}
	if got := readFixture(t, list, "new file.txt"); got.Old != "" || got.New != "new\n" {
		t.Fatal(got)
	}
	after, err := os.ReadFile(indexPath)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("viewing changes modified the index", err)
	}
}

func TestRenamesDeletionsAndLiteralPaths(t *testing.T) {
	oldName, newName := "old\tname.txt", "new\nname.txt"
	if runtime.GOOS == "windows" {
		oldName, newName = "old name.txt", "new name.txt"
	}
	root := repoFixture(t, map[string]string{oldName: "rename\n", "gone.txt": "deleted\n", "[ab].txt": "literal\n"})
	gitFixture(t, root, "", "mv", "--", oldName, newName)
	if err := os.Remove(filepath.Join(root, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, "[ab].txt", "changed\n")
	list := listFixture(t, root)
	if len(list.Files) != 3 {
		t.Fatal(list.Files)
	}
	rename := list.Files[2]
	if rename.Path != newName || rename.OldPath != oldName || rename.Status != "Renamed" {
		t.Fatal(rename)
	}
	if got := readFixture(t, list, rename.Path); got.Old != "rename\n" || got.New != "rename\n" {
		t.Fatal(got)
	}
	if got := readFixture(t, list, "gone.txt"); got.Old != "deleted\n" || got.New != "" {
		t.Fatal(got)
	}
	if got := readFixture(t, list, "[ab].txt"); got.New != "changed\n" {
		t.Fatal(got)
	}
}

func TestUnbornHeadAndLinkedWorktree(t *testing.T) {
	t.Run("unborn", func(t *testing.T) {
		root := t.TempDir()
		gitFixture(t, root, "", "init", "-b", "main")
		writeFixture(t, root, "first.txt", "first\n")
		writeFixture(t, root, "staged.txt", "staged\n")
		gitFixture(t, root, "", "add", "--", "staged.txt")
		list := listFixture(t, root)
		if len(list.Files) != 2 {
			t.Fatal(list.Files)
		}
		if got := readFixture(t, list, "staged.txt"); got.Old != "" || got.New != "staged\n" {
			t.Fatal(got)
		}
	})
	t.Run("worktree", func(t *testing.T) {
		root := repoFixture(t, map[string]string{"file.txt": "old\n"})
		linked := filepath.Join(t.TempDir(), "linked")
		gitFixture(t, root, "", "worktree", "add", "-b", "feature", linked)
		writeFixture(t, linked, "file.txt", "new\n")
		list := listFixture(t, linked)
		if got := readFixture(t, list, "file.txt"); got.Old != "old\n" || got.New != "new\n" {
			t.Fatal(got)
		}
		if len(listFixture(t, root).Files) != 0 {
			t.Fatal("read the parent checkout instead of the worktree")
		}
	})
}

func TestLargeBinaryAndNonUTF8Files(t *testing.T) {
	root := repoFixture(t, map[string]string{"large.txt": strings.Repeat("unchanged line\n", 80000)})
	writeFixture(t, root, "large.txt", strings.Repeat("unchanged line\n", 80000)+"changed\n")
	writeFixture(t, root, "binary", "a\x00b")
	writeFixture(t, root, "encoding", "\xff\xfe")
	list := listFixture(t, root)
	if got := readFixture(t, list, "large.txt"); got.Patch == "" || got.Old != "" || got.New != "" {
		t.Fatal("large file did not use a patch", got.Notice)
	}
	for _, name := range []string{"binary", "encoding"} {
		if got := readFixture(t, list, name); got.Notice == "" || got.Old != "" || got.New != "" {
			t.Fatal(name, got)
		}
	}
}

func TestPathBoundariesAndSymlinks(t *testing.T) {
	root := repoFixture(t, map[string]string{"file.txt": "old\n"})
	outside := t.TempDir()
	writeFixture(t, outside, "private.txt", "outside\n")
	for _, name := range []string{"../private.txt", ".git/config", ".GIT/config", filepath.Join(outside, "private.txt")} {
		if _, err := localPath(root, name); err == nil {
			t.Fatal("accepted unsafe path", name)
		}
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, err := localPath(root, "escape/private.txt"); err == nil {
		t.Fatal("followed a directory link outside the repository")
	}
	if err := os.Symlink(filepath.Join(outside, "private.txt"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	got := readFixture(t, listFixture(t, root), "link")
	if got.New == "outside\n" || !strings.Contains(got.New, "private.txt") {
		t.Fatal("read a symlink's target instead of its link text", got)
	}
}

func TestCancelledAndNonRepositoryRequests(t *testing.T) {
	root := repoFixture(t, map[string]string{"file.txt": "old\n"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := List(ctx, root); err == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := List(context.Background(), t.TempDir()); err == nil {
		t.Fatal("accepted a non-repository")
	}
	if _, err := List(context.Background(), "relative"); err == nil {
		t.Fatal("accepted a relative workspace")
	}
}
