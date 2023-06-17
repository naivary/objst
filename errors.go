package main

import "errors"

var (
	ErrContentTypeNotExist = errors.New("missing content type metadata")
	ErrEmptyPayload        = errors.New("object doesn't contain any payload")
	ErrObjectIsImmutable   = errors.New("object is immutable. Create a new object")
)