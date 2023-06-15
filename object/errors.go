package object

import "errors"

var (
	ErrContentTypeNotExist = errors.New("missing content type metadata")
)
