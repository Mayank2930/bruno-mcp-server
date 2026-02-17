package workspace

import "errors"

var (
	ErrNotFound    = errors.New("workspace not found")
	ErrInvalidName = errors.New("invalid workspace name")
	ErrInvalidPath = errors.New("invalid workspace path")
)
