package objst

import (
	"errors"
	"fmt"
)

// Object errors
var (
	ErrContentTypeNotExist     = errors.New("missing content type metadata")
	ErrEmptyPayload            = errors.New("object doesn't contain any payload")
	ErrObjectIsImmutable       = errors.New("object is immutable. Create a new object")
	ErrMustIncludeOwnerAndName = errors.New("object is immutable. Create a new object")
	ErrInvalidNamePattern      = fmt.Errorf("object name must match the following regex pattern: %s", objectNamePattern)
)

// HTTP errors
var (
	ErrMissingOwner      = errors.New("missing owner in the request context")
	ErrUknownContentType = errors.New("content type of the file is not an official mime-type and no contentType key could be found in the form")
)

// Query errors
var (
	ErrEmptyQuery          = errors.New("empty query")
	ErrNameOwnerCtxMissing = errors.New("name is set but missing owner")
)
