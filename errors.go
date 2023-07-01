package objst

import "errors"

// Object errors
var (
	ErrContentTypeNotExist     = errors.New("missing content type metadata")
	ErrEmptyPayload            = errors.New("object doesn't contain any payload")
	ErrObjectIsImmutable       = errors.New("object is immutable. Create a new object")
	ErrMustIncludeOwnerAndName = errors.New("object is immutable. Create a new object")
	ErrInvalidNamePattern      = errors.New("object name must match the following regex pattern: ^[a-zA-Z0-9_.-]+$")
)

// Bucket errors
var (
	ErrUnauthorized = errors.New("owner is not authorized to access the object")
)

// Query errors
var (
	ErrInvalidQuery = errors.New("query is invalid. Missing owner")
)
