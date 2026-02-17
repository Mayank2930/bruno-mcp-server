package bruno

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Client struct {
	brunoPath string
}

func NewClient() *Client {
	path, _ := exec.LookPath("bru")
	return &Client{brunoPath: path}
}

func (c *Client) hasCli() bool { return c.brunoPath != "" }

type CommandError struct {
	Cmd    []string
	Stderr string
	Err    error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("bruno command error: %v: %v", e.Cmd, e.Err)
}

func (c *Client) run(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	if c.brunoPath == "" {
		return "", "", errors.New("bru not found in file PATH")
	}

	cmd := exec.CommandContext(ctx, c.brunoPath, args...)
	outB, err := cmd.Output()

	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return string(outB), string(ee.Stderr), &CommandError{
				Cmd:    append([]string{c.brunoPath}, args...),
				Stderr: string(ee.Stderr),
				Err:    err,
			}
		}
		return string(outB), "", &CommandError{
			Cmd: append([]string{c.brunoPath}, args...),
			Err: err,
		}
	}
	return string(outB), "", nil
}

func (c *Client) ListCollections(workspaceDir string) ([]string, error) {
	workspaceDir = filepath.Clean(workspaceDir)

	info, err := os.Stat(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("workspace path stat failed: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace path is not a directory: %q", workspaceDir)
	}

	var cols []string

	// If the workspace itself is a collection root (has bruno.json), include it.
	if fileExists(filepath.Join(workspaceDir, "bruno.json")) {
		cols = append(cols, filepath.Base(workspaceDir))
	}

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("read workspace dir failed: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == ".git" || name == "node_modules" {
			continue
		}
		dir := filepath.Join(workspaceDir, name)
		if fileExists(filepath.Join(dir, "bruno.json")) {
			cols = append(cols, name)
		}
	}

	sort.Strings(cols)
	return cols, nil
}

func (c *Client) ListRequests(workspaceDir, collection string) ([]string, error) {
	workspaceDir = filepath.Clean(workspaceDir)
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return nil, errors.New("collection is required")
	}

	// Resolve collection root:
	// - if workspace itself is a collection and collection matches base name => use workspaceDir
	// - else treat collection as a subdir under workspaceDir (safe join)
	var colRoot string
	if fileExists(filepath.Join(workspaceDir, "bruno.json")) && collection == filepath.Base(workspaceDir) {
		colRoot = workspaceDir
	} else {
		j, err := safeJoin(workspaceDir, collection)
		if err != nil {
			return nil, err
		}
		colRoot = j
	}

	info, err := os.Stat(colRoot)
	if err != nil {
		return nil, fmt.Errorf("collection path stat failed: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("collection path is not a directory: %q", colRoot)
	}
	if !fileExists(filepath.Join(colRoot, "bruno.json")) {
		return nil, fmt.Errorf("not a Bruno collection (missing bruno.json): %q", colRoot)
	}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"environments": true, // env files are .bru too; skip them for “requests”
	}

	var reqs []string
	err = filepath.WalkDir(colRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}

		if filepath.Ext(d.Name()) != ".bru" {
			return nil
		}
		base := d.Name()
		if base == "collection.bru" || base == "folder.bru" {
			return nil
		}

		rel, err := filepath.Rel(colRoot, path)
		if err != nil {
			return err
		}
		reqs = append(reqs, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk collection failed: %w", err)
	}

	sort.Strings(reqs)
	return reqs, nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func safeJoin(base, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid collection path (absolute): %q", rel)
	}
	clean := filepath.Clean(rel)
	if clean == "." || clean == string(filepath.Separator) {
		return "", fmt.Errorf("invalid collection path: %q", rel)
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid collection path (traversal): %q", rel)
	}
	joined := filepath.Join(base, clean)

	r, err := filepath.Rel(base, joined)
	if err != nil {
		return "", err
	}
	if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid collection path (escapes workspace): %q", rel)
	}
	return joined, nil
}
