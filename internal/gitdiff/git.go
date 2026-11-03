// Package gitdiff reads uncommitted changes and commit history without
// modifying the index or working tree. Git commands and file reads run
// outside the UI thread.
package gitdiff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const maxText = 1 << 20
const maxOutput = 4 << 20

// MaxCommits is how many recent commits the history list loads.
const MaxCommits = 100

type File struct {
	Path, OldPath, Status string
	Added, Removed        int
	Untracked, Binary     bool
}

type Changes struct {
	Root, Base string
	// Commit is set when Files are that commit's changes against its first
	// parent. Empty means the working tree compared with Base.
	Commit string
	Files  []File
}

// Commit is one entry from git log. Hash is the full object name.
type Commit struct {
	Hash, Subject, Author string
	Time                  int64
}

type Request struct {
	Root, Base string
	// Commit, when set, reads blobs at Base and Commit instead of the working tree.
	Commit string
	File   File
}

type Text struct {
	Old, New, Patch, Notice string
}

var errOutputLimit = errors.New("Git output is too large to display")

type limitedOutput struct{ bytes.Buffer }

func (b *limitedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > maxOutput {
		return 0, errOutputLimit
	}
	return b.Buffer.Write(p)
}

func run(ctx context.Context, root string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"--no-optional-locks", "--literal-pathspecs", "-c", "color.ui=false", "-c", "core.quotepath=false", "-C", root}, args...)...)
	hideCommand(cmd)
	cmd.WaitDelay = 100 * time.Millisecond
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if !strings.HasPrefix(key, "GIT_") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	var stdout, stderr limitedOutput
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		// Diagnostics can contain repository configuration and paths. Show a
		// bounded, fixed error instead of forwarding subprocess output.
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errors.New("Git is not installed or is not on PATH")
		}
		return nil, fmt.Errorf("git %s: %w", args[0], err)
	}
	return stdout.Bytes(), nil
}

// List compares HEAD with the working tree, including staged and untracked
// files. Base is resolved once so selecting files uses the same revision.
func List(ctx context.Context, dir string) (Changes, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if !filepath.IsAbs(dir) {
		return Changes{}, errors.New("select a workspace with a local directory")
	}
	root, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Changes{}, fmt.Errorf("cannot read a Git repository in this workspace: %w", err)
	}
	res := Changes{Root: strings.TrimSuffix(strings.TrimSuffix(string(root), "\n"), "\r")}
	base, err := run(ctx, res.Root, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return Changes{}, err
		}
		// hash-object without -w computes an empty tree's id without writing
		// an object, including in a repository with no commits.
		base, err = run(ctx, res.Root, "hash-object", "-t", "tree", "--stdin")
		if err != nil {
			return Changes{}, err
		}
	}
	res.Base = strings.TrimSpace(string(base))
	if err := fillDiff(&res, ctx, res.Base); err != nil {
		return Changes{}, err
	}
	out, err := run(ctx, res.Root, "ls-files", "--others", "--exclude-standard", "-z", "--")
	if err != nil {
		return Changes{}, err
	}
	seen := make(map[string]bool, len(res.Files))
	for _, f := range res.Files {
		seen[f.Path] = true
	}
	for _, path := range strings.Split(string(out), "\x00") {
		if path != "" && !seen[path] {
			res.Files = append(res.Files, File{Path: path, Status: "Untracked", Untracked: true})
		}
	}
	sort.Slice(res.Files, func(i, j int) bool { return res.Files[i].Path < res.Files[j].Path })
	return res, nil
}

// Log returns the newest commits, up to MaxCommits. A repository with no
// commits yet returns an empty slice.
func Log(ctx context.Context, dir string) ([]Commit, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	root, err := toplevel(ctx, dir)
	if err != nil {
		return nil, err
	}
	if _, err := run(ctx, root, "rev-parse", "--verify", "--quiet", "HEAD^{commit}"); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return []Commit{}, nil
		}
		return nil, err
	}
	out, err := run(ctx, root, "log", "--no-show-signature", "-n", strconv.Itoa(MaxCommits), "--pretty=format:%H%x1f%s%x1f%an%x1f%at%x00")
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, rec := range bytes.Split(out, []byte{0}) {
		rec = bytes.Trim(rec, "\r\n")
		if len(rec) == 0 {
			continue
		}
		parts := bytes.Split(rec, []byte{0x1f})
		if len(parts) != 4 || !validRev(string(parts[0])) {
			continue
		}
		when, err := strconv.ParseInt(string(bytes.TrimSpace(parts[3])), 10, 64)
		if err != nil || when < 0 {
			when = 0
		}
		commits = append(commits, Commit{
			Hash:    string(parts[0]),
			Subject: oneLine(string(parts[1]), 200),
			Author:  oneLine(string(parts[2]), 80),
			Time:    when,
		})
	}
	return commits, nil
}

