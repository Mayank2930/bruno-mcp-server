package bruno

import "errors"

var (
	ErrBruNotFound   = errors.New("bru cli not found")
	ErrCommandFailed = errors.New("bruno command failed")

	ErrInvalidCollectionName = errors.New("invalid collection name")
	ErrInvalidRequestPath    = errors.New("invalid request path")

	ErrAlreadyExists     = errors.New("already exists")
	ErrNotACollection    = errors.New("not a bruno collection")
	ErrCollectionMissing = errors.New("collection not found")
)
