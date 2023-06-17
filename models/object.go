package models

import "net/url"

type Object struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Owner   string     `json:"owner"`
	Meta    url.Values `json:"metadata"`
	Payload []byte     `json:"payload"`
}
