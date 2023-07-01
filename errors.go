package objst

import "errors"

var (
	ErrContentTypeNotExist     = errors.New("missing content type metadata")
	ErrEmptyPayload            = errors.New("object doesn't contain any payload")
	ErrObjectIsImmutable       = errors.New("object is immutable. Create a new object")
	ErrMustIncludeOwnerAndName = errors.New("object is immutable. Create a new object")
	ErrInvalidNamePattern      = errors.New("object name must match the following regex pattern: ^[a-zA-Z0-9_.-]+$")
)

var (
	ErrUnauthorized = errors.New("owner is not authorized to access the object")
)

var (
	ErrInvalidQuery = errors.New("query is invalid. Missing owner")
)