// ListCommit returns the files a commit changed relative to its first parent.
// The root commit is compared with an empty tree. The working tree is not read.
func ListCommit(ctx context.Context, dir, hash string) (Changes, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if !validRev(hash) {
		return Changes{}, errors.New("invalid commit")
	}
	root, err := toplevel(ctx, dir)
	if err != nil {
		return Changes{}, err
	}
	got, err := run(ctx, root, "rev-parse", "--verify", "--quiet", hash+"^{commit}")
	if err != nil || strings.TrimSpace(string(got)) != hash {
		return Changes{}, errors.New("commit is not in this repository")
	}
	parent, err := run(ctx, root, "rev-parse", "--verify", "--quiet", hash+"^")
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return Changes{}, err
		}
		parent, err = run(ctx, root, "hash-object", "-t", "tree", "--stdin")
		if err != nil {
			return Changes{}, err
		}
	}
	base := strings.TrimSpace(string(parent))
	if !validRev(base) {
		return Changes{}, errors.New("invalid commit")
	}
	res := Changes{Root: root, Base: base, Commit: hash}
	if err := fillDiff(&res, ctx, base, hash); err != nil {
		return Changes{}, err
	}
	return res, nil
}

func toplevel(ctx context.Context, dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", errors.New("select a workspace with a local directory")
	}
	root, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("cannot read a Git repository in this workspace: %w", err)
	}
	rootPath := strings.TrimSuffix(strings.TrimSuffix(string(root), "\n"), "\r")
	if !filepath.IsAbs(rootPath) {
		return "", errors.New("cannot read a Git repository in this workspace")
	}
	return rootPath, nil
}

func fillDiff(res *Changes, ctx context.Context, revs ...string) error {
	args := append([]string{"diff", "--no-ext-diff", "--no-textconv", "-M", "--name-status", "-z"}, revs...)
	out, err := run(ctx, res.Root, append(args, "--")...)
	if err != nil {
		return err
	}
	files, err := parseNameStatus(out)
	if err != nil {
		return err
	}
	res.Files = files
	stats, err := run(ctx, res.Root, append(append([]string{"diff", "--no-ext-diff", "--no-textconv", "-M", "--numstat", "-z"}, revs...), "--")...)
	if err != nil {
		return err
	}
	applyNumstat(res.Files, stats)
	sort.Slice(res.Files, func(i, j int) bool { return res.Files[i].Path < res.Files[j].Path })
	return nil
}

func parseNameStatus(out []byte) ([]File, error) {
	var files []File
	fields := strings.Split(string(out), "\x00")
	for i := 0; i+1 < len(fields); {
		code, path := fields[i], fields[i+1]
		i += 2
		if code == "" {
			break
		}
		f := File{Path: path, Status: "Modified"}
		switch code[0] {
		case 'A':
			f.Status = "Added"
		case 'D':
			f.Status = "Deleted"
		case 'R', 'C':
			if i >= len(fields) {
				return nil, errors.New("incomplete Git rename result")
			}
			f.OldPath, f.Path, f.Status = path, fields[i], "Renamed"
			i++
		case 'T':
			f.Status = "Type changed"
		}
		files = append(files, f)
	}
	return files, nil
}

func applyNumstat(files []File, stats []byte) {
	byPath := make(map[string]int, len(files))
	for i, f := range files {
		byPath[f.Path] = i
	}
	fields := strings.Split(string(stats), "\x00")
	for i := 0; i < len(fields); i++ {
		parts := strings.SplitN(fields[i], "\t", 3)
		if len(parts) != 3 {
			continue
		}
		path := parts[2]
		if path == "" && i+2 < len(fields) {
			path = fields[i+2]
			i += 2
		}
		if n, ok := byPath[path]; ok {
			files[n].Added, _ = strconv.Atoi(parts[0])
			files[n].Removed, _ = strconv.Atoi(parts[1])
			files[n].Binary = parts[0] == "-"
		}
	}
}

func validRev(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	return strings.Trim(s, "0123456789abcdef") == ""
}

