package main

import "errors"

var (
	ErrContentTypeNotExist = errors.New("missing content type metadata")
	ErrEmptyPayload        = errors.New("object doesn't contain any payload")
)
