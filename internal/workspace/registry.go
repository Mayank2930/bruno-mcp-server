package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Workspace struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Registry struct {
	mu sync.RWMutex
	m  map[string]Workspace
}

func NewRegistry() *Registry {
	return &Registry{m: make(map[string]Workspace)}
}

var nameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func validateName(name string) error {
	n := strings.TrimSpace(name)
	if n == "" {
		return fmt.Errorf("%w: empty", ErrInvalidName)
	}
	if strings.Contains(n, "..") {
		return fmt.Errorf("%w: contains \"..\": %q", ErrInvalidName, name)
	}
	if strings.ContainsAny(n, `/\`) {
		return fmt.Errorf("%w: contains path separator: %q", ErrInvalidName, name)
	}
	if !nameRe.MatchString(n) {
		return fmt.Errorf("%w: must match %s: %q", ErrInvalidName, nameRe.String(), name)
	}
	return nil
}

func validateAbsPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("%w: empty", ErrInvalidPath)
	}
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return "", fmt.Errorf("%w: must be absolute: %q", ErrInvalidPath, path)
	}
	return clean, nil
}

func (r *Registry) Register(name, path string, createIfMissing bool) (Workspace, error) {
	if err := validateName(name); err != nil {
		return Workspace{}, err
	}

	cleanPath, err := validateAbsPath(path)
	if err != nil {
		return Workspace{}, err
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			if !createIfMissing {
				return Workspace{}, fmt.Errorf("%w: does not exist: %q", ErrInvalidPath, cleanPath)
			}
			if err := os.MkdirAll(cleanPath, 0o755); err != nil {
				return Workspace{}, fmt.Errorf("%w: failed to create %q: %v", ErrInvalidPath, cleanPath, err)
			}
			info, err = os.Stat(cleanPath)
			if err != nil {
				return Workspace{}, fmt.Errorf("%w: failed to stat after create %q: %v", ErrInvalidPath, cleanPath, err)
			}
		} else {
			return Workspace{}, fmt.Errorf("%w: failed to stat %q: %v", ErrInvalidPath, cleanPath, err)
		}
	}

	if !info.IsDir() {
		return Workspace{}, fmt.Errorf("%w: not a directory: %q", ErrInvalidPath, cleanPath)
	}

	ws := Workspace{Name: name, Path: cleanPath}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Idempotent if same name+path; conflict if same name different path.
	if existing, ok := r.m[name]; ok {
		if existing.Path != ws.Path {
			// See note below about adding a dedicated conflict error.
			return Workspace{}, fmt.Errorf("%w: name %q already registered as %q", ErrInvalidPath, name, existing.Path)
		}
		return existing, nil
	}

	r.m[name] = ws
	return ws, nil
}

func (r *Registry) Get(name string) (Workspace, error) {
	if err := validateName(name); err != nil {
		return Workspace{}, err
	}

	r.mu.RLock()
	ws, ok := r.m[name]
	r.mu.RUnlock()

	if !ok {
		return Workspace{}, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	return ws, nil
}

func (r *Registry) List() []Workspace {
	r.mu.RLock()
	out := make([]Workspace, 0, len(r.m))
	for _, ws := range r.m {
		out = append(out, ws)
	}
	r.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Optional helper if you want callers to check "not found" without importing errors.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
