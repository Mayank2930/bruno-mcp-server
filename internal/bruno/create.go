package bruno

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CreateCollectionOptions struct {
	Overwrite bool
	Ignore    []string
}

func (c *Client) CreateCollection(workspaceDir, name string, opts CreateCollectionOptions) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: empty", ErrInvalidCollectionName)
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("%w: %q", ErrInvalidCollectionName, name)
	}

	collectionDir, err := safeJoin(workspaceDir, name)
	if err != nil {
		return "", err
	}

	// If exists:
	if st, err := os.Stat(collectionDir); err == nil {
		if !st.IsDir() {
			return "", fmt.Errorf("%w: path exists but is not a directory: %q", ErrAlreadyExists, collectionDir)
		}
		if !opts.Overwrite {
			return "", fmt.Errorf("%w: %q", ErrAlreadyExists, collectionDir)
		}
	} else if err := os.MkdirAll(collectionDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir collection: %w", err)
	}

	if len(opts.Ignore) == 0 {
		opts.Ignore = []string{"node_modules", ".git"}
	}
	brunoJSON := map[string]any{
		"version": "1",
		"name":    name,
		"type":    "collection",
		"ignore":  opts.Ignore,
	}
	b, _ := json.MarshalIndent(brunoJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(collectionDir, "bruno.json"), b, 0o644); err != nil {
		return "", fmt.Errorf("write bruno.json: %w", err)
	}

	_ = os.WriteFile(filepath.Join(collectionDir, "collection.bru"), []byte(""), 0o644)

	return collectionDir, nil
}

type CreateRequestOptions struct {
	Overwrite bool
	Seq       int
}

func (c *Client) CreateRequest(workspaceDir, collection, relRequestPath, method, url string, opts CreateRequestOptions) (string, error) {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return "", fmt.Errorf("%w: empty collection", ErrInvalidRequestPath)
	}

	relRequestPath = strings.TrimSpace(relRequestPath)
	if relRequestPath == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidRequestPath)
	}
	if filepath.IsAbs(relRequestPath) {
		return "", fmt.Errorf("%w: absolute path: %q", ErrInvalidRequestPath, relRequestPath)
	}
	if !strings.HasSuffix(strings.ToLower(relRequestPath), ".bru") {
		relRequestPath += ".bru"
	}

	var colRoot string
	if fileExists(filepath.Join(workspaceDir, "bruno.json")) && collection == filepath.Base(workspaceDir) {
		colRoot = workspaceDir
	} else {
		j, err := safeJoin(workspaceDir, collection)
		if err != nil {
			return "", err
		}
		colRoot = j
	}

	if !fileExists(filepath.Join(colRoot, "bruno.json")) {
		return "", fmt.Errorf("not a Bruno collection (missing bruno.json): %q", colRoot)
	}

	fullPath, err := safeJoin(colRoot, relRequestPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir request parent: %w", err)
	}

	if _, err := os.Stat(fullPath); err == nil && !opts.Overwrite {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExists, fullPath)
	}

	method = strings.ToLower(strings.TrimSpace(method))
	if !isAllowedMethod(method) {
		return "", fmt.Errorf("%w: unsupported method: %q", ErrInvalidRequestPath, method)
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("%w: empty url", ErrInvalidRequestPath)
	}

	seq := opts.Seq
	if seq <= 0 {
		seq = nextSeq(filepath.Dir(fullPath))
	}

	// Minimal .bru format using tags from docs :contentReference[oaicite:3]{index=3}
	displayName := strings.TrimSuffix(filepath.Base(relRequestPath), ".bru")
	content := fmt.Sprintf(
		"meta {\n  name: %s\n  type: http\n  seq: %d\n}\n\n%s {\n  url: %s\n}\n",
		displayName, seq, method, url,
	)

	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write request: %w", err)
	}

	// Return normalized relative path
	return filepath.ToSlash(relRequestPath), nil
}

func isAllowedMethod(m string) bool {
	switch m {
	case "get", "post", "put", "delete", "head", "options", "trace", "connect", "patch":
		return true
	default:
		return false
	}
}

func nextSeq(dir string) int {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}
	n := 0
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".bru") {
			continue
		}
		if name == "collection.bru" || name == "folder.bru" {
			continue
		}
		n++
	}
	return n + 1
}
