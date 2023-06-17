package models

import "net/url"

type Object struct {
	ID string
	Name  string
	Owner string
	Meta    url.Values
	Payload []byte
}