func oneLine(s string, limit int) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			r = ' '
		}
		n++
		if n > limit {
			break
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// Read returns a file's old and current text. Large tracked files fall back to
// a bounded patch; binary files, submodules and non-UTF-8 files show a message.
func Read(ctx context.Context, req Request) (Text, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if !filepath.IsAbs(req.Root) || !validRev(req.Base) || req.Commit != "" && !validRev(req.Commit) {
		return Text{}, errors.New("invalid repository or revision")
	}
	full, err := localPath(req.Root, req.File.Path)
	if err != nil {
		return Text{}, err
	}
	oldPath := req.File.OldPath
	if oldPath == "" {
		oldPath = req.File.Path
	}
	if _, err := localPath(req.Root, oldPath); err != nil {
		return Text{}, err
	}
	var old, now []byte
	large := false
	if !req.File.Untracked && req.File.Status != "Added" {
		var notice string
		old, large, notice, err = readBlob(ctx, req.Root, req.Base, oldPath)
		if err != nil {
			return Text{}, err
		}
		if notice != "" {
			return Text{Notice: notice}, nil
		}
	}
	if req.File.Status != "Deleted" {
		if req.Commit != "" {
			var notice string
			var blobLarge bool
			now, blobLarge, notice, err = readBlob(ctx, req.Root, req.Commit, req.File.Path)
			if err != nil {
				return Text{}, err
			}
			if notice != "" {
				return Text{Notice: notice}, nil
			}
			large = large || blobLarge
		} else {
			now, err = readWorking(full)
			if errors.Is(err, errOutputLimit) {
				large = true
			} else if err != nil {
				return Text{}, err
			}
		}
	}
	if req.File.Binary || bytes.IndexByte(old, 0) >= 0 || bytes.IndexByte(now, 0) >= 0 {
		return Text{Notice: "Binary file. No text preview is available."}, nil
	}
	if large {
		t := Text{Notice: "File exceeds 1 MiB. Showing its patch when available."}
		if !req.File.Untracked {
			paths := []string{req.File.Path}
			if req.File.OldPath != "" {
				paths = append(paths, req.File.OldPath)
			}
			diffArgs := []string{"diff", "--no-ext-diff", "--no-textconv", "--no-renames", req.Base}
			if req.Commit != "" {
				diffArgs = append(diffArgs, req.Commit)
			}
			patch, err := run(ctx, req.Root, append(append(diffArgs, "--"), paths...)...)
			if ctx.Err() != nil {
				return Text{}, ctx.Err()
			}
			if err == nil && utf8.Valid(patch) {
				t.Patch = string(patch)
			}
		}
		return t, nil
	}
	if !utf8.Valid(old) || !utf8.Valid(now) {
		return Text{Notice: "This file is not UTF-8 text. No preview is available."}, nil
	}
	return Text{Old: string(old), New: string(now)}, nil
}

// readBlob loads one Git blob. large means the object is over the text limit
// and data is empty. notice is set when the object is not a blob.
func readBlob(ctx context.Context, root, rev, path string) ([]byte, bool, string, error) {
	object := rev + ":" + path
	kind, err := run(ctx, root, "cat-file", "-t", object)
	if err != nil {
		return nil, false, "", err
	}
	if strings.TrimSpace(string(kind)) != "blob" {
		return nil, false, "Submodule changes cannot be previewed as text.", nil
	}
	size, err := run(ctx, root, "cat-file", "-s", object)
	if err != nil {
		return nil, false, "", err
	}
	n, err := strconv.ParseInt(strings.TrimSpace(string(size)), 10, 64)
	if err != nil {
		return nil, false, "", err
	}
	if n > maxText {
		return nil, true, "", nil
	}
	data, err := run(ctx, root, "cat-file", "blob", object)
	return data, false, "", err
}

func localPath(root, path string) (string, error) {
	if path == "" || strings.ContainsRune(path, 0) || filepath.IsAbs(path) || !filepath.IsLocal(filepath.FromSlash(path)) {
		return "", errors.New("file is outside the working tree")
	}
	for _, part := range strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' }) {
		if strings.EqualFold(part, ".git") {
			return "", errors.New("Git metadata cannot be previewed")
		}
	}
	full := filepath.Join(root, filepath.FromSlash(path))
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(full)
	for {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			rel, err := filepath.Rel(canonicalRoot, resolved)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return "", errors.New("file is outside the working tree")
			}
			return full, nil
		}
		if !os.IsNotExist(err) || filepath.Dir(parent) == parent {
			return "", err
		}
		parent = filepath.Dir(parent)
	}
}

func readWorking(full string) ([]byte, error) {
	fi, err := os.Lstat(full)
	if os.IsNotExist(err) {
		return nil, errors.New("file changed or disappeared; refresh Changes")
	}
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(full)
		return []byte(target), err
	}
	if !fi.Mode().IsRegular() {
		return nil, errors.New("this file type cannot be previewed as text")
	}
	if fi.Size() > maxText {
		return nil, errOutputLimit
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxText+1))
	if len(b) > maxText {
		return nil, errOutputLimit
	}
	return b, err
}
